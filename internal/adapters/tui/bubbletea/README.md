# BubbleTea Terminal UI

This directory contains the Terminal User Interface (TUI) implementation for Tusk, built using the BubbleTea framework.

## Structure

```plaintext
bubbletea/
├── app/                # Main application container
│   └── app.go          # Application initialization and coordination
├── components/         # Reusable UI components
│   ├── input/          # Input field components
│   ├── layout/         # Layout components
│   │   ├── footer.go   # Application footer component
│   │   └── header.go   # Application header component
│   ├── list/           # List view components
│   ├── panels/         # Panel implementations
│   │   ├── create_form.go  # Task creation form panel
│   │   ├── task_details.go # Task details panel
│   │   ├── task_list.go    # Task list panel
│   │   └── timeline.go     # Timeline/calendar view panel
│   └── shared/         # Shared component utilities
│       ├── panel.go    # Base panel implementation
│       ├── scrollable_panel.go # Scrollable content panel
│       └── styles.go   # Shared styling constants
├── legacy/             # Legacy implementation (for reference)
│   ├── model.go        # Original model definition
│   ├── styles.go       # Original styling approach
│   ├── update.go       # Original update logic
│   └── view.go         # Original view rendering
├── messages/           # Message definitions for component communication
│   └── messages.go     # Message type definitions
├── styles/             # Global UI styling
│   └── styles.go       # Theme definitions and style constants
└── views/              # Full-screen views
    ├── mainmenu/       # Main menu view
    └── settings/       # Settings view
```

## Architecture

The TUI follows a Model-View-Update (MVU) architecture pattern implemented through BubbleTea:

1. **Model**: Represents the application state
2. **View**: Renders the state as terminal output
3. **Update**: Processes messages and updates the model accordingly

Components communicate through message passing, which allows for loosely coupled interactions between different parts of the UI.

## Key Components

### App Container (`app/app.go`)

The entry point for the TUI implementation, responsible for:

- Initializing the BubbleTea program
- Managing global state and component coordination
- Handling top-level keyboard shortcuts
- Managing view transitions

### Panels (`components/panels/`)

Focused UI components that handle specific functionality:

- **Task List Panel**: Displays tasks with filtering and sorting options
- **Task Details Panel**: Shows comprehensive information about a selected task
- **Create Form Panel**: Provides form fields for creating and editing tasks
- **Timeline Panel**: Visualizes tasks on a calendar or timeline view

### Shared Components (`components/shared/`)

Base components and utilities that are reused across the TUI:

- **Panel**: Base panel implementation with common functionality
- **ScrollablePanel**: Extension for panels that need scrollable content
- **Styles**: Shared styling constants and helper functions

### Messages (`messages/messages.go`)

Message types used for communication between components:

- Task selection events
- Form submission events
- Error notifications
- Status updates

## Keyboard Navigation

The TUI implements a comprehensive keyboard navigation system:

- **↑/↓** or **k/j**: Navigate between items
- **Tab**: Cycle between panels
- **Space** or **c**: Toggle task completion
- **Enter**: Select item or confirm action
- **Esc**: Go back or dismiss dialogs
- **q**: Quit the application
- **?**: Show help overlay with keyboard shortcut reference

## Development Guidelines

When working with the TUI:

1. **Component Isolation**: Each component should manage its own state and view logic
2. **Message Passing**: Components should communicate through messages, not direct calls
3. **Consistent Styling**: Use the shared styles for consistent visual appearance
4. **Keyboard Focus**: Ensure keyboard navigation works consistently
5. **Error Handling**: Provide meaningful feedback for errors
6. **Accessibility**: Consider terminal limitations and provide good navigation options

### Adding a New Panel

To create a new panel:

1. Create a new file in the appropriate directory (usually `components/panels/`)
2. Embed the base `Panel` or `ScrollablePanel` struct
3. Implement the required BubbleTea methods (Init, Update, View)
4. Define message types if needed in `messages/`
5. Add the panel to the appropriate container in `app/app.go`

### Styling Guidelines

- Use `lipgloss` for all styling needs
- Reference colors from the theme definitions in `styles/styles.go`
- Consider terminal color limitations
- Use margin and padding consistently
- Ensure readability with appropriate contrast
