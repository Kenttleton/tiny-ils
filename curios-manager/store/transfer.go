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

type TransferStore struct {
	db     *pgxpool.Pool
	nodeID string // this node's fingerprint; used to decide which copy updates to apply
}

func NewTransferStore(db *pgxpool.Pool, nodeID string) *TransferStore {
	return &TransferStore{db: db, nodeID: nodeID}
}

// homeNode returns the home-node fingerprint portion of a global_copy_id ("{home}/{copy_uuid}").
func homeNode(globalCopyID string) string {
	if i := strings.LastIndex(globalCopyID, "/"); i > 0 {
		return globalCopyID[:i]
	}
	return ""
}

// copyUUID returns the UUID portion of a global_copy_id ("{home}/{copy_uuid}").
func copyUUID(globalCopyID string) (uuid.UUID, bool) {
	if i := strings.LastIndex(globalCopyID, "/"); i > 0 {
		id, err := uuid.Parse(globalCopyID[i+1:])
		return id, err == nil
	}
	return uuid.Nil, false
}

// isLocalCopy reports whether the copy in this transfer record lives in this node's
// physical_copies table (i.e. this node is the home/source node).
func (s *TransferStore) isLocalCopy(globalCopyID string) bool {
	return homeNode(globalCopyID) == s.nodeID
}

// Create inserts a new transfer ledger entry.
// id is the pre-assigned structured transfer ID; generated via uuid.NewV7() if empty.
// globalCopyID is "{home_node}/{copy_uuid}"; if the home node matches this node, the
// physical copy's status is updated to REQUESTED within the same transaction.
func (s *TransferStore) Create(ctx context.Context, id, globalCopyID string, initiatedBy uuid.UUID, transferType models.TransferType, sourceNode, destNode, notes string) (*models.CopyTransfer, error) {
	if id == "" {
		v7, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("generate transfer id: %w", err)
		}
		id = v7.String()
	}

	t := &models.CopyTransfer{
		ID:           id,
		GlobalCopyID: globalCopyID,
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
		  (id, global_copy_id, transfer_type, source_node, dest_node, initiated_by, status, notes, requested_at)
		  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		t.ID, t.GlobalCopyID, string(t.TransferType), t.SourceNode, t.DestNode,
		t.InitiatedBy, string(t.Status), t.Notes, t.RequestedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create transfer: %w", err)
	}

	// Only mark the copy as REQUESTED when this node owns the physical copy.
	if s.isLocalCopy(globalCopyID) {
		copyID, ok := copyUUID(globalCopyID)
		if !ok {
			return nil, fmt.Errorf("invalid global_copy_id: %q", globalCopyID)
		}
		if _, err := tx.Exec(ctx,
			"UPDATE physical_copies SET status = 'REQUESTED' WHERE id = $1",
			copyID,
		); err != nil {
			return nil, fmt.Errorf("mark copy requested: %w", err)
		}
	}

	return t, tx.Commit(ctx)
}

func (s *TransferStore) Get(ctx context.Context, id string) (*models.CopyTransfer, error) {
	t := &models.CopyTransfer{}
	var approvedBy *uuid.UUID
	var approvedAt, shippedAt, receivedAt *time.Time
	err := s.db.QueryRow(ctx,
		`SELECT id, global_copy_id, transfer_type, source_node, dest_node,
		        initiated_by, approved_by, status, notes,
		        requested_at, approved_at, shipped_at, received_at
		 FROM copy_transfers WHERE id = $1`,
		id,
	).Scan(
		&t.ID, &t.GlobalCopyID, &t.TransferType, &t.SourceNode, &t.DestNode,
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
	query := `SELECT id, global_copy_id, transfer_type, source_node, dest_node,
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
			&t.ID, &t.GlobalCopyID, &t.TransferType, &t.SourceNode, &t.DestNode,
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

// scan helper shared by all UPDATE...RETURNING queries.
func scanTransfer(row interface {
	Scan(dest ...any) error
}) (*models.CopyTransfer, error) {
	t := &models.CopyTransfer{}
	var approvedBy *uuid.UUID
	var approvedAt, shippedAt, receivedAt *time.Time
	err := row.Scan(
		&t.ID, &t.GlobalCopyID, &t.TransferType, &t.SourceNode, &t.DestNode,
		&t.InitiatedBy, &approvedBy, &t.Status, &t.Notes,
		&t.RequestedAt, &approvedAt, &shippedAt, &receivedAt,
	)
	if err != nil {
		return nil, err
	}
	t.ApprovedBy = approvedBy
	t.ApprovedAt = approvedAt
	t.ShippedAt = shippedAt
	t.ReceivedAt = receivedAt
	return t, nil
}

const transferCols = `id, global_copy_id, transfer_type, source_node, dest_node,
		           initiated_by, approved_by, status, notes,
		           requested_at, approved_at, shipped_at, received_at`

// Approve transitions PENDING → APPROVED (copy stays REQUESTED on source node).
func (s *TransferStore) Approve(ctx context.Context, id string, actorID uuid.UUID) (*models.CopyTransfer, error) {
	now := time.Now()
	t, err := scanTransfer(s.db.QueryRow(ctx,
		`UPDATE copy_transfers
		 SET status = 'APPROVED', approved_by = $1, approved_at = $2
		 WHERE id = $3 AND status = 'PENDING'
		 RETURNING `+transferCols,
		actorID, now, id,
	))
	if err != nil {
		return nil, fmt.Errorf("approve transfer: %w", err)
	}
	return t, nil
}

// Reject transitions PENDING → REJECTED and restores the copy to AVAILABLE on the source node.
func (s *TransferStore) Reject(ctx context.Context, id string, actorID uuid.UUID) (*models.CopyTransfer, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	t, err := scanTransfer(tx.QueryRow(ctx,
		`UPDATE copy_transfers SET status = 'REJECTED'
		 WHERE id = $1 AND status = 'PENDING'
		 RETURNING `+transferCols,
		id,
	))
	if err != nil {
		return nil, fmt.Errorf("reject transfer: %w", err)
	}

	if s.isLocalCopy(t.GlobalCopyID) {
		copyID, _ := copyUUID(t.GlobalCopyID)
		if _, err := tx.Exec(ctx,
			"UPDATE physical_copies SET status = 'AVAILABLE' WHERE id = $1",
			copyID,
		); err != nil {
			return nil, fmt.Errorf("restore copy status: %w", err)
		}
	}
	return t, tx.Commit(ctx)
}

// MarkShipped transitions APPROVED → IN_TRANSIT and sets the copy to IN_TRANSIT on the source node.
func (s *TransferStore) MarkShipped(ctx context.Context, id string, actorID uuid.UUID) (*models.CopyTransfer, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	t, err := scanTransfer(tx.QueryRow(ctx,
		`UPDATE copy_transfers SET status = 'IN_TRANSIT', shipped_at = $1
		 WHERE id = $2 AND status = 'APPROVED'
		 RETURNING `+transferCols,
		now, id,
	))
	if err != nil {
		return nil, fmt.Errorf("mark shipped: %w", err)
	}

	if s.isLocalCopy(t.GlobalCopyID) {
		copyID, _ := copyUUID(t.GlobalCopyID)
		if _, err := tx.Exec(ctx,
			"UPDATE physical_copies SET status = 'IN_TRANSIT' WHERE id = $1",
			copyID,
		); err != nil {
			return nil, fmt.Errorf("mark copy in-transit: %w", err)
		}
	}
	return t, tx.Commit(ctx)
}

// ConfirmReceived transitions IN_TRANSIT → RECEIVED, sets copy to AVAILABLE on the source node.
// For PERMANENT transfers: also updates node_id and home_node_id to dest_node.
func (s *TransferStore) ConfirmReceived(ctx context.Context, id string, actorID uuid.UUID) (*models.CopyTransfer, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	t, err := scanTransfer(tx.QueryRow(ctx,
		`UPDATE copy_transfers SET status = 'RECEIVED', approved_by = $1, received_at = $2
		 WHERE id = $3 AND status = 'IN_TRANSIT'
		 RETURNING `+transferCols,
		actorID, now, id,
	))
	if err != nil {
		return nil, fmt.Errorf("confirm received: %w", err)
	}

	if s.isLocalCopy(t.GlobalCopyID) {
		copyID, _ := copyUUID(t.GlobalCopyID)
		var copyErr error
		if t.TransferType == models.TransferTypePermanent {
			_, copyErr = tx.Exec(ctx,
				"UPDATE physical_copies SET status = 'AVAILABLE', node_id = $1, home_node_id = $1 WHERE id = $2",
				t.DestNode, copyID,
			)
		} else {
			_, copyErr = tx.Exec(ctx,
				"UPDATE physical_copies SET status = 'AVAILABLE' WHERE id = $1",
				copyID,
			)
		}
		if copyErr != nil {
			return nil, fmt.Errorf("finalize copy: %w", copyErr)
		}
	}
	return t, tx.Commit(ctx)
}

// Cancel transitions PENDING or APPROVED → CANCELLED and restores the copy to AVAILABLE on the source node.
func (s *TransferStore) Cancel(ctx context.Context, id string, actorID uuid.UUID) (*models.CopyTransfer, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	t, err := scanTransfer(tx.QueryRow(ctx,
		`UPDATE copy_transfers SET status = 'CANCELLED'
		 WHERE id = $1 AND status IN ('PENDING','APPROVED')
		 RETURNING `+transferCols,
		id,
	))
	if err != nil {
		return nil, fmt.Errorf("cancel transfer: %w", err)
	}

	if s.isLocalCopy(t.GlobalCopyID) {
		copyID, _ := copyUUID(t.GlobalCopyID)
		if _, err := tx.Exec(ctx,
			"UPDATE physical_copies SET status = 'AVAILABLE' WHERE id = $1",
			copyID,
		); err != nil {
			return nil, fmt.Errorf("restore copy status: %w", err)
		}
	}
	return t, tx.Commit(ctx)
}
