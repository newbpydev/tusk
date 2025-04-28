package main

import (
	"context"
	"fmt"
	"os"

	"github.com/newbpydev/tusk/internal/adapters/db"
	"github.com/newbpydev/tusk/internal/cli"
	"github.com/newbpydev/tusk/internal/config"
	"github.com/newbpydev/tusk/internal/util/logging"
	"go.uber.org/zap"
)

func main() {
	// Load the configuration from environment variables and .env file
	cfg := config.Load()

	// Initialize the logging system with console output disabled for better UX
	var err error
	err = logging.InitWithOptions(cfg, true)

	if err != nil {
		// If we can't initialize logging, fall back to basic console output
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Use defer to flush any buffered log entries when main returns
	defer logging.Sync()

	// Log startup information to log file only, not to console
	logging.Logger.Info("Tusk CLI starting up",
		zap.String("version", "0.1.0"),
		zap.String("environment", cfg.AppEnv))

	ctx := context.Background() // Create a context for the database connection

	// Initialize the database connection pool
	if err := db.Connect(ctx); err != nil {
		// Log to file and show minimal error on screen
		logging.Logger.Error("Failed to connect to database", zap.Error(err))
		fmt.Println("Error: Could not connect to database. Check logs for details.")
		os.Exit(1)
	}

	// Log silently to the log file
	logging.CLILogger.Info("Starting CLI execution")

	// Execute CLI commands - services will be initialized inside
	cli.Execute()
}
