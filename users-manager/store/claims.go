package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"tiny-ils/shared/models"
)

type ClaimsStore struct {
	db *pgxpool.Pool
}

func NewClaimsStore(db *pgxpool.Pool) *ClaimsStore {
	return &ClaimsStore{db: db}
}

func (s *ClaimsStore) Grant(ctx context.Context, userID uuid.UUID, nodeID string, role models.Role, grantedBy uuid.UUID) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO node_claims (user_id, node_id, role, granted_by, granted_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id, node_id) DO UPDATE SET role = EXCLUDED.role, granted_by = EXCLUDED.granted_by, granted_at = EXCLUDED.granted_at`,
		userID, nodeID, string(role), grantedBy, time.Now(),
	)
	return err
}

func (s *ClaimsStore) Revoke(ctx context.Context, userID uuid.UUID, nodeID string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM node_claims WHERE user_id = $1 AND node_id = $2", userID, nodeID)
	return err
}

func (s *ClaimsStore) ListForNode(ctx context.Context, nodeID string) ([]*models.NodeClaim, error) {
	rows, err := s.db.Query(ctx,
		"SELECT user_id, node_id, role, granted_by, granted_at FROM node_claims WHERE node_id = $1",
		nodeID,
	)
	if err != nil {
		return nil, fmt.Errorf("list claims: %w", err)
	}
	defer rows.Close()

	var claims []*models.NodeClaim
	for rows.Next() {
		c := &models.NodeClaim{}
		if err := rows.Scan(&c.UserID, &c.NodeID, &c.Role, &c.GrantedBy, &c.GrantedAt); err != nil {
			return nil, err
		}
		claims = append(claims, c)
	}
	return claims, rows.Err()
}

func (s *ClaimsStore) ListForUser(ctx context.Context, userID uuid.UUID) ([]*models.NodeClaim, error) {
	rows, err := s.db.Query(ctx,
		"SELECT user_id, node_id, role, granted_by, granted_at FROM node_claims WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list user claims: %w", err)
	}
	defer rows.Close()

	var claims []*models.NodeClaim
	for rows.Next() {
		c := &models.NodeClaim{}
		if err := rows.Scan(&c.UserID, &c.NodeID, &c.Role, &c.GrantedBy, &c.GrantedAt); err != nil {
			return nil, err
		}
		claims = append(claims, c)
	}
	return claims, rows.Err()
}
