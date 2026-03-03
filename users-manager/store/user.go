package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"tiny-ils/shared/models"
)

type UserStore struct {
	db *pgxpool.Pool
}

func NewUserStore(db *pgxpool.Pool) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, u *models.User, passwordHash string) (*models.User, error) {
	u.ID = uuid.New()
	u.CreatedAt = time.Now()
	_, err := s.db.Exec(ctx,
		"INSERT INTO users (id, email, display_name, password_hash, sso_provider, sso_subject, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		u.ID, u.Email, u.DisplayName, passwordHash, u.SSOProvider, u.SSOSubject, u.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*models.User, string, error) {
	u := &models.User{}
	var hash string
	err := s.db.QueryRow(ctx,
		"SELECT id, email, display_name, password_hash, sso_provider, sso_subject, created_at FROM users WHERE email = $1",
		email,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &hash, &u.SSOProvider, &u.SSOSubject, &u.CreatedAt)
	if err != nil {
		return nil, "", fmt.Errorf("get user by email: %w", err)
	}
	return u, hash, nil
}

func (s *UserStore) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u := &models.User{}
	err := s.db.QueryRow(ctx,
		"SELECT id, email, display_name, sso_provider, sso_subject, created_at FROM users WHERE id = $1",
		id,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.SSOProvider, &u.SSOSubject, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

// GetByIDWithHash fetches a user and their password hash.
// hash is empty string when the account has no local password (SSO-only).
func (s *UserStore) GetByIDWithHash(ctx context.Context, id uuid.UUID) (*models.User, string, error) {
	u := &models.User{}
	var hash string
	err := s.db.QueryRow(ctx,
		"SELECT id, email, display_name, password_hash, sso_provider, sso_subject, created_at FROM users WHERE id = $1",
		id,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &hash, &u.SSOProvider, &u.SSOSubject, &u.CreatedAt)
	if err != nil {
		return nil, "", fmt.Errorf("get user: %w", err)
	}
	return u, hash, nil
}

// Update partially updates a user's profile. Only non-empty string fields are written.
// If unlinkSSO is true, sso_provider and sso_subject are cleared.
func (s *UserStore) Update(ctx context.Context, id uuid.UUID, displayName, email, passwordHash string, unlinkSSO bool) (*models.User, error) {
	u := &models.User{}
	err := s.db.QueryRow(ctx, `
		UPDATE users SET
			display_name  = CASE WHEN $2 != '' THEN $2 ELSE display_name END,
			email         = CASE WHEN $3 != '' THEN $3 ELSE email END,
			password_hash = CASE WHEN $4 != '' THEN $4 ELSE password_hash END,
			sso_provider  = CASE WHEN $5 THEN '' ELSE sso_provider END,
			sso_subject   = CASE WHEN $5 THEN '' ELSE sso_subject END
		WHERE id = $1
		RETURNING id, email, display_name, sso_provider, sso_subject, created_at`,
		id, displayName, email, passwordHash, unlinkSSO,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.SSOProvider, &u.SSOSubject, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return u, nil
}

func (s *UserStore) UpsertSSO(ctx context.Context, provider, subject, email, displayName string) (*models.User, error) {
	u := &models.User{}
	err := s.db.QueryRow(ctx,
		`INSERT INTO users (id, email, display_name, sso_provider, sso_subject, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, now())
		 ON CONFLICT (sso_provider, sso_subject) DO UPDATE
		   SET email = EXCLUDED.email, display_name = EXCLUDED.display_name
		 RETURNING id, email, display_name, sso_provider, sso_subject, created_at`,
		email, displayName, provider, subject,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.SSOProvider, &u.SSOSubject, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert sso user: %w", err)
	}
	return u, nil
}

// LinkSSO associates an SSO identity with an existing user.
// Returns an error if that SSO identity is already linked to a different account.
func (s *UserStore) LinkSSO(ctx context.Context, id uuid.UUID, provider, subject string) (*models.User, string, error) {
	u := &models.User{}
	var hash string
	err := s.db.QueryRow(ctx, `
		UPDATE users SET sso_provider = $2, sso_subject = $3
		WHERE id = $1
		RETURNING id, email, display_name, password_hash, sso_provider, sso_subject, created_at`,
		id, provider, subject,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &hash, &u.SSOProvider, &u.SSOSubject, &u.CreatedAt)
	if err != nil {
		return nil, "", fmt.Errorf("link sso: %w", err)
	}
	return u, hash, nil
}

// UpsertGuest creates or updates a thin local record for a cross-node user.
// The UUID is preserved from the home node. A synthetic email is used since
// the user has no account on this node.
func (s *UserStore) UpsertGuest(ctx context.Context, id uuid.UUID, homeNodeID, displayName string) (*models.User, error) {
	// Store home-node identity via sso fields so the record is identifiable later.
	provider := "node:" + homeNodeID
	subject := id.String()
	email := id.String() + "@" + homeNodeID + ".cross-node"

	u := &models.User{}
	err := s.db.QueryRow(ctx,
		`INSERT INTO users (id, email, display_name, sso_provider, sso_subject, created_at)
		 VALUES ($1, $2, $3, $4, $5, now())
		 ON CONFLICT (id) DO UPDATE
		   SET display_name = EXCLUDED.display_name,
		       sso_provider  = EXCLUDED.sso_provider,
		       sso_subject   = EXCLUDED.sso_subject
		 RETURNING id, email, display_name, sso_provider, sso_subject, created_at`,
		id, email, displayName, provider, subject,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.SSOProvider, &u.SSOSubject, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert guest user: %w", err)
	}
	return u, nil
}

// ListUsers returns all users with their role on the given node, paginated.
func (s *UserStore) ListUsers(ctx context.Context, nodeID string, limit, offset int) ([]*models.UserWithRole, int, error) {
	var total int
	if err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT u.id, u.email, u.display_name, u.password_hash, u.sso_provider, u.sso_subject, u.created_at,
		       COALESCE(nc.role, '') AS role
		FROM users u
		LEFT JOIN node_claims nc ON nc.user_id = u.id AND nc.node_id = $1
		ORDER BY u.created_at ASC
		LIMIT $2 OFFSET $3`,
		nodeID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var result []*models.UserWithRole
	for rows.Next() {
		var u models.User
		var role string
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.SSOProvider, &u.SSOSubject, &u.CreatedAt, &role); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		result = append(result, &models.UserWithRole{User: u, Role: role})
	}
	return result, total, nil
}

// DeleteUser removes a user and all their node claims.
func (s *UserStore) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, "DELETE FROM node_claims WHERE user_id = $1", id)
	if err != nil {
		return fmt.Errorf("delete claims: %w", err)
	}
	_, err = s.db.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (s *UserStore) HasAnyManager(ctx context.Context, nodeID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM node_claims WHERE node_id = $1 AND role = 'MANAGER')",
		nodeID,
	).Scan(&exists)
	return exists, err
}
