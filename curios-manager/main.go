package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	pb "tiny-ils/gen/curiospb"
	networkpb "tiny-ils/gen/networkpb"
	"tiny-ils/curios-manager/lcp"
	"tiny-ils/curios-manager/service"
	"tiny-ils/curios-manager/store"
	"tiny-ils/shared/db"
)

func main() {
	ctx := context.Background()

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	curioStore := store.NewCurioStore(pool)
	loanStore := store.NewLoanStore(pool)
	leaseStore := store.NewLeaseStore(pool)
	nodeID := envOr("NODE_ID", "")
	transferStore := store.NewTransferStore(pool, nodeID)

	var lcpClient *lcp.Client
	if lcpURL := os.Getenv("LCP_SERVER_URL"); lcpURL != "" {
		lsdURL := os.Getenv("LSD_SERVER_URL")
		lcpClient = lcp.NewClient(lcpURL, lsdURL)
		log.Printf("LCP client configured: lcp=%s lsd=%s", lcpURL, lsdURL)
	}

	networkAddr := envOr("NETWORK_GRPC", "localhost:50153")
	networkConn, err := grpc.NewClient(networkAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("network-manager client: %v", err)
	}
	defer networkConn.Close()
	networkClient := networkpb.NewNetworkManagerClient(networkConn)

	svc := service.NewCuriosService(curioStore, loanStore, leaseStore, transferStore, lcpClient, networkClient, nodeID)

	port := envOr("GRPC_PORT", "50151")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterCuriosManagerServer(srv, svc)
	reflection.Register(srv) // enables grpcurl

	log.Printf("curios-manager listening on :%s", port)
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
