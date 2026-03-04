package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	internalpb "tiny-ils/gen/internalpb"
	pb "tiny-ils/gen/userspb"
	"tiny-ils/shared/db"
	"tiny-ils/shared/identity"
	"tiny-ils/users-manager/service"
	"tiny-ils/users-manager/store"
)

const (
	keyPath    = "/data/node.key"
	pubKeyPath = "/data/node.pub"
)

func main() {
	ctx := context.Background()

	// Load or generate node keypair
	privKey, pubKey, err := identity.LoadOrCreate(
		envOr("NODE_KEY_PATH", keyPath),
		envOr("NODE_PUBKEY_PATH", pubKeyPath),
	)
	if err != nil {
		log.Fatalf("node keypair: %v", err)
	}
	nodeID := identity.Fingerprint(pubKey)

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	userStore := store.NewUserStore(pool)
	claimsStore := store.NewClaimsStore(pool)
	settingsStore := store.NewSettingsStore(pool)
	svc := service.NewUsersService(userStore, claimsStore, settingsStore, nodeID, privKey)

	port := envOr("GRPC_PORT", "50152")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	selfAddr := envOr("REGISTER_ADDRESS", "users-manager:"+port)

	srv := grpc.NewServer()
	pb.RegisterUsersManagerServer(srv, svc)
	internalpb.RegisterCapabilityServiceServer(srv, &capSvc{name: "users", addr: selfAddr})
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
	reflection.Register(srv)

	// Connect to network-manager internal port and announce.
	networkAddr := envOr("NETWORK_GRPC", "localhost:50154")
	go func() {
		networkConn, err := grpc.NewClient(networkAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("users-manager: network-manager client error: %v", err)
			return
		}
		defer networkConn.Close()
		dirClient := internalpb.NewLocalDirectoryClient(networkConn)
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err := dirClient.Announce(ctx, &internalpb.LocalServiceInfo{
				Name:    "users",
				Address: selfAddr,
			})
			cancel()
			if err == nil {
				log.Printf("announced users to network-manager @ %s", networkAddr)
				return
			}
			log.Printf("waiting for network-manager (%s): %v", networkAddr, err)
			time.Sleep(2 * time.Second)
		}
	}()

	log.Printf("users-manager listening on :%s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

// capSvc implements internalpb.CapabilityServiceServer for self-identification.
type capSvc struct {
	internalpb.UnimplementedCapabilityServiceServer
	name string
	addr string
}

func (c *capSvc) WhoAmI(_ context.Context, _ *internalpb.Empty) (*internalpb.ServiceInfo, error) {
	return &internalpb.ServiceInfo{Name: c.name, Address: c.addr}, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
