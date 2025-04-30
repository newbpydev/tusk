package app

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/core/task"
	taskService "github.com/newbpydev/tusk/internal/service/task"

	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/layout"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/panels"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
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
	m.initCollapsibleSections()

	return m
}

// Init implements tea.Model Init.
func (m *Model) Init() tea.Cmd {
	m.currentTime = time.Now()
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return messages.TickMsg(t)
	})
}

// Update implements tea.Model Update, handling all message types.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		newModel, cmd := m.handleKeyPress(msg)
		return newModel, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case messages.TickMsg:
		m.currentTime = time.Time(msg)
		if (!m.statusExpiry.IsZero()) && time.Now().After(m.statusExpiry) {
			m.statusMessage = ""
			m.statusType = ""
			m.statusExpiry = time.Time{}
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return messages.TickMsg(t)
		})

	case messages.StatusUpdateErrorMsg:
		m.err = msg.Err
		m.setErrorStatus(fmt.Sprintf("Error updating task '%s': %v", msg.TaskTitle, msg.Err))
		return m, m.refreshTasks()

	case messages.StatusUpdateSuccessMsg:
		m.setSuccessStatus(msg.Message)
		// Update local task entry only
		for i := range m.tasks {
			if m.tasks[i].ID == msg.Task.ID {
				m.tasks[i] = msg.Task
				break
			}
		}
		// No reposition here; toggleTaskCompletion already handled cursor movement
		return m, nil

	case messages.TasksRefreshedMsg:
		m.tasks = msg.Tasks
		if m.cursor >= len(m.tasks) {
			m.cursor = max(0, len(m.tasks)-1)
		}
		m.clearLoadingStatus()
		m.initCollapsibleSections()
		return m, nil

	case messages.ErrorMsg:
		m.err = error(msg)
		m.setErrorStatus(fmt.Sprintf("Error: %v", error(msg)))
		return m, nil

	default:
		return m, nil
	}
}

// handleKeyPress processes keyboard input by viewMode
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.viewMode {
	case "list":
		return m.handleListViewKeys(msg)
	case "detail":
		return m.handleDetailViewKeys(msg)
	case "edit":
		return m.handleEditViewKeys(msg)
	case "create":
		return m.handleCreateFormKeys(msg)
	default:
		return m, nil
	}
}

// initCollapsibleSections initializes the sections in the task list
func (m *Model) initCollapsibleSections() {
	if m.collapsibleManager == nil {
		m.collapsibleManager = hooks.NewCollapsibleManager()
	}

	// Count todo and completed tasks
	var todoCount, completedCount int
	for _, t := range m.tasks {
		if t.Status == task.StatusDone {
			completedCount++
		} else {
			todoCount++
		}
	}

	// Clear and reinitialize sections
	m.collapsibleManager.ClearSections()

	// Add our sections - order matters as it determines the index
	m.collapsibleManager.AddSection(hooks.SectionTypeTodo, "Todo", todoCount, 0)
	m.collapsibleManager.AddSection(hooks.SectionTypeProjects, "Projects", 0, todoCount)
	m.collapsibleManager.AddSection(hooks.SectionTypeCompleted, "Completed", completedCount, todoCount)

	// Initialize visual cursor at the right position
	m.updateVisualCursorFromTaskCursor()
}

// Refactor categorizeTasks to update slices directly
func (m *Model) categorizeTasks(tasks []task.Task) {
	m.todoTasks = nil
	m.projectTasks = nil
	m.completedTasks = nil

	for _, t := range tasks {
		if t.Status == task.StatusDone {
			m.completedTasks = append(m.completedTasks, t)
		} else if t.ParentID != nil {
			m.projectTasks = append(m.projectTasks, t)
		} else {
			m.todoTasks = append(m.todoTasks, t)
		}
	}

	// Update collapsible sections
	m.collapsibleManager.ClearSections()
	m.collapsibleManager.AddSection(hooks.SectionTypeTodo, "Todo", len(m.todoTasks), 0)
	m.collapsibleManager.AddSection(hooks.SectionTypeProjects, "Projects", len(m.projectTasks), len(m.todoTasks))
	m.collapsibleManager.AddSection(hooks.SectionTypeCompleted, "Completed", len(m.completedTasks), len(m.todoTasks)+len(m.projectTasks))
}

// updateVisualCursorFromTaskCursor translates the task index (cursor) to the visual cursor position
func (m *Model) updateVisualCursorFromTaskCursor() {
	if m.collapsibleManager == nil {
		m.visualCursor = m.cursor
		return
	}

	// Find which section contains this task
	m.visualCursor = m.collapsibleManager.GetVisibleIndexFromTaskIndex(m.cursor)
	m.cursorOnHeader = false
}

// updateTaskCursorFromVisualCursor translates the visual cursor to the actual task index
func (m *Model) updateTaskCursorFromVisualCursor() {
	if m.collapsibleManager == nil {
		m.cursor = m.visualCursor
		return
	}

	// Check if we're on a section header
	if m.collapsibleManager.IsSectionHeader(m.visualCursor) {
		// We're on a header, cursor shouldn't point to a task
		m.cursorOnHeader = true
		return
	}

	// Get the actual task index
	taskIndex := m.collapsibleManager.GetActualTaskIndex(m.visualCursor)
	if taskIndex != -1 {
		m.cursor = taskIndex
		m.cursorOnHeader = false
	}
}

// handleListViewKeys processes keyboard input in list view with collapsible sections support
func (m *Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Initialize sections if needed
	if m.collapsibleManager == nil {
		m.initCollapsibleSections()
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		switch m.activePanel {
		case 0: // Task list panel
			if m.collapsibleManager != nil {
				prevVisual := m.visualCursor

				// Move up in the visual list (includes section headers)
				m.visualCursor = m.collapsibleManager.GetNextCursorPosition(m.visualCursor, -1)

				// Auto-scroll if cursor moves out of view
				if m.visualCursor < m.taskListOffset {
					m.taskListOffset = m.visualCursor
				}

				// Update the task cursor based on the new visual position
				m.updateTaskCursorFromVisualCursor()

				// Reset detail & timeline scroll offsets if selection changed
				if prevVisual != m.visualCursor {
					m.taskDetailsOffset = 0
					m.timelineOffset = 0
				}
			} else {
				// Legacy behavior without collapsible sections
				if m.cursor > 0 {
					prevCursor := m.cursor
					m.cursor--
					if m.cursor < m.taskListOffset {
						m.taskListOffset = m.cursor
					}
					if prevCursor != m.cursor {
						m.taskDetailsOffset = 0
						m.timelineOffset = 0
					}
				}
			}
		case 1: // Task Details panel - instant scroll
			viewportHeight := 10
			m.taskDetailsOffset = max(0, m.taskDetailsOffset-viewportHeight)
		case 2: // Timeline panel - instant scroll
			viewportHeight := 10
			m.timelineOffset = max(0, m.timelineOffset-viewportHeight)
		}
		return m, nil

	case "down", "j":
		switch m.activePanel {
		case 0: // Task list panel
			if m.collapsibleManager != nil {
				prevVisual := m.visualCursor

				// Move down in the visual list (includes section headers)
				m.visualCursor = m.collapsibleManager.GetNextCursorPosition(m.visualCursor, 1)

				// Auto-scroll if cursor moves out of view
				viewportHeight := 10 // Approximate visible lines
				if m.visualCursor >= m.taskListOffset+viewportHeight {
					m.taskListOffset = m.visualCursor - viewportHeight + 1
				}

				// Update the task cursor based on the new visual position
				m.updateTaskCursorFromVisualCursor()

				// Reset detail & timeline scroll offsets if selection changed
				if prevVisual != m.visualCursor {
					m.taskDetailsOffset = 0
					m.timelineOffset = 0
				}
			} else {
				// Legacy behavior without collapsible sections
				if m.cursor < len(m.tasks)-1 {
					prevCursor := m.cursor
					m.cursor++
					viewportHeight := 10
					if m.cursor >= m.taskListOffset+viewportHeight {
						m.taskListOffset = m.cursor - viewportHeight + 1
					}
					if prevCursor != m.cursor {
						m.taskDetailsOffset = 0
						m.timelineOffset = 0
					}
				}
			}
		case 1: // Task Details panel - instant scroll
			viewportHeight := 10
			maxOffset := 15
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].Description != nil {
				maxOffset += len(*m.tasks[m.cursor].Description) / 30
			}
			m.taskDetailsOffset = min(m.taskDetailsOffset+viewportHeight, maxOffset)
		case 2: // Timeline panel - instant scroll
			viewportHeight := 10
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOffset := len(overdue) + len(today) + len(upcoming) + 15
			m.timelineOffset = min(m.timelineOffset+viewportHeight, maxOffset)
		}
		return m, nil

	case "page-up", "ctrl+b":
		pageSize := 20 // Larger page size for details and timeline
		switch m.activePanel {
		case 0:
			if m.collapsibleManager != nil {
				// Move up by a page in the visual list
				prevVisual := m.visualCursor

				m.visualCursor = max(0, m.visualCursor-pageSize)
				m.taskListOffset = max(0, m.taskListOffset-pageSize)

				// Update the task cursor based on the new visual position
				m.updateTaskCursorFromVisualCursor()

				// Reset detail & timeline scroll offsets if selection changed
				if prevVisual != m.visualCursor {
					m.taskDetailsOffset = 0
					m.timelineOffset = 0
				}
			} else {
				// Legacy behavior
				prev := m.cursor
				m.taskListOffset -= pageSize
				if m.taskListOffset < 0 {
					m.taskListOffset = 0
				}
				if m.cursor >= m.taskListOffset+pageSize {
					m.cursor = m.taskListOffset
				}
				if prev != m.cursor {
					m.taskDetailsOffset = 0
					m.timelineOffset = 0
				}
			}
		case 1:
			m.taskDetailsOffset = max(0, m.taskDetailsOffset-pageSize)
		case 2:
			m.timelineOffset = max(0, m.timelineOffset-pageSize)
		}
		return m, nil

	case "page-down", "ctrl+f":
		pageSize := 20 // Larger page size for details and timeline
		var maxOffset int
		switch m.activePanel {
		case 0:
			if m.collapsibleManager != nil {
				// Move down by a page in the visual list
				prevVisual := m.visualCursor
				totalItems := m.collapsibleManager.GetItemCount()

				// Calculate maximum offset and cursor positions
				maxOffset = max(0, totalItems-pageSize)
				m.taskListOffset += pageSize
				if m.taskListOffset > maxOffset {
					m.taskListOffset = maxOffset
				}

				m.visualCursor = min(totalItems-1, m.visualCursor+pageSize)

				// Update the task cursor based on the new visual position
				m.updateTaskCursorFromVisualCursor()

				// Reset detail & timeline scroll offsets if selection changed
				if prevVisual != m.visualCursor {
					m.taskDetailsOffset = 0
					m.timelineOffset = 0
				}
			} else {
				// Legacy behavior
				maxOffset = max(0, len(m.tasks)-pageSize)
				m.taskListOffset += pageSize
				if m.taskListOffset > maxOffset {
					m.taskListOffset = maxOffset
				}
				if m.cursor < m.taskListOffset {
					m.cursor = m.taskListOffset
				}
			}
		case 1:
			maxOffset = 15
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].Description != nil {
				maxOffset += len(*m.tasks[m.cursor].Description) / 30
			}
			m.taskDetailsOffset = min(m.taskDetailsOffset+pageSize, maxOffset)
		case 2:
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOffset = len(overdue) + len(today) + len(upcoming) + 15
			m.timelineOffset = min(m.timelineOffset+pageSize, maxOffset)
		}
		return m, nil

	case "home", "g":
		switch m.activePanel {
		case 0:
			if m.collapsibleManager != nil {
				// Go to the first item (should be the first section header)
				prevVisual := m.visualCursor
				m.visualCursor = 0
				m.taskListOffset = 0

				// Update the task cursor
				m.updateTaskCursorFromVisualCursor()

				// Reset scroll positions if cursor changed
				if prevVisual != m.visualCursor {
					m.taskDetailsOffset = 0
					m.timelineOffset = 0
				}
			} else {
				// Legacy behavior
				m.taskListOffset = 0
				m.cursor = 0
			}
		case 1:
			m.taskDetailsOffset = 0
		case 2:
			m.timelineOffset = 0
		}
		return m, nil

	case "end", "G":
		pageSize := 10
		switch m.activePanel {
		case 0:
			if m.collapsibleManager != nil {
				// Go to the last item in the visual list
				prevVisual := m.visualCursor
				totalItems := m.collapsibleManager.GetItemCount()

				if totalItems > 0 {
					m.visualCursor = totalItems - 1
					m.taskListOffset = max(0, m.visualCursor-pageSize+1)

					// Update the task cursor
					m.updateTaskCursorFromVisualCursor()

					// Reset scroll positions if cursor changed
					if prevVisual != m.visualCursor {
						m.taskDetailsOffset = 0
						m.timelineOffset = 0
					}
				}
			} else {
				// Legacy behavior
				if len(m.tasks) > 0 {
					m.cursor = len(m.tasks) - 1
					m.taskListOffset = max(0, m.cursor-pageSize+1)
				}
			}
		case 1:
			maxOff := 15
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].Description != nil {
				maxOff += len(*m.tasks[m.cursor].Description) / 30
			}
			m.taskDetailsOffset = max(0, maxOff-pageSize)
		case 2:
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOff := len(overdue) + len(today) + len(upcoming) + 15
			m.timelineOffset = max(0, maxOff-pageSize)
		}
		return m, nil

	case "enter", " ":
		// Only respond when task list panel is active
		if m.activePanel == 0 {
			// Check if we're on a section header
			if m.cursorOnHeader && m.collapsibleManager != nil {
				// Toggle the section's expanded state
				section := m.collapsibleManager.GetSectionAtIndex(m.visualCursor)
				if section != nil {
					m.collapsibleManager.ToggleSection(section.Type)
					// Don't reset cursor state after toggling
					return m, nil
				}
				return m, nil
			} else if !m.cursorOnHeader && len(m.tasks) > 0 && m.cursor < len(m.tasks) {
				// We're on a task (not a header), go to detail view
				m.viewMode = "detail"
				return m, nil
			}
		}
		return m, nil

	case "c":
		// Only toggle completion status when on a task (not section header)
		if m.activePanel == 0 && !m.cursorOnHeader && len(m.tasks) > 0 && m.cursor < len(m.tasks) {
			return m, m.toggleTaskCompletion()
		}
		return m, nil

	// Keep the rest of the key bindings unchanged
	case "r":
		return m, m.refreshTasks()

	case "n":
		m.viewMode = "create"
		m.formStatus = string(task.StatusTodo)
		m.formPriority = string(task.PriorityLow)
		return m, nil

	case "1":
		m.showTaskList = !m.showTaskList
		if m.activePanel == 0 && !m.showTaskList {
			if m.showTaskDetails {
				m.activePanel = 1
			} else if m.showTimeline {
				m.activePanel = 2
			}
		}
		return m, nil

	case "2":
		m.showTaskDetails = !m.showTaskDetails
		if m.activePanel == 1 && !m.showTaskDetails {
			if m.showTimeline {
				m.activePanel = 2
			} else if m.showTaskList {
				m.activePanel = 0
			}
		}
		return m, nil

	case "3":
		m.showTimeline = !m.showTimeline
		if m.activePanel == 2 && !m.showTimeline {
			if m.showTaskList {
				m.activePanel = 0
			} else if m.showTaskDetails {
				m.activePanel = 1
			}
		}
		return m, nil

	case "right", "l":
		// Move focus to the next visible panel
		visiblePanels := []bool{m.showTaskList, m.showTaskDetails, m.showTimeline}
		originalPanel := m.activePanel

		// Find next visible panel
		for i := 0; i < 3; i++ {
			m.activePanel = (m.activePanel + 1) % 3
			if visiblePanels[m.activePanel] {
				break
			}
		}

		// If no other panels are visible, keep the original panel
		if !visiblePanels[m.activePanel] {
			m.activePanel = originalPanel
		}
		return m, nil

	case "left", "h":
		// Move focus to the previous visible panel
		visiblePanels := []bool{m.showTaskList, m.showTaskDetails, m.showTimeline}
		originalPanel := m.activePanel

		// Find previous visible panel
		for i := 0; i < 3; i++ {
			m.activePanel = (m.activePanel + 2) % 3 // +2 is equivalent to -1 in modulo 3
			if visiblePanels[m.activePanel] {
				break
			}
		}

		// If no other panels are visible, keep the original panel
		if !visiblePanels[m.activePanel] {
			m.activePanel = originalPanel
		}
		return m, nil
	}
	return m, nil
}

// handleDetailViewKeys processes keyboard input in detail view
func (m *Model) handleDetailViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.viewMode = "list"
		return m, nil

	case "e":
		if m.activePanel == 0 {
			m.viewMode = "edit"
		}
		return m, nil

	case "d":
		if m.activePanel == 0 {
			return m, m.deleteCurrentTask()
		}
		return m, nil

	case "c":
		if m.activePanel == 0 {
			return m, m.toggleTaskCompletion()
		}
		return m, nil

	case "n":
		m.viewMode = "create"
		m.formStatus = string(task.StatusTodo)
		m.formPriority = string(task.PriorityLow)
		return m, nil

	case "up", "k":
		// reuse list scrolling logic
		return m.handleListViewKeys(msg)
	case "down", "j":
		return m.handleListViewKeys(msg)
	case "page-up", "ctrl+b":
		return m.handleListViewKeys(msg)
	case "page-down", "ctrl+f":
		return m.handleListViewKeys(msg)
	case "home", "g":
		return m.handleListViewKeys(msg)
	case "end", "G":
		return m.handleListViewKeys(msg)
	case "right", "l":
		return m.handleListViewKeys(msg)
	case "left", "h":
		return m.handleListViewKeys(msg)
	}
	return m, nil
}

// handleCreateFormKeys processes keyboard input in create form
func (m *Model) handleCreateFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.viewMode = "list"
		m.formTitle = ""
		m.formDescription = ""
		m.formPriority = ""
		m.formDueDate = ""
		m.formStatus = ""
		m.activeField = 0
		return m, nil

	case tea.KeyTab:
		m.activeField = (m.activeField + 1) % 5
		return m, nil

	case tea.KeyShiftTab:
		m.activeField = (m.activeField - 1) % 5
		if m.activeField < 0 {
			m.activeField = 4
		}
		return m, nil

	case tea.KeyEnter:
		if m.activeField == 4 {
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				return m, nil
			}
			return m, m.createNewTask()
		}
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	}

	switch m.activeField {
	case 0:
		return m.handleInputField(msg, &m.formTitle)
	case 1:
		return m.handleInputField(msg, &m.formDescription)
	case 2:
		if msg.String() == " " {
			switch m.formPriority {
			case string(task.PriorityLow):
				m.formPriority = string(task.PriorityMedium)
			case string(task.PriorityMedium):
				m.formPriority = string(task.PriorityHigh)
			default:
				m.formPriority = string(task.PriorityLow)
			}
		}
		return m, nil
	case 3:
		return m.handleDateField(msg, &m.formDueDate)
	}
	return m, nil
}

// handleInputField handles text input in a string field
func (m *Model) handleInputField(msg tea.KeyMsg, field *string) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		*field += string(msg.Runes)
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		if len(*field) > 0 {
			*field = (*field)[:len(*field)-1]
		}
		return m, nil
	case tea.KeyEsc:
		m.viewMode = "list"
		m.formTitle = ""
		m.formDescription = ""
		m.formPriority = ""
		m.formDueDate = ""
		m.formStatus = ""
		m.activeField = 0
		return m, nil
	case tea.KeyEnter:
		if m.activeField == 4 {
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				return m, nil
			}
			return m, m.createNewTask()
		}
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case tea.KeyTab:
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case tea.KeyShiftTab:
		m.activeField = (m.activeField - 1) % 5
		if m.activeField < 0 {
			m.activeField = 4
		}
		return m, nil
	}
	return m, nil
}

// handleDateField handles date input with basic validation
func (m *Model) handleDateField(msg tea.KeyMsg, field *string) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			if (r >= '0' && r <= '9') || r == '-' {
				*field += string(r)
			}
		}
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		if len(*field) > 0 {
			*field = (*field)[:len(*field)-1]
		}
		return m, nil
	case tea.KeyEsc:
		m.viewMode = "list"
		m.formTitle = ""
		m.formDescription = ""
		m.formPriority = ""
		m.formDueDate = ""
		m.formStatus = ""
		m.activeField = 0
		return m, nil
	case tea.KeyEnter:
		if m.activeField == 4 {
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				return m, nil
			}
			return m, m.createNewTask()
		}
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case tea.KeyTab:
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case tea.KeyShiftTab:
		m.activeField = (m.activeField - 1) % 5
		if m.activeField < 0 {
			m.activeField = 4
		}
		return m, nil
	}
	return m, nil
}

// handleEditViewKeys processes keyboard input in edit view
func (m *Model) handleEditViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = "list"
		return m, nil
	case "tab":
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case "shift+tab":
		m.activeField = (m.activeField - 1) % 5
		if m.activeField < 0 {
			m.activeField = 4
		}
		return m, nil
	case "enter":
		if m.activeField == 4 {
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				return m, nil
			}
			return m, m.createNewTask()
		}
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	}

	switch m.activeField {
	case 0:
		return m.handleInputField(msg, &m.formTitle)
	case 1:
		return m.handleInputField(msg, &m.formDescription)
	case 2:
		if msg.String() == " " {
			switch m.formPriority {
			case string(task.PriorityLow):
				m.formPriority = string(task.PriorityMedium)
			case string(task.PriorityMedium):
				m.formPriority = string(task.PriorityHigh)
			default:
				m.formPriority = string(task.PriorityLow)
			}
		}
		return m, nil
	case 3:
		return m.handleDateField(msg, &m.formDueDate)
	}
	return m, nil
}

// setStatusMessage sets a status message with a type and expiry duration
func (m *Model) setStatusMessage(msg, msgType string, duration time.Duration) {
	m.statusMessage = msg
	m.statusType = msgType
	m.statusExpiry = time.Now().Add(duration)
}

// setSuccessStatus is a helper to set success status messages
func (m *Model) setSuccessStatus(msg string) {
	m.setStatusMessage(msg, "success", 5*time.Second)
}

// setErrorStatus is a helper to set error status messages
func (m *Model) setErrorStatus(msg string) {
	m.setStatusMessage(msg, "error", 10*time.Second)
}

// setInfoStatus is a helper to set informational status messages
func (m *Model) setInfoStatus(msg string) {
	m.setStatusMessage(msg, "info", 3*time.Second)
}

// setLoadingStatus sets the app in loading state with a message
func (m *Model) setLoadingStatus(msg string) {
	m.setStatusMessage(msg, "loading", 30*time.Second)
	m.isLoading = true
}

// clearLoadingStatus clears the loading state
func (m *Model) clearLoadingStatus() {
	m.isLoading = false
	if m.statusType == "loading" {
		m.statusMessage = ""
		m.statusType = ""
	}
}

// Update refreshTasks to categorize tasks
func (m *Model) refreshTasks() tea.Cmd {
	m.setLoadingStatus("Loading tasks...")
	return func() tea.Msg {
		tasks, err := m.taskSvc.List(m.ctx, m.userID)
		if err != nil {
			return messages.ErrorMsg(fmt.Errorf("failed to refresh tasks: %v", err))
		}
		m.categorizeTasks(tasks)
		return messages.TasksRefreshedMsg{Tasks: tasks}
	}
}

// Update toggleTaskCompletion to ensure synchronization between visualCursor and taskCursor
func (m *Model) toggleTaskCompletion() tea.Cmd {
	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) {
		return nil
	}

	// Get current task and its ID
	curr := m.tasks[m.cursor]
	toggledID := curr.ID

	// Determine new status
	var newStatus task.Status
	if curr.Status != task.StatusDone {
		newStatus = task.StatusDone
	} else {
		newStatus = task.StatusTodo
	}

	// Update local task
	m.tasks[m.cursor].Status = newStatus
	m.tasks[m.cursor].IsCompleted = (newStatus == task.StatusDone)

	// Re-categorize tasks
	m.categorizeTasks(m.tasks)

	// Update visual cursor and task cursor
	m.updateVisualCursorFromTaskCursor()
	m.updateTaskCursorFromVisualCursor()

	// Call server to update
	return func() tea.Msg {
		updatedTask, err := m.taskSvc.ChangeStatus(m.ctx, int64(toggledID), newStatus)
		if err != nil {
			return messages.StatusUpdateErrorMsg{TaskIndex: m.cursor, TaskTitle: curr.Title, Err: err}
		}
		return messages.StatusUpdateSuccessMsg{Task: updatedTask, Message: "Task status updated successfully"}
	}
}

// deleteCurrentTask deletes the currently selected task
func (m *Model) deleteCurrentTask() tea.Cmd {
	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) {
		return nil
	}
	taskTitle := m.tasks[m.cursor].Title
	taskID := int64(m.tasks[m.cursor].ID)
	taskIndex := m.cursor
	m.setLoadingStatus("Deleting task...")
	return func() tea.Msg {
		err := m.taskSvc.Delete(m.ctx, taskID)
		if err != nil {
			return messages.StatusUpdateErrorMsg{TaskIndex: taskIndex, TaskTitle: taskTitle, Err: err}
		}
		tasks, err2 := m.taskSvc.List(m.ctx, m.userID)
		if err2 != nil {
			return messages.ErrorMsg(fmt.Errorf("task deleted but failed to refresh: %v", err2))
		}
		m.viewMode = "list"
		return messages.TasksRefreshedMsg{Tasks: tasks}
	}
}

// createNewTask creates a new task from the form fields
func (m *Model) createNewTask() tea.Cmd {
	if m.formTitle == "" {
		m.err = fmt.Errorf("title is required")
		m.setErrorStatus("Title is required")
		return nil
	}
	m.setLoadingStatus("Creating new task...")
	var dueDate *time.Time
	if m.formDueDate != "" {
		date, err := time.Parse("2006-01-02", m.formDueDate)
		if err == nil {
			dueDate = &date
		} else {
			m.setErrorStatus(fmt.Sprintf("Invalid date format: %v", err))
			m.err = fmt.Errorf("invalid date format: %v", err)
			return nil
		}
	}
	priority := task.PriorityLow
	if m.formPriority == string(task.PriorityMedium) {
		priority = task.PriorityMedium
	} else if m.formPriority == string(task.PriorityHigh) {
		priority = task.PriorityHigh
	}
	title := m.formTitle
	description := m.formDescription
	m.formTitle = ""
	m.formDescription = ""
	m.formPriority = ""
	m.formDueDate = ""
	m.formStatus = ""
	m.activeField = 0
	m.viewMode = "list"
	return func() tea.Msg {
		_, err := m.taskSvc.Create(m.ctx, m.userID, nil, title, description, dueDate, priority, []string{})
		if err != nil {
			return messages.ErrorMsg(fmt.Errorf("failed to create task: %v", err))
		}
		tasks, err2 := m.taskSvc.List(m.ctx, m.userID)
		if err2 != nil {
			return messages.ErrorMsg(fmt.Errorf("task created but failed to refresh: %v", err2))
		}
		return messages.TasksRefreshedMsg{Tasks: tasks}
	}
}

// getTasksByTimeCategory organizes tasks into overdue, today, and upcoming categories
func (m *Model) getTasksByTimeCategory() ([]task.Task, []task.Task, []task.Task) {
	var overdue, todayTasks, upcoming []task.Task
	now := time.Now()
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := todayDate.AddDate(0, 0, 1)
	for _, t := range m.tasks {
		if t.DueDate == nil {
			continue
		}
		dueDate := *t.DueDate
		if dueDate.Before(todayDate) {
			overdue = append(overdue, t)
		} else if dueDate.Before(tomorrow) {
			todayTasks = append(todayTasks, t)
		} else {
			upcoming = append(upcoming, t)
		}
	}
	return overdue, todayTasks, upcoming
}

// Utility functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

// View implements tea.Model View, composing header, panels, and footer.
func (m *Model) View() string {
	sharedStyles := &shared.Styles{
		Title:          m.styles.Title,
		SelectedItem:   m.styles.SelectedItem,
		Help:           m.styles.Help,
		ActiveBorder:   m.styles.ActiveBorder,
		Todo:           m.styles.Todo,
		InProgress:     m.styles.InProgress,
		Done:           m.styles.Done,
		LowPriority:    m.styles.LowPriority,
		MediumPriority: m.styles.MediumPriority,
		HighPriority:   m.styles.HighPriority,
	}

	// Initialize collapsible sections if needed
	if m.collapsibleManager == nil {
		m.initCollapsibleSections()
	}

	// Render header
	header := layout.RenderHeader(layout.HeaderProps{
		Width:         m.width,
		CurrentTime:   m.currentTime,
		StatusMessage: m.statusMessage,
		StatusType:    m.statusType,
		IsLoading:     m.isLoading,
	})

	if m.viewMode == "create" {
		createForm := panels.RenderCreateForm(panels.CreateFormProps{
			FormTitle:       m.formTitle,
			FormDescription: m.formDescription,
			FormPriority:    m.formPriority,
			FormDueDate:     m.formDueDate,
			ActiveField:     m.activeField,
			Error:           m.err,
			Styles:          sharedStyles,
		})
		return lipgloss.JoinVertical(lipgloss.Left, header, createForm)
	}

	// Default multi-panel view
	var visiblePanelCount int
	const headerHeight = 5
	const headerGap = 0
	const footerHeight = 1
	const footerGap = 0
	const totalOffset = headerHeight + headerGap + footerHeight + footerGap
	panelHeight := m.height - totalOffset
	if m.showTaskList {
		visiblePanelCount++
	}
	if m.showTaskDetails {
		visiblePanelCount++
	}
	if m.showTimeline {
		visiblePanelCount++
	}
	availableWidth := m.width
	columnWidth := availableWidth / max(1, visiblePanelCount)
	var columns []string

	// Task List Panel
	if m.showTaskList {
		contentWidth := columnWidth - 2
		list := panels.RenderTaskList(panels.TaskListProps{
			Tasks:          m.tasks,
			Cursor:         m.cursor,
			VisualCursor:   m.visualCursor,
			Offset:         m.taskListOffset,
			Width:          contentWidth,
			Height:         panelHeight - 2,
			Styles:         sharedStyles,
			IsActive:       m.activePanel == 0,
			Error:          m.err,
			SuccessMsg:     m.successMsg,
			ClearSuccess:   func() { m.successMsg = "" },
			CursorOnHeader: m.cursorOnHeader,
			CollapsibleMgr: m.collapsibleManager,
		})
		wrapped := shared.RenderPanel(shared.PanelProps{
			Content:     list,
			Width:       columnWidth,
			Height:      panelHeight,
			IsActive:    m.activePanel == 0,
			BorderColor: shared.ColorBorder,
		})
		columns = append(columns, wrapped)
	}

	// Task Details Panel
	if m.showTaskDetails {
		contentWidth := columnWidth - 2
		details := panels.RenderTaskDetails(panels.TaskDetailsProps{
			Tasks:          m.tasks,
			Cursor:         m.cursor,
			Offset:         m.taskDetailsOffset,
			Width:          contentWidth,
			Height:         panelHeight - 2,
			Styles:         sharedStyles,
			IsActive:       m.activePanel == 1,
			CursorOnHeader: m.cursorOnHeader,
		})
		wrapped := shared.RenderPanel(shared.PanelProps{
			Content:     details,
			Width:       columnWidth,
			Height:      panelHeight,
			IsActive:    m.activePanel == 1,
			BorderColor: shared.ColorBorder,
		})
		columns = append(columns, wrapped)
	}

	// Timeline Panel
	if m.showTimeline {
		contentWidth := columnWidth - 2
		timeline := panels.RenderTimeline(panels.TimelineProps{
			Tasks:    m.tasks,
			Offset:   m.timelineOffset,
			Width:    contentWidth,
			Height:   panelHeight - 2,
			Styles:   sharedStyles,
			IsActive: m.activePanel == 2,
		})
		wrapped := shared.RenderPanel(shared.PanelProps{
			Content:     timeline,
			Width:       columnWidth,
			Height:      panelHeight,
			IsActive:    m.activePanel == 2,
			BorderColor: shared.ColorBorder,
		})
		columns = append(columns, wrapped)
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	footer := layout.RenderFooter(layout.FooterProps{
		Width:          m.width,
		ViewMode:       m.viewMode,
		HelpStyle:      m.styles.Help,
		CursorOnHeader: m.cursorOnHeader,
	})

	sections := []string{header, content, footer}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
