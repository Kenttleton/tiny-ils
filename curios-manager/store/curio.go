package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"tiny-ils/shared/models"
)

type CurioStore struct {
	db        *pgxpool.Pool
	serviceID uuid.UUID
}

func NewCurioStore(db *pgxpool.Pool, serviceID uuid.UUID) *CurioStore {
	return &CurioStore{db: db, serviceID: serviceID}
}

type ListFilter struct {
	Query      string
	MediaType  string
	FormatType string
	Tags       []string
	Limit      int32
	Offset     int32
}

func (s *CurioStore) List(ctx context.Context, f ListFilter) ([]*models.Curio, int32, error) {
	conditions := []string{"1=1"}
	args := []any{}
	i := 1

	if f.Query != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", i, i+1))
		q := "%" + f.Query + "%"
		args = append(args, q, q)
		i += 2
	}
	if f.MediaType != "" {
		conditions = append(conditions, fmt.Sprintf("media_type = $%d", i))
		args = append(args, f.MediaType)
		i++
	}
	if f.FormatType != "" {
		conditions = append(conditions, fmt.Sprintf("format_type = $%d", i))
		args = append(args, f.FormatType)
		i++
	}
	if len(f.Tags) > 0 {
		conditions = append(conditions, fmt.Sprintf("tags @> $%d", i))
		args = append(args, f.Tags)
		i++
	}

	where := strings.Join(conditions, " AND ")
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	var total int32
	if err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM curios WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count curios: %w", err)
	}

	args = append(args, limit, offset)
	rows, err := s.db.Query(ctx,
		fmt.Sprintf("SELECT id, title, description, media_type, format_type, tags, barcode, qr_code, created_at, updated_at FROM curios WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d", where, i, i+1),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list curios: %w", err)
	}
	defer rows.Close()

	var curios []*models.Curio
	for rows.Next() {
		c := &models.Curio{}
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.MediaType, &c.FormatType, &c.Tags, &c.Barcode, &c.QRCode, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		curios = append(curios, c)
	}
	return curios, total, rows.Err()
}

func (s *CurioStore) Get(ctx context.Context, id uuid.UUID) (*models.Curio, error) {
	c := &models.Curio{}
	err := s.db.QueryRow(ctx,
		"SELECT id, title, description, media_type, format_type, tags, barcode, qr_code, created_at, updated_at FROM curios WHERE id = $1",
		id,
	).Scan(&c.ID, &c.Title, &c.Description, &c.MediaType, &c.FormatType, &c.Tags, &c.Barcode, &c.QRCode, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get curio: %w", err)
	}
	return c, nil
}

func (s *CurioStore) Create(ctx context.Context, c *models.Curio) (*models.Curio, error) {
	c.ID = saltedID(s.serviceID)
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now

	_, err := s.db.Exec(ctx,
		"INSERT INTO curios (id, title, description, media_type, format_type, tags, barcode, qr_code, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)",
		c.ID, c.Title, c.Description, c.MediaType, c.FormatType, c.Tags, c.Barcode, c.QRCode, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create curio: %w", err)
	}
	return c, nil
}

func (s *CurioStore) Update(ctx context.Context, c *models.Curio) (*models.Curio, error) {
	c.UpdatedAt = time.Now()
	_, err := s.db.Exec(ctx,
		"UPDATE curios SET title=$2, description=$3, format_type=$4, tags=$5, barcode=$6, qr_code=$7, updated_at=$8 WHERE id=$1",
		c.ID, c.Title, c.Description, c.FormatType, c.Tags, c.Barcode, c.QRCode, c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update curio: %w", err)
	}
	return c, nil
}

func (s *CurioStore) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, "DELETE FROM curios WHERE id = $1", id)
	return err
}
