// Package cli provides the CLI commands and functionality for the Tusk application.
package cli

import (
	"context"

	"github.com/newbpydev/tusk/internal/adapters/db"
	tservice "github.com/newbpydev/tusk/internal/service/task"
	uservice "github.com/newbpydev/tusk/internal/service/user"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "tusk",
		Short: "Tusk: Your Tasks, Tamed with Go",
		Long:  "A terminal-based task manager with nested subtasks, kanban support, and more.",
	}

	// Service instances
	userSvc uservice.Service
	taskSvc tservice.Service
)

// Execute runs the root command and handles errors.
// It initializes the database connection and repositories before executing the command.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

// init initializes the database connection and repositories for the CLI application.
// It sets up the user and task services using the database repositories.
func init() {
	// Initialize the database connection and repositories
	// This will use the DSN from the environment variable DB_URL.
	ctx := context.Background()

	// Connect to the database
	// This will load the configuration from the environment variables and the config file.
	cobra.CheckErr(db.Connect(ctx))
	defer db.Close()

	uRepo := db.NewSQLUserRepo(db.Pool)
	tRepo := db.NewSQLTaskRepository(db.Pool)

	userSvc = uservice.NewUserService(uRepo)
	taskSvc = tservice.NewTaskService(tRepo)
}
