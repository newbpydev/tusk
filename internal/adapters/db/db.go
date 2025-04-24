package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is the global database connection pool
// It is initialized in the Connect function and used throughout the application.
var DB *pgxpool.Pool

// Connect initializes the database connection pool using the DSN from the environment variable DB_URL.
// It returns an error if the connection fails or if the DSN is not set.
func Connect() error {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		return fmt.Errorf("DB_URL environment variable is not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = pool
	return nil
}

// Close closes the database connection pool.
// It should be called when the application is shutting down to release resources.
func Close() {
	if DB != nil {
		DB.Close()
	} else {
		fmt.Println("DB is nil, nothing to close")
	}
}
