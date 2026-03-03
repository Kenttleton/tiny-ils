package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"tiny-ils/shared/models"
)

type LeaseStore struct {
	db *pgxpool.Pool
}

func NewLeaseStore(db *pgxpool.Pool) *LeaseStore {
	return &LeaseStore{db: db}
}

// ─── Digital assets ───────────────────────────────────────────────────────────

func (s *LeaseStore) GetAsset(ctx context.Context, curioID uuid.UUID) (*models.DigitalAsset, error) {
	a := &models.DigitalAsset{}
	err := s.db.QueryRow(ctx,
		`SELECT id, curio_id, format, file_ref, checksum, max_concurrent,
		        COALESCE(lcp_content_id, ''), COALESCE(storage_backend, 'local'), COALESCE(encrypted, false)
		 FROM digital_assets WHERE curio_id = $1`,
		curioID,
	).Scan(&a.ID, &a.CurioID, &a.Format, &a.FileRef, &a.Checksum, &a.MaxConcurrent,
		&a.LCPContentID, &a.StorageBackend, &a.Encrypted)
	if err != nil {
		return nil, fmt.Errorf("get asset: %w", err)
	}
	return a, nil
}

func (s *LeaseStore) CreateAsset(ctx context.Context, a *models.DigitalAsset) (*models.DigitalAsset, error) {
	a.ID = uuid.New()
	_, err := s.db.Exec(ctx,
		`INSERT INTO digital_assets
		   (id, curio_id, format, file_ref, checksum, max_concurrent, lcp_content_id, storage_backend, encrypted)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 ON CONFLICT (curio_id) DO UPDATE SET
		   format = EXCLUDED.format,
		   file_ref = EXCLUDED.file_ref,
		   checksum = EXCLUDED.checksum,
		   max_concurrent = EXCLUDED.max_concurrent,
		   lcp_content_id = EXCLUDED.lcp_content_id,
		   storage_backend = EXCLUDED.storage_backend,
		   encrypted = EXCLUDED.encrypted`,
		a.ID, a.CurioID, a.Format, a.FileRef, a.Checksum, a.MaxConcurrent,
		nilIfEmpty(a.LCPContentID), a.StorageBackend, a.Encrypted,
	)
	if err != nil {
		return nil, fmt.Errorf("create asset: %w", err)
	}
	return a, nil
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// ─── Digital leases ───────────────────────────────────────────────────────────

// CountActiveLeases returns the number of non-revoked, non-expired leases for assetID.
func (s *LeaseStore) CountActiveLeases(ctx context.Context, assetID uuid.UUID) (int, error) {
	var count int
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM digital_leases
		 WHERE asset_id = $1 AND revoked = false AND expires_at > NOW()`,
		assetID,
	).Scan(&count)
	return count, err
}

func (s *LeaseStore) IssueLease(ctx context.Context, assetID, userID uuid.UUID, userNodeID string, accessToken string, expiresAt time.Time) (*models.DigitalLease, error) {
	lease := &models.DigitalLease{
		ID:          uuid.New(),
		AssetID:     assetID,
		UserID:      userID,
		UserNodeID:  userNodeID,
		AccessToken: accessToken,
		IssuedAt:    time.Now(),
		ExpiresAt:   expiresAt,
	}
	_, err := s.db.Exec(ctx,
		`INSERT INTO digital_leases
		   (id, asset_id, user_id, user_node_id, access_token, issued_at, expires_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		lease.ID, lease.AssetID, lease.UserID, lease.UserNodeID, lease.AccessToken, lease.IssuedAt, lease.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("issue lease: %w", err)
	}
	return lease, nil
}

func (s *LeaseStore) RevokeLease(ctx context.Context, leaseID uuid.UUID) error {
	_, err := s.db.Exec(ctx, "UPDATE digital_leases SET revoked = true WHERE id = $1", leaseID)
	return err
}

// ListLeases returns digital leases for a specific user (identified by home node + user ID).
func (s *LeaseStore) ListLeases(ctx context.Context, userID uuid.UUID, userNodeID string, activeOnly bool) ([]*models.DigitalLease, error) {
	q := `SELECT id, asset_id, user_id, user_node_id, access_token, issued_at, expires_at, revoked
	      FROM digital_leases
	      WHERE user_id = $1 AND user_node_id = $2`
	if activeOnly {
		q += " AND revoked = false AND expires_at > NOW()"
	}
	q += " ORDER BY issued_at DESC"

	rows, err := s.db.Query(ctx, q, userID, userNodeID)
	if err != nil {
		return nil, fmt.Errorf("list leases: %w", err)
	}
	defer rows.Close()

	var leases []*models.DigitalLease
	for rows.Next() {
		l := &models.DigitalLease{}
		if err := rows.Scan(&l.ID, &l.AssetID, &l.UserID, &l.UserNodeID, &l.AccessToken, &l.IssuedAt, &l.ExpiresAt, &l.Revoked); err != nil {
			return nil, err
		}
		leases = append(leases, l)
	}
	return leases, rows.Err()
}

// GetLease returns a single lease by ID.
func (s *LeaseStore) GetLease(ctx context.Context, leaseID uuid.UUID) (*models.DigitalLease, error) {
	l := &models.DigitalLease{}
	err := s.db.QueryRow(ctx,
		`SELECT id, asset_id, user_id, user_node_id, access_token, issued_at, expires_at, revoked
		 FROM digital_leases WHERE id = $1`,
		leaseID,
	).Scan(&l.ID, &l.AssetID, &l.UserID, &l.UserNodeID, &l.AccessToken, &l.IssuedAt, &l.ExpiresAt, &l.Revoked)
	if err != nil {
		return nil, fmt.Errorf("get lease: %w", err)
	}
	return l, nil
}
