# TUI Refactoring Guide

This guide provides recommendations for refactoring the Terminal User Interface (TUI) implementation to improve maintainability, readability, and adherence to best practices.

## 1. Code Organization Improvements

### 1.1 File Structure Reorganization

- Reorganize the code in `app/` directory to separate concerns more clearly:
  - Split `model.go` into smaller files focused on specific functionality
  - Create `timeline.go` for timeline-specific state management
  - Move view-rendering helper functions to appropriate files
  - Extract all panel initialization logic into a dedicated `panels.go`

### 1.2 Component Extraction

- Extract reusable components from panel implementations:
  - Create a `widgets/` directory for small, reusable UI elements
  - Implement proper component interfaces with `Init()`, `Update()`, and `View()` methods
  - Extract duplicated rendering logic into shared utility functions

### 1.3 Message Passing Cleanup

- Consolidate message types in `messages/messages.go`:
  - Group related messages into logical categories
  - Document each message type's purpose and payload
  - Ensure consistent naming conventions for all messages

## 2. State Management Improvements

### 2.1 Model Refactoring

- Break down the monolithic `Model` struct:
  - Create smaller, focused sub-models for each major feature area
  - Implement proper state containers for form, navigation, and task management
  - Use composition to combine these into the main application model

```go
// Example state organization
type Model struct {
    // Core application state
    ctx         context.Context
    width       int
    height      int

    // Feature-specific state containers
    tasks       TaskState
    navigation  NavigationState
    form        FormState
    ui          UIState
    timeline    TimelineState
    status      StatusState

    // Services
    taskSvc     taskService.Service
}

type NavigationState struct {
    activePanel       int
    cursorOnHeader    bool
    visualCursor      int
    cursor            int
    // Panel-specific scroll positions
    taskListOffset    int
    taskDetailsOffset int
    timelineOffset    int
}

type UIState struct {
    viewMode          string
    showTaskList      bool
    showTaskDetails   bool
    showTimeline      bool
    styles            *styles.Styles
}
```

### 2.2 Cursor Management

- Simplify cursor management logic:
  - Create a dedicated cursor manager to handle all cursor operations
  - Implement navigation logic that's aware of the collapsible sections
  - Add validation to ensure cursor positions are always valid
  - Reduce duplication of cursor positioning logic

### 2.3 View Mode Management

- Implement a proper state machine for view modes:
  - Define explicit transitions between modes
  - Document valid state transitions
  - Add validation for state transitions

## 3. Handler Logic Improvements

### 3.1 Event Handler Decomposition

- Decompose large handler functions:
  - Refactor `handleKeyPress()` using the command pattern
  - Map keys to commands instead of directly implementing behavior
  - Isolate side effects from pure state transitions

```go
// Example command pattern implementation
type Command func(m *Model) (tea.Model, tea.Cmd)

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // Get command for the key in the current context
    cmd := m.getCommandForKey(m.viewMode, m.activePanel, msg)
    if cmd != nil {
        return cmd(m)
    }
    return m, nil
}

func (m *Model) getCommandForKey(viewMode string, activePanel int, msg tea.KeyMsg) Command {
    commands := map[string]map[int]map[string]Command{
        "list": {
            0: { // Task list panel
                "j":         m.navigateDown,
                "down":      m.navigateDown,
                "k":         m.navigateUp,
                "up":        m.navigateUp,
                "tab":       m.nextPanel,
                "space":     m.toggleTaskCompletion,
                // ...
            },
            // Other panels...
        },
        // Other view modes...
    }

    if panelCommands, ok := commands[viewMode][activePanel]; ok {
        if command, ok := panelCommands[msg.String()]; ok {
            return command
        }
    }
    return nil
}
```

### 3.2 Update Logic Separation

- Separate different types of updates:
  - Move business logic updates to service layer functions
  - Keep UI state updates in the TUI layer
  - Replace long switch statements with map-based dispatch

## 4. Rendering Improvements

### 4.1 View Logic Decomposition

- Decompose rendering functions using render props pattern:
  - Create dedicated render props for each component
  - Pass only necessary data to rendering functions
  - Ensure rendering functions are pure and don't modify state

### 4.2 Style Management

- Centralize styling logic:
  - Move all style definitions to a theme system
  - Support multiple themes with consistent application
  - Implement helper functions for common styling patterns

### 4.3 Layout Management

- Improve layout calculations:
  - Create dedicated layout managers for different view modes
  - Implement responsive layouts that adapt to terminal size
  - Extract layout calculations from render functions

## 5. Error Handling and Validation

### 5.1 Consistent Error Handling

- Implement consistent error handling patterns:
  - Define error types for different categories of errors
  - Use structured error messages with context
  - Implement proper error recovery in the UI

### 5.2 Input Validation

- Add robust input validation:
  - Validate form inputs as they are entered
  - Provide immediate feedback for invalid inputs
  - Create reusable validation functions

## 6. Testing Strategy

### 6.1 Unit Testing Components

- Add unit tests for pure functions and components:
  - Test rendering logic with snapshot tests
  - Test state transitions in isolation
  - Mock dependencies for deterministic testing

### 6.2 Integration Testing

- Add integration tests for key workflows:
  - Test complete user interactions
  - Verify state consistency across operations
  - Test edge cases and error handling

## 7. Performance Optimizations

### 7.1 Rendering Optimization

- Optimize rendering performance:
  - Implement view caching for expensive renders
  - Only re-render parts of the UI that have changed
  - Add debouncing for frequent updates (e.g., loading indicators)

### 7.2 State Update Optimization

- Optimize state updates:
  - Batch related state updates
  - Add optimistic UI updates for better responsiveness
  - Implement proper loading states for async operations

## 8. Documentation

### 8.1 Code Documentation

- Improve code documentation:
  - Add consistent comments to all exported functions
  - Document complex algorithms and state management
  - Add diagrams for the overall component architecture

### 8.2 User Documentation

- Update user documentation:
  - Document all keyboard shortcuts
  - Create help screens for complex features
  - Add tooltips and status messages for better discoverability

## Implementation Strategy

When implementing these refactorings, follow this approach:

1. **Start small**: Begin with focused, isolated improvements
2. **Test continuously**: Add tests before major refactorings
3. **Refactor in stages**: Complete one category of changes before moving to the next
4. **Update documentation**: Keep documentation in sync with code changes
5. **Get feedback**: Review changes with team members regularly

## Code Examples

### Example: Panel Interface

```go
// Panel is the interface that all panels must implement
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
    HandleKey(msg tea.KeyMsg) (Panel, tea.Cmd)
}

// BasePanel provides common functionality for all panels
type BasePanel struct {
    width      int
    height     int
    isActive   bool
    styles     *shared.Styles
}

// NewTaskListPanel creates a new task list panel
func NewTaskListPanel(styles *shared.Styles) *TaskListPanel {
    return &TaskListPanel{
        BasePanel: BasePanel{
            styles: styles,
            isActive: false,
        },
        tasks: []task.Task{},
        cursor: 0,
        offset: 0,
        collapsibleManager: hooks.NewCollapsibleManager(),
    }
}
```

### Example: Command Pattern Implementation

```go
// Command represents a user action that modifies the model
type Command func(m *Model) (tea.Model, tea.Cmd)

// commandMap maps key presses to commands based on context
type commandMap map[string]map[int]map[string]Command

// getCommand returns the appropriate command for a key press
func (m *Model) getCommand(key string) Command {
    viewMode := m.viewMode
    activePanel := m.activePanel

    // Global commands that work in any context
    globalCommands := map[string]Command{
        "q":      m.quit,
        "ctrl+c": m.quit,
        "r":      m.refreshTasks,
    }

    // Check for global command first
    if cmd, ok := globalCommands[key]; ok {
        return cmd
    }

    // View-specific commands
    commands := commandMap{
        "list": {
            0: { // Task list panel
                "j":     m.navigateDown,
                "down":  m.navigateDown,
                "k":     m.navigateUp,
                "up":    m.navigateUp,
                // ...more commands
            },
            // ...more panels
        },
        // ...more view modes
    }

    // Find the command in the context-specific map
    if viewCommands, ok := commands[viewMode]; ok {
        if panelCommands, ok := viewCommands[activePanel]; ok {
            if cmd, ok := panelCommands[key]; ok {
                return cmd
            }
        }
    }

    // No command found
    return nil
}
```

### Example: State Container Implementation

```go
// TimelineState manages the state for the timeline feature
type TimelineState struct {
    // Tasks in different time categories
    overdueTasks  []task.Task
    todayTasks    []task.Task
    upcomingTasks []task.Task

    // UI state
    cursor        int
    cursorOnHeader bool
    offset        int

    // Section management
    collapsibleManager *hooks.CollapsibleManager
}

// NewTimelineState creates a new timeline state
func NewTimelineState() *TimelineState {
    ts := &TimelineState{
        cursor: 0,
        cursorOnHeader: true,
        offset: 0,
        collapsibleManager: hooks.NewCollapsibleManager(),
    }

    // Initialize default sections
    ts.collapsibleManager.AddSection(hooks.SectionTypeOverdue, "Overdue", 0, 0)
    ts.collapsibleManager.AddSection(hooks.SectionTypeToday, "Today", 0, 0)
    ts.collapsibleManager.AddSection(hooks.SectionTypeUpcoming, "Upcoming", 0, 0)

    return ts
}

// Update updates the timeline state with new tasks
func (ts *TimelineState) Update(tasks []task.Task) {
    // Categorize tasks
    ts.overdueTasks, ts.todayTasks, ts.upcomingTasks = categorizeTasks(tasks)

    // Update section item counts
    ts.collapsibleManager.UpdateSectionItemCount(hooks.SectionTypeOverdue, len(ts.overdueTasks))
    ts.collapsibleManager.UpdateSectionItemCount(hooks.SectionTypeToday, len(ts.todayTasks))
    ts.collapsibleManager.UpdateSectionItemCount(hooks.SectionTypeUpcoming, len(ts.upcomingTasks))
}

// Helper methods for cursor management, navigation, etc.
```

Use these patterns and examples to guide your refactoring efforts towards a more maintainable and testable TUI implementation.
