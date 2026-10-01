// Package tui implements the interactive adapter. Only Update prepares frames;
// View is a read of a completed immutable string.
package tui

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

// Options contains per-session dependencies. Callers provide resolved terminal
// and calendar facts; construction never queries the environment or a terminal.
type Options struct {
	Context   context.Context
	Load      func(context.Context) ([]*core.TaskNode, error)
	Now       func() time.Time
	Wait      func(context.Context, time.Duration) error
	Location  *time.Location
	Profile   termenv.Profile
	DayBounds func(string, time.Time, *time.Location) (time.Time, time.Time, error)
	History   func(context.Context, string) ([]ports.TaskEvent, error)
	Render    MarkdownRenderer
	ParseDue  DueParser
	Mutate    func(context.Context, mutationRequest) (*core.Task, error)
	Preview   func(context.Context, string) (ports.DeletePreview, error)
	Delete    func(context.Context, ports.DeleteTaskCommand) (ports.DeleteResult, error)
	Recover   func(context.Context, string) (recoverySnapshot, error)
}

type loadState uint8

const (
	loading loadState = iota
	loaded
	loadFailed
)

type Model struct {
	options                   Options
	renderer                  *lipgloss.Renderer
	notes                     textarea.Model
	help                      help.Model
	frame                     string
	width, height             int
	state                     loadState
	operation                 uint64
	forest                    []*core.TaskNode
	exitErr                   error
	focus                     panelFocus
	helpOpen                  bool
	helpScroll                int
	rows                      []taskRow
	selected                  int
	listOffset, detailsScroll int
	detailsEnd                bool
	collapsed                 map[string]bool
	filter                    core.TaskFilter
	workspaceTab              int // 0 all, 1 today, 2 done, -1 custom filters
	dueStart, dueEnd          *time.Time
	now                       time.Time
	searching                 bool
	searchDraft               string
	searchBefore              searchSnapshot
	restore                   *selectionRestore
	searchToken               uint64
	searchCancel              context.CancelFunc
	owner, generation         uint64
	busy, refreshPending      bool
	stale, recoveryNeeded     bool
	tickToken                 uint64
	timerStarted              bool
	filters                   *filterDraft
	dueExpression, dueLabel   string
	detailRef                 taskRef
	markdown                  markdownState
	history                   historyState
	historyRefresh            bool
	form                      *taskForm
	formSequence              uint64
	saving, awaitingRead      bool
	pendingMutation           *mutationRequest
	readContext               context.Context
	readCancel                context.CancelFunc
	readInterrupted           bool
	notice                    string
	confirmation              *deleteDialog
	committedKind             mutationKind
	uncertain                 *mutationRequest
	recovery                  *recoveryState
}

func New(options Options) *Model {
	if options.Context == nil {
		options.Context = context.Background()
	}
	if options.Location == nil {
		options.Location = time.UTC
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.Wait == nil {
		options.Wait = Wait
	}
	if options.Render == nil {
		options.Render = RenderMarkdown
	}
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
	m := &Model{options: options, renderer: r, notes: notes, help: h, operation: 1, busy: true, selected: -1, collapsed: map[string]bool{}}
	m.startRead()
	m.rebuildRows()
	m.prepareFrame()
	return m
}

func (m *Model) Init() tea.Cmd {
	// Copy every command input now. The closure must never capture m.
	ctx, load, now, operation, owner, generation := m.readContext, m.options.Load, m.options.Now, m.operation, m.owner, m.generation
	return safeCommand(func() tea.Msg {
		forest, err := load(ctx)
		return forestMsg{operation: operation, owner: owner, generation: generation, now: now(), forest: forest, err: err}
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
	case previewMsg:
		cmd = m.acceptPreview(msg)
	case recoveryMsg:
		cmd = m.acceptRecovery(msg)
	case mutationMsg:
		cmd = m.acceptMutation(msg)
	case notesMsg:
		m.acceptNotes(msg)
	case historyMsg:
		cmd = m.acceptHistory(msg)
	case tickMsg:
		if msg.token != m.tickToken || msg.err != nil {
			return m, nil
		}
		m.now = msg.now
		cmd = tea.Batch(m.nextTick(), m.requestRefresh())
	case searchMsg:
		if !m.searching || msg.token != m.searchToken || msg.err != nil {
			return m, nil
		}
		m.filter.SearchTerm = strings.TrimSpace(msg.query)
	case fatalMsg:
		m.exitErr = msg.err
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.exitErr = context.Canceled
			return m, tea.Quit
		}
		if !measure(m.width, m.height).usable() {
			if (m.recoveryNeeded || (m.form == nil && m.confirmation == nil && !m.saving)) && msg.String() == "q" {
				return m, tea.Quit
			}
			return m, nil
		}
		if m.recoveryNeeded {
			return m.finish(m.recoveryKey(msg))
		}
		if m.confirmation != nil {
			return m.finish(m.confirmKey(msg))
		}
		if m.form != nil {
			return m.finish(m.formKey(msg))
		}
		if m.saving {
			return m.finish(nil)
		}
		if m.helpOpen {
			switch msg.String() {
			case "?", "esc", "q":
				m.helpOpen = false
			case "j", "down":
				m.helpScroll++
			case "k", "up":
				m.helpScroll--
			case "pgdown":
				m.helpScroll += max(1, measure(m.width, m.height).modal.height-6)
			case "pgup":
				m.helpScroll -= max(1, measure(m.width, m.height).modal.height-6)
			case "g", "home":
				m.helpScroll = 0
			case "G", "end":
				m.helpScroll = len(helpLines())
			}
			return m.finish(nil)
		}
		if m.searching {
			cmd = m.searchKey(msg)
			return m.finish(cmd)
		}
		if m.filters != nil {
			m.filterKey(msg)
			return m.finish(nil)
		}
		switch msg.String() {
		case "1", "2", "3":
			m.chooseWorkspaceTab(int(msg.Runes[0] - '1'))
		case "d":
			m.beginDelete()
		case "a":
			m.beginForm(false)
		case "e":
			m.beginForm(true)
		case "x", " ":
			cmd = m.toggleTask()
		case "q":
			return m, tea.Quit
		case "tab", "shift+tab":
			m.focus = 1 - m.focus
		case "?":
			m.helpOpen = true
			m.helpScroll = 0
		case "/":
			m.beginSearch()
		case "f":
			m.beginFilters()
		case "esc":
			m.clearFilters()
		case "r":
			cmd = m.requestRefresh()
		default:
			m.navigate(msg.String())
		}
	case tea.WindowSizeMsg:
		m.width, m.height = max(0, msg.Width), max(0, msg.Height)
	case forestMsg:
		cmd = m.acceptForest(msg)
	}
	return m.finish(cmd)
}
func (m *Model) View() string { return m.frame }

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
