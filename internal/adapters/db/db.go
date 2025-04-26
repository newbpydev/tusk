package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newbpydev/tusk/internal/config"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Pool is the global database connection pool
// It is initialized in the Connect function and used throughout the application.
var Pool *pgxpool.Pool
var logger *zap.Logger

func init() {
	// Initialize logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		// Fallback to basic logging if zap initialization fails
		panic("failed to initialize logger: " + err.Error())
	}
}

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
	logger.Info("Database connection established")
	return nil
}

// Close closes the database connection pool.
// It should be called when the application is shutting down to release resources.
func Close() {
	if Pool != nil {
		logger.Info("Closing database connection")
		Pool.Close()
	} else {
		logger.Warn("DB is nil, nothing to close")
	}
}
