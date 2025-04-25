package cli

import (
	"context"
	"time"

	"github.com/newbpydev/tusk/internal/core/task"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks in the Tusk application",
	Long:  "This command allows you to create, update, delete, and list tasks in the Tusk application.",
}

var createTaskCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	Long:  "This command creates a new task in the Tusk application.",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetInt64("user-id")
		if err != nil {
			return err
		}

		title, err := cmd.Flags().GetString("title")
		if err != nil {
			return err
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		dueDateStr, err := cmd.Flags().GetString("due-date")
		if err != nil {
			return err
		}

		var dueDate *time.Time
		if dueDateStr != "" {
			dueDateParsed, err := time.Parse(time.RFC3339, dueDateStr)
			if err != nil {
				return err
			}
			dueDate = &dueDateParsed
		}

		priority, err := cmd.Flags().GetString("priority")
		if err != nil {
			return err
		}

		tags, err := cmd.Flags().GetStringArray("tags")
		if err != nil {
			return err
		}

		// Added parentID parameter with nil value for root tasks
		var parentID *int64

		t, err := taskSvc.Create(context.Background(), userID, parentID, title, description, dueDate, task.Priority(priority), tags)
		if err != nil {
			return err
		}

		cmd.Printf("Created task %d (%s)\n", t.ID, t.Title)
		return nil
	},
}

var showTaskCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show a task's details and subtasks",
	Long:  "This command shows the details of a task and its subtasks.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := cmd.Flags().GetInt64("id")
		if err != nil {
			return err
		}

		// Changed Get to Show to match the service interface
		t, err := taskSvc.Show(context.Background(), id)
		if err != nil {
			return err
		}

		cmd.Printf("Task ID: %d\n", t.ID)
		cmd.Printf("Title: %s\n", t.Title)

		// Fixed the description format for pointer to string
		desc := "No description"
		if t.Description != nil {
			desc = *t.Description
		}
		cmd.Printf("Description: %s\n", desc)

		// Check if DueDate is nil before formatting
		dueDate := "Not set"
		if t.DueDate != nil {
			dueDate = t.DueDate.Format(time.RFC3339)
		}
		cmd.Printf("Due Date: %s\n", dueDate)

		cmd.Printf("Priority: %s\n", t.Priority)
		cmd.Printf("Tags: %v\n", t.Tags)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(createTaskCmd, showTaskCmd)

	createTaskCmd.Flags().Int64P("user-id", "u", 0, "User ID of the task owner")
	createTaskCmd.Flags().StringP("title", "t", "", "Title of the task")
	createTaskCmd.Flags().StringP("description", "d", "", "Description of the task")
	createTaskCmd.Flags().StringP("due", "D", "", "Due date of the task in RFC3339 format")
	createTaskCmd.Flags().StringP("priority", "p", "normal", "Priority of the task")
	createTaskCmd.Flags().StringArrayP("tags", "g", []string{}, "Tags for the task")

	createTaskCmd.MarkFlagRequired("user-id")
	createTaskCmd.MarkFlagRequired("title")
}
