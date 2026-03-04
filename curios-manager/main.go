package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "tiny-ils/gen/curiospb"
	internalpb "tiny-ils/gen/internalpb"
	networkpb "tiny-ils/gen/networkpb"
	"tiny-ils/curios-manager/lcp"
	"tiny-ils/curios-manager/service"
	"tiny-ils/curios-manager/store"
	"tiny-ils/shared/db"
	"tiny-ils/shared/identity"
)

func main() {
	ctx := context.Background()

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	privKey, pubKey, err := identity.LoadOrCreate(
		envOr("NODE_KEY_PATH", "/data/node.key"),
		envOr("NODE_PUBKEY_PATH", "/data/node.pub"),
	)
	if err != nil {
		log.Printf("warning: node keypair unavailable, cross-node features disabled: %v", err)
	}
	_ = privKey // available for future use
	nodeID := identity.Fingerprint(pubKey)

	port := envOr("GRPC_PORT", "50151")
	selfAddr := envOr("REGISTER_ADDRESS", "curios-manager:"+port)

	// Derive a stable, opaque service UUID from node identity + registered address.
	// UUIDv5(namespace=UUIDv5(X500, nodeID), name=selfAddr) is unique per
	// (node, address) pair and cannot be reversed to recover either input.
	nodeNS := uuid.NewSHA1(uuid.NameSpaceX500, []byte(nodeID))
	serviceID := uuid.NewSHA1(nodeNS, []byte(selfAddr))

	curioStore := store.NewCurioStore(pool, serviceID)
	loanStore := store.NewLoanStore(pool, nodeID, serviceID)
	leaseStore := store.NewLeaseStore(pool, serviceID)
	transferStore := store.NewTransferStore(pool, nodeID)

	var lcpClient *lcp.Client
	if lcpURL := os.Getenv("LCP_SERVER_URL"); lcpURL != "" {
		lsdURL := os.Getenv("LSD_SERVER_URL")
		lcpClient = lcp.NewClient(lcpURL, lsdURL)
		log.Printf("LCP client configured: lcp=%s lsd=%s", lcpURL, lsdURL)
	}

	// Connect to network-manager internal port (insecure, Docker-private).
	networkAddr := envOr("NETWORK_GRPC", "localhost:50154")
	networkConn, err := grpc.NewClient(networkAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("network-manager client: %v", err)
	}
	defer networkConn.Close()
	networkClient := networkpb.NewNetworkManagerClient(networkConn)

	svc := service.NewCuriosService(curioStore, loanStore, leaseStore, transferStore, lcpClient, networkClient, nodeID)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterCuriosManagerServer(srv, svc)
	internalpb.RegisterCapabilityServiceServer(srv, &capSvc{name: "curios", addr: selfAddr})
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
	reflection.Register(srv)

	// Announce to network-manager's LocalDirectory. Retries until successful.
	go func() {
		dirClient := internalpb.NewLocalDirectoryClient(networkConn)
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err := dirClient.Announce(ctx, &internalpb.LocalServiceInfo{
				Name:    "curios",
				Address: selfAddr,
			})
			cancel()
			if err == nil {
				log.Printf("announced curios to network-manager @ %s", networkAddr)
				return
			}
			log.Printf("waiting for network-manager (%s): %v", networkAddr, err)
			time.Sleep(2 * time.Second)
		}
	}()

	log.Printf("curios-manager listening on :%s", port)
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
