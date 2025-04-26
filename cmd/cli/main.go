package main

import (
	"context"
	"fmt"
	"os"

	"github.com/newbpydev/tusk/internal/adapters/db"
	"github.com/newbpydev/tusk/internal/cli"
	"github.com/newbpydev/tusk/internal/config"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	// Initialize logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync() // Flush any buffered log entries

	// Load the configuration from environment variables and .env file
	config.Load()
}

func main() {
	ctx := context.Background() // Create a context for the database connection

	// Initialize the database connection pool
	if err := db.Connect(ctx); err != nil {
		logger.Error("Failed to connect to database",
			zap.Error(err)) // Using zap for structured logging with error wrapping
		os.Exit(1)
	}
	defer db.Close() // Now properly deferred until the end of main()

	cli.Execute()
}
