package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newbpydev/tusk/internal/config"
	"github.com/newbpydev/tusk/internal/util/logging"
	"github.com/pkg/errors"
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
		return errors.New("DB_URL is not configured")
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return errors.Wrap(err, "failed to parse database URL")
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	Pool = pool

	// Use safe logging approach to avoid nil pointer dereference
	if logging.DBLogger != nil {
		logging.DBLogger.Info("Database connection established")
	} else {
		// Fallback if logger is not yet initialized
		fmt.Println("Database connection established")
	}
	return nil
}

// Close closes the database connection pool.
// It should be called when the application is shutting down to release resources.
func Close() {
	if Pool != nil {
		// Use safe logging approach
		if logging.DBLogger != nil {
			logging.DBLogger.Info("Closing database connection")
		} else {
			fmt.Println("Closing database connection")
		}
		Pool.Close()
	} else {
		if logging.DBLogger != nil {
			logging.DBLogger.Warn("DB is nil, nothing to close")
		} else {
			fmt.Println("DB is nil, nothing to close")
		}
	}
}
