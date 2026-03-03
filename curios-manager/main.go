package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "tiny-ils/gen/curiospb"
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
	transferStore := store.NewTransferStore(pool)

	var lcpClient *lcp.Client
	if lcpURL := os.Getenv("LCP_SERVER_URL"); lcpURL != "" {
		lsdURL := os.Getenv("LSD_SERVER_URL")
		lcpClient = lcp.NewClient(lcpURL, lsdURL)
		log.Printf("LCP client configured: lcp=%s lsd=%s", lcpURL, lsdURL)
	}

	svc := service.NewCuriosService(curioStore, loanStore, leaseStore, transferStore, lcpClient)

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50151"
	}

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
