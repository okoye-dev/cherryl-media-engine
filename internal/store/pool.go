package store

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool initializes a pgxpool optimized for direct Postgres and Supabase/PgBouncer poolers.
func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	// If using Supabase / PgBouncer pooler (port 6543 or pooler hostname), disable statement cache
	// by switching to Simple Protocol so prepared statements don't conflict across transactions.
	if strings.Contains(connString, "pooler.supabase.com") || strings.Contains(connString, ":6543") {
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	}

	return pgxpool.NewWithConfig(ctx, config)
}
