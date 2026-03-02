package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Peer struct {
	NodeID      string
	PublicKey   string
	Address     string
	DisplayName string
	FirstSeen   time.Time
	LastSeen    time.Time
}

type PeerStore struct {
	db *pgxpool.Pool
}

func NewPeerStore(db *pgxpool.Pool) *PeerStore {
	return &PeerStore{db: db}
}

func (s *PeerStore) Upsert(ctx context.Context, p *Peer) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO peers (node_id, public_key, address, display_name, first_seen, last_seen)
		 VALUES ($1, $2, $3, $4, now(), now())
		 ON CONFLICT (node_id) DO UPDATE
		   SET public_key = EXCLUDED.public_key,
		       address = EXCLUDED.address,
		       display_name = EXCLUDED.display_name,
		       last_seen = now()`,
		p.NodeID, p.PublicKey, p.Address, p.DisplayName,
	)
	return err
}

func (s *PeerStore) List(ctx context.Context) ([]*Peer, error) {
	rows, err := s.db.Query(ctx,
		"SELECT node_id, public_key, address, display_name, first_seen, last_seen FROM peers ORDER BY display_name",
	)
	if err != nil {
		return nil, fmt.Errorf("list peers: %w", err)
	}
	defer rows.Close()

	var peers []*Peer
	for rows.Next() {
		p := &Peer{}
		if err := rows.Scan(&p.NodeID, &p.PublicKey, &p.Address, &p.DisplayName, &p.FirstSeen, &p.LastSeen); err != nil {
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
