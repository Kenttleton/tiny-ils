package service

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	curiospb "tiny-ils/gen/curiospb"
	pb "tiny-ils/gen/networkpb"
	"tiny-ils/network-manager/store"
	"tiny-ils/shared/identity"
)

type NetworkService struct {
	pb.UnimplementedNetworkManagerServer
	peers        *store.PeerStore
	nodeID       string
	pubKey       ed25519.PublicKey
	privKey      ed25519.PrivateKey
	pubKeyB64    string
	curiosClient curiospb.CuriosManagerClient
}

func NewNetworkService(peers *store.PeerStore, nodeID string, pub ed25519.PublicKey, priv ed25519.PrivateKey, curiosClient curiospb.CuriosManagerClient) *NetworkService {
	return &NetworkService{
		peers:        peers,
		nodeID:       nodeID,
		pubKey:       pub,
		privKey:      priv,
		pubKeyB64:    base64.StdEncoding.EncodeToString(pub),
		curiosClient: curiosClient,
	}
}

// ─── Peer registry ───────────────────────────────────────────────────────────

func (s *NetworkService) RegisterPeer(ctx context.Context, req *pb.PeerInfo) (*pb.PeerAck, error) {
	if req.NodeId == "" || req.PublicKey == "" || req.Address == "" {
		return nil, status.Errorf(codes.InvalidArgument, "node_id, public_key, and address are required")
	}
	if err := s.peers.Upsert(ctx, &store.Peer{
		NodeID:      req.NodeId,
		PublicKey:   req.PublicKey,
		Address:     req.Address,
		DisplayName: req.DisplayName,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "register peer: %v", err)
	}
	return &pb.PeerAck{NodeId: s.nodeID, PublicKey: s.pubKeyB64}, nil
}

func (s *NetworkService) ListPeers(ctx context.Context, _ *pb.Empty) (*pb.PeerList, error) {
	peers, err := s.peers.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list peers: %v", err)
	}
	var pbPeers []*pb.PeerInfo
	for _, p := range peers {
		pbPeers = append(pbPeers, &pb.PeerInfo{
			NodeId:      p.NodeID,
			PublicKey:   p.PublicKey,
			Address:     p.Address,
			DisplayName: p.DisplayName,
		})
	}
	return &pb.PeerList{Peers: pbPeers}, nil
}

// ─── Cross-node search (streaming) ───────────────────────────────────────────

// SearchNetwork fans out to all known peers and streams results back as each responds.
func (s *NetworkService) SearchNetwork(req *pb.NetworkSearchRequest, stream pb.NetworkManager_SearchNetworkServer) error {
	ctx := stream.Context()
	peers, err := s.peers.List(ctx)
	if err != nil {
		return status.Errorf(codes.Internal, "list peers: %v", err)
	}

	results := make(chan *pb.NetworkSearchResult, len(peers))
	var wg sync.WaitGroup

	for _, p := range peers {
		wg.Add(1)
		go func(peer *store.Peer) {
			defer wg.Done()
			results <- queryPeer(ctx, peer, req)
		}(p)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if err := stream.Send(result); err != nil {
			return err
		}
	}
	return nil
}

// queryPeer opens a gRPC connection to a peer's network-manager and requests a
// catalog search. The remote node's network-manager forwards the request to its
// local curios-manager.
// TODO: implement the remote CuriosManager query once service-to-service
// routing is finalized. Currently returns an empty result as a stub.
func queryPeer(ctx context.Context, peer *store.Peer, req *pb.NetworkSearchRequest) *pb.NetworkSearchResult {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(peer.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return &pb.NetworkSearchResult{
			NodeId:   peer.NodeID,
			NodeName: peer.DisplayName,
			Error:    fmt.Sprintf("connect: %v", err),
		}
	}
	defer conn.Close()

	// Stub: fan-out search to the remote network-manager is defined in the proto
	// but remote curio querying requires the remote node to expose a search endpoint.
	// This will be wired up in a follow-on implementation.
	_ = req
	return &pb.NetworkSearchResult{
		NodeId:   peer.NodeID,
		NodeName: peer.DisplayName,
		Curios:   nil, // TODO: wire up remote curios query
	}
}

// ─── Catalog federation ───────────────────────────────────────────────────────

func (s *NetworkService) ShareCatalog(_ context.Context, req *pb.CatalogSnapshot) (*pb.SyncAck, error) {
	log.Printf("received catalog snapshot from node %s: %d curios", req.NodeId, len(req.Curios))
	// TODO: persist/index remote catalog for local search augmentation
	return &pb.SyncAck{Received: int32(len(req.Curios))}, nil
}

// ─── Cross-node borrowing ─────────────────────────────────────────────────────

func (s *NetworkService) RequestBorrow(ctx context.Context, req *pb.BorrowRequest) (*pb.BorrowResponse, error) {
	if req.UserJwt == "" {
		return nil, status.Errorf(codes.Unauthenticated, "user_jwt required")
	}
	if _, err := s.verifyForeignJWT(ctx, req.UserJwt, req.RequestingNode); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	// TODO: delegate to local curios-manager gRPC client once wired in
	return &pb.BorrowResponse{
		LoanId:  "stub-loan-id",
		DueDate: time.Now().Add(14 * 24 * time.Hour).Unix(),
	}, nil
}

func (s *NetworkService) ReturnCurio(ctx context.Context, req *pb.ReturnRequest) (*pb.ReturnResponse, error) {
	if req.UserJwt == "" {
		return nil, status.Errorf(codes.Unauthenticated, "user_jwt required")
	}
	if _, err := s.verifyForeignJWT(ctx, req.UserJwt, ""); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	// TODO: delegate to local curios-manager
	return &pb.ReturnResponse{ReturnedAt: time.Now().Unix()}, nil
}

// ─── Identity verification ────────────────────────────────────────────────────

func (s *NetworkService) VerifyUser(ctx context.Context, req *pb.UserToken) (*pb.VerificationResult, error) {
	claims, err := s.verifyForeignJWT(ctx, req.Jwt, "")
	if err != nil {
		return &pb.VerificationResult{Valid: false, Error: err.Error()}, nil
	}
	return &pb.VerificationResult{
		Valid:    true,
		UserId:   claims.UserID,
		HomeNode: claims.Issuer,
	}, nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// verifyForeignJWT verifies a JWT issued by a peer node.
// It extracts the issuer from the unverified payload, looks up the peer's
// public key, then performs full signature verification.
func (s *NetworkService) verifyForeignJWT(ctx context.Context, tokenStr, claimedNodeID string) (*identity.NodeClaims, error) {
	issuer, err := extractJWTIssuer(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("parse token issuer: %w", err)
	}
	if claimedNodeID != "" && claimedNodeID != issuer {
		return nil, fmt.Errorf("token issuer %q does not match claimed node %q", issuer, claimedNodeID)
	}
	pubKeyB64, err := s.peers.GetPublicKey(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("unknown node %q: %w", issuer, err)
	}
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyB64)
	if err != nil {
		return nil, fmt.Errorf("decode peer public key: %w", err)
	}
	return identity.VerifyToken(tokenStr, ed25519.PublicKey(pubKeyBytes))
}

// extractJWTIssuer decodes the JWT payload segment (without signature verification)
// to read the "iss" claim, used only to identify which peer key to look up.
func extractJWTIssuer(tokenStr string) (string, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("malformed JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decode JWT payload: %w", err)
	}
	var claims struct {
		Iss string `json:"iss"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("unmarshal JWT payload: %w", err)
	}
	if claims.Iss == "" {
		return "", fmt.Errorf("JWT missing iss claim")
	}
	return claims.Iss, nil
}
