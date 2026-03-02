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

	curiospb "tiny-ils/gen/curiospb"
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
	log.Printf("node identity: %s", nodeID)

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	// Connect to local curios-manager for cross-node transfer delegation
	curiosAddr := envOr("CURIOS_GRPC", "localhost:50051")
	curiosConn, err := grpc.NewClient(curiosAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("curios-manager client: %v", err)
	}
	defer curiosConn.Close()
	curiosClient := curiospb.NewCuriosManagerClient(curiosConn)

	peerStore := store.NewPeerStore(pool)
	svc := service.NewNetworkService(peerStore, nodeID, pubKey, privKey, curiosClient)

	port := envOr("GRPC_PORT", "50053")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterNetworkManagerServer(srv, svc)
	reflection.Register(srv)

	log.Printf("network-manager listening on :%s", port)
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
