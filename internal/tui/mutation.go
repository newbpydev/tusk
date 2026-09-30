package tui

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

type mutationMsg struct {
	owner, operation uint64
	request          mutationRequest
	task             *core.Task
	err              error
	deleted          ports.DeleteResult
}

func (s *Session) Mutate(ctx context.Context, request mutationRequest) (*core.Task, error) {
	value, err := s.Call(ctx, true, func(ctx context.Context, svc ports.TaskService) (any, error) {
		var task *core.Task
		var err error
		switch request.kind {
		case mutationCreate:
			task, err = svc.CreateTask(ctx, request.create)
		case mutationEdit:
			task, err = svc.UpdateTask(ctx, request.update)
		case mutationComplete:
			task, err = svc.CompleteTask(ctx, request.task)
		case mutationReopen:
			task, err = svc.ReopenTask(ctx, ports.ReopenTaskCommand{ID: request.task.ID, Base: request.task.Base, Status: core.StatusTodo})
		default:
			return nil, ports.ErrInvalidRecord
		}
		if task == nil {
			return nil, err
		}
		return task, err
	})
	if err != nil {
		return nil, err
	}
	return value.(*core.Task), nil
}

func (m *Model) startRead() {
	m.readContext, m.readCancel = context.WithCancel(m.options.Context)
	m.readInterrupted = false
}

// Finish resource ownership before checking display freshness. A requested save
// can cancel a read, but it cannot enter the service until that read returns.
func (m *Model) completeRead(err error) (tea.Cmd, bool) {
	m.busy = false
	if m.readCancel != nil {
		m.readCancel()
		m.readCancel = nil
	}
	interrupted := m.readInterrupted
	m.readInterrupted = false
	if IsUnknown(err) {
		m.freezeWrites()
		return nil, true
	}
	if errors.Is(err, errRuntime) {
		m.exitErr = errRuntime
		return tea.Quit, true
	}
	if m.pendingMutation != nil {
		if err == nil || (interrupted && errors.Is(err, context.Canceled)) {
			return m.dispatchMutation(), true
		}
		m.stale = true
		m.abandonMutation("Refresh failed. Ctrl+R refresh before saving.")
	}
	return nil, false
}

func (m *Model) freezeWrites() {
	m.helpOpen = false
	m.filters = nil
	m.searching = false
	m.cancelSearchTimer()
	m.recoveryNeeded = true
	m.stale = true
	m.refreshPending = false
	m.history.pending = false
	m.history.events = nil
	m.history.state = loading
	if m.confirmation != nil {
		m.confirmation.pending = false
		m.confirmation.preview = nil
	}
	m.abandonMutation("Outcome unknown. Draft retained; reload is required.")
	m.beginRecovery()
}

func (m *Model) abandonMutation(message string) {
	m.pendingMutation = nil
	m.saving = false
	m.notice = message
	if m.form != nil {
		m.form.saving = false
		m.form.err = message
	}
}

func (m *Model) canWrite() bool {
	return m.state == loaded && !m.stale && !m.recoveryNeeded && !m.awaitingRead && !m.saving
}

func (m *Model) submitForm() tea.Cmd {
	f := m.form
	if f.conflict || !m.canWrite() {
		return nil
	}
	request, field, message := f.draft.command(m.now, m.options.Location, m.options.ParseDue)
	if message != "" {
		f.err = message
		m.focusForm(field)
		return nil
	}
	if request.kind == mutationNone {
		m.closeForm()
		m.notice = "No changes"
		return nil
	}
	if m.options.Mutate == nil {
		f.err = "Saving is unavailable."
		return nil
	}
	f.err = ""
	f.saving = true
	return m.admitMutation(request)
}

func (m *Model) toggleTask() tea.Cmd {
	task := m.selectedTask()
	if task == nil || !m.canWrite() || m.options.Mutate == nil {
		return nil
	}
	base := task.Clone()
	request := mutationRequest{kind: mutationComplete, task: ports.TaskCommand{ID: base.ID, Base: &base}}
	if base.Status == core.StatusDone {
		request.kind = mutationReopen
	}
	return m.admitMutation(request)
}

func (m *Model) admitMutation(request mutationRequest) tea.Cmd {
	m.saving = true
	m.pendingMutation = &request
	m.notice = "Saving…"
	if m.busy {
		m.readInterrupted = true
		if m.readCancel != nil {
			m.readCancel()
		}
		return nil
	}
	return m.dispatchMutation()
}

func (m *Model) dispatchMutation() tea.Cmd {
	request := *m.pendingMutation
	m.pendingMutation = nil
	m.busy = true
	m.operation++
	m.refreshPending = false
	owner, operation, ctx, mutate := m.owner, m.operation, m.options.Context, m.options.Mutate
	remove := m.options.Delete
	return safeCommand(func() tea.Msg {
		if request.kind == mutationDelete {
			result, err := remove(ctx, request.deletion)
			return mutationMsg{owner: owner, operation: operation, request: request, deleted: result, err: err}
		}
		task, err := mutate(ctx, request)
		return mutationMsg{owner: owner, operation: operation, request: request, task: task, err: err}
	})
}

func mutationError(err error) string {
	switch {
	case errors.Is(err, ports.ErrConflict):
		return "Task changed elsewhere. Keep this draft or reload the task."
	case errors.Is(err, core.ErrTaskNotFound):
		return "Task or parent is missing. Reload before saving."
	case errors.Is(err, core.ErrCyclicDependency):
		return "That parent would create a cycle. Choose another parent."
	case errors.Is(err, core.ErrMaxDepthExceeded):
		return "That parent exceeds the nesting limit."
	case errors.Is(err, ports.ErrBusy):
		return "Database busy. Your draft is safe; save again to retry."
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return "Save did not finish. Your draft is safe; save again to retry."
	default:
		return "Could not save. Your draft is safe; review fields and retry."
	}
}

func (m *Model) acceptMutation(msg mutationMsg) tea.Cmd {
	if msg.owner != m.owner || msg.operation != m.operation || !m.busy || !m.saving {
		return nil
	}
	m.busy = false
	m.saving = false
	if IsUnknown(msg.err) {
		request := msg.request
		m.uncertain = &request
		m.freezeWrites()
		return nil
	}
	if msg.request.kind == mutationDelete {
		return m.deleteFinished(msg)
	}
	f := m.form
	if f != nil && f.draft.id == msg.request.formID {
		f.saving = false
	} else {
		f = nil
	}
	if msg.err != nil {
		m.notice = mutationError(msg.err)
		if f == nil {
			m.notice = "Action failed; refresh before retrying"
			if errors.Is(msg.err, ports.ErrConflict) || errors.Is(msg.err, core.ErrTaskNotFound) {
				m.notice = "Task changed; refresh before retrying"
			}
		}
		if f != nil {
			f.err = m.notice
			if errors.Is(msg.err, core.ErrTaskNotFound) && f.draft.base == nil {
				f.err = "Parent is missing. Choose another parent or clear it."
				m.focusForm(fieldParent)
			} else if errors.Is(msg.err, ports.ErrConflict) || errors.Is(msg.err, core.ErrTaskNotFound) {
				f.conflict = true
				f.prompt = promptConflict
				f.confirm = false
			}
		}
		return nil
	}
	if f != nil {
		m.closeForm()
	}
	m.notice = "Saved"
	m.committedKind = msg.request.kind
	m.awaitingRead = true
	return m.requestRefresh()
}

func (m *Model) writeNotice() string {
	if m.committedKind == mutationDelete {
		return "Deleted"
	}
	return "Saved"
}

func (m *Model) reloadForm() {
	f := m.form
	if f == nil || !f.reloading {
		return
	}
	f.reloading = false
	var current *core.Task
	children := false
	var walk func([]*core.TaskNode)
	walk = func(nodes []*core.TaskNode) {
		for _, n := range nodes {
			if n.Task.ID == f.draft.base.ID {
				current = &n.Task
				children = len(n.Children) > 0
			}
			walk(n.Children)
		}
	}
	walk(m.forest)
	if current == nil {
		f.err = "Task is missing. Draft retained; Esc discards."
		return
	}
	m.formSequence++
	f.draft = newTaskDraft(current, children, m.options.Location)
	f.draft.id = m.formSequence
	for i := range f.inputs {
		m.loadFormField(f, i)
	}
	f.readonly[fieldProgress] = f.draft.progressLocked
	f.conflict = false
	f.err = ""
	m.focusForm(fieldTitle)
}
