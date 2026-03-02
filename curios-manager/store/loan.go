package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"tiny-ils/shared/models"
)

type LoanStore struct {
	db *pgxpool.Pool
}

func NewLoanStore(db *pgxpool.Pool) *LoanStore {
	return &LoanStore{db: db}
}

func (s *LoanStore) ListCopies(ctx context.Context, curioID uuid.UUID) ([]*models.PhysicalCopy, error) {
	rows, err := s.db.Query(ctx,
		"SELECT id, curio_id, condition, location, node_id, status, created_at FROM physical_copies WHERE curio_id = $1 ORDER BY created_at",
		curioID,
	)
	if err != nil {
		return nil, fmt.Errorf("list copies: %w", err)
	}
	defer rows.Close()

	var copies []*models.PhysicalCopy
	for rows.Next() {
		c := &models.PhysicalCopy{}
		if err := rows.Scan(&c.ID, &c.CurioID, &c.Condition, &c.Location, &c.NodeID, &c.Status, &c.CreatedAt); err != nil {
			return nil, err
		}
		copies = append(copies, c)
	}
	return copies, rows.Err()
}

func (s *LoanStore) Checkout(ctx context.Context, copyID, userID uuid.UUID, userNodeID string, dueDate time.Time, requestingNode string) (*models.PhysicalLoan, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var copyStatus string
	if err := tx.QueryRow(ctx, "SELECT status FROM physical_copies WHERE id = $1 FOR UPDATE", copyID).Scan(&copyStatus); err != nil {
		return nil, fmt.Errorf("get copy: %w", err)
	}
	if copyStatus != string(models.CopyStatusAvailable) {
		return nil, fmt.Errorf("copy %s is not available (status: %s)", copyID, copyStatus)
	}

	if _, err := tx.Exec(ctx, "UPDATE physical_copies SET status = 'ON_LOAN' WHERE id = $1", copyID); err != nil {
		return nil, fmt.Errorf("mark on-loan: %w", err)
	}

	loan := &models.PhysicalLoan{
		ID:             uuid.New(),
		CopyID:         copyID,
		UserID:         userID,
		UserNodeID:     userNodeID,
		CheckedOut:     time.Now(),
		DueDate:        dueDate,
		RequestingNode: requestingNode,
	}
	_, err = tx.Exec(ctx,
		"INSERT INTO physical_loans (id, copy_id, user_id, user_node_id, checked_out, due_date, requesting_node) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		loan.ID, loan.CopyID, loan.UserID, loan.UserNodeID, loan.CheckedOut, loan.DueDate, loan.RequestingNode,
	)
	if err != nil {
		return nil, fmt.Errorf("create loan: %w", err)
	}

	return loan, tx.Commit(ctx)
}

func (s *LoanStore) Return(ctx context.Context, copyID uuid.UUID) (*models.PhysicalLoan, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	loan := &models.PhysicalLoan{}
	err = tx.QueryRow(ctx,
		"UPDATE physical_loans SET returned_at = $1 WHERE copy_id = $2 AND returned_at IS NULL RETURNING id, copy_id, user_id, user_node_id, checked_out, due_date, returned_at, requesting_node",
		now, copyID,
	).Scan(&loan.ID, &loan.CopyID, &loan.UserID, &loan.UserNodeID, &loan.CheckedOut, &loan.DueDate, &loan.ReturnedAt, &loan.RequestingNode)
	if err != nil {
		return nil, fmt.Errorf("update loan: %w", err)
	}

	if _, err := tx.Exec(ctx, "UPDATE physical_copies SET status = 'AVAILABLE' WHERE id = $1", copyID); err != nil {
		return nil, fmt.Errorf("mark available: %w", err)
	}

	return loan, tx.Commit(ctx)
}

func (s *LoanStore) PlaceHold(ctx context.Context, curioID, userID uuid.UUID, userNodeID string) (*models.Hold, error) {
	hold := &models.Hold{
		ID:       uuid.New(),
		CurioID:  curioID,
		UserID:   userID,
		PlacedAt: time.Now(),
	}
	_, err := s.db.Exec(ctx,
		"INSERT INTO holds (id, curio_id, user_id, user_node_id, placed_at) VALUES ($1,$2,$3,$4,$5)",
		hold.ID, hold.CurioID, hold.UserID, userNodeID, hold.PlacedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("place hold: %w", err)
	}
	return hold, nil
}

func (s *LoanStore) CancelHold(ctx context.Context, holdID uuid.UUID) error {
	_, err := s.db.Exec(ctx, "DELETE FROM holds WHERE id = $1", holdID)
	return err
}
