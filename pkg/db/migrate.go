package db

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrate runs pending SQL files from migrationsFS in lexicographic order.
// Each migration is recorded in schema_migrations using "<dir>/<filename>" as
// the version key, so multiple packages can share the same table safely.
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrationsFS fs.ReadDirFS, dir string) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT        PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
	); err != nil {
		return fmt.Errorf("migrate: bootstrap: %w", err)
	}

	entries, err := migrationsFS.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("migrate: read dir %q: %w", dir, err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version := dir + "/" + entry.Name()

		var exists bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`,
			version,
		).Scan(&exists); err != nil {
			return fmt.Errorf("migrate: check %s: %w", version, err)
		}
		if exists {
			continue
		}

		sql, err := fs.ReadFile(migrationsFS, dir+"/"+entry.Name())
		if err != nil {
			return fmt.Errorf("migrate: read %s: %w", version, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("migrate: exec %s: %w", version, err)
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, version,
		); err != nil {
			return fmt.Errorf("migrate: record %s: %w", version, err)
		}
	}
	return nil
}
