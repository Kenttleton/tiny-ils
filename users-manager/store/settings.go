package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingsStore struct {
	db *pgxpool.Pool
}

func NewSettingsStore(db *pgxpool.Pool) *SettingsStore {
	return &SettingsStore{db: db}
}

func (s *SettingsStore) Set(ctx context.Context, key, value string) error {
	_, err := s.db.Exec(ctx,
		"INSERT INTO app_settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value",
		key, value,
	)
	if err != nil {
		return fmt.Errorf("set setting %q: %w", key, err)
	}
	return nil
}

// Get returns the value for key, or "" if not found.
func (s *SettingsStore) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRow(ctx, "SELECT value FROM app_settings WHERE key = $1", key).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get setting %q: %w", key, err)
	}
	return value, nil
}
