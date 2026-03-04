package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	internalpb "tiny-ils/gen/internalpb"
	pb "tiny-ils/gen/networkpb"
	"tiny-ils/network-manager/service"
	"tiny-ils/network-manager/store"
	"tiny-ils/shared/db"
	"tiny-ils/shared/identity"
)

func main() {
	ctx := context.Background()

	privKey, pubKey, err := identity.LoadOrCreate(
		envOr("NODE_KEY_PATH", "/data/node.key"),
		envOr("NODE_PUBKEY_PATH", "/data/node.pub"),
	)
	if err != nil {
		log.Fatalf("node keypair: %v", err)
	}
	nodeID := identity.Fingerprint(pubKey)
	pubKeyB64 := base64.StdEncoding.EncodeToString(pubKey)

	// Generate an in-memory self-signed TLS certificate from the node keypair.
	nodeCert, err := identity.SelfSignedCert(nodeID, privKey)
	if err != nil {
		log.Fatalf("self-signed cert: %v", err)
	}

	// Write the cert to disk so it can be shared with other services if needed.
	certPath := envOr("NODE_CERT_PATH", "/data/node.crt")
	if err := identity.WriteCertPEM(certPath, nodeCert); err != nil {
		log.Printf("warning: write cert PEM: %v", err)
	}

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	peerStore := store.NewPeerStore(pool)
	settings := store.NewSettingsStore(pool)
	registry := service.NewLocalDirectoryService()

	externalPort := envOr("GRPC_PORT", "50153")
	selfAddr := detectOrLoadAddress(ctx, settings, externalPort)

	svc := service.NewNetworkService(peerStore, settings, nodeID, pubKey, privKey, nodeCert, registry, selfAddr)

	// ─── Internal server (Docker-private, no auth) ────────────────────────────
	// Hosts LocalDirectory (service registration + lookup) and NetworkManager
	// (for admin ops and relay calls from local curios-manager / frontend).
	internalPort := envOr("INTERNAL_GRPC_PORT", "50154")
	internalLis, err := net.Listen("tcp", fmt.Sprintf(":%s", internalPort))
	if err != nil {
		log.Fatalf("internal listen: %v", err)
	}
	internalSrv := grpc.NewServer()
	internalpb.RegisterLocalDirectoryServer(internalSrv, registry)
	pb.RegisterNetworkManagerServer(internalSrv, svc)
	grpc_health_v1.RegisterHealthServer(internalSrv, health.NewServer())
	go func() {
		log.Printf("network-manager internal listening on :%s (insecure)", internalPort)
		if err := internalSrv.Serve(internalLis); err != nil {
			log.Fatalf("internal serve: %v", err)
		}
	}()

	// ─── External server (mTLS, host-exposed, peer-to-peer only) ─────────────
	externalLis, err := net.Listen("tcp", fmt.Sprintf(":%s", externalPort))
	if err != nil {
		log.Fatalf("external listen: %v", err)
	}
	tlsCfg := service.ServerTLSConfig(nodeCert)
	externalSrv := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsCfg)),
		grpc.ChainUnaryInterceptor(service.TrustInterceptor(peerStore, pubKeyB64)),
		grpc.ChainStreamInterceptor(service.TrustStreamInterceptor(peerStore, pubKeyB64)),
	)
	pb.RegisterNetworkManagerServer(externalSrv, svc)
	reflection.Register(externalSrv)

	// Start HTTP interop server (digital passthrough, ISO 18626 stub, SRU stub).
	httpPort := envOr("HTTP_PORT", "8153")
	go func() {
		log.Printf("network-manager HTTP interop listening on :%s", httpPort)
		if err := service.StartHTTPServer(fmt.Sprintf(":%s", httpPort), svc); err != nil {
			log.Fatalf("http interop: %v", err)
		}
	}()

	log.Printf("network-manager external listening on :%s (mTLS)", externalPort)
	if err := externalSrv.Serve(externalLis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

// detectOrLoadAddress returns the gRPC address to advertise to peers.
// Admin-persisted value in app_settings wins; falls back to interface enumeration.
func detectOrLoadAddress(ctx context.Context, settings *store.SettingsStore, port string) string {
	if addr, err := settings.Get(ctx, "grpc_address"); err == nil && addr != "" {
		log.Printf("network-manager grpc address (from db): %s", addr)
		return addr
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("warning: enumerate interfaces: %v", err)
		return ""
	}

	var private string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			addr := fmt.Sprintf("%s:%s", ip.String(), port)
			if !isRFC1918(ip) {
				log.Printf("network-manager grpc address (auto-detected, public): %s", addr)
				return addr
			}
			if private == "" {
				private = addr
			}
		}
	}

	if private != "" {
		log.Printf("network-manager grpc address (auto-detected, private): %s", private)
		return private
	}
	log.Printf("warning: could not auto-detect grpc address; configure via admin settings")
	return ""
}

func isRFC1918(ip net.IP) bool {
	for _, cidr := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"} {
		_, block, _ := net.ParseCIDR(cidr)
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
