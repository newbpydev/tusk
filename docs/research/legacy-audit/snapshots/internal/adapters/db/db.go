package db

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newbpydev/tusk/internal/config"
	"github.com/newbpydev/tusk/internal/util/logging"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Pool is the global database connection pool
// It is initialized in the Connect function and used throughout the application.
var Pool *pgxpool.Pool

// Logger is a specialized file-only logger for database operations
var Logger *zap.Logger

// Connect initializes the database connection pool using the DSN from the configuration.
// It returns an error if the connection fails.
func Connect(ctx context.Context) error {
	cfg := config.Load()
	dsn := cfg.DBURL

	// Initialize the file-only logger for database operations
	Logger = logging.GetFileOnlyLogger("db")

	Logger.Info("Connecting to database...",
		zap.String("dsn", strings.Replace(dsn, ":", ":*****@", 1))) // Mask password in logs

	if dsn == "" {
		Logger.Error("DB_URL is not configured")
		return errors.New("DB_URL is not configured")
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		Logger.Error("Failed to parse database URL", zap.Error(err))
		return errors.Wrap(err, "failed to parse database URL")
	}

	// Configure the pool with reasonable defaults
	poolCfg.MaxConns = 10
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = 1 * time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		Logger.Error("Failed to connect to database", zap.Error(err))
		return errors.Wrap(err, "failed to connect to database")
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		Logger.Error("Failed to ping database", zap.Error(err))
		return errors.Wrap(err, "failed to ping database")
	}

	Pool = pool

	Logger.Info("Database connection established successfully",
		zap.Int("max_connections", 10))

	return nil
}

// Close closes the database connection pool.
// It should be called when the application is shutting down to release resources.
func Close() {
	if Pool != nil {
		Logger.Info("Closing database connection")
		Pool.Close()
	} else {
		Logger.Warn("DB is nil, nothing to close")
	}
}
