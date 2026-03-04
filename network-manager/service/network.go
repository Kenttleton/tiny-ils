package service

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	curiospb "tiny-ils/gen/curiospb"
	pb "tiny-ils/gen/networkpb"
	userspb "tiny-ils/gen/userspb"
	"tiny-ils/network-manager/store"
	"tiny-ils/shared/identity"
)

type NetworkService struct {
	pb.UnimplementedNetworkManagerServer
	peers     *store.PeerStore
	nodeID    string
	pubKey    ed25519.PublicKey
	privKey   ed25519.PrivateKey
	pubKeyB64 string
	nodeCert  tls.Certificate
	registry  *LocalDirectoryService
}

func NewNetworkService(peers *store.PeerStore, nodeID string, pub ed25519.PublicKey, priv ed25519.PrivateKey, nodeCert tls.Certificate, registry *LocalDirectoryService) *NetworkService {
	return &NetworkService{
		peers:     peers,
		nodeID:    nodeID,
		pubKey:    pub,
		privKey:   priv,
		pubKeyB64: base64.StdEncoding.EncodeToString(pub),
		nodeCert:  nodeCert,
		registry:  registry,
	}
}

// firstCuriosSvc returns the first registered curios client, or nil if none registered.
func (s *NetworkService) firstCuriosSvc() curiospb.CuriosManagerClient {
	svcs := s.registry.CuriosSvcs()
	if len(svcs) == 0 {
		return nil
	}
	return svcs[0].client
}

// ─── Peer registry ───────────────────────────────────────────────────────────

func (s *NetworkService) GetNodeInfo(_ context.Context, _ *pb.Empty) (*pb.PeerAck, error) {
	return &pb.PeerAck{NodeId: s.nodeID, PublicKey: s.pubKeyB64, Capabilities: s.registry.Capabilities()}, nil
}

// RegisterPeer handles inbound connection requests from remote peer nodes.
// The calling node is stored as PENDING unless this node's admin has already
// pre-registered them (which upgrades them directly to CONNECTED).
func (s *NetworkService) RegisterPeer(ctx context.Context, req *pb.PeerInfo) (*pb.PeerAck, error) {
	if req.NodeId == "" || req.PublicKey == "" || req.Address == "" {
		return nil, status.Errorf(codes.InvalidArgument, "node_id, public_key, and address are required")
	}
	// Verify that the presented mTLS cert matches the claimed public key.
	cert := clientCert(ctx)
	if cert != nil {
		pub, ok := cert.PublicKey.(ed25519.PublicKey)
		if !ok || base64.StdEncoding.EncodeToString(pub) != req.PublicKey {
			return nil, status.Errorf(codes.PermissionDenied, "cert public key does not match claimed public_key")
		}
	}
	if _, err := s.peers.UpsertInbound(ctx, &store.Peer{
		NodeID:       req.NodeId,
		PublicKey:    req.PublicKey,
		Address:      req.Address,
		DisplayName:  req.DisplayName,
		Capabilities: req.Capabilities,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "register peer: %v", err)
	}
	return &pb.PeerAck{NodeId: s.nodeID, PublicKey: s.pubKeyB64, Capabilities: s.registry.Capabilities()}, nil
}

// ConnectPeer is the admin-initiated outbound connection flow.
// It stores the target peer as CONNECTED locally (explicit admin trust), then
// calls the remote node's RegisterPeer to let them know about us.
func (s *NetworkService) ConnectPeer(ctx context.Context, req *pb.PeerInfo) (*pb.PeerAck, error) {
	if req.NodeId == "" || req.PublicKey == "" || req.Address == "" {
		return nil, status.Errorf(codes.InvalidArgument, "node_id, public_key, and address are required")
	}
	// Store the peer as CONNECTED locally (admin explicitly initiated this).
	if err := s.peers.Upsert(ctx, &store.Peer{
		NodeID:       req.NodeId,
		PublicKey:    req.PublicKey,
		Address:      req.Address,
		DisplayName:  req.DisplayName,
		Capabilities: req.Capabilities,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "store peer: %v", err)
	}
	// Attempt to register ourselves with the remote node.
	ack, err := s.callRemoteRegister(req.Address)
	if err != nil {
		// Remote call failed, but we still stored them CONNECTED locally.
		log.Printf("ConnectPeer: remote RegisterPeer on %s failed: %v", req.Address, err)
		return &pb.PeerAck{NodeId: s.nodeID, PublicKey: s.pubKeyB64}, nil
	}
	return ack, nil
}

// ApprovePeer upgrades a PENDING inbound peer to CONNECTED.
func (s *NetworkService) ApprovePeer(ctx context.Context, req *pb.NodeId) (*pb.Empty, error) {
	if req.NodeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "node_id is required")
	}
	if err := s.peers.Approve(ctx, req.NodeId); err != nil {
		return nil, status.Errorf(codes.Internal, "approve peer: %v", err)
	}
	return &pb.Empty{}, nil
}

func (s *NetworkService) ListPeers(ctx context.Context, _ *pb.Empty) (*pb.PeerList, error) {
	peers, err := s.peers.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list peers: %v", err)
	}
	var pbPeers []*pb.PeerInfo
	for _, p := range peers {
		pbPeers = append(pbPeers, &pb.PeerInfo{
			NodeId:       p.NodeID,
			PublicKey:    p.PublicKey,
			Address:      p.Address,
			DisplayName:  p.DisplayName,
			Status:       p.Status,
			Capabilities: p.Capabilities,
		})
	}
	return &pb.PeerList{Peers: pbPeers}, nil
}

// ─── Cross-node search (streaming) ───────────────────────────────────────────

// SearchNetwork fans out to CONNECTED peers that offer the "curios" capability.
func (s *NetworkService) SearchNetwork(req *pb.NetworkSearchRequest, stream pb.NetworkManager_SearchNetworkServer) error {
	ctx := stream.Context()
	peers, err := s.peers.ListConnectedWithCapability(ctx, "curios")
	if err != nil {
		return status.Errorf(codes.Internal, "list peers: %v", err)
	}

	results := make(chan *pb.NetworkSearchResult, len(peers))
	var wg sync.WaitGroup

	for _, p := range peers {
		wg.Add(1)
		go func(peer *store.Peer) {
			defer wg.Done()
			results <- s.queryPeer(ctx, peer, req)
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

func (s *NetworkService) queryPeer(ctx context.Context, peer *store.Peer, req *pb.NetworkSearchRequest) *pb.NetworkSearchResult {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(peer.Address, PeerDialOptions(s.nodeCert)...)
	if err != nil {
		return &pb.NetworkSearchResult{
			NodeId:   peer.NodeID,
			NodeName: peer.DisplayName,
			Error:    fmt.Sprintf("connect: %v", err),
		}
	}
	defer conn.Close()

	result, err := pb.NewNetworkManagerClient(conn).SearchCatalog(ctx, req)
	if err != nil {
		return &pb.NetworkSearchResult{
			NodeId:   peer.NodeID,
			NodeName: peer.DisplayName,
			Error:    fmt.Sprintf("search: %v", err),
		}
	}
	result.NodeId = peer.NodeID
	result.NodeName = peer.DisplayName
	return result
}

// SearchCatalog searches this node's local catalog and returns a single result.
// It is called by remote peer nodes as part of their SearchNetwork fan-out.
// Returns an empty result (not an error) when no curios service is registered.
func (s *NetworkService) SearchCatalog(ctx context.Context, req *pb.NetworkSearchRequest) (*pb.NetworkSearchResult, error) {
	curios := s.firstCuriosSvc()
	if curios == nil {
		return &pb.NetworkSearchResult{}, nil
	}
	resp, err := curios.ListCurios(ctx, &curiospb.ListCuriosRequest{
		Query:     req.Query,
		MediaType: req.MediaType,
		Limit:     50,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search catalog: %v", err)
	}
	result := &pb.NetworkSearchResult{}
	for _, c := range resp.Curios {
		result.Curios = append(result.Curios, &pb.CurioSummary{
			Id:         c.Id,
			Title:      c.Title,
			MediaType:  c.MediaType,
			FormatType: c.FormatType,
			Tags:       c.Tags,
			Available:  true,
			NodeId:     s.nodeID,
		})
	}
	return result, nil
}

// ─── Catalog federation ───────────────────────────────────────────────────────

func (s *NetworkService) ShareCatalog(_ context.Context, req *pb.CatalogSnapshot) (*pb.SyncAck, error) {
	log.Printf("received catalog snapshot from node %s: %d curios", req.NodeId, len(req.Curios))
	return &pb.SyncAck{Received: int32(len(req.Curios))}, nil
}

// ─── Cross-node borrowing ─────────────────────────────────────────────────────

func (s *NetworkService) RequestBorrow(ctx context.Context, req *pb.BorrowRequest) (*pb.BorrowResponse, error) {
	if req.UserJwt == "" || req.CurioId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_jwt and curio_id are required")
	}
	claims, err := s.verifyForeignJWT(ctx, req.UserJwt, req.RequestingNode)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	curios := s.firstCuriosSvc()
	if curios == nil {
		return nil, status.Errorf(codes.Unavailable, "no curios service registered")
	}

	copyID := req.CopyId

	// If no specific copy requested, look for an available physical copy.
	if copyID == "" {
		copies, err := curios.ListCopies(ctx, &curiospb.CurioId{Id: req.CurioId})
		if err == nil {
			for _, c := range copies.Copies {
				if c.Status == "AVAILABLE" {
					copyID = c.Id
					break
				}
			}
		}
	}

	if copyID != "" {
		// Physical checkout (ISO 18626: item supply)
		var dueDate int64
		if req.NeedBefore > 0 {
			dueDate = req.NeedBefore
		}
		loan, err := curios.CheckoutCopy(ctx, &curiospb.CheckoutRequest{
			CopyId:     copyID,
			UserId:     claims.UserID,
			UserNodeId: claims.Issuer,
			DueDate:    dueDate,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "checkout: %v", err)
		}
		return &pb.BorrowResponse{
			LoanId:    loan.Id,
			DueDate:   loan.DueDate,
			CopyId:    copyID,
			IsDigital: false,
		}, nil
	}

	// No physical copy available — try digital lease.
	lease, err := curios.IssueLease(ctx, &curiospb.LeaseRequest{
		CurioId:    req.CurioId,
		UserId:     claims.UserID,
		UserNodeId: claims.Issuer,
	})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no available copy or digital seat: %v", err)
	}
	return &pb.BorrowResponse{
		LoanId:      lease.Id,
		IsDigital:   true,
		ExpiresAt:   lease.ExpiresAt,
		AccessToken: lease.AccessToken,
	}, nil
}

func (s *NetworkService) ReturnCurio(ctx context.Context, req *pb.ReturnRequest) (*pb.ReturnResponse, error) {
	if req.UserJwt == "" {
		return nil, status.Errorf(codes.Unauthenticated, "user_jwt required")
	}
	if _, err := s.verifyForeignJWT(ctx, req.UserJwt, ""); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	curios := s.firstCuriosSvc()
	if curios == nil {
		return nil, status.Errorf(codes.Unavailable, "no curios service registered")
	}

	if req.CopyId != "" {
		// Physical return.
		loan, err := curios.ReturnCopy(ctx, &curiospb.ReturnRequest{CopyId: req.CopyId})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "return copy: %v", err)
		}
		return &pb.ReturnResponse{ReturnedAt: loan.ReturnedAt}, nil
	}

	// Digital revoke — early return supported; RevokeLease sets revoked=true regardless of expiry.
	if req.LoanId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "copy_id or loan_id is required")
	}
	if _, err := curios.RevokeLease(ctx, &curiospb.LeaseId{Id: req.LoanId}); err != nil {
		return nil, status.Errorf(codes.Internal, "revoke lease: %v", err)
	}
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

// ─── Cross-node user authentication ──────────────────────────────────────────

// IssueGuestToken is called by a CONNECTED peer that wants to authenticate one
// of this node's users on their node. It verifies the user exists locally and
// returns a short-lived JWT scoped to the requesting node only.
func (s *NetworkService) IssueGuestToken(ctx context.Context, req *pb.GuestTokenRequest) (*pb.GuestTokenResponse, error) {
	if req.UserId == "" || req.RequestingNode == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id and requesting_node are required")
	}
	usersSvc, err := s.registry.GetAnyUsersSvc()
	if err != nil || usersSvc == nil {
		return nil, status.Errorf(codes.Unavailable, "no users service registered")
	}
	// Verify the user exists on this node.
	user, err := usersSvc.client.GetMe(ctx, &userspb.UserId{Id: req.UserId})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}
	// Issue an audience-scoped JWT — only valid for the requesting node.
	token, err := identity.IssueTokenForAudience(user.Id, s.nodeID, req.RequestingNode, s.privKey, time.Hour)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "issue token: %v", err)
	}
	return &pb.GuestTokenResponse{Token: token, DisplayName: user.DisplayName}, nil
}

// AuthenticateGuest is called by the local BFF to log in a cross-node user.
// It dials the home node, obtains a guest token, creates a thin local user
// record (via users-manager), and returns a local session JWT.
func (s *NetworkService) AuthenticateGuest(ctx context.Context, req *pb.AuthenticateGuestRequest) (*pb.AuthenticateGuestResponse, error) {
	if req.UserId == "" || req.HomeNodeAddress == "" || req.HomeNodeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id, home_node_address, and home_node_id are required")
	}
	// Verify the home node is CONNECTED — cross-node auth requires full trust.
	homePeer, err := s.peers.Get(ctx, req.HomeNodeId)
	if err != nil || homePeer == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "home node %q is not a CONNECTED partner", req.HomeNodeId)
	}
	// Capability guard: nodes without a users capability cannot issue guest tokens.
	if len(homePeer.Capabilities) > 0 && !slices.Contains(homePeer.Capabilities, "users") {
		return nil, status.Errorf(codes.FailedPrecondition, "home node %q does not have the users capability", req.HomeNodeId)
	}

	// Dial the home node and request a guest token scoped to this node.
	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(req.HomeNodeAddress, PeerDialOptions(s.nodeCert)...)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "dial home node: %v", err)
	}
	defer conn.Close()

	homeClient := pb.NewNetworkManagerClient(conn)
	guestResp, err := homeClient.IssueGuestToken(dialCtx, &pb.GuestTokenRequest{
		UserId:        req.UserId,
		RequestingNode: s.nodeID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "home node rejected token request: %v", err)
	}

	// Create or update a thin local user record and issue a local session JWT.
	usersSvc, usersErr := s.registry.GetAnyUsersSvc()
	if usersErr != nil || usersSvc == nil {
		return nil, status.Errorf(codes.Unavailable, "no users service registered")
	}
	authResp, err := usersSvc.client.UpsertGuestUser(ctx, &userspb.UpsertGuestUserRequest{
		UserId:      req.UserId,
		HomeNodeId:  req.HomeNodeId,
		DisplayName: guestResp.DisplayName,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "upsert guest user: %v", err)
	}
	return &pb.AuthenticateGuestResponse{
		Token:       authResp.Token,
		DisplayName: guestResp.DisplayName,
	}, nil
}

// ─── Cross-node loans / leases ────────────────────────────────────────────────

// GetUserLoans streams loan/lease records for a user across this node and all
// CONNECTED peers — one UserLoansResult message per responding node.
func (s *NetworkService) GetUserLoans(req *pb.UserLoansRequest, stream pb.NetworkManager_GetUserLoansServer) error {
	ctx := stream.Context()

	// Send local results first.
	local, localErr := s.localUserLoans(ctx, req)
	local.NodeId = s.nodeID
	local.NodeName = "this node"
	if err := stream.Send(local); err != nil {
		return err
	}
	if localErr != nil {
		// localErr already captured in local.Error; continue to peers.
		log.Printf("GetUserLoans: local query error: %v", localErr)
	}

	// Fan out to CONNECTED peers that hold catalog/loan records.
	peers, err := s.peers.ListConnectedWithCapability(ctx, "curios")
	if err != nil {
		return status.Errorf(codes.Internal, "list peers: %v", err)
	}

	results := make(chan *pb.UserLoansResult, len(peers))
	var wg sync.WaitGroup
	for _, p := range peers {
		wg.Add(1)
		go func(peer *store.Peer) {
			defer wg.Done()
			results <- s.queryPeerLoans(ctx, peer, req)
		}(p)
	}
	go func() { wg.Wait(); close(results) }()

	for result := range results {
		if err := stream.Send(result); err != nil {
			return err
		}
	}
	return nil
}

// localUserLoans queries the local curios-manager for a user's physical loans
// and digital leases, returning them as a UserLoansResult.
// Returns an empty result when no curios service is registered.
func (s *NetworkService) localUserLoans(ctx context.Context, req *pb.UserLoansRequest) (*pb.UserLoansResult, error) {
	result := &pb.UserLoansResult{}
	curios := s.firstCuriosSvc()
	if curios == nil {
		return result, nil
	}

	loanResp, err := curios.ListLoans(ctx, &curiospb.ListLoansRequest{
		UserId:     req.UserId,
		UserNodeId: req.UserNodeId,
		ActiveOnly: req.ActiveOnly,
		Limit:      200,
	})
	if err != nil {
		result.Error = fmt.Sprintf("list loans: %v", err)
		return result, err
	}
	for _, l := range loanResp.Loans {
		result.Loans = append(result.Loans, &pb.RemoteLoan{
			LoanId:     l.Id,
			CurioId:    l.CurioId,
			CurioTitle: l.CurioTitle,
			IsDigital:  false,
			IssuedAt:   l.CheckedOut,
			DueDate:    l.DueDate,
			Closed:     l.ReturnedAt != 0,
		})
	}

	leaseResp, err := curios.ListLeases(ctx, &curiospb.ListLeasesRequest{
		UserId:     req.UserId,
		UserNodeId: req.UserNodeId,
		ActiveOnly: req.ActiveOnly,
	})
	if err != nil {
		result.Error = fmt.Sprintf("list leases: %v", err)
		return result, err
	}
	for _, l := range leaseResp.Leases {
		result.Loans = append(result.Loans, &pb.RemoteLoan{
			LoanId:    l.Id,
			CurioId:   l.AssetId,
			IsDigital: true,
			IssuedAt:  l.IssuedAt,
			ExpiresAt: l.ExpiresAt,
			Closed:    l.Revoked,
		})
	}
	return result, nil
}

// queryPeerLoans calls GetUserLoans on a remote peer and collects the first result.
func (s *NetworkService) queryPeerLoans(ctx context.Context, peer *store.Peer, req *pb.UserLoansRequest) *pb.UserLoansResult {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(peer.Address, PeerDialOptions(s.nodeCert)...)
	if err != nil {
		return &pb.UserLoansResult{NodeId: peer.NodeID, NodeName: peer.DisplayName, Error: fmt.Sprintf("connect: %v", err)}
	}
	defer conn.Close()

	stream, err := pb.NewNetworkManagerClient(conn).GetUserLoans(ctx, req)
	if err != nil {
		return &pb.UserLoansResult{NodeId: peer.NodeID, NodeName: peer.DisplayName, Error: fmt.Sprintf("rpc: %v", err)}
	}
	// We only want the first message from the remote (its local result).
	msg, err := stream.Recv()
	if err != nil {
		return &pb.UserLoansResult{NodeId: peer.NodeID, NodeName: peer.DisplayName, Error: fmt.Sprintf("recv: %v", err)}
	}
	return msg
}

// ─── Cross-node digital leasing ──────────────────────────────────────────────

func (s *NetworkService) RequestDigitalLease(ctx context.Context, req *pb.DigitalLeaseRequest) (*pb.DigitalLeaseResponse, error) {
	if req.UserJwt == "" {
		return nil, status.Errorf(codes.Unauthenticated, "user_jwt required")
	}
	claims, err := s.verifyForeignJWT(ctx, req.UserJwt, req.RequestingNode)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	curios := s.firstCuriosSvc()
	if curios == nil {
		return nil, status.Errorf(codes.Unavailable, "no curios service registered")
	}
	lease, err := curios.IssueLease(ctx, &curiospb.LeaseRequest{
		CurioId:    req.CurioId,
		UserId:     claims.UserID,
		UserNodeId: claims.Issuer,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "issue lease: %v", err)
	}
	return &pb.DigitalLeaseResponse{
		LeaseId:     lease.Id,
		AccessToken: lease.AccessToken,
		LicenseUrl:  lease.LicenseUrl,
		ExpiresAt:   lease.ExpiresAt,
	}, nil
}

func (s *NetworkService) RevokeDigitalLease(ctx context.Context, req *pb.LeaseRef) (*pb.Empty, error) {
	if req.UserJwt == "" {
		return nil, status.Errorf(codes.Unauthenticated, "user_jwt required")
	}
	if _, err := s.verifyForeignJWT(ctx, req.UserJwt, ""); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	curios := s.firstCuriosSvc()
	if curios == nil {
		return nil, status.Errorf(codes.Unavailable, "no curios service registered")
	}
	if _, err := curios.RevokeLease(ctx, &curiospb.LeaseId{Id: req.LeaseId}); err != nil {
		return nil, status.Errorf(codes.Internal, "revoke lease: %v", err)
	}
	return &pb.Empty{}, nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// callRemoteRegister dials the remote node and calls its RegisterPeer with our identity.
func (s *NetworkService) callRemoteRegister(address string) (*pb.PeerAck, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(address, PeerDialOptions(s.nodeCert)...)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()
	client := pb.NewNetworkManagerClient(conn)
	return client.RegisterPeer(ctx, &pb.PeerInfo{
		NodeId:       s.nodeID,
		PublicKey:    s.pubKeyB64,
		Address:      "", // remote cannot call back via this field; use address stored at registration
		Capabilities: s.registry.Capabilities(),
	})
}

// verifyForeignJWT verifies a JWT issued by a peer node.
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

// extractJWTIssuer decodes the JWT payload segment (without verification)
// to read the "iss" claim, used to identify which peer key to look up.
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
