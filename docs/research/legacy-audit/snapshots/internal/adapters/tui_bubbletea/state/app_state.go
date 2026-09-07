// Package state contains state management logic for the TUI application.
// It centralizes state updates and provides type-safe reducers.
package state

import (
	"time"

	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/core/task"
)

// AppState represents the centralized state of the application
type AppState struct {
	// View state
	Width        int
	Height       int
	ViewMode     string
	ActivePanel  int
	CurrentTime  time.Time
	
	// Task data
	Tasks           []task.Task
	TodoTasks       []task.Task
	ProjectTasks    []task.Task
	CompletedTasks  []task.Task
	OverdueTasks    []task.Task
	TodayTasks      []task.Task
	UpcomingTasks   []task.Task
	
	// Cursor state
	Cursor                  int
	VisualCursor            int
	CursorOnHeader          bool
	TaskListOffset          int
	TaskDetailsOffset       int
	TimelineCursor          int
	TimelineCursorOnHeader  bool
	
	// Status state
	StatusMessage  string
	StatusType     string
	StatusExpiry   time.Time
	IsLoading      bool
	LoadingMessage string
	Error          error
	
	// Form state
	FormTitle       string
	FormDescription string
	FormPriority    string
	FormDueDate     string
	FormIsCompleted bool
	FormParentID    *int32
	FormTaskID      *int32
	FormFocused     string
	
	// Panel visibility
	ShowTaskList    bool
	ShowTaskDetails bool
	ShowTimeline    bool
	
	// Collapsible state
	CollapsibleManager        *hooks.CollapsibleManager
	TimelineCollapsibleManager *hooks.CollapsibleManager
}

// NewAppState creates a new application state with default values
func NewAppState() *AppState {
	return &AppState{
		ViewMode:      "list",
		ActivePanel:   0,
		CurrentTime:   time.Now(),
		ShowTaskList:  true,
		ShowTaskDetails: true,
		ShowTimeline: true,
	}
}

// StateUpdater provides an interface for components that update state
type StateUpdater interface {
	UpdateState(state *AppState) *AppState
}

// TasksUpdated updates the task lists and recategorizes them
func (s *AppState) TasksUpdated(tasks []task.Task) *AppState {
	s.Tasks = tasks
	
	// This would use the TaskCategorizationService to sort tasks
	// For now, we'll use a simplified approach
	var todoTasks, projectTasks, completedTasks []task.Task
	for _, t := range tasks {
		if t.Status == task.StatusDone {
			completedTasks = append(completedTasks, t)
		} else if t.ParentID != nil {
			projectTasks = append(projectTasks, t)
		} else {
			todoTasks = append(todoTasks, t)
		}
	}
	
	s.TodoTasks = todoTasks
	s.ProjectTasks = projectTasks
	s.CompletedTasks = completedTasks
	
	// Also recategorize timeline tasks
	var overdueTasks, todayTasks, upcomingTasks []task.Task
	now := time.Now()
	
	for _, t := range tasks {
		if t.DueDate == nil || t.Status == task.StatusDone {
			continue
		}
		
		y1, m1, d1 := t.DueDate.Date()
		y2, m2, d2 := now.Date()
		
		if y1 < y2 || (y1 == y2 && m1 < m2) || (y1 == y2 && m1 == m2 && d1 < d2) {
			// Overdue
			overdueTasks = append(overdueTasks, t)
		} else if y1 == y2 && m1 == m2 && d1 == d2 {
			// Today
			todayTasks = append(todayTasks, t)
		} else {
			// Upcoming
			upcomingTasks = append(upcomingTasks, t)
		}
	}
	
	s.OverdueTasks = overdueTasks
	s.TodayTasks = todayTasks
	s.UpcomingTasks = upcomingTasks
	
	return s
}

// SetStatusMessage updates the status message with type and expiry
func (s *AppState) SetStatusMessage(message, statusType string, duration time.Duration) *AppState {
	s.StatusMessage = message
	s.StatusType = statusType
	
	if duration > 0 {
		s.StatusExpiry = time.Now().Add(duration)
	} else {
		s.StatusExpiry = time.Time{}
	}
	
	return s
}

// SetLoadingStatus sets the loading status
func (s *AppState) SetLoadingStatus(message string) *AppState {
	s.IsLoading = true
	s.LoadingMessage = message
	return s
}

// ClearLoadingStatus clears the loading status
func (s *AppState) ClearLoadingStatus() *AppState {
	s.IsLoading = false
	s.LoadingMessage = ""
	return s
}

// SetError updates the error state
func (s *AppState) SetError(err error) *AppState {
	s.Error = err
	return s
}

// ResetForm clears the form fields
func (s *AppState) ResetForm() *AppState {
	s.FormTitle = ""
	s.FormDescription = ""
	s.FormPriority = string(task.PriorityLow)
	s.FormDueDate = ""
	s.FormIsCompleted = false
	s.FormParentID = nil
	s.FormTaskID = nil
	s.FormFocused = "title"
	return s
}

// LoadTaskIntoForm loads a task's data into the form
func (s *AppState) LoadTaskIntoForm(t task.Task) *AppState {
	s.FormTitle = t.Title
	// Handle nil Description pointer
	if t.Description != nil {
		s.FormDescription = *t.Description
	} else {
		s.FormDescription = ""
	}
	s.FormPriority = string(t.Priority)
	
	if t.DueDate != nil {
		s.FormDueDate = t.DueDate.Format("2006-01-02")
	} else {
		s.FormDueDate = ""
	}
	
	s.FormIsCompleted = t.IsCompleted
	s.FormParentID = t.ParentID
	s.FormTaskID = &t.ID
	s.FormFocused = "title"
	
	return s
}
