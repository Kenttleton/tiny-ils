package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

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

	srv := grpc.NewServer()
	pb.RegisterUsersManagerServer(srv, svc)
	reflection.Register(srv)

	log.Printf("users-manager listening on :%s", port)
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
