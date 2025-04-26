// Package cli provides the CLI commands and functionality for the Tusk application.
package cli

import (
	"github.com/newbpydev/tusk/internal/adapters/db"
	tservice "github.com/newbpydev/tusk/internal/service/task"
	uservice "github.com/newbpydev/tusk/internal/service/user"
	"github.com/spf13/cobra"
)

var (
	// rootCmd is the main command for the Tusk CLI application.
	// It serves as the entry point for the command-line interface.
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
	// Initialize the services before running the command
	initServices()

	err := rootCmd.Execute()

	// Close the database connection after command execution
	db.Close()

	cobra.CheckErr(err)
}

// initServices initializes the database connection and repositories for the CLI application.
// It sets up the user and task services using the database repositories.
func initServices() {
	// The database connection should already be established in main.go
	uRepo := db.NewSQLUserRepo(db.Pool)
	tRepo := db.NewSQLTaskRepository(db.Pool)

	userSvc = uservice.NewUserService(uRepo)
	taskSvc = tservice.NewTaskService(tRepo)
}

// init is only used to set up commands, not for database connections
func init() {
	// rootCmd and all subcommand registrations happen in their respective files (task.go, user.go, tui.go)
	// The CLI commands are registered in their respective files via their own init() functions
	// No database operations occur during command registration
}
