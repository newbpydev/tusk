# BubbleTea TUI Implementation Guide

This guide provides recommendations for implementing and refactoring the Tusk Terminal User Interface (TUI) using the BubbleTea framework. It focuses on improving code organization, separation of concerns, and following best practices for terminal applications.

## Architecture Overview

The TUI implementation should follow the hexagonal architecture pattern used in the rest of the Tusk application:

1. **Core Domain**: Task and User models remain independent of UI concerns
2. **Ports**: TUI components interact with core via service interfaces
3. **Adapter**: BubbleTea implementation serves as adapter for terminal UI
4. **Services**: Orchestrate user actions between UI and business logic

## Component Structure

Organize the BubbleTea implementation into the following component hierarchy:

```
internal/adapters/tui/bubbletea/
├── app/                 # Main application model and orchestration
│   ├── state/           # State management for different functional areas
│   ├── commands/        # Command pattern implementations
│   └── events/          # Event handling and propagation
├── components/          # Reusable UI components
│   ├── input/           # Input components (text fields, toggles, etc.)
│   ├── layout/          # Layout components (header, footer, etc.)
│   ├── list/            # List-based components
│   └── panels/          # Panel implementations
├── hooks/               # Reusable state management hooks
├── messages/            # Message type definitions
├── styles/              # UI styling definitions
└── views/               # Complete view implementations
```

## Model Organization

Break down the monolithic Model struct into smaller, focused sub-models:

```go
// app/model.go
type Model struct {
    // Core application state
    ctx         context.Context
    ready       bool

    // Feature-specific state containers
    tasks       *state.TaskState
    navigation  *state.NavigationState
    form        *state.FormState
    ui          *state.UIState
    timeline    *state.TimelineState
    status      *state.StatusState

    // Services
    taskSvc     taskService.Service
}

// app/state/task_state.go
type TaskState struct {
    allTasks           []task.Task
    filteredTasks      []task.Task
    taskMap            map[int]task.Task
    selectedTaskID     int

    collapsibleManager *hooks.CollapsibleManager
    filter             TaskFilter
    sorter             TaskSorter
}

// app/state/navigation_state.go
type NavigationState struct {
    activePanel       int
    previousPanel     int
    cursorOnHeader    bool
    visualCursor      int
    cursor            int

    // Panel-specific scroll positions
    panelOffsets      map[int]int
}

// app/state/ui_state.go
type UIState struct {
    viewMode     ViewMode
    panelStates  map[int]bool  // Map of panel ID to visibility
    width        int
    height       int
    styles       *styles.Styles
    theme        *styles.Theme
}
```

## Command Pattern Implementation

Use the command pattern for handling user interactions:

```go
// app/commands/command.go
type Command func(m *Model) (tea.Model, tea.Cmd)

// app/handlers.go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if cmd := m.getCommand(msg); cmd != nil {
            return cmd(&m)
        }
    // ...handle other message types
    }
    return m, nil
}

func (m *Model) getCommand(msg tea.KeyMsg) Command {
    // Global commands available in all contexts
    globalCommands := map[string]Command{
        "q":      commands.Quit,
        "ctrl+c": commands.Quit,
        "?":      commands.ShowHelp,
    }

    if cmd, ok := globalCommands[msg.String()]; ok {
        return cmd
    }

    // Context-specific commands based on active panel and view mode
    contextCommands := m.getContextCommands()
    key := msg.String()

    if cmd, ok := contextCommands[key]; ok {
        return cmd
    }

    return nil
}
```

## Panel Interface

Implement a consistent interface for all panels:

```go
// components/shared/panel.go
type Panel interface {
    // Init initializes the panel and returns any commands to run
    Init() tea.Cmd

    // Update processes messages and returns updated panel and commands
    Update(msg tea.Msg) (Panel, tea.Cmd)

    // View renders the panel as a string
    View() string

    // SetSize updates the panel's dimensions
    SetSize(width, height int)

    // SetActive updates the panel's active state
    SetActive(active bool)

    // IsActive returns whether the panel is currently active
    IsActive() bool

    // HandleKey processes keyboard input specific to this panel
    HandleKey(msg tea.KeyMsg) (tea.Cmd, bool)
}

// components/shared/base_panel.go
type BasePanel struct {
    width      int
    height     int
    isActive   bool
    styles     *styles.Styles
}

func (b *BasePanel) SetSize(width, height int) {
    b.width = width
    b.height = height
}

func (b *BasePanel) SetActive(active bool) {
    b.isActive = active
}

func (b *BasePanel) IsActive() bool {
    return b.isActive
}
```

## State Management

Implement clean state management for complex UI elements:

```go
// hooks/collapsible.go
type CollapsibleManager struct {
    sections       map[string]*Section
    sectionOrder   []string
}

type Section struct {
    ID            string
    Title         string
    Expanded      bool
    ItemCount     int
    StartIndex    int
}

func (cm *CollapsibleManager) AddSection(id, title string, itemCount int) {
    // Add a new collapsible section
}

func (cm *CollapsibleManager) ToggleSection(id string) bool {
    // Toggle a section's expanded state
}

func (cm *CollapsibleManager) GetVisibleItemCount() int {
    // Calculate the total visible items (considering expanded state)
}

func (cm *CollapsibleManager) GetItemIndexForCursor(cursor int) (sectionID string, itemIndex int) {
    // Convert cursor position to section and item index
}
```

## Message System

Define clear message types for communication between components:

```go
// messages/messages.go

// ErrorMsg represents an error that occurred during operation
type ErrorMsg struct {
    Err error
}

// StatusMsg represents a status update to display
type StatusMsg struct {
    Message string
    Timeout time.Duration
    Type    StatusType // Info, Success, Warning, Error
}

// TaskMsg carries task-related data between components
type TaskMsg struct {
    Task     task.Task
    Action   TaskAction // Created, Updated, Deleted, Completed
}

// TasksLoadedMsg indicates tasks have been loaded
type TasksLoadedMsg struct {
    Tasks []task.Task
}

// CursorMsg requests cursor movement
type CursorMsg struct {
    Direction CursorDirection
    Amount    int
}

// ViewChangeMsg requests a change in view mode
type ViewChangeMsg struct {
    ViewMode ViewMode
}
```

## View Management

Implement a view registry for managing different application views:

```go
// app/view_registry.go
type ViewMode string

const (
    ViewModeList     ViewMode = "list"
    ViewModeKanban   ViewMode = "kanban"
    ViewModeCalendar ViewMode = "calendar"
    ViewModeSettings ViewMode = "settings"
)

type ViewRegistry struct {
    views map[ViewMode]View
}

type View interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (tea.Cmd, bool)
    View() string
}

func NewViewRegistry() *ViewRegistry {
    return &ViewRegistry{
        views: make(map[ViewMode]View),
    }
}

func (vr *ViewRegistry) RegisterView(mode ViewMode, view View) {
    vr.views[mode] = view
}

func (vr *ViewRegistry) GetView(mode ViewMode) (View, bool) {
    view, ok := vr.views[mode]
    return view, ok
}
```

## Style Management

Centralize styling with a theme system:

```go
// styles/styles.go
type Theme struct {
    // Colors
    TextColor         lipgloss.Color
    AccentColor       lipgloss.Color
    HighlightColor    lipgloss.Color
    ErrorColor        lipgloss.Color
    SuccessColor      lipgloss.Color
    WarningColor      lipgloss.Color
    InfoColor         lipgloss.Color

    // Border styles
    BorderNormal      lipgloss.Border
    BorderActive      lipgloss.Border

    // Priority colors
    PriorityHigh      lipgloss.Color
    PriorityMedium    lipgloss.Color
    PriorityLow       lipgloss.Color

    // Task status colors
    CompletedTask     lipgloss.Color
    OverdueTask       lipgloss.Color
}

type Styles struct {
    theme *Theme

    // Layout styles
    App              lipgloss.Style
    Header           lipgloss.Style
    Footer           lipgloss.Style

    // Panel styles
    Panel            lipgloss.Style
    PanelActive      lipgloss.Style
    PanelTitle       lipgloss.Style
    PanelTitleActive lipgloss.Style

    // Text styles
    TextNormal       lipgloss.Style
    TextHighlight    lipgloss.Style
    TextDimmed       lipgloss.Style

    // Task styles
    TaskTitle        lipgloss.Style
    TaskCompleted    lipgloss.Style
    TaskOverdue      lipgloss.Style
    TaskDueToday     lipgloss.Style

    // Custom component styles
    // ...
}

func NewDefaultStyles() *Styles {
    theme := &Theme{
        TextColor:      lipgloss.Color("#FFFFFF"),
        AccentColor:    lipgloss.Color("#7D56F4"),
        // ...more theme settings
    }

    return NewStylesWithTheme(theme)
}

func NewStylesWithTheme(theme *Theme) *Styles {
    s := &Styles{
        theme: theme,
    }

    // Initialize all styles based on theme
    s.App = lipgloss.NewStyle().
        Background(lipgloss.Color("#333333"))

    s.Panel = lipgloss.NewStyle().
        Border(theme.BorderNormal).
        BorderForeground(theme.TextColor)

    s.PanelActive = s.Panel.Copy().
        Border(theme.BorderActive).
        BorderForeground(theme.AccentColor)

    // ...initialize other styles

    return s
}
```

## Layout Management

Create a responsive layout system:

```go
// components/layout/main_layout.go
type MainLayout struct {
    width    int
    height   int
    styles   *styles.Styles

    header   *Header
    footer   *Footer
    panels   []shared.Panel
}

func (l *MainLayout) Init() tea.Cmd {
    var cmds []tea.Cmd

    // Initialize all components
    if cmd := l.header.Init(); cmd != nil {
        cmds = append(cmds, cmd)
    }

    for _, panel := range l.panels {
        if cmd := panel.Init(); cmd != nil {
            cmds = append(cmds, cmd)
        }
    }

    if cmd := l.footer.Init(); cmd != nil {
        cmds = append(cmds, cmd)
    }

    return tea.Batch(cmds...)
}

func (l *MainLayout) Update(msg tea.Msg) (tea.Cmd, bool) {
    // Update components and collect commands
    // ...
}

func (l *MainLayout) View() string {
    // Render the complete layout
    headerView := l.header.View()

    // Calculate panel dimensions based on layout
    // ...

    // Render panels
    // ...

    footerView := l.footer.View()

    // Combine all views
    return lipgloss.JoinVertical(
        lipgloss.Left,
        headerView,
        panelsView,
        footerView,
    )
}

func (l *MainLayout) SetSize(width, height int) {
    l.width = width
    l.height = height

    // Update component sizes
    headerHeight := 1
    footerHeight := 1
    panelHeight := height - headerHeight - footerHeight

    l.header.SetSize(width, headerHeight)
    l.footer.SetSize(width, footerHeight)

    // Calculate panel dimensions
    // ...
}
```

## Error Handling

Implement consistent error handling:

```go
// app/error_handling.go
func handleError(err error) tea.Cmd {
    return func() tea.Msg {
        return messages.ErrorMsg{Err: err}
    }
}

func showErrorStatus(msg string) tea.Cmd {
    return func() tea.Msg {
        return messages.StatusMsg{
            Message: msg,
            Type:    messages.StatusError,
            Timeout: 5 * time.Second,
        }
    }
}

// app/update.go
func (m Model) handleErrorMsg(msg messages.ErrorMsg) (Model, tea.Cmd) {
    // Log the error
    log.Printf("Error: %v", msg.Err)

    // Update status
    statusMsg := messages.StatusMsg{
        Message: fmt.Sprintf("Error: %v", msg.Err),
        Type:    messages.StatusError,
        Timeout: 5 * time.Second,
    }

    newStatus, cmd := m.status.Update(statusMsg)
    m.status = newStatus

    return m, cmd
}
```

## Form Management

Create a robust form management system:

```go
// components/input/form.go
type Field struct {
    ID          string
    Label       string
    Value       string
    Placeholder string
    Required    bool
    Validator   func(string) error

    // UI state
    Focused     bool
    Error       string
}

type Form struct {
    fields      []*Field
    activeField int
    validated   bool
    submitted   bool

    styles      *styles.Styles
    width       int
    height      int
}

func (f *Form) AddField(field *Field) {
    f.fields = append(f.fields, field)
}

func (f *Form) NextField() {
    f.activeField = (f.activeField + 1) % len(f.fields)
    f.fields[f.activeField].Focused = true
}

func (f *Form) PrevField() {
    f.activeField = (f.activeField - 1 + len(f.fields)) % len(f.fields)
    f.fields[f.activeField].Focused = true
}

func (f *Form) Validate() bool {
    valid := true

    for _, field := range f.fields {
        if field.Required && field.Value == "" {
            field.Error = "This field is required"
            valid = false
            continue
        }

        if field.Validator != nil {
            if err := field.Validator(field.Value); err != nil {
                field.Error = err.Error()
                valid = false
            } else {
                field.Error = ""
            }
        }
    }

    f.validated = true
    return valid
}

func (f *Form) GetValues() map[string]string {
    values := make(map[string]string)
    for _, field := range f.fields {
        values[field.ID] = field.Value
    }
    return values
}

func (f *Form) HandleKeyPress(msg tea.KeyMsg) (tea.Cmd, bool) {
    // Handle form navigation and field editing
    // ...
}
```

## Testing Strategy

Implement a comprehensive testing approach:

1. **Unit Testing**:

   - Test individual components in isolation
   - Test state transitions for each component
   - Test key handlers and commands

2. **Integration Testing**:

   - Test interaction between components
   - Test complete workflows (e.g., task creation)
   - Test navigation and cursor management

3. **View Testing**:

   - Implement snapshot testing for view rendering
   - Test responsive layout behavior

4. **End-to-End Testing**:
   - Test complete application workflows
   - Use mocked services for deterministic testing

## Performance Considerations

Implement performance optimizations:

1. **Viewport Management**:

   - Only render visible content
   - Implement proper scrolling for large datasets
   - Use pagination when appropriate

2. **Rendering Optimization**:

   - Cache rendered content when possible
   - Only re-render changed components
   - Use lipgloss efficiently to minimize string operations

3. **Async Operations**:
   - Use tea.Cmds for asynchronous operations
   - Implement proper loading states
   - Use optimistic UI updates for better responsiveness

## Implementation Strategy

Follow these steps when implementing or refactoring TUI components:

1. **Start with interfaces**:

   - Define clear component interfaces
   - Create model structs with explicit state

2. **Implement core logic**:

   - Focus on state management first
   - Implement update logic with clear state transitions
   - Write tests for state transitions

3. **Add rendering**:

   - Implement View() methods
   - Focus on layout and structure before styling
   - Use consistent styling patterns

4. **Connect components**:
   - Implement message passing between components
   - Create composition hierarchy
   - Test integration points

The goal is to create a maintainable, testable, and performant TUI implementation that follows the principles of separation of concerns and clean architecture.
