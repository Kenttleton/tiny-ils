// migrate applies all SQL schema files in order against the DATABASE_URL.
// It tracks applied files in a schema_migrations table.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx)

	if err := ensureMigrationsTable(ctx, conn); err != nil {
		log.Fatalf("ensure migrations table: %v", err)
	}

	schemaDir := "schema"
	if len(os.Args) > 1 {
		schemaDir = os.Args[1]
	}

	files, err := filepath.Glob(filepath.Join(schemaDir, "*.sql"))
	if err != nil {
		log.Fatalf("glob schema: %v", err)
	}
	sort.Strings(files)

	applied := 0
	for _, f := range files {
		name := filepath.Base(f)
		ran, err := hasRun(ctx, conn, name)
		if err != nil {
			log.Fatalf("check migration %s: %v", name, err)
		}
		if ran {
			fmt.Printf("  skip  %s\n", name)
			continue
		}

		sql, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("read %s: %v", f, err)
		}

		if _, err := conn.Exec(ctx, string(sql)); err != nil {
			log.Fatalf("run %s: %v", name, err)
		}

		if err := markRun(ctx, conn, name); err != nil {
			log.Fatalf("mark %s: %v", name, err)
		}

		fmt.Printf("  apply %s\n", name)
		applied++
	}

	fmt.Printf("schema applied: %d files\n", applied)
}

func ensureMigrationsTable(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name       TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	return err
}

func hasRun(ctx context.Context, conn *pgx.Conn, name string) (bool, error) {
	var exists bool
	err := conn.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)",
		strings.TrimSuffix(name, ".sql"),
	).Scan(&exists)
	return exists, err
}

func markRun(ctx context.Context, conn *pgx.Conn, name string) error {
	_, err := conn.Exec(ctx,
		"INSERT INTO schema_migrations (name) VALUES ($1)",
		strings.TrimSuffix(name, ".sql"),
	)
	return err
}
