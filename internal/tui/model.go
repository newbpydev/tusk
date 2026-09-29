// Package tui implements the interactive adapter. Only Update prepares frames;
// View is a read of a completed immutable string.
package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/core"
)

// Options contains per-session dependencies. Callers provide resolved terminal
// and calendar facts; construction never queries the environment or a terminal.
type Options struct {
	Context  context.Context
	Load     func(context.Context) ([]*core.TaskNode, error)
	Now      func() time.Time
	Wait     func(context.Context, time.Duration) error
	Location *time.Location
	Profile  termenv.Profile
}

type loadState uint8

const (
	loading loadState = iota
	loaded
	loadFailed
)

type Model struct {
	options       Options
	renderer      *lipgloss.Renderer
	notes         textarea.Model
	help          help.Model
	frame         string
	width, height int
	state         loadState
	operation     uint64
	forest        []*core.TaskNode
	exitErr       error
}

func New(options Options) *Model {
	r := lipgloss.NewRenderer(io.Discard, termenv.WithProfile(options.Profile))
	r.SetColorProfile(options.Profile)
	r.SetHasDarkBackground(true)
	plain := r.NewStyle()
	notes := textarea.New()
	notes.FocusedStyle = textarea.Style{Base: plain, CursorLine: plain, CursorLineNumber: plain, EndOfBuffer: plain, LineNumber: plain, Placeholder: plain, Prompt: plain, Text: plain}
	notes.BlurredStyle = notes.FocusedStyle
	notes.Cursor.Style, notes.Cursor.TextStyle = plain, plain
	notes.Cursor.SetMode(cursor.CursorStatic)
	notes.KeyMap.Paste.SetEnabled(false)
	notes.Blur()
	h := help.New()
	h.Styles = help.Styles{Ellipsis: plain, ShortKey: plain, ShortDesc: plain, ShortSeparator: plain, FullKey: plain, FullDesc: plain, FullSeparator: plain}
	m := &Model{options: options, renderer: r, notes: notes, help: h, operation: 1, width: 80, height: 24}
	m.prepareFrame()
	return m
}

func (m *Model) Init() tea.Cmd {
	// Copy every command input now. The closure must never capture m.
	ctx, load, operation := m.options.Context, m.options.Load, m.operation
	return safeCommand(func() tea.Msg {
		forest, err := load(ctx)
		return forestMsg{operation: operation, forest: forest, err: err}
	})
}

func (m *Model) Update(msg tea.Msg) (next tea.Model, cmd tea.Cmd) {
	defer func() {
		if recover() != nil {
			m.exitErr = errRuntime
			m.frame = "Interactive session failed."
			next, cmd = m, tea.Quit
		}
	}()
	switch msg := msg.(type) {
	case fatalMsg:
		m.exitErr = msg.err
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "ctrl+c":
			m.exitErr = context.Canceled
			return m, tea.Quit
		case "r":
			if m.state == loadFailed {
				m.operation++
				m.state = loading
				m.prepareFrame()
				return m, m.Init()
			}
		}
	case tea.WindowSizeMsg:
		m.width, m.height = max(1, msg.Width), max(1, msg.Height)
	case forestMsg:
		if msg.operation != m.operation {
			return m, nil
		}
		if errors.Is(msg.err, errRuntime) {
			m.exitErr = errRuntime
			return m, tea.Quit
		}
		if msg.err != nil {
			m.state = loadFailed
		} else {
			m.state = loaded
			m.forest = msg.forest
		}
	}
	m.prepareFrame()
	return m, nil
}
func (m *Model) View() string { return m.frame }
func (m *Model) prepareFrame() {
	text := "Loading tasks…"
	switch m.state {
	case loaded:
		text = "No tasks"
		if len(m.forest) > 0 {
			text = fmt.Sprintf("%d task roots", len(m.forest))
		}
	case loadFailed:
		text = "Could not load tasks. Press r to retry."
	}
	// Child components may maintain caches while rendering. They run here only,
	// under constructor/Update ownership, never from root View.
	if m.notes.Focused() {
		text += "\n" + m.notes.View()
	}
	help := m.help.ShortHelpView([]key.Binding{key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit"))})
	lines := []string{m.renderer.NewStyle().Bold(true).Render("TUSK"), text, help}
	m.frame = strings.Join(lines, "\n")
}

// Wait is a cancellable timer seam. It performs work only inside a command.
func Wait(ctx context.Context, delay time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
