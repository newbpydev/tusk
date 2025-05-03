package app

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core/task"
	taskService "github.com/newbpydev/tusk/internal/service/task"

	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/handlers"
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

	// Collapsible sections organization
	collapsibleManager     *hooks.CollapsibleManager
	timelineCollapsibleMgr *hooks.CollapsibleManager
	// The visual position of the cursor in the task list with sections
	visualCursor int
	// Whether the current task is a section header (not a task item)
	cursorOnHeader bool
	
	// Timeline cursor state
	timelineCursor      int  // Visual cursor position in the timeline
	timelineCursorOnHeader bool // Whether the timeline cursor is on a section header

	// Add separate slices for todo, projects, and completed tasks
	todoTasks, projectTasks, completedTasks []task.Task
	
	// Timeline specific task categories
	overdueTasks, todayTasks, upcomingTasks []task.Task

	// View registry for managing different views
	viewRegistry ViewRegistry
	
	// Date input handler for interactive date fields
	dateInputHandler *handlers.DateInputHandler
}

// NewModel initializes the bubbletea application model.
func NewModel(ctx context.Context, svc taskService.Service, userID int64) *Model {
	roots, err := svc.List(ctx, userID)

	m := &Model{
		ctx:                   ctx,
		tasks:                 roots,
		cursor:                0,
		err:                   err,
		taskSvc:               svc,
		userID:                userID,
		viewMode:              "list",
		styles:                styles.ActiveStyles,
		showTaskList:          true,
		showTaskDetails:       true,
		showTimeline:          true,
		activePanel:           0,
		collapsibleManager:    hooks.NewCollapsibleManager(),
		timelineCollapsibleMgr: hooks.NewCollapsibleManager(),
		dateInputHandler:      handlers.NewDateInputHandler(),
	}

	// Setup initial collapsible sections
	m.initCollapsibleSections() // Note: initCollapsibleSections will be in sections.go
	m.initTimelineCollapsibleSections() // Initialize timeline sections

	return m
}

// Init implements tea.Model Init.
func (m *Model) Init() tea.Cmd {
	m.currentTime = time.Now()
	
	// Register the due date field with the date input handler
	m.dateInputHandler.RegisterInput("dueDate", "Due Date")
	
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return messages.TickMsg(t)
	})
}

// Apply size changes based on window resize event
func (m *Model) handleWindowResize(msg tea.WindowSizeMsg) {
	// Store previous dimensions to detect changes
	prevHeight := m.height
	
	// Update dimensions
	m.width = msg.Width
	m.height = msg.Height
	
	// CRITICAL FIX: Ensure cursor visibility after resize
	// This guarantees that the cursor stays visible when window dimensions change
	if m.height != prevHeight {
		// Timeline view specific adjustments - use the numerical panel index (2 for timeline)
		if m.activePanel == 2 && m.timelineCollapsibleMgr.GetItemCount() > 0 {
			// Calculate visible height based on new window size
			visibleHeight := (m.height - 10) / 2
			
			// Current cursor position
			cursorPos := m.timelineCursor
			currentOffset := m.timelineOffset
			
			// Check if cursor is outside visible area after resize
			if cursorPos < currentOffset {
				// Cursor is above visible area, adjust offset
				m.timelineOffset = max(0, cursorPos)
			} else if cursorPos >= (currentOffset + visibleHeight) {
				// Cursor is below visible area, adjust offset
				m.timelineOffset = max(0, cursorPos - visibleHeight + 1)
			}
			
			// Final bounds check
			maxOffsetValue := max(0, m.timelineCollapsibleMgr.GetItemCount() - visibleHeight)
			m.timelineOffset = min(m.timelineOffset, maxOffsetValue)
		}
	}
}
