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
}

// NewModel initializes the bubbletea application model.
func NewModel(ctx context.Context, svc taskService.Service, userID int64) *Model {
	roots, err := svc.List(ctx, userID)

	m := &Model{
		ctx:             ctx,
		tasks:           roots,
		cursor:          0,
		err:             err,
		taskSvc:         svc,
		userID:          userID,
		viewMode:        "list",
		styles:          styles.ActiveStyles,
		showTaskList:    true,
		showTaskDetails: true,
		showTimeline:    true,
		activePanel:     0,
	}

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
		for i := range m.tasks {
			if m.tasks[i].ID == msg.Task.ID {
				m.tasks[i] = msg.Task
				break
			}
		}
		return m, nil

	case messages.TasksRefreshedMsg:
		m.tasks = msg.Tasks
		if m.cursor >= len(m.tasks) && len(m.tasks) > 0 {
			m.cursor = len(m.tasks) - 1
		}
		m.clearLoadingStatus()
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

// handleListViewKeys processes keyboard input in list view
func (m *Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		switch m.activePanel {
		case 0:
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
		case 1:
			// For Task Details panel, scroll more efficiently
			m.taskDetailsOffset = max(0, m.taskDetailsOffset-3)
		case 2:
			// For Timeline panel, scroll more efficiently
			m.timelineOffset = max(0, m.timelineOffset-3)
		}
		return m, nil

	case "down", "j":
		switch m.activePanel {
		case 0:
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
		case 1:
			// For Task Details panel, scroll more efficiently
			// Calculate a rough estimate of content length
			maxOffset := 15
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].Description != nil {
				maxOffset += len(*m.tasks[m.cursor].Description) / 30
			}
			m.taskDetailsOffset = min(m.taskDetailsOffset+3, maxOffset)
		case 2:
			// For Timeline panel, scroll more efficiently
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOffset := len(overdue) + len(today) + len(upcoming) + 10
			m.timelineOffset = min(m.timelineOffset+3, maxOffset)
		}
		return m, nil

	// Additional key bindings
	case "page-up", "ctrl+b":
		pageSize := 10
		switch m.activePanel {
		case 0:
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
		case 1:
			m.taskDetailsOffset -= pageSize
			if m.taskDetailsOffset < 0 {
				m.taskDetailsOffset = 0
			}
		case 2:
			m.timelineOffset -= pageSize
			if m.timelineOffset < 0 {
				m.timelineOffset = 0
			}
		}
		return m, nil

	case "page-down", "ctrl+f":
		pageSize := 10
		var maxOffset int
		switch m.activePanel {
		case 0:
			maxOffset = max(0, len(m.tasks)-pageSize)
			m.taskListOffset += pageSize
			if m.taskListOffset > maxOffset {
				m.taskListOffset = maxOffset
			}
			if m.cursor < m.taskListOffset {
				m.cursor = m.taskListOffset
			}
		case 1:
			maxOffset = 15
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].Description != nil {
				maxOffset += len(*m.tasks[m.cursor].Description) / 30
			}
			maxOffset = max(0, maxOffset-pageSize)
			m.taskDetailsOffset += pageSize
			if m.taskDetailsOffset > maxOffset {
				m.taskDetailsOffset = maxOffset
			}
		case 2:
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOffset = len(overdue) + len(today) + len(upcoming) - pageSize
			maxOffset = max(0, maxOffset)
			m.timelineOffset += pageSize
			if m.timelineOffset > maxOffset {
				m.timelineOffset = maxOffset
			}
		}
		return m, nil

	case "home", "g":
		switch m.activePanel {
		case 0:
			m.taskListOffset = 0
			m.cursor = 0
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
			if len(m.tasks) > 0 {
				m.cursor = len(m.tasks) - 1
				m.taskListOffset = max(0, m.cursor-pageSize+1)
			}
		case 1:
			maxOff := 15
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].Description != nil {
				maxOff += len(*m.tasks[m.cursor].Description) / 30
			}
			m.taskDetailsOffset = max(0, maxOff-pageSize)
		case 2:
			overdue, today, upcoming := m.getTasksByTimeCategory()
			m.timelineOffset = max(0, len(overdue)+len(today)+len(upcoming)-pageSize)
		}
		return m, nil

	case "right", "l":
		visible := []bool{m.showTaskList, m.showTaskDetails, m.showTimeline}
		orig := m.activePanel
		for i := 0; i < 3; i++ {
			m.activePanel = (m.activePanel + 1) % 3
			if visible[m.activePanel] {
				break
			}
		}
		if !visible[m.activePanel] {
			m.activePanel = orig
		}
		return m, nil

	case "left", "h":
		visible := []bool{m.showTaskList, m.showTaskDetails, m.showTimeline}
		orig := m.activePanel
		for i := 0; i < 3; i++ {
			m.activePanel = (m.activePanel + 2) % 3
			if visible[m.activePanel] {
				break
			}
		}
		if !visible[m.activePanel] {
			m.activePanel = orig
		}
		return m, nil

	case "enter":
		if m.activePanel == 0 && len(m.tasks) > 0 {
			m.viewMode = "detail"
			return m, nil
		}

	case "c", " ":
		if m.activePanel == 0 && len(m.tasks) > 0 {
			return m, m.toggleTaskCompletion()
		}
		return m, nil

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

// refreshTasks reloads the task list from the service
func (m *Model) refreshTasks() tea.Cmd {
	m.setLoadingStatus("Loading tasks...")
	return func() tea.Msg {
		tasks, err := m.taskSvc.List(m.ctx, m.userID)
		if err != nil {
			return messages.ErrorMsg(fmt.Errorf("failed to refresh tasks: %v", err))
		}
		return messages.TasksRefreshedMsg{Tasks: tasks}
	}
}

// toggleTaskCompletion toggles the completion status of the selected task
func (m *Model) toggleTaskCompletion() tea.Cmd {
	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) {
		return nil
	}
	currentTask := m.tasks[m.cursor]
	taskID := int64(currentTask.ID)
	taskTitle := currentTask.Title
	var newStatus task.Status
	if currentTask.Status == task.StatusDone {
		newStatus = task.StatusTodo
		m.tasks[m.cursor].Status = newStatus
		m.tasks[m.cursor].IsCompleted = false
		m.setSuccessStatus(fmt.Sprintf("Task '%s' marked as TODO", taskTitle))
	} else {
		newStatus = task.StatusDone
		m.tasks[m.cursor].Status = newStatus
		m.tasks[m.cursor].IsCompleted = true
		m.setSuccessStatus(fmt.Sprintf("Task '%s' marked as DONE", taskTitle))
	}
	m.setInfoStatus("Saving changes...")
	return func() tea.Msg {
		updatedTask, err := m.taskSvc.ChangeStatus(m.ctx, taskID, newStatus)
		if err != nil {
			return messages.StatusUpdateErrorMsg{TaskIndex: m.cursor, TaskTitle: taskTitle, Err: err}
		}
		return messages.StatusUpdateSuccessMsg{Task: updatedTask, Message: fmt.Sprintf("Task '%s' status updated successfully", taskTitle)}
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
			Tasks:        m.tasks,
			Cursor:       m.cursor,
			Offset:       m.taskListOffset,
			Width:        contentWidth,
			Height:       panelHeight - 2,
			Styles:       sharedStyles,
			IsActive:     m.activePanel == 0,
			Error:        m.err,
			SuccessMsg:   m.successMsg,
			ClearSuccess: func() { m.successMsg = "" },
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
			Tasks:    m.tasks,
			Cursor:   m.cursor,
			Offset:   m.taskDetailsOffset,
			Width:    contentWidth,
			Height:   panelHeight - 2,
			Styles:   sharedStyles,
			IsActive: m.activePanel == 1,
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
		Width:     m.width,
		ViewMode:  m.viewMode,
		HelpStyle: m.styles.Help,
	})

	sections := []string{header, content, footer}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
