package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/terminaltext"
)

type taskRef struct {
	id      string
	created time.Time
}

func taskIdentity(t *core.Task) taskRef {
	if t == nil {
		return taskRef{}
	}
	return taskRef{t.ID, t.CreatedAt.UTC()}
}

type notesKey struct {
	target     taskRef
	source     string
	width      int
	profile    termenv.Profile
	generation uint64
}
type markdownState struct {
	wanted                 notesKey
	token                  uint64
	active, pending, plain bool
	lines                  []string
}
type notesMsg struct {
	key   notesKey
	token uint64
	lines []string
	plain bool
}

func (m *Model) finish(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	if m.exitErr != nil {
		return m, cmd
	}
	m.rebuildRows()
	details := m.syncDetails()
	m.prepareFrame()
	return m, tea.Batch(cmd, details)
}

func (m *Model) syncDetails() tea.Cmd {
	task := m.selectedTask()
	ref := taskIdentity(task)
	if ref != m.detailRef {
		m.detailRef = ref
		m.detailsScroll = 0
		m.detailsEnd = false
		m.history.events = nil
		m.history.state = loading
		m.historyRefresh = true
	}
	if m.historyRefresh {
		m.historyRefresh = false
		m.history.key = ref
		m.history.token++
		m.history.pending = task != nil && m.options.History != nil
	}
	var key notesKey
	if task != nil {
		key = notesKey{ref, task.Description, max(1, measure(m.width, m.height).detailsWidth-6), m.options.Profile, m.generation}
	}
	if key != m.markdown.wanted {
		keep := key.target == m.markdown.wanted.target && key.source == m.markdown.wanted.source
		m.markdown.wanted = key
		if !keep {
			m.markdown.lines = nil
			m.markdown.plain = false
		}
		m.markdown.pending = task != nil && key.source != ""
		if task != nil && key.source == "" {
			m.markdown.lines = []string{"No notes"}
		}
	}
	var render tea.Cmd
	if m.markdown.pending && !m.markdown.active {
		m.markdown.active = true
		m.markdown.pending = false
		m.markdown.token++
		token, ctx, renderer := m.markdown.token, m.options.Context, m.options.Render
		render = safeCommand(func() tea.Msg {
			lines, plain := renderNoteText(ctx, renderer, key.source, key.width, key.profile)
			return notesMsg{key, token, lines, plain}
		})
	}
	return tea.Batch(render, m.dispatchHistory())
}

func (m *Model) acceptNotes(msg notesMsg) {
	if !m.markdown.active || msg.token != m.markdown.token {
		return
	}
	m.markdown.active = false
	if msg.key == m.markdown.wanted {
		m.markdown.lines = msg.lines
		m.markdown.plain = msg.plain
	}
}

func (m *Model) detailLines(w int) []string {
	t := m.selectedTask()
	if t == nil {
		return []string{"", " Select a task to see its details."}
	}
	width := max(1, w-4)
	var lines []string
	add := func(text string) {
		for _, line := range strings.Split(ansi.Wrap(text, width, ""), "\n") {
			lines = append(lines, "  "+line)
		}
	}
	label := func(text string) { add(m.paint(text, textColor, true)) }
	date := func(v *time.Time) string {
		if v == nil || v.IsZero() {
			return "Not set"
		}
		return v.In(m.options.Location).Format("2006-01-02 15:04:05 MST (-07:00)")
	}
	add("")
	label(terminaltext.Scalar(t.Title))
	if t.ParentID == nil {
		add(m.paint("Root task", mutedColor, false))
	} else {
		add(m.paint("Subtask of "+terminaltext.Scalar(*t.ParentID), mutedColor, false))
	}
	add("")
	add("Status    " + terminaltext.Scalar(string(t.Status)))
	add("Priority  " + t.Priority.String())
	add(fmt.Sprintf("Progress  %d%%", t.Progress))
	add("Due       " + date(t.DueDate))
	tags := make([]string, len(t.Tags))
	for i, tag := range t.Tags {
		tags[i] = terminaltext.Scalar(string(tag))
	}
	tagText := strings.Join(tags, ", ")
	if tagText == "" {
		tagText = "None"
	}
	add("Tags      " + tagText)
	add("")
	add("ID        " + terminaltext.Scalar(t.ID))
	add("Created   " + date(&t.CreatedAt))
	add("Updated   " + date(&t.UpdatedAt))
	add("Completed " + date(t.CompletedAt))
	add("")
	label("Notes")
	if m.markdown.plain {
		add(m.paint("Plain text · large or unsupported note", mutedColor, false))
	}
	if m.markdown.lines == nil {
		add("Rendering notes…")
	} else {
		for _, line := range m.markdown.lines {
			lines = append(lines, "  "+line)
		}
	}
	if m.notes.Focused() {
		lines = append(lines, strings.Split(m.notes.View(), "\n")...)
	}
	add("")
	add(m.paint(strings.Repeat("─", width), borderColor, false))
	label("Activity")
	add(m.paint("History is refreshed separately", mutedColor, false))
	switch {
	case m.options.History == nil:
		add("History unavailable")
	case m.history.state == loadFailed:
		add("Could not load history · r retries")
	case m.history.state == loading:
		add("Loading history…")
	case len(m.history.events) == 0:
		add("No history")
	default:
		for _, event := range m.history.events {
			add("")
			add(fmt.Sprintf("#%d  %s", event.Sequence, terminaltext.Scalar(string(event.Kind))))
			fields := make([]string, len(event.ChangedFields))
			for i, f := range event.ChangedFields {
				fields[i] = terminaltext.Scalar(f)
			}
			add("Changed " + strings.Join(fields, ", "))
			add(m.paint(date(&event.OccurredAt), mutedColor, false))
		}
	}
	return lines
}
