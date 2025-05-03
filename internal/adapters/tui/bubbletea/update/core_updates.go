// Package update contains message handlers for the bubbletea update loop.
// This package separates update logic into distinct handlers for better maintainability.
package update

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/state"
)

// Handler defines a common interface for message handlers
type Handler interface {
	// HandleMessage processes a tea.Msg and returns updated state and command
	HandleMessage(msg tea.Msg, appState *state.AppState) (*state.AppState, tea.Cmd)
	
	// CanHandle checks if this handler can process the given message type
	CanHandle(msg tea.Msg) bool
}

// Dispatcher routes messages to appropriate handlers
type Dispatcher struct {
	handlers []Handler
}

// NewDispatcher creates a message dispatcher with the provided handlers
func NewDispatcher(handlers ...Handler) *Dispatcher {
	return &Dispatcher{
		handlers: handlers,
	}
}

// RegisterHandler adds a handler to the dispatcher
func (d *Dispatcher) RegisterHandler(handler Handler) {
	d.handlers = append(d.handlers, handler)
}

// Dispatch sends a message to the appropriate handler
func (d *Dispatcher) Dispatch(msg tea.Msg, appState *state.AppState) (*state.AppState, tea.Cmd) {
	for _, handler := range d.handlers {
		if handler.CanHandle(msg) {
			return handler.HandleMessage(msg, appState)
		}
	}
	
	// Default handler for unhandled message types
	return appState, nil
}

// KeyHandler processes keyboard input messages
type KeyHandler struct {
	// Dependencies that would be injected
	// keyboardHandlers would be an interface to the handlers package
}

// CanHandle checks if this handler can process the message
func (h *KeyHandler) CanHandle(msg tea.Msg) bool {
	_, ok := msg.(tea.KeyMsg)
	return ok
}

// HandleMessage processes a tea.KeyMsg
func (h *KeyHandler) HandleMessage(msg tea.Msg, appState *state.AppState) (*state.AppState, tea.Cmd) {
	keyMsg := msg.(tea.KeyMsg)
	
	// Global key handlers that work in any mode
	switch keyMsg.String() {
	case "q", "ctrl+c":
		return appState, tea.Quit
	}
	
	// Delegate to view-specific handlers based on current view mode
	switch appState.ViewMode {
	case "list":
		// Handle list view keys based on active panel
		// This would call into the handlers package
		return appState, nil
	case "detail":
		// Handle detail view keys
		return appState, nil
	case "create", "edit":
		// Handle form view keys
		return appState, nil
	default:
		return appState, nil
	}
}

// WindowSizeHandler processes window size change messages
type WindowSizeHandler struct{}

// CanHandle checks if this handler can process the message
func (h *WindowSizeHandler) CanHandle(msg tea.Msg) bool {
	_, ok := msg.(tea.WindowSizeMsg)
	return ok
}

// HandleMessage processes a tea.WindowSizeMsg
func (h *WindowSizeHandler) HandleMessage(msg tea.Msg, appState *state.AppState) (*state.AppState, tea.Cmd) {
	sizeMsg := msg.(tea.WindowSizeMsg)
	
	// Update window dimensions
	appState.Width = sizeMsg.Width
	appState.Height = sizeMsg.Height
	
	return appState, nil
}

// TickHandler processes timer tick messages
type TickHandler struct{}

// CanHandle checks if this handler can process the message
func (h *TickHandler) CanHandle(msg tea.Msg) bool {
	_, ok := msg.(messages.TickMsg)
	return ok
}

// HandleMessage processes a messages.TickMsg
func (h *TickHandler) HandleMessage(msg tea.Msg, appState *state.AppState) (*state.AppState, tea.Cmd) {
	// Update current time with the CURRENT system time, not the msg time
	// This ensures we always display the latest time and prevents drift
	appState.CurrentTime = time.Now()
	
	// Clear expired status messages
	if (!appState.StatusExpiry.IsZero()) && time.Now().After(appState.StatusExpiry) {
		appState.StatusMessage = ""
		appState.StatusType = ""
		appState.StatusExpiry = time.Time{}
	}
	
	// Schedule next tick
	return appState, tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return messages.TickMsg(t)
	})
}

// ErrorHandler processes error messages
type ErrorHandler struct{}

// CanHandle checks if this handler can process the message
func (h *ErrorHandler) CanHandle(msg tea.Msg) bool {
	_, ok := msg.(messages.ErrorMsg)
	return ok
}

// HandleMessage processes a messages.ErrorMsg
func (h *ErrorHandler) HandleMessage(msg tea.Msg, appState *state.AppState) (*state.AppState, tea.Cmd) {
	errorMsg := msg.(messages.ErrorMsg)
	
	// Update error state
	appState.Error = error(errorMsg)
	
	// Set error status message
	appState = appState.SetStatusMessage(
		fmt.Sprintf("Error: %v", error(errorMsg)),
		"error", 
		5*time.Second,
	)
	
	return appState, nil
}
