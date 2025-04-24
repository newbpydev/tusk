package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the global database connection pool
// It is initialized in the Connect function and used throughout the application.
var Pool *pgxpool.Pool

// Connect initializes the database connection pool using the DSN from the environment variable DB_URL.
// It returns an error if the connection fails or if the DSN is not set.
func Connect(ctx context.Context) error {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		return fmt.Errorf("DB_URL environment variable is not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
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
