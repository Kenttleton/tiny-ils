package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	curiospb "tiny-ils/gen/curiospb"
	pb "tiny-ils/gen/networkpb"
	userspb "tiny-ils/gen/userspb"
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

	// Write the cert to disk so the local BFF can load it for mTLS.
	certPath := envOr("NODE_CERT_PATH", "/data/node.crt")
	if err := identity.WriteCertPEM(certPath, nodeCert); err != nil {
		log.Printf("warning: write cert PEM: %v", err)
	}

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	// Connect to local curios-manager (internal, no TLS needed).
	curiosAddr := envOr("CURIOS_GRPC", "localhost:50151")
	curiosConn, err := grpc.NewClient(curiosAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("curios-manager client: %v", err)
	}
	defer curiosConn.Close()
	curiosClient := curiospb.NewCuriosManagerClient(curiosConn)

	usersAddr := envOr("USERS_GRPC", "localhost:50152")
	usersConn, err := grpc.NewClient(usersAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("users-manager client: %v", err)
	}
	defer usersConn.Close()
	usersClient := userspb.NewUsersManagerClient(usersConn)

	// Parse node capabilities from env; default to full stack.
	capStr := envOr("NODE_CAPABILITIES", "curios,users,ui")
	capabilities := strings.Split(capStr, ",")
	for i, c := range capabilities {
		capabilities[i] = strings.TrimSpace(c)
	}

	peerStore := store.NewPeerStore(pool)
	svc := service.NewNetworkService(peerStore, nodeID, pubKey, privKey, nodeCert, capabilities, curiosClient, usersClient)

	port := envOr("GRPC_PORT", "50153")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	// mTLS: client certs requested but not required.
	// The trust interceptor handles authorization based on cert identity.
	tlsCfg := service.ServerTLSConfig(nodeCert)
	srv := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsCfg)),
		grpc.ChainUnaryInterceptor(service.TrustInterceptor(peerStore, pubKeyB64)),
		grpc.ChainStreamInterceptor(service.TrustStreamInterceptor(peerStore, pubKeyB64)),
	)
	pb.RegisterNetworkManagerServer(srv, svc)
	reflection.Register(srv)

	// Start HTTP interop server (digital passthrough, ISO 18626 stub, SRU stub).
	httpPort := envOr("HTTP_PORT", "8153")
	go func() {
		log.Printf("network-manager HTTP interop listening on :%s", httpPort)
		if err := service.StartHTTPServer(fmt.Sprintf(":%s", httpPort), svc); err != nil {
			log.Fatalf("http interop: %v", err)
		}
	}()

	log.Printf("network-manager listening on :%s (mTLS)", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
