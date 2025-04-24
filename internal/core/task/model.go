package task

import "time"

// Tag represents a tag associated with a task.
// Tags can be used to categorize tasks and make them easier to find.
type Tag struct {
	Name string `json:"name"`
	// Color string `json:"color"` // TODO: add color to tag
	// Icon  string `json:"icon"`  // TODO: add icon to tag
}

// Status represents the status of a task.
// It can be one of the following values: "todo", "in_progress", or "done".
type Status string

// Priority represents the priority of a task.
// It can be one of the following values: "low", "medium", or "high".
type Priority string

const (
	// StatusTodo represents a task that is yet to be started.
	StatusTodo Status = "todo"
	// StatusInProgress represents a task that is currently being worked on.
	StatusInProgress Status = "in_progress"
	// StatusDone represents a task that has been completed.
	StatusDone Status = "done"

	// PriorityLow represents a task with low priority.
	PriorityLow Priority = "low"
	// PriorityMedium represents a task with medium priority.
	PriorityMedium Priority = "medium"
	// PriorityHigh represents a task with high priority.
	PriorityHigh Priority = "high"
)

// Task represents a task in the system.
// It includes fields for the task's ID, user ID, parent ID, title, description,
// created and updated timestamps, due date, status, priority, tags, and display order.
// It also includes a list of sub-tasks and computed fields for total count, completed count, and progress.
type Task struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	ParentID     *int64     `json:"parent_id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DueDate      *time.Time `json:"due_date"`
	Status       Status     `json:"status"`
	Priority     Priority   `json:"priority"`
	Tags         []Tag      `json:"tags"`
	DisplayOrder int        `json:"display_order"`

	// Children hierarchical tasks
	SubTasks []Task `json:"sub_tasks"`

	// Computed fields
	TotalCount     int     `json:"total_count"`
	CompletedCount int     `json:"completed_count"`
	Progress       float64 `json:"progress"` // CompletedCount / TotalCount * (0.0-1.0)
}
