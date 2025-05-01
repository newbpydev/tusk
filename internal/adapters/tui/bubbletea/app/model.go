package app

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core/task"
	taskService "github.com/newbpydev/tusk/internal/service/task"

	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/styles"
)

// Model represents the state of the TUI application.
// It contains context, tasks, cursor, view state, and styling.
type Model struct {
	ctx      context.Context
	tasks    []task.Task
	cursor   int
	err      error
	taskSvc  taskService.Service
	userID   int64
	viewMode string
	width    int
	height   int
	styles   *styles.Styles

	// Form fields
	formTitle       string
	formDescription string
	formPriority    string
	formDueDate     string
	formStatus      string
	activeField     int

	// Panel visibility and focus
	showTaskList    bool
	showTaskDetails bool
	showTimeline    bool
	activePanel     int

	// Scroll offsets
	taskListOffset    int
	taskDetailsOffset int
	timelineOffset    int

	// Header status
	currentTime   time.Time
	statusMessage string
	statusType    string
	statusExpiry  time.Time
	isLoading     bool

	// Success message
	successMsg string

	// Collapsible section management
	collapsibleManager *hooks.CollapsibleManager
	// The visual position of the cursor in the task list with sections
	visualCursor int
	// Whether the current task is a section header (not a task item)
	cursorOnHeader bool

	// Add separate slices for todo, projects, and completed tasks
	todoTasks, projectTasks, completedTasks []task.Task

	// View registry for managing different views
	viewRegistry ViewRegistry
}

// NewModel initializes the bubbletea application model.
func NewModel(ctx context.Context, svc taskService.Service, userID int64) *Model {
	roots, err := svc.List(ctx, userID)

	m := &Model{
		ctx:                ctx,
		tasks:              roots,
		cursor:             0,
		err:                err,
		taskSvc:            svc,
		userID:             userID,
		viewMode:           "list",
		styles:             styles.ActiveStyles,
		showTaskList:       true,
		showTaskDetails:    true,
		showTimeline:       true,
		activePanel:        0,
		collapsibleManager: hooks.NewCollapsibleManager(),
	}

	// Setup initial collapsible sections
	m.initCollapsibleSections() // Note: initCollapsibleSections will be in sections.go

	return m
}

// Init implements tea.Model Init.
func (m *Model) Init() tea.Cmd {
	m.currentTime = time.Now()
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return messages.TickMsg(t)
	})
}
