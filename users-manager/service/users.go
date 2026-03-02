package service

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "tiny-ils/gen/userspb"
	"tiny-ils/shared/identity"
	"tiny-ils/shared/models"
	"tiny-ils/users-manager/store"
)

const tokenTTL = 24 * time.Hour

type UsersService struct {
	pb.UnimplementedUsersManagerServer
	users       *store.UserStore
	claims      *store.ClaimsStore
	nodeID      string            // public key fingerprint of this node
	privKey     ed25519.PrivateKey
}

func NewUsersService(users *store.UserStore, claims *store.ClaimsStore, nodeID string, privKey ed25519.PrivateKey) *UsersService {
	return &UsersService{users: users, claims: claims, nodeID: nodeID, privKey: privKey}
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

func (s *UsersService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email and password are required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "hash password: %v", err)
	}
	u := &models.User{Email: req.Email, DisplayName: req.DisplayName}
	created, err := s.users.Create(ctx, u, string(hash))
	if err != nil {
		return nil, status.Errorf(codes.AlreadyExists, "register: %v", err)
	}
	// New users get a USER claim on this node by default
	if err := s.claims.Grant(ctx, created.ID, s.nodeID, models.RoleUser, created.ID); err != nil {
		return nil, status.Errorf(codes.Internal, "grant default claim: %v", err)
	}
	return s.authResponse(ctx, created)
}

func (s *UsersService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email and password are required")
	}
	u, hash, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}
	return s.authResponse(ctx, u)
}

func (s *UsersService) UpsertSSOUser(ctx context.Context, req *pb.SSOProfile) (*pb.AuthResponse, error) {
	if req.Provider == "" || req.Subject == "" || req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "provider, subject, and email are required")
	}
	u, err := s.users.UpsertSSO(ctx, req.Provider, req.Subject, req.Email, req.DisplayName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "upsert sso user: %v", err)
	}
	// Ensure SSO users have at least a USER claim on this node
	_ = s.claims.Grant(ctx, u.ID, s.nodeID, models.RoleUser, u.ID)
	return s.authResponse(ctx, u)
}

func (s *UsersService) GetMe(ctx context.Context, req *pb.UserId) (*pb.User, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}
	return toPBUser(u), nil
}

// ─── Claims ──────────────────────────────────────────────────────────────────

func (s *UsersService) GrantClaim(ctx context.Context, req *pb.ClaimRequest) (*pb.Empty, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}
	if err := s.claims.Grant(ctx, userID, req.NodeId, models.Role(req.Role), uuid.Nil); err != nil {
		return nil, status.Errorf(codes.Internal, "grant claim: %v", err)
	}
	return &pb.Empty{}, nil
}

func (s *UsersService) RevokeClaim(ctx context.Context, req *pb.ClaimRequest) (*pb.Empty, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}
	if err := s.claims.Revoke(ctx, userID, req.NodeId); err != nil {
		return nil, status.Errorf(codes.Internal, "revoke claim: %v", err)
	}
	return &pb.Empty{}, nil
}

func (s *UsersService) ListClaims(ctx context.Context, req *pb.NodeId) (*pb.ClaimList, error) {
	claims, err := s.claims.ListForNode(ctx, req.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list claims: %v", err)
	}
	var pbClaims []*pb.NodeClaim
	for _, c := range claims {
		pbClaims = append(pbClaims, &pb.NodeClaim{
			UserId:    c.UserID.String(),
			NodeId:    c.NodeID,
			Role:      string(c.Role),
			GrantedBy: c.GrantedBy.String(),
			GrantedAt: c.GrantedAt.Unix(),
		})
	}
	return &pb.ClaimList{Claims: pbClaims}, nil
}

// ─── Bootstrap ───────────────────────────────────────────────────────────────

func (s *UsersService) BootstrapManager(ctx context.Context, req *pb.BootstrapRequest) (*pb.AuthResponse, error) {
	hasManager, err := s.users.HasAnyManager(ctx, s.nodeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check managers: %v", err)
	}
	if hasManager {
		return nil, status.Errorf(codes.FailedPrecondition, "bootstrap already complete: a manager already exists on this node")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "hash password: %v", err)
	}
	u := &models.User{Email: req.Email, DisplayName: req.DisplayName}
	created, err := s.users.Create(ctx, u, string(hash))
	if err != nil {
		// User may already exist — look them up and promote
		existing, _, lookupErr := s.users.GetByEmail(ctx, req.Email)
		if lookupErr != nil {
			return nil, status.Errorf(codes.Internal, "bootstrap create user: %v", err)
		}
		created = existing
	}
	if err := s.claims.Grant(ctx, created.ID, s.nodeID, models.RoleManager, created.ID); err != nil {
		return nil, status.Errorf(codes.Internal, "grant manager claim: %v", err)
	}
	return s.authResponse(ctx, created)
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

func (s *UsersService) authResponse(ctx context.Context, u *models.User) (*pb.AuthResponse, error) {
	userClaims, err := s.claims.ListForUser(ctx, u.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "load claims: %v", err)
	}
	var jwtClaims []models.JWTClaim
	for _, c := range userClaims {
		jwtClaims = append(jwtClaims, models.JWTClaim{Node: c.NodeID, Role: c.Role})
	}
	token, err := identity.IssueToken(u.ID.String(), s.nodeID, jwtClaims, s.privKey, tokenTTL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("issue token: %v", err))
	}
	return &pb.AuthResponse{Token: token, User: toPBUser(u)}, nil
}

func toPBUser(u *models.User) *pb.User {
	return &pb.User{
		Id:          u.ID.String(),
		Email:       u.Email,
		DisplayName: u.DisplayName,
		SsoProvider: u.SSOProvider,
		CreatedAt:   u.CreatedAt.Unix(),
	}
}
