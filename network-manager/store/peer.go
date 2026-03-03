package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Peer struct {
	NodeID       string
	PublicKey    string
	Address      string
	DisplayName  string
	Status       string // PENDING | CONNECTED
	Capabilities []string
	FirstSeen    time.Time
	LastSeen     time.Time
}

type PeerStore struct {
	db *pgxpool.Pool
}

func NewPeerStore(db *pgxpool.Pool) *PeerStore {
	return &PeerStore{db: db}
}

// Upsert stores a peer as CONNECTED (admin-initiated — explicit trust).
func (s *PeerStore) Upsert(ctx context.Context, p *Peer) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO peers (node_id, public_key, address, display_name, capabilities, status, first_seen, last_seen)
		 VALUES ($1, $2, $3, $4, $5, 'CONNECTED', now(), now())
		 ON CONFLICT (node_id) DO UPDATE
		   SET public_key    = EXCLUDED.public_key,
		       address       = EXCLUDED.address,
		       display_name  = EXCLUDED.display_name,
		       capabilities  = EXCLUDED.capabilities,
		       status        = 'CONNECTED',
		       last_seen     = now()`,
		p.NodeID, p.PublicKey, p.Address, p.DisplayName, p.Capabilities,
	)
	return err
}

// UpsertInbound stores an inbound peer-initiated connection.
// If the peer is unknown → PENDING. If already PENDING → CONNECTED (mutual
// registration happened — this node was pre-registered by admin). If already
// CONNECTED → unchanged. Returns the resulting status.
func (s *PeerStore) UpsertInbound(ctx context.Context, p *Peer) (string, error) {
	var status string
	err := s.db.QueryRow(ctx,
		`INSERT INTO peers (node_id, public_key, address, display_name, capabilities, status, first_seen, last_seen)
		 VALUES ($1, $2, $3, $4, $5, 'PENDING', now(), now())
		 ON CONFLICT (node_id) DO UPDATE
		   SET public_key    = EXCLUDED.public_key,
		       address       = EXCLUDED.address,
		       capabilities  = EXCLUDED.capabilities,
		       last_seen     = now(),
		       status        = CASE
		           WHEN peers.status = 'PENDING' THEN 'CONNECTED'
		           ELSE peers.status
		       END
		 RETURNING status`,
		p.NodeID, p.PublicKey, p.Address, p.DisplayName, p.Capabilities,
	).Scan(&status)
	return status, err
}

// Approve upgrades a PENDING peer to CONNECTED.
func (s *PeerStore) Approve(ctx context.Context, nodeID string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE peers SET status = 'CONNECTED', last_seen = now() WHERE node_id = $1`,
		nodeID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("peer %q not found", nodeID)
	}
	return nil
}

func (s *PeerStore) List(ctx context.Context) ([]*Peer, error) {
	rows, err := s.db.Query(ctx,
		`SELECT node_id, public_key, address, display_name, capabilities, status, first_seen, last_seen
		 FROM peers ORDER BY display_name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list peers: %w", err)
	}
	defer rows.Close()

	var peers []*Peer
	for rows.Next() {
		p := &Peer{}
		if err := rows.Scan(&p.NodeID, &p.PublicKey, &p.Address, &p.DisplayName, &p.Capabilities, &p.Status, &p.FirstSeen, &p.LastSeen); err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return peers, rows.Err()
}

// ListConnected returns only CONNECTED peers (used for fan-out operations).
func (s *PeerStore) ListConnected(ctx context.Context) ([]*Peer, error) {
	rows, err := s.db.Query(ctx,
		`SELECT node_id, public_key, address, display_name, capabilities, status, first_seen, last_seen
		 FROM peers WHERE status = 'CONNECTED' ORDER BY display_name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list connected peers: %w", err)
	}
	defer rows.Close()

	var peers []*Peer
	for rows.Next() {
		p := &Peer{}
		if err := rows.Scan(&p.NodeID, &p.PublicKey, &p.Address, &p.DisplayName, &p.Capabilities, &p.Status, &p.FirstSeen, &p.LastSeen); err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return peers, rows.Err()
}

// ListConnectedWithCapability returns CONNECTED peers that advertise the given capability.
// Peers with an empty capabilities array are included for backwards compatibility
// with nodes that predate the capabilities protocol.
func (s *PeerStore) ListConnectedWithCapability(ctx context.Context, capability string) ([]*Peer, error) {
	rows, err := s.db.Query(ctx,
		`SELECT node_id, public_key, address, display_name, capabilities, status, first_seen, last_seen
		 FROM peers
		 WHERE status = 'CONNECTED'
		   AND (capabilities = '{}' OR $1 = ANY(capabilities))
		 ORDER BY display_name`,
		capability,
	)
	if err != nil {
		return nil, fmt.Errorf("list peers with capability %q: %w", capability, err)
	}
	defer rows.Close()

	var peers []*Peer
	for rows.Next() {
		p := &Peer{}
		if err := rows.Scan(&p.NodeID, &p.PublicKey, &p.Address, &p.DisplayName, &p.Capabilities, &p.Status, &p.FirstSeen, &p.LastSeen); err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return peers, rows.Err()
}

func (s *PeerStore) GetPublicKey(ctx context.Context, nodeID string) (string, error) {
	var pubKey string
	err := s.db.QueryRow(ctx, "SELECT public_key FROM peers WHERE node_id = $1", nodeID).Scan(&pubKey)
	if err != nil {
		return "", fmt.Errorf("get peer public key: %w", err)
	}
	return pubKey, nil
}

// Get returns a CONNECTED peer by node ID, or nil if not found or not connected.
func (s *PeerStore) Get(ctx context.Context, nodeID string) (*Peer, error) {
	p := &Peer{}
	err := s.db.QueryRow(ctx,
		`SELECT node_id, public_key, address, display_name, capabilities, status, first_seen, last_seen
		 FROM peers WHERE node_id = $1 AND status = 'CONNECTED'`,
		nodeID,
	).Scan(&p.NodeID, &p.PublicKey, &p.Address, &p.DisplayName, &p.Capabilities, &p.Status, &p.FirstSeen, &p.LastSeen)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get peer: %w", err)
	}
	return p, nil
}

// GetByPublicKey looks up a peer by its base64-encoded public key.
// Returns nil, nil if not found.
func (s *PeerStore) GetByPublicKey(ctx context.Context, pubKeyB64 string) (*Peer, error) {
	p := &Peer{}
	err := s.db.QueryRow(ctx,
		`SELECT node_id, public_key, address, display_name, capabilities, status, first_seen, last_seen
		 FROM peers WHERE public_key = $1`,
		pubKeyB64,
	).Scan(&p.NodeID, &p.PublicKey, &p.Address, &p.DisplayName, &p.Capabilities, &p.Status, &p.FirstSeen, &p.LastSeen)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get peer by public key: %w", err)
	}
	return p, nil
}
