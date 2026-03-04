package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingsStore struct {
	db *pgxpool.Pool
}

func NewSettingsStore(db *pgxpool.Pool) *SettingsStore {
	return &SettingsStore{db: db}
}

// Get returns the value for key, or "" when absent.
func (s *SettingsStore) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRow(ctx, "SELECT value FROM app_settings WHERE key = $1", key).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return value, err
}

// Set upserts a key-value pair in app_settings.
func (s *SettingsStore) Set(ctx context.Context, key, value string) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO app_settings (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
		key, value,
	)
	return err
}
