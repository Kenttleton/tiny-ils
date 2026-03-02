package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"tiny-ils/shared/models"
)

type TransferStore struct {
	db *pgxpool.Pool
}

func NewTransferStore(db *pgxpool.Pool) *TransferStore {
	return &TransferStore{db: db}
}

func (s *TransferStore) Create(ctx context.Context, copyID, initiatedBy uuid.UUID, transferType models.TransferType, sourceNode, destNode, notes string) (*models.CopyTransfer, error) {
	t := &models.CopyTransfer{
		ID:           uuid.New(),
		CopyID:       copyID,
		TransferType: transferType,
		SourceNode:   sourceNode,
		DestNode:     destNode,
		InitiatedBy:  initiatedBy,
		Status:       models.TransferStatusPending,
		Notes:        notes,
		RequestedAt:  time.Now(),
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO copy_transfers
		  (id, copy_id, transfer_type, source_node, dest_node, initiated_by, status, notes, requested_at)
		  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		t.ID, t.CopyID, string(t.TransferType), t.SourceNode, t.DestNode,
		t.InitiatedBy, string(t.Status), t.Notes, t.RequestedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create transfer: %w", err)
	}

	// Mark copy as REQUESTED
	if _, err := tx.Exec(ctx,
		"UPDATE physical_copies SET status = 'REQUESTED' WHERE id = $1",
		copyID,
	); err != nil {
		return nil, fmt.Errorf("mark copy requested: %w", err)
	}

	return t, tx.Commit(ctx)
}

func (s *TransferStore) Get(ctx context.Context, id uuid.UUID) (*models.CopyTransfer, error) {
	t := &models.CopyTransfer{}
	var approvedBy *uuid.UUID
	var approvedAt, shippedAt, receivedAt *time.Time
	err := s.db.QueryRow(ctx,
		`SELECT id, copy_id, transfer_type, source_node, dest_node,
		        initiated_by, approved_by, status, notes,
		        requested_at, approved_at, shipped_at, received_at
		 FROM copy_transfers WHERE id = $1`,
		id,
	).Scan(
		&t.ID, &t.CopyID, &t.TransferType, &t.SourceNode, &t.DestNode,
		&t.InitiatedBy, &approvedBy, &t.Status, &t.Notes,
		&t.RequestedAt, &approvedAt, &shippedAt, &receivedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get transfer: %w", err)
	}
	t.ApprovedBy = approvedBy
	t.ApprovedAt = approvedAt
	t.ShippedAt = shippedAt
	t.ReceivedAt = receivedAt
	return t, nil
}

func (s *TransferStore) List(ctx context.Context, statusFilter, nodeID, transferType string) ([]*models.CopyTransfer, error) {
	query := `SELECT id, copy_id, transfer_type, source_node, dest_node,
	                 initiated_by, approved_by, status, notes,
	                 requested_at, approved_at, shipped_at, received_at
	          FROM copy_transfers WHERE 1=1`
	args := []any{}
	n := 1
	if statusFilter != "" {
		query += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, statusFilter)
		n++
	}
	if nodeID != "" {
		query += fmt.Sprintf(" AND (source_node = $%d OR dest_node = $%d)", n, n)
		args = append(args, nodeID)
		n++
	}
	if transferType != "" {
		query += fmt.Sprintf(" AND transfer_type = $%d", n)
		args = append(args, transferType)
		n++
	}
	query += " ORDER BY requested_at DESC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list transfers: %w", err)
	}
	defer rows.Close()

	var out []*models.CopyTransfer
	for rows.Next() {
		t := &models.CopyTransfer{}
		var approvedBy *uuid.UUID
		var approvedAt, shippedAt, receivedAt *time.Time
		if err := rows.Scan(
			&t.ID, &t.CopyID, &t.TransferType, &t.SourceNode, &t.DestNode,
			&t.InitiatedBy, &approvedBy, &t.Status, &t.Notes,
			&t.RequestedAt, &approvedAt, &shippedAt, &receivedAt,
		); err != nil {
			return nil, err
		}
		t.ApprovedBy = approvedBy
		t.ApprovedAt = approvedAt
		t.ShippedAt = shippedAt
		t.ReceivedAt = receivedAt
		out = append(out, t)
	}
	return out, rows.Err()
}

// Approve transitions PENDING → APPROVED (copy stays REQUESTED).
func (s *TransferStore) Approve(ctx context.Context, id, actorID uuid.UUID) (*models.CopyTransfer, error) {
	now := time.Now()
	t := &models.CopyTransfer{}
	var approvedBy *uuid.UUID
	var approvedAt, shippedAt, receivedAt *time.Time
	err := s.db.QueryRow(ctx,
		`UPDATE copy_transfers
		 SET status = 'APPROVED', approved_by = $1, approved_at = $2
		 WHERE id = $3 AND status = 'PENDING'
		 RETURNING id, copy_id, transfer_type, source_node, dest_node,
		           initiated_by, approved_by, status, notes,
		           requested_at, approved_at, shipped_at, received_at`,
		actorID, now, id,
	).Scan(
		&t.ID, &t.CopyID, &t.TransferType, &t.SourceNode, &t.DestNode,
		&t.InitiatedBy, &approvedBy, &t.Status, &t.Notes,
		&t.RequestedAt, &approvedAt, &shippedAt, &receivedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("approve transfer: %w", err)
	}
	t.ApprovedBy = approvedBy
	t.ApprovedAt = approvedAt
	t.ShippedAt = shippedAt
	t.ReceivedAt = receivedAt
	return t, nil
}

// Reject transitions PENDING → REJECTED and restores copy to AVAILABLE.
func (s *TransferStore) Reject(ctx context.Context, id, actorID uuid.UUID) (*models.CopyTransfer, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	t := &models.CopyTransfer{}
	var approvedBy *uuid.UUID
	var approvedAt, shippedAt, receivedAt *time.Time
	err = tx.QueryRow(ctx,
		`UPDATE copy_transfers SET status = 'REJECTED'
		 WHERE id = $1 AND status = 'PENDING'
		 RETURNING id, copy_id, transfer_type, source_node, dest_node,
		           initiated_by, approved_by, status, notes,
		           requested_at, approved_at, shipped_at, received_at`,
		id,
	).Scan(
		&t.ID, &t.CopyID, &t.TransferType, &t.SourceNode, &t.DestNode,
		&t.InitiatedBy, &approvedBy, &t.Status, &t.Notes,
		&t.RequestedAt, &approvedAt, &shippedAt, &receivedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("reject transfer: %w", err)
	}
	t.ApprovedBy = approvedBy
	t.ApprovedAt = approvedAt
	t.ShippedAt = shippedAt
	t.ReceivedAt = receivedAt

	if _, err := tx.Exec(ctx,
		"UPDATE physical_copies SET status = 'AVAILABLE' WHERE id = $1",
		t.CopyID,
	); err != nil {
		return nil, fmt.Errorf("restore copy status: %w", err)
	}
	return t, tx.Commit(ctx)
}

// MarkShipped transitions APPROVED → IN_TRANSIT and sets copy to IN_TRANSIT.
func (s *TransferStore) MarkShipped(ctx context.Context, id, actorID uuid.UUID) (*models.CopyTransfer, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	t := &models.CopyTransfer{}
	var approvedBy *uuid.UUID
	var approvedAt, shippedAt, receivedAt *time.Time
	err = tx.QueryRow(ctx,
		`UPDATE copy_transfers SET status = 'IN_TRANSIT', shipped_at = $1
		 WHERE id = $2 AND status = 'APPROVED'
		 RETURNING id, copy_id, transfer_type, source_node, dest_node,
		           initiated_by, approved_by, status, notes,
		           requested_at, approved_at, shipped_at, received_at`,
		now, id,
	).Scan(
		&t.ID, &t.CopyID, &t.TransferType, &t.SourceNode, &t.DestNode,
		&t.InitiatedBy, &approvedBy, &t.Status, &t.Notes,
		&t.RequestedAt, &approvedAt, &shippedAt, &receivedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("mark shipped: %w", err)
	}
	t.ApprovedBy = approvedBy
	t.ApprovedAt = approvedAt
	t.ShippedAt = shippedAt
	t.ReceivedAt = receivedAt

	if _, err := tx.Exec(ctx,
		"UPDATE physical_copies SET status = 'IN_TRANSIT' WHERE id = $1",
		t.CopyID,
	); err != nil {
		return nil, fmt.Errorf("mark copy in-transit: %w", err)
	}
	return t, tx.Commit(ctx)
}

// ConfirmReceived transitions IN_TRANSIT → RECEIVED, sets copy to AVAILABLE,
// and updates node_id for PERMANENT transfers.
func (s *TransferStore) ConfirmReceived(ctx context.Context, id, actorID uuid.UUID) (*models.CopyTransfer, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	t := &models.CopyTransfer{}
	var approvedBy *uuid.UUID
	var approvedAt, shippedAt, receivedAt *time.Time
	err = tx.QueryRow(ctx,
		`UPDATE copy_transfers SET status = 'RECEIVED', approved_by = $1, received_at = $2
		 WHERE id = $3 AND status = 'IN_TRANSIT'
		 RETURNING id, copy_id, transfer_type, source_node, dest_node,
		           initiated_by, approved_by, status, notes,
		           requested_at, approved_at, shipped_at, received_at`,
		actorID, now, id,
	).Scan(
		&t.ID, &t.CopyID, &t.TransferType, &t.SourceNode, &t.DestNode,
		&t.InitiatedBy, &approvedBy, &t.Status, &t.Notes,
		&t.RequestedAt, &approvedAt, &shippedAt, &receivedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("confirm received: %w", err)
	}
	t.ApprovedBy = approvedBy
	t.ApprovedAt = approvedAt
	t.ShippedAt = shippedAt
	t.ReceivedAt = receivedAt

	if t.TransferType == models.TransferTypePermanent {
		_, err = tx.Exec(ctx,
			"UPDATE physical_copies SET status = 'AVAILABLE', node_id = $1 WHERE id = $2",
			t.DestNode, t.CopyID,
		)
	} else {
		_, err = tx.Exec(ctx,
			"UPDATE physical_copies SET status = 'AVAILABLE' WHERE id = $1",
			t.CopyID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("finalize copy: %w", err)
	}
	return t, tx.Commit(ctx)
}

// Cancel transitions PENDING or APPROVED → CANCELLED and restores copy to AVAILABLE.
func (s *TransferStore) Cancel(ctx context.Context, id, actorID uuid.UUID) (*models.CopyTransfer, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	t := &models.CopyTransfer{}
	var approvedBy *uuid.UUID
	var approvedAt, shippedAt, receivedAt *time.Time
	err = tx.QueryRow(ctx,
		`UPDATE copy_transfers SET status = 'CANCELLED'
		 WHERE id = $1 AND status IN ('PENDING','APPROVED')
		 RETURNING id, copy_id, transfer_type, source_node, dest_node,
		           initiated_by, approved_by, status, notes,
		           requested_at, approved_at, shipped_at, received_at`,
		id,
	).Scan(
		&t.ID, &t.CopyID, &t.TransferType, &t.SourceNode, &t.DestNode,
		&t.InitiatedBy, &approvedBy, &t.Status, &t.Notes,
		&t.RequestedAt, &approvedAt, &shippedAt, &receivedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("cancel transfer: %w", err)
	}
	t.ApprovedBy = approvedBy
	t.ApprovedAt = approvedAt
	t.ShippedAt = shippedAt
	t.ReceivedAt = receivedAt

	if _, err := tx.Exec(ctx,
		"UPDATE physical_copies SET status = 'AVAILABLE' WHERE id = $1",
		t.CopyID,
	); err != nil {
		return nil, fmt.Errorf("restore copy status: %w", err)
	}
	return t, tx.Commit(ctx)
}
