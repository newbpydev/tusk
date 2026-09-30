package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/terminaltext"
)

type deleteDialog struct {
	id                 uint64
	target             taskRef
	title              string
	preview            *ports.DeletePreview
	pending, recursive bool
	field              int
	scroll             int
	err                string
}
type previewMsg struct {
	owner, operation, dialogID uint64
	target                     taskRef
	preview                    ports.DeletePreview
	err                        error
}

func (s *Session) Preview(ctx context.Context, id string) (ports.DeletePreview, error) {
	value, err := s.Call(ctx, false, func(ctx context.Context, svc ports.TaskService) (any, error) { return svc.PreviewDeleteTask(ctx, id) })
	if err != nil {
		return ports.DeletePreview{}, err
	}
	return value.(ports.DeletePreview), nil
}
func (s *Session) Delete(ctx context.Context, c ports.DeleteTaskCommand) (ports.DeleteResult, error) {
	value, err := s.Call(ctx, true, func(ctx context.Context, svc ports.TaskService) (any, error) {
		result, err := svc.DeleteTask(ctx, c)
		if err != nil {
			return nil, err
		}
		if !result.Deleted || result.ID != c.ID || result.DeletedCount == 0 || result.DeletedCount != len(result.DeletedIDs) {
			return nil, ports.NewTransactionError("delete", ports.ErrInvalidRecord)
		}
		return result, nil
	})
	if err != nil {
		return ports.DeleteResult{}, err
	}
	return value.(ports.DeleteResult), nil
}

func (m *Model) beginDelete() {
	task := m.selectedTask()
	if task == nil || !m.canWrite() || m.options.Preview == nil {
		return
	}
	m.formSequence++
	m.confirmation = &deleteDialog{id: m.formSequence, target: taskIdentity(task), title: task.Title, pending: true}
}

func (m *Model) dispatchPreview() tea.Cmd {
	d := m.confirmation
	if d == nil || !d.pending || m.busy || m.recoveryNeeded {
		return nil
	}
	d.pending = false
	m.busy = true
	m.operation++
	m.startRead()
	owner, operation, id, target, ctx, preview := m.owner, m.operation, d.id, d.target, m.readContext, m.options.Preview
	return safeCommand(func() tea.Msg {
		result, err := preview(ctx, target.id)
		return previewMsg{owner, operation, id, target, result, err}
	})
}

func (m *Model) acceptPreview(msg previewMsg) tea.Cmd {
	if msg.owner != m.owner || msg.operation != m.operation || !m.busy {
		return nil
	}
	if cmd, handled := m.completeRead(msg.err); handled {
		return cmd
	}
	d := m.confirmation
	if d == nil || d.id != msg.dialogID || d.target != msg.target {
		return nil
	}
	if errors.Is(msg.err, core.ErrTaskNotFound) {
		m.confirmation = nil
		m.notice = "Task no longer exists"
		return m.requestRefresh()
	}
	if msg.err != nil {
		d.err = "Could not load deletion preview. Press r to retry."
		return nil
	}
	if taskIdentity(&msg.preview.Target) != d.target {
		m.confirmation = nil
		m.notice = "Task identity changed. Review it before deleting."
		return m.requestRefresh()
	}
	preview := clonePreview(msg.preview)
	d.preview = &preview
	d.title = preview.Target.Title
	d.field = 0
	d.scroll = 0
	d.recursive = false
	return nil
}

func clonePreview(p ports.DeletePreview) ports.DeletePreview {
	return ports.DeletePreview{Target: p.Target.Clone(), IDs: append([]string(nil), p.IDs...)}
}

func (m *Model) confirmKey(key tea.KeyMsg) tea.Cmd {
	d := m.confirmation
	if m.saving {
		return nil
	}
	count := 2
	parent := d.preview != nil && len(d.preview.IDs) > 1
	if parent {
		count = 3
	}
	switch key.String() {
	case "esc":
		m.confirmation = nil
	case "tab":
		d.field = (d.field + 1) % count
		d.scroll = 1 << 30
		if d.field == 0 {
			d.scroll = 0
		}
	case "shift+tab":
		d.field = (d.field + count - 1) % count
		d.scroll = 1 << 30
		if d.field == 0 {
			d.scroll = 0
		}
	case "home":
		d.scroll = 0
	case "end":
		d.scroll = 1 << 30
	case "down":
		d.scroll++
	case "up":
		d.scroll--
	case "pgdown":
		d.scroll += max(1, measure(m.width, m.height).modal.height-6)
	case "pgup":
		d.scroll -= max(1, measure(m.width, m.height).modal.height-6)
	case "r":
		if d.preview == nil {
			d.pending = true
			d.field = 0
			d.recursive = false
			d.err = ""
		}
	case " ":
		if parent && d.field == 1 {
			d.recursive = !d.recursive
		}
	case "enter":
		if d.field == 0 {
			m.confirmation = nil
			return nil
		}
		if d.field != count-1 || d.preview == nil || (parent && !d.recursive) || m.options.Delete == nil || !m.canWrite() {
			return nil
		}
		preview := clonePreview(*d.preview)
		return m.admitMutation(mutationRequest{kind: mutationDelete, deleteID: d.id, deletion: ports.DeleteTaskCommand{ID: d.target.id, Expected: &preview, Recursive: d.recursive, Force: false}})
	}
	return nil
}

func (m *Model) deleteFinished(msg mutationMsg) tea.Cmd {
	d := m.confirmation
	if d == nil || d.id != msg.request.deleteID {
		return nil
	}
	if msg.err != nil {
		if errors.Is(msg.err, core.ErrTaskNotFound) {
			m.confirmation = nil
			m.notice = "Task no longer exists"
			return m.requestRefresh()
		}
		d.preview = nil
		d.recursive = false
		d.field = 0
		d.pending = true
		d.err = "Delete did not complete. Review the new preview and confirm again."
		if errors.Is(msg.err, ports.ErrConflict) {
			d.err = "Task or subtree changed. Review the new preview and confirm again."
		}
		return nil
	}
	m.confirmation = nil
	m.committedKind = mutationDelete
	m.notice = "Deleted"
	m.awaitingRead = true
	return m.requestRefresh()
}

func (m *Model) confirmContent(width int) ([]string, string) {
	d := m.confirmation
	lines := []string{"", m.paint("Delete this task?", textColor, true), ""}
	lines = append(lines, strings.Split(ansi.Wrap(terminaltext.Scalar(d.title), width-2, ""), "\n")...)
	lines = append(lines, strings.Split(ansi.Wrap("ID: "+terminaltext.Scalar(d.target.id), width-2, ""), "\n")...)
	lines = append(lines, "", "Cannot undo. Task history will be removed.", "")
	cancel := m.actionButton("Cancel", d.field == 0, false, false)
	if d.preview == nil {
		if d.err != "" {
			lines = append(lines, d.err)
		} else {
			lines = append(lines, "Loading deletion preview…")
		}
		return lines, cancel + " · Esc cancel · r retry"
	}
	total := len(d.preview.IDs)
	taskNoun, descendantNoun := "tasks", "descendants"
	if total == 1 {
		taskNoun = "task"
	}
	if total == 2 {
		descendantNoun = "descendant"
	}
	lines = append(lines, fmt.Sprintf("%d %s total · %d %s", total, taskNoun, total-1, descendantNoun), "")
	if total > 1 {
		checkbox := m.checkbox("Delete entire subtree", d.recursive, d.field == 1)
		lines = append(lines, checkbox, "Space toggles the focused checkbox.")
	}
	deleteField := 1
	if total > 1 {
		deleteField = 2
	}
	remove := m.actionButton("Delete", d.field == deleteField, false, total > 1 && !d.recursive)
	if d.err != "" {
		lines = append(lines, "", d.err)
	}
	if m.saving {
		return lines, "Deleting… · Ctrl+C exits safely"
	}
	return lines, cancel + "  " + remove
}
