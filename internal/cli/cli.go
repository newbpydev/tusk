// Package cli implements the command-line interface for the Tusk application
package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/newbpydev/tusk/internal/adapters/db"
	"github.com/newbpydev/tusk/internal/service/task"
	"github.com/newbpydev/tusk/internal/service/user"
	"github.com/newbpydev/tusk/internal/util/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	taskSvc      task.Service
	asyncTaskSvc *task.AsyncTaskService
	userSvc      user.Service
	rootCmd      = &cobra.Command{
		Use:   "tusk",
		Short: "Tusk - Task Management System",
		Long: `Tusk is a task management system for organizing your work and personal projects.
It supports hierarchical tasks, priorities, due dates, and more.`,
		// Run: func(cmd *cobra.Command, args []string) {
		// 	// If no commands or flags, show help by default
		// 	cmd.Help()
		// },
	}
)

// initServices initializes all application services
func initServices() {
	// Connect to database
	if err := db.Connect(context.Background()); err != nil {
		logging.Logger.Error("Failed to connect to database", zap.Error(err))
		fmt.Println("Error: Could not connect to database. Check logs for details.")
		os.Exit(1)
	}
	logger := logging.Logger

	// Initialize the task repository using the global DB pool
	taskRepo := db.NewSQLTaskRepository(db.Pool)

	// Initialize the regular task service
	regularTaskSvc := task.NewTaskService(taskRepo)

	// Wrap in async service for non-blocking operations
	asyncTaskSvc = task.NewAsyncTaskService(regularTaskSvc, logger)

	// Expose as the global task service
	taskSvc = asyncTaskSvc

	// Initialize the user repository using the global DB pool
	userRepo := db.NewSQLUserRepo(db.Pool)

	// Initialize the user service
	userSvc = user.NewUserService(userRepo)
}

// Execute runs the root command
func Execute() {
	// Initialize services
	initServices()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Clean shutdown of async services
	if asyncTaskSvc != nil {
		asyncTaskSvc.Close()
	}
}
