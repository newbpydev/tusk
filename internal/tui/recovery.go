package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/terminaltext"
)

const errRetireFailed core.Error = "previous storage owner could not be closed"

type recoverySnapshot struct {
	forest  []*core.TaskNode
	history []ports.TaskEvent
}
type recoveryState struct {
	target        taskRef
	state         loadState
	field, scroll int
	err           string
	blocked       bool
	observed      *core.Task
	history       []ports.TaskEvent
}
type recoveryMsg struct {
	owner, operation uint64
	snapshot         recoverySnapshot
	now              time.Time
	err              error
}

func (s *Session) Recover(ctx context.Context, id string) (recoverySnapshot, error) {
	value, err := s.call(ctx, false, true, func(ctx context.Context, svc ports.TaskService) (any, error) {
		forest, err := svc.GetTaskTree(ctx, "")
		if err != nil {
			return nil, err
		}
		result := recoverySnapshot{forest: forest}
		if findTask(forest, id) != nil {
			result.history, err = svc.GetTaskHistory(ctx, id)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	})
	if err != nil {
		return recoverySnapshot{}, err
	}
	return value.(recoverySnapshot), nil
}

func findTask(forest []*core.TaskNode, id string) *core.Task {
	if id == "" {
		return nil
	}
	for _, node := range forest {
		if node.Task.ID == id {
			return &node.Task
		}
		if found := findTask(node.Children, id); found != nil {
			return found
		}
	}
	return nil
}

func (m *Model) beginRecovery() {
	if m.recovery != nil {
		return
	}
	target := taskIdentity(m.selectedTask())
	if m.confirmation != nil {
		target = m.confirmation.target
	}
	if m.form != nil {
		target = taskIdentity(m.form.draft.base)
	}
	if request := m.uncertain; request != nil {
		target = taskIdentity(request.baseTask())
	}
	m.recovery = &recoveryState{target: target, state: loadFailed}
}

func (m *Model) requestRecovery() tea.Cmd {
	r := m.recovery
	if m.busy || r.blocked {
		return nil
	}
	if m.options.Recover == nil {
		r.err = "Reload is unavailable. Quit and inspect the data with the CLI."
		return nil
	}
	m.owner++
	m.generation++
	m.operation++
	m.busy = true
	m.refreshPending = false
	m.history.pending = false
	r.state = loading
	r.field = 0
	r.err = ""
	owner, operation, ctx, recover, target, now := m.owner, m.operation, m.options.Context, m.options.Recover, r.target.id, m.options.Now
	return safeCommand(func() tea.Msg {
		snapshot, err := recover(ctx, target)
		return recoveryMsg{owner, operation, snapshot, now(), err}
	})
}

func (m *Model) acceptRecovery(msg recoveryMsg) tea.Cmd {
	if msg.owner != m.owner || msg.operation != m.operation || !m.busy || !m.recoveryNeeded || m.recovery == nil {
		return nil
	}
	m.busy = false
	r := m.recovery
	if msg.err != nil {
		r.state = loadFailed
		r.err = "Readback failed. Writes stay paused; r retries only the readback."
		if errors.Is(msg.err, errRetireFailed) {
			r.blocked = true
			r.err = "Storage could not close. Quit, then inspect data with the CLI."
		}
		return nil
	}
	m.forest = msg.snapshot.forest
	m.now = msg.now
	m.state = loaded
	m.stale = false
	m.pruneCollapsed()
	r.state = loaded
	r.field = 0
	r.scroll = 0
	r.observed = nil
	r.history = append([]ports.TaskEvent(nil), msg.snapshot.history...)
	if task := findTask(m.forest, r.target.id); task != nil {
		clone := task.Clone()
		r.observed = &clone
	}
	// Keep write admission frozen until the user acknowledges the observed state.
	if !m.timerStarted {
		m.timerStarted = true
		return m.nextTick()
	}
	return nil
}

func (m *Model) recoveryKey(key tea.KeyMsg) tea.Cmd {
	if m.recovery == nil {
		m.beginRecovery()
	}
	r := m.recovery
	if key.String() == "q" || (r.blocked && key.String() == "enter") {
		return tea.Quit
	}
	if m.busy {
		return nil
	}
	count := 2
	if r.state == loaded {
		count = 3
	}
	switch key.String() {
	case "r":
		return m.requestRecovery()
	case "tab":
		if r.blocked {
			return nil
		}
		r.field = (r.field + 1) % count
	case "shift+tab":
		if r.blocked {
			return nil
		}
		r.field = (r.field + count - 1) % count
	case "down", "j":
		r.scroll++
	case "up", "k":
		r.scroll--
	case "pgdown":
		r.scroll += max(1, measure(m.width, m.height).modal.height-4)
	case "pgup":
		r.scroll -= max(1, measure(m.width, m.height).modal.height-4)
	case "home":
		r.scroll = 0
	case "end":
		r.scroll = 1 << 30
	case "enter":
		if r.field == 0 {
			return m.requestRecovery()
		}
		if r.field == count-1 {
			return tea.Quit
		}
		if r.state == loaded {
			if m.form != nil {
				m.closeForm()
			}
			m.confirmation = nil
			m.recovery = nil
			m.uncertain = nil
			m.recoveryNeeded = false
			m.awaitingRead = false
			m.refreshPending = false
			m.historyRefresh = true
			m.notice = "Readback acknowledged"
		}
	}
	return nil
}

func (m *Model) recoveryContent(width, area int) ([]string, int, string) {
	r := m.recovery
	lines := []string{}
	add := func(value string) {
		lines = append(lines, strings.Split(ansi.Wrap(value, max(1, width-2), ""), "\n")...)
	}
	add("Outcome unknown — do not resubmit.")
	if r.state == loaded {
		add("Saved data reloaded. Review the current state below.")
	} else {
		add("Writes are paused. Reload checks what is currently saved.")
	}
	add("")
	if m.uncertain != nil && m.uncertain.kind == mutationCreate {
		add("A task may have been created, but its identity is unconfirmed.")
		add("The app will not retry this action. Review the current task list.")
	}
	if r.err != "" {
		add("")
		add(r.err)
	}
	if r.target.id != "" {
		add("Task ID: " + terminaltext.Scalar(r.target.id))
	}
	if request := m.uncertain; request != nil {
		if base := request.baseTask(); base != nil {
			label := "Edit task"
			switch request.kind {
			case mutationDelete:
				label = "Delete task"
			case mutationComplete:
				label = "Complete subtree"
			case mutationReopen:
				label = "Reopen task"
			}
			add("Attempted action: " + label)
			add(terminaltext.Scalar(base.Title))
		}
	}
	if m.form != nil {
		add("")
		add("Your draft · read-only")
		add(terminaltext.Scalar(m.form.draft.fields[fieldTitle]))
		add(terminaltext.Multiline(m.form.draft.fields[fieldNotes]))
		labels := []string{"Priority", "Due", "Tags", "Parent", "Status", "Progress"}
		for i := fieldPriority; i < m.form.count(); i++ {
			value := m.form.draft.fields[i]
			if value == "" {
				value = "(empty)"
			}
			add(labels[i-fieldPriority] + ": " + terminaltext.Scalar(value))
		}
	}
	if r.state == loaded {
		add("")
		add("Currently saved · other apps may also have made changes")
		if r.target.id == "" {
			add("Workspace refreshed. Review the current tasks before continuing.")
			var walk func([]*core.TaskNode)
			walk = func(nodes []*core.TaskNode) {
				for _, n := range nodes {
					add(terminaltext.Scalar(n.Task.Title) + " · " + terminaltext.Scalar(n.Task.ID))
					walk(n.Children)
				}
			}
			walk(m.forest)
		} else if r.observed == nil {
			add("Target is absent: " + terminaltext.Scalar(r.target.id))
			add("There is no stored history for an absent task.")
		} else {
			task := r.observed
			add(terminaltext.Scalar(task.Title))
			add("ID: " + terminaltext.Scalar(task.ID))
			if taskIdentity(task) != r.target {
				add("A different task now uses this ID.")
			}
			add(fmt.Sprintf("Status: %s · Progress: %d%%", terminaltext.Scalar(string(task.Status)), task.Progress))
			add("Priority: " + task.Priority.String())
			tags := make([]string, len(task.Tags))
			for i, tag := range task.Tags {
				tags[i] = string(tag)
			}
			add("Tags: " + terminaltext.Scalar(strings.Join(tags, ", ")))
			parent := "Root"
			if task.ParentID != nil {
				parent = terminaltext.Scalar(*task.ParentID)
			}
			add("Parent: " + parent)
			due := "Not set"
			if task.DueDate != nil {
				due = task.DueDate.In(m.options.Location).Format(time.RFC3339Nano)
			}
			add("Due: " + due)
			add("Created: " + task.CreatedAt.In(m.options.Location).Format(time.RFC3339Nano))
			add("Updated: " + task.UpdatedAt.In(m.options.Location).Format(time.RFC3339Nano))
			if task.CompletedAt != nil {
				add("Completed: " + task.CompletedAt.In(m.options.Location).Format(time.RFC3339Nano))
			}
			add(terminaltext.Multiline(task.Description))
			add(fmt.Sprintf("%d stored activity events", len(r.history)))
			for _, event := range r.history {
				add(fmt.Sprintf("#%d %s · %s", event.Sequence, terminaltext.Scalar(string(event.Kind)), event.OccurredAt.In(m.options.Location).Format(time.RFC3339)))
			}
		}
		add("")
		add("Acknowledge to discard uncertain intent and enable new actions.")
	}
	r.scroll = max(0, min(r.scroll, max(0, len(lines)-area)))
	if m.busy {
		return lines, r.scroll, "Reloading… · q quit · Ctrl+C exits safely"
	}
	if r.blocked {
		return lines, r.scroll, "[ Quit ] Enter / q · inspect saved data with the CLI"
	}
	labels := []string{"Reload", "Quit"}
	if r.state == loaded {
		labels = []string{"Reload", "Acknowledge / discard", "Quit"}
	}
	labels[r.field] = "[ " + labels[r.field] + " ]"
	return lines, r.scroll, strings.Join(labels, "   ")
}
