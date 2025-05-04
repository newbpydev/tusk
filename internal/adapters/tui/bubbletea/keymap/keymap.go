// Package keymap provides centralized keyboard shortcut definitions and handling
// for the TUI application, ensuring consistency across different panels and views.
package keymap

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// KeyMap defines a collection of key bindings for a specific context
type KeyMap struct {
	keys     []key.Binding
	context  string
	helpText string
}

// NewKeyMap creates a new key map for a specific context
func NewKeyMap(context string) *KeyMap {
	return &KeyMap{
		keys:    []key.Binding{},
		context: context,
	}
}

// SetHelpText sets custom help text for this keymap
func (k *KeyMap) SetHelpText(text string) {
	k.helpText = text
}

// Add adds a new key binding to the map
func (k *KeyMap) Add(binding key.Binding) {
	k.keys = append(k.keys, binding)
}

// Keys returns all key bindings in this map
func (k *KeyMap) Keys() []key.Binding {
	return k.keys
}

// ShortHelp returns abbreviated help text for the footer
func (k *KeyMap) ShortHelp() []key.Binding {
	// Return a subset of keys that should be shown in short help
	// Sort by help text to ensure consistent display
	bindings := k.keys
	sort.Slice(bindings, func(i, j int) bool {
		return bindings[i].Help().Desc < bindings[j].Help().Desc
	})

	// Only show up to 8 keys in the short help
	if len(bindings) > 8 {
		bindings = bindings[:8]
	}
	return bindings
}

// FullHelp returns complete help text for the help screen
func (k *KeyMap) FullHelp() [][]key.Binding {
	// Group keys by category if needed
	// Here we just return all keys as one group
	return [][]key.Binding{k.keys}
}

// HelpText returns the help text for this keymap
func (k *KeyMap) HelpText() string {
	if k.helpText != "" {
		return k.helpText
	}
	return k.context
}

// HandleKey processes a key press against this keymap and returns true if handled
func (k *KeyMap) HandleKey(msg tea.KeyMsg) bool {
	// Check if any key in the map matches
	for _, binding := range k.keys {
		if key.Matches(msg, binding) {
			return true
		}
	}
	return false
}

// FindKeyHelp returns the help text for the matching key
func (k *KeyMap) FindKeyHelp(msg tea.KeyMsg) (string, string) {
	for _, binding := range k.keys {
		if key.Matches(msg, binding) {
			help := binding.Help()
			return help.Key, help.Desc
		}
	}
	return "", ""
}

// GlobalKeyMap contains application-wide key bindings
var GlobalKeyMap = &KeyMap{
	context: "Global",
	keys: []key.Binding{
		key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "Quit"),
		),
		key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "Quit"),
		),
		key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "Create New Task"),
		),
		key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "Show Sample Modal"),
		),
		key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "Toggle Help"),
		),
	},
}

// TaskListKeyMap contains key bindings for the task list panel
var TaskListKeyMap = &KeyMap{
	context: "Task List",
	keys: []key.Binding{
		key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j/↓", "Move Down"),
		),
		key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k/↑", "Move Up"),
		),
		key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("", "Down"),
		),
		key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("", "Up"),
		),
		key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "New Task"),
		),
		key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "Delete Task"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Select"),
		),
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "Next Panel"),
		),
	},
}

// TaskDetailsKeyMap contains key bindings for the task details panel
var TaskDetailsKeyMap = &KeyMap{
	context: "Task Details",
	keys: []key.Binding{
		key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "Edit Task"),
		),
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "Next Panel"),
		),
		key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "Previous Panel"),
		),
		key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "Toggle Status"),
		),
	},
}

// TimelineKeyMap contains key bindings for the timeline panel
var TimelineKeyMap = &KeyMap{
	context: "Timeline",
	keys: []key.Binding{
		key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j/↓", "Move Down"),
		),
		key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k/↑", "Move Up"),
		),
		key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("", "Down"),
		),
		key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("", "Up"),
		),
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "Next Panel"),
		),
		key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "Previous Panel"),
		),
	},
}

// FormKeyMap contains key bindings for forms
var FormKeyMap = &KeyMap{
	context: "Form",
	keys: []key.Binding{
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "Next Field"),
		),
		key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "Previous Field"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Submit"),
		),
		key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "Cancel"),
		),
	},
}

// ModalKeyMap contains key bindings for modal dialogs
var ModalKeyMap = &KeyMap{
	context: "Modal",
	keys: []key.Binding{
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab/→", "Next Option"),
		),
		key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab/←", "Previous Option"),
		),
		key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("", "Next Option"),
		),
		key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("", "Previous Option"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Confirm"),
		),
		key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "Cancel"),
		),
	},
}

// GetKeyMapForContext returns the appropriate keymap for a given context
func GetKeyMapForContext(context string) *KeyMap {
	switch strings.ToLower(context) {
	case "global":
		return GlobalKeyMap
	case "task list", "tasklist":
		return TaskListKeyMap
	case "task details", "taskdetails":
		return TaskDetailsKeyMap
	case "timeline":
		return TimelineKeyMap
	case "form":
		return FormKeyMap
	case "modal":
		return ModalKeyMap
	default:
		return GlobalKeyMap
	}
}

// KeyMapContexts returns all available context names
func KeyMapContexts() []string {
	return []string{
		"Global",
		"Task List",
		"Task Details",
		"Timeline",
		"Form",
		"Modal",
	}
}
