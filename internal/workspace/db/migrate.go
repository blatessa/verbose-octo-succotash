package workspacedb

import (
	"context"
	"embed"

	pkgdb "github.com/blatessa/verbose-octo-succotash/pkg/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate runs the workspace schema migrations.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	return pkgdb.Migrate(ctx, pool, migrations, "migrations")
}
