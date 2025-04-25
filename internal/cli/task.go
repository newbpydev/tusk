package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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

		dueDateStr, err := cmd.Flags().GetString("due")
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

		// Get parent-id flag value for subtasks
		var parentID *int64
		parentIDValue, err := cmd.Flags().GetInt64("parent-id")
		if err != nil {
			return err
		}

		// Only set parentID if it was explicitly provided (non-zero value)
		if parentIDValue > 0 {
			parentID = &parentIDValue
		}

		// Get status for the task if provided
		status, err := cmd.Flags().GetString("status")
		if err != nil {
			return err
		}

		// First create the task with default status
		t, err := taskSvc.Create(context.Background(), userID, parentID, title, description, dueDate, task.Priority(priority), tags)
		if err != nil {
			return err
		}

		// Then update the status if it was provided and is different from default
		if status != "" && status != "todo" {
			t, err = taskSvc.ChangeStatus(context.Background(), int64(t.ID), task.Status(status))
			if err != nil {
				cmd.Printf("Warning: Created task but failed to set status: %v\n", err)
			}
		}

		cmd.Printf("Created task %d (%s)\n", t.ID, t.Title)
		if parentID != nil {
			cmd.Printf("This is a subtask of task %d\n", *parentID)
		}
		return nil
	},
}

var showTaskCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show a task's details and subtasks",
	Long:  "This command shows the details of a task and its subtasks.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var id int64
		var err error

		// Check if an ID was provided as a flag
		id, err = cmd.Flags().GetInt64("id")
		if err != nil || id == 0 {
			// If not provided as a flag, check if it was provided as an argument
			if len(args) > 0 {
				// Try to parse the ID from the argument
				id, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid task ID: %v", err)
				}
			} else {
				return fmt.Errorf("task ID is required (use --id flag or provide as argument)")
			}
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

var searchByTitleCmd = &cobra.Command{
	Use:   "search",
	Short: "Search tasks by title",
	Long:  "Search for tasks containing the specified text in their title.",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetInt64("user-id")
		if err != nil {
			return err
		}

		searchTerm, err := cmd.Flags().GetString("query")
		if err != nil {
			return err
		}

		if searchTerm == "" {
			return fmt.Errorf("search query cannot be empty")
		}

		tasks, err := taskSvc.SearchByTitle(context.Background(), userID, searchTerm)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			cmd.Printf("No tasks found matching '%s'\n", searchTerm)
			return nil
		}

		cmd.Printf("Found %d tasks matching '%s':\n\n", len(tasks), searchTerm)
		for i, t := range tasks {
			cmd.Printf("%d. [%s] %s (Priority: %s)\n", i+1, t.Status, t.Title, t.Priority)
		}

		return nil
	},
}

var listByTagCmd = &cobra.Command{
	Use:   "by-tag",
	Short: "List tasks with a specific tag",
	Long:  "List all tasks that have been tagged with the specified tag.",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetInt64("user-id")
		if err != nil {
			return err
		}

		tag, err := cmd.Flags().GetString("tag")
		if err != nil {
			return err
		}

		if tag == "" {
			return fmt.Errorf("tag cannot be empty")
		}

		tasks, err := taskSvc.SearchByTag(context.Background(), userID, tag)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			cmd.Printf("No tasks found with tag '%s'\n", tag)
			return nil
		}

		cmd.Printf("Found %d tasks with tag '%s':\n\n", len(tasks), tag)
		for i, t := range tasks {
			cmd.Printf("%d. [%s] %s (Priority: %s)\n", i+1, t.Status, t.Title, t.Priority)
		}

		return nil
	},
}

var listByStatusCmd = &cobra.Command{
	Use:   "by-status",
	Short: "List tasks with a specific status",
	Long:  "List all tasks that have the specified status (todo, in-progress, done).",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetInt64("user-id")
		if err != nil {
			return err
		}

		status, err := cmd.Flags().GetString("status")
		if err != nil {
			return err
		}

		if status == "" {
			return fmt.Errorf("status cannot be empty")
		}

		tasks, err := taskSvc.ListByStatus(context.Background(), userID, task.Status(status))
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			cmd.Printf("No tasks found with status '%s'\n", status)
			return nil
		}

		cmd.Printf("Found %d tasks with status '%s':\n\n", len(tasks), status)
		for i, t := range tasks {
			priority := string(t.Priority)
			cmd.Printf("%d. %s (Priority: %s)\n", i+1, t.Title, priority)
		}

		return nil
	},
}

var dueTodayCmd = &cobra.Command{
	Use:   "due-today",
	Short: "List tasks due today",
	Long:  "List all tasks that are due today and not completed.",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetInt64("user-id")
		if err != nil {
			return err
		}

		tasks, err := taskSvc.ListTasksDueToday(context.Background(), userID)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			cmd.Printf("No tasks due today\n")
			return nil
		}

		cmd.Printf("Tasks due today (%d):\n\n", len(tasks))
		for i, t := range tasks {
			priority := string(t.Priority)
			cmd.Printf("%d. [%s] %s (Priority: %s)\n", i+1, t.Status, t.Title, priority)
		}

		return nil
	},
}

var overdueCmd = &cobra.Command{
	Use:   "overdue",
	Short: "List overdue tasks",
	Long:  "List all tasks that are past their due date and not completed.",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetInt64("user-id")
		if err != nil {
			return err
		}

		tasks, err := taskSvc.ListOverdueTasks(context.Background(), userID)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			cmd.Printf("No overdue tasks\n")
			return nil
		}

		cmd.Printf("Overdue tasks (%d):\n\n", len(tasks))
		for i, t := range tasks {
			priority := string(t.Priority)
			dueDate := "Unknown"
			if t.DueDate != nil {
				dueDate = t.DueDate.Format("2006-01-02")
			}
			cmd.Printf("%d. [%s] %s (Priority: %s, Due: %s)\n", i+1, t.Status, t.Title, priority, dueDate)
		}

		return nil
	},
}

var upcomingCmd = &cobra.Command{
	Use:   "upcoming",
	Short: "List upcoming tasks",
	Long:  "List all tasks that are due within the next 7 days.",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetInt64("user-id")
		if err != nil {
			return err
		}

		tasks, err := taskSvc.ListTasksDueSoon(context.Background(), userID)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			cmd.Printf("No upcoming tasks in the next 7 days\n")
			return nil
		}

		cmd.Printf("Upcoming tasks in the next 7 days (%d):\n\n", len(tasks))
		for i, t := range tasks {
			priority := string(t.Priority)
			dueDate := "Unknown"
			if t.DueDate != nil {
				dueDate = t.DueDate.Format("2006-01-02")
			}
			cmd.Printf("%d. [%s] %s (Priority: %s, Due: %s)\n", i+1, t.Status, t.Title, priority, dueDate)
		}

		return nil
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show task statistics",
	Long:  "Show statistics about tasks, including counts by status and priority.",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetInt64("user-id")
		if err != nil {
			return err
		}

		statusCounts, err := taskSvc.GetTaskCountsByStatus(context.Background(), userID)
		if err != nil {
			return err
		}

		priorityCounts, err := taskSvc.GetTaskCountsByPriority(context.Background(), userID)
		if err != nil {
			return err
		}

		cmd.Printf("Task Statistics for User %d\n\n", userID)

		cmd.Printf("Status counts:\n")
		cmd.Printf("  Todo: %d\n", statusCounts.TodoCount)
		cmd.Printf("  In Progress: %d\n", statusCounts.InProgressCount)
		cmd.Printf("  Done: %d\n", statusCounts.DoneCount)
		cmd.Printf("  Total: %d\n\n", statusCounts.TotalCount)

		cmd.Printf("Priority counts (incomplete tasks):\n")
		cmd.Printf("  Low: %d\n", priorityCounts.LowCount)
		cmd.Printf("  Medium: %d\n", priorityCounts.MediumCount)
		cmd.Printf("  High: %d\n", priorityCounts.HighCount)

		// Calculate completion rate
		completionRate := 0.0
		if statusCounts.TotalCount > 0 {
			completionRate = float64(statusCounts.DoneCount) / float64(statusCounts.TotalCount) * 100
		}
		cmd.Printf("\nCompletion rate: %.1f%%\n", completionRate)

		return nil
	},
}

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags",
	Long:  "List all unique tags used across all tasks for a user.",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, err := cmd.Flags().GetInt64("user-id")
		if err != nil {
			return err
		}

		tags, err := taskSvc.GetAllTags(context.Background(), userID)
		if err != nil {
			return err
		}

		if len(tags) == 0 {
			cmd.Printf("No tags found\n")
			return nil
		}

		cmd.Printf("All tags (%d):\n\n", len(tags))
		cmd.Println(strings.Join(tags, ", "))

		return nil
	},
}

var bulkUpdateStatusCmd = &cobra.Command{
	Use:   "bulk-update-status",
	Short: "Update status for multiple tasks at once",
	Long:  "Update the status of multiple tasks in a single operation",
	RunE: func(cmd *cobra.Command, args []string) error {
		// The userID variable is not used in this function, so we can remove it
		// We still check for errors from flag parsing to maintain consistency

		status, err := cmd.Flags().GetString("status")
		if err != nil || status == "" {
			return fmt.Errorf("status is required")
		}

		idsStr, err := cmd.Flags().GetStringArray("task-ids")
		if err != nil {
			return err
		}

		if len(idsStr) == 0 {
			return fmt.Errorf("at least one task ID is required")
		}

		// Convert string IDs to int32 slice
		taskIDs := make([]int32, 0, len(idsStr))
		for _, idStr := range idsStr {
			id, err := strconv.ParseInt(idStr, 10, 32)
			if err != nil {
				return fmt.Errorf("invalid task ID '%s': %v", idStr, err)
			}
			taskIDs = append(taskIDs, int32(id))
		}

		// Call the service to bulk update tasks
		err = taskSvc.BulkUpdateStatus(context.Background(), taskIDs, task.Status(status))
		if err != nil {
			return err
		}

		cmd.Printf("Successfully updated %d tasks to status '%s'\n", len(taskIDs), status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)

	// Add all task-related commands
	taskCmd.AddCommand(
		createTaskCmd,
		showTaskCmd,
		searchByTitleCmd,
		listByTagCmd,
		listByStatusCmd,
		dueTodayCmd,
		overdueCmd,
		upcomingCmd,
		statsCmd,
		tagsCmd,
		bulkUpdateStatusCmd,
	)

	// Flags for createTaskCmd
	createTaskCmd.Flags().Int64P("user-id", "u", 0, "User ID of the task owner")
	createTaskCmd.Flags().StringP("title", "t", "", "Title of the task")
	createTaskCmd.Flags().StringP("description", "d", "", "Description of the task")
	createTaskCmd.Flags().StringP("due", "D", "", "Due date of the task in RFC3339 format")
	createTaskCmd.Flags().StringP("priority", "p", "medium", "Priority of the task (low, medium, high)")
	createTaskCmd.Flags().StringArrayP("tags", "g", []string{}, "Tags for the task")
	createTaskCmd.Flags().Int64P("parent-id", "P", 0, "Parent task ID for creating subtasks")
	createTaskCmd.Flags().StringP("status", "s", "todo", "Status of the task (todo, in-progress, done)")

	createTaskCmd.MarkFlagRequired("user-id")
	createTaskCmd.MarkFlagRequired("title")

	// Flags for showTaskCmd
	showTaskCmd.Flags().Int64P("id", "i", 0, "Task ID to show")

	// Search and filtering commands flags
	searchByTitleCmd.Flags().Int64P("user-id", "u", 0, "User ID of the task owner")
	searchByTitleCmd.Flags().StringP("query", "q", "", "Text to search for in task titles")
	searchByTitleCmd.MarkFlagRequired("user-id")
	searchByTitleCmd.MarkFlagRequired("query")

	listByTagCmd.Flags().Int64P("user-id", "u", 0, "User ID of the task owner")
	listByTagCmd.Flags().StringP("tag", "t", "", "Tag to search for")
	listByTagCmd.MarkFlagRequired("user-id")
	listByTagCmd.MarkFlagRequired("tag")

	listByStatusCmd.Flags().Int64P("user-id", "u", 0, "User ID of the task owner")
	listByStatusCmd.Flags().StringP("status", "s", "", "Status to filter by (todo, in-progress, done)")
	listByStatusCmd.MarkFlagRequired("user-id")
	listByStatusCmd.MarkFlagRequired("status")

	// Due date commands flags
	dueTodayCmd.Flags().Int64P("user-id", "u", 0, "User ID of the task owner")
	dueTodayCmd.MarkFlagRequired("user-id")

	overdueCmd.Flags().Int64P("user-id", "u", 0, "User ID of the task owner")
	overdueCmd.MarkFlagRequired("user-id")

	upcomingCmd.Flags().Int64P("user-id", "u", 0, "User ID of the task owner")
	upcomingCmd.MarkFlagRequired("user-id")

	// Stats command flags
	statsCmd.Flags().Int64P("user-id", "u", 0, "User ID of the task owner")
	statsCmd.MarkFlagRequired("user-id")

	// Tags command flags
	tagsCmd.Flags().Int64P("user-id", "u", 0, "User ID of the task owner")
	tagsCmd.MarkFlagRequired("user-id")

	// Bulk update command flags
	bulkUpdateStatusCmd.Flags().Int64P("user-id", "u", 0, "User ID of the task owner")
	bulkUpdateStatusCmd.Flags().StringP("status", "s", "", "Status to set (todo, in-progress, done)")
	bulkUpdateStatusCmd.Flags().StringArrayP("task-ids", "t", []string{}, "Task IDs to update")
	bulkUpdateStatusCmd.MarkFlagRequired("user-id")
	bulkUpdateStatusCmd.MarkFlagRequired("status")
	bulkUpdateStatusCmd.MarkFlagRequired("task-ids")
}

// Helper functions to convert between types
func int32PtrToInt32Ptr(i64 *int64) *int32 {
	if i64 == nil {
		return nil
	}
	i32 := int32(*i64)
	return &i32
}

func stringPtrFromString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringsToTags(strs []string) []task.Tag {
	tags := make([]task.Tag, 0, len(strs))
	for _, s := range strs {
		if s != "" {
			tags = append(tags, task.Tag{Name: s})
		}
	}
	return tags
}
