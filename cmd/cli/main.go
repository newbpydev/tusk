package main

import (
	"context"

	"github.com/newbpydev/tusk/internal/adapters/db"
	"github.com/newbpydev/tusk/internal/cli"
	"github.com/newbpydev/tusk/internal/config"
)

func init() {
	config.Load() // Load the configuration from environment variables and .env file

	ctx := context.Background() // Create a context for the database connection

	// Initialize the database connection pool
	if err := db.Connect(ctx); err != nil {
		panic(err) // Handle the error appropriately in a real application
	}
	defer db.Close() // Ensure the database connection is closed when the application exits

}

func main() {
	cli.Execute()
}
