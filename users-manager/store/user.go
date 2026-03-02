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

func (s *UserStore) HasAnyManager(ctx context.Context, nodeID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM node_claims WHERE node_id = $1 AND role = 'MANAGER')",
		nodeID,
	).Scan(&exists)
	return exists, err
}
