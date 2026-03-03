package service

import (
	"context"
	"crypto/ed25519"
	"strings"
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
	users    *store.UserStore
	claims   *store.ClaimsStore
	settings *store.SettingsStore
	nodeID   string // public key fingerprint of this node
	privKey  ed25519.PrivateKey
}

func NewUsersService(users *store.UserStore, claims *store.ClaimsStore, settings *store.SettingsStore, nodeID string, privKey ed25519.PrivateKey) *UsersService {
	return &UsersService{users: users, claims: claims, settings: settings, nodeID: nodeID, privKey: privKey}
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
	return s.authResponse(ctx, created, true)
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
	return s.authResponse(ctx, u, hash != "")
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
	return s.authResponse(ctx, u, false)
}

func (s *UsersService) LinkSSO(ctx context.Context, req *pb.LinkSSORequest) (*pb.User, error) {
	if req.Provider == "" || req.Subject == "" {
		return nil, status.Errorf(codes.InvalidArgument, "provider and subject are required")
	}
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}
	u, hash, err := s.users.LinkSSO(ctx, id, req.Provider, req.Subject)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, status.Errorf(codes.AlreadyExists, "that %s account is already linked to another user", req.Provider)
		}
		return nil, status.Errorf(codes.Internal, "link sso: %v", err)
	}
	return toPBUser(u, hash != ""), nil
}

func (s *UsersService) GetMe(ctx context.Context, req *pb.UserId) (*pb.User, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}
	u, hash, err := s.users.GetByIDWithHash(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}
	return toPBUser(u, hash != ""), nil
}

func (s *UsersService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}
	u, hash, err := s.users.GetByIDWithHash(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	newHash := ""
	if req.NewPassword != "" {
		if len(req.NewPassword) < 8 {
			return nil, status.Errorf(codes.InvalidArgument, "password must be at least 8 characters")
		}
		if hash != "" {
			// Already has a password — verify current one first
			if req.CurrentPassword == "" {
				return nil, status.Errorf(codes.InvalidArgument, "current password is required")
			}
			if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.CurrentPassword)); err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "current password is incorrect")
			}
		}
		h, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "hash password: %v", err)
		}
		newHash = string(h)
	}

	// Email changes require a password (existing or being set in this request)
	if req.Email != "" && hash == "" && newHash == "" {
		return nil, status.Errorf(codes.FailedPrecondition, "set a password before changing email on an SSO-only account")
	}

	if req.UnlinkSso {
		if u.SSOProvider == "" {
			return nil, status.Errorf(codes.FailedPrecondition, "no SSO account is linked")
		}
		if hash == "" && newHash == "" {
			return nil, status.Errorf(codes.FailedPrecondition, "set a password before unlinking SSO")
		}
	}

	updated, err := s.users.Update(ctx, id, req.DisplayName, req.Email, newHash, req.UnlinkSso)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update user: %v", err)
	}
	effectiveHash := hash
	if newHash != "" {
		effectiveHash = newHash
	}
	return toPBUser(updated, effectiveHash != ""), nil
}

// ─── Admin user management ────────────────────────────────────────────────────

func (s *UsersService) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.UserList, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50
	}
	users, total, err := s.users.ListUsers(ctx, s.nodeID, limit, int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list users: %v", err)
	}
	var pbUsers []*pb.User
	for _, u := range users {
		pu := toPBUser(&u.User, u.PasswordHash != "")
		pu.Role = u.Role
		pbUsers = append(pbUsers, pu)
	}
	return &pb.UserList{Users: pbUsers, Total: int32(total)}, nil
}

func (s *UsersService) DeleteUser(ctx context.Context, req *pb.UserId) (*pb.Empty, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}
	if err := s.users.DeleteUser(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "delete user: %v", err)
	}
	return &pb.Empty{}, nil
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

// ─── Bootstrap / Setup ───────────────────────────────────────────────────────

func (s *UsersService) HasSetup(ctx context.Context, _ *pb.Empty) (*pb.SetupStatus, error) {
	has, err := s.users.HasAnyManager(ctx, s.nodeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check setup: %v", err)
	}
	return &pb.SetupStatus{HasManager: has}, nil
}

func (s *UsersService) GetNodeID(_ context.Context, _ *pb.Empty) (*pb.NodeId, error) {
	return &pb.NodeId{NodeId: s.nodeID}, nil
}

func (s *UsersService) GetAppConfig(ctx context.Context, _ *pb.Empty) (*pb.AppConfig, error) {
	url, err := s.settings.Get(ctx, "public_url")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get app config: %v", err)
	}
	allowLocalhostStr, err := s.settings.Get(ctx, "allow_localhost")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get app config: %v", err)
	}
	// Default to true when the key is absent.
	allowLocalhost := allowLocalhostStr != "false"
	return &pb.AppConfig{PublicUrl: url, AllowLocalhost: &allowLocalhost}, nil
}

func (s *UsersService) UpdateAppConfig(ctx context.Context, req *pb.AppConfig) (*pb.Empty, error) {
	if req.PublicUrl != "" {
		if err := s.settings.Set(ctx, "public_url", req.PublicUrl); err != nil {
			return nil, status.Errorf(codes.Internal, "update public_url: %v", err)
		}
	}
	if req.AllowLocalhost != nil {
		val := "true"
		if !*req.AllowLocalhost {
			val = "false"
		}
		if err := s.settings.Set(ctx, "allow_localhost", val); err != nil {
			return nil, status.Errorf(codes.Internal, "update allow_localhost: %v", err)
		}
	}
	return &pb.Empty{}, nil
}

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
	if req.PublicUrl != "" {
		if err := s.settings.Set(ctx, "public_url", req.PublicUrl); err != nil {
			return nil, status.Errorf(codes.Internal, "save public url: %v", err)
		}
	}
	return s.authResponse(ctx, created, true)
}

// ─── Cross-node identity ──────────────────────────────────────────────────────

// UpsertGuestUser creates (or refreshes) a thin local record for a cross-node user,
// ensures they have a default USER claim on this node, and issues a local session JWT.
// The user's UUID is preserved from the home node so loan records remain consistent.
func (s *UsersService) UpsertGuestUser(ctx context.Context, req *pb.UpsertGuestUserRequest) (*pb.AuthResponse, error) {
	if req.UserId == "" || req.HomeNodeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id and home_node_id are required")
	}
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id")
	}
	u, err := s.users.UpsertGuest(ctx, id, req.HomeNodeId, req.DisplayName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "upsert guest user: %v", err)
	}
	// Grant USER claim on this node if none already exists (won't downgrade MANAGER).
	if err := s.claims.GrantDefault(ctx, u.ID, s.nodeID); err != nil {
		return nil, status.Errorf(codes.Internal, "grant default claim: %v", err)
	}
	// Build local session JWT that carries the home_node so the BFF can acquire
	// cross-node tokens from the right home node for subsequent operations.
	userClaims, err := s.claims.ListForUser(ctx, u.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "load claims: %v", err)
	}
	var jwtClaims []models.JWTClaim
	for _, c := range userClaims {
		jwtClaims = append(jwtClaims, models.JWTClaim{Node: c.NodeID, Role: c.Role})
	}
	token, err := identity.IssueTokenWithHomeNode(u.ID.String(), s.nodeID, req.HomeNodeId, jwtClaims, s.privKey, tokenTTL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "issue token: %v", err)
	}
	return &pb.AuthResponse{Token: token, User: toPBUser(u, false)}, nil
}

// IssueCrossNodeToken issues a short-lived, audience-scoped JWT for a user on this node.
// Called by the network-manager when a CONNECTED peer requests auth for one of this
// node's users; the network-manager handles peer validation before forwarding here.
func (s *UsersService) IssueCrossNodeToken(ctx context.Context, req *pb.CrossNodeTokenRequest) (*pb.AuthResponse, error) {
	if req.UserId == "" || req.AudienceNode == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id and audience_node are required")
	}
	id, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id")
	}
	u, _, err := s.users.GetByIDWithHash(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}
	token, err := identity.IssueTokenForAudience(u.ID.String(), s.nodeID, req.AudienceNode, s.privKey, time.Hour)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "issue token: %v", err)
	}
	return &pb.AuthResponse{Token: token, User: toPBUser(u, false)}, nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

func (s *UsersService) authResponse(ctx context.Context, u *models.User, hasPassword bool) (*pb.AuthResponse, error) {
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
		return nil, status.Errorf(codes.Internal, "issue token: %v", err)
	}
	return &pb.AuthResponse{Token: token, User: toPBUser(u, hasPassword)}, nil
}

func toPBUser(u *models.User, hasPassword bool) *pb.User {
	return &pb.User{
		Id:          u.ID.String(),
		Email:       u.Email,
		DisplayName: u.DisplayName,
		SsoProvider: u.SSOProvider,
		CreatedAt:   u.CreatedAt.Unix(),
		HasPassword: hasPassword,
	}
}
