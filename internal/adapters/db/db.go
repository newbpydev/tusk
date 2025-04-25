package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newbpydev/tusk/internal/config"
)

// Pool is the global database connection pool
// It is initialized in the Connect function and used throughout the application.
var Pool *pgxpool.Pool

// Connect initializes the database connection pool using the DSN from the configuration.
// It returns an error if the connection fails.
func Connect(ctx context.Context) error {
	cfg := config.Load()
	dsn := cfg.DBURL

	if dsn == "" {
		return fmt.Errorf("DB_URL is not configured")
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	Pool = pool
	return nil
}

// Close closes the database connection pool.
// It should be called when the application is shutting down to release resources.
func Close() {
	if Pool != nil {
		Pool.Close()
	} else {
		fmt.Println("DB is nil, nothing to close")
	}
}
