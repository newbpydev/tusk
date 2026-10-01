package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func TestMutation_DrainDuplicateAndReadbackBarrier(t *testing.T) {
	m := loadedModel()
	writes := 0
	m.options.Mutate = func(_ context.Context, r mutationRequest) (*core.Task, error) {
		writes++
		if r.create.Title != "New task" {
			t.Fatal("wrong request")
		}
		return &core.Task{ID: "created"}, nil
	}
	press(m, "a")
	press(m, "New task")
	read := m.requestRefresh()
	ctx := m.readContext
	cmd := press(m, "ctrl+s")
	operation := m.operation
	m.Update(tea.WindowSizeMsg{Width: 79, Height: 23})
	if !m.saving || ctx.Err() != context.Canceled || writes != 0 {
		t.Fatal("write before read drained")
	}
	press(m, "ctrl+s")
	press(m, "enter")
	press(m, "esc")
	if m.form == nil || !m.form.saving {
		t.Fatal("saving form escaped")
	}
	if m.operation != operation {
		t.Fatal("resize changed admitted operation")
	}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd != nil {
		deliverUI(m, cmd)
	}
	msg := read().(forestMsg)
	msg.err = context.Canceled
	_, cmd = m.Update(msg)
	// Extract one mutation without also consuming its post-write refresh.
	completion := findMutation(t, cmd)
	if writes != 1 {
		t.Fatal(writes)
	}
	_, readback := m.Update(completion)
	if m.form != nil || !m.awaitingRead {
		t.Fatal("commit not acknowledged")
	}
	press(m, "a")
	press(m, "x")
	if m.form != nil || writes != 1 {
		t.Fatal("readback barrier bypassed")
	}
	m.options.Load = func(context.Context) ([]*core.TaskNode, error) { return nil, ports.ErrStorage }
	// The already captured read is successful: replace its result with a failure.
	forest := findForest(t, readback)
	forest.err = ports.ErrStorage
	m.Update(forest)
	if !m.awaitingRead || !m.stale || m.notice != "Saved; refresh failed" {
		t.Fatal("saved refresh failure")
	}
	m.options.Load = func(context.Context) ([]*core.TaskNode, error) { return nil, nil }
	deliverUI(m, m.requestRefresh())
	if m.awaitingRead || m.stale || writes != 1 || m.notice != "Saved" {
		t.Fatal("replayed commit")
	}
}

func findMutation(t *testing.T, cmd tea.Cmd) mutationMsg {
	t.Helper()
	if cmd == nil {
		t.Fatal("no mutation command")
	}
	msg := cmd()
	if v, ok := msg.(mutationMsg); ok {
		return v
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c == nil {
				continue
			}
			v := c()
			if r, ok := v.(mutationMsg); ok {
				return r
			}
			if inner, ok := v.(tea.BatchMsg); ok {
				return findMutation(t, tea.Batch(inner...))
			}
		}
	}
	t.Fatalf("mutation missing: %T", msg)
	return mutationMsg{}
}
func findForest(t *testing.T, cmd tea.Cmd) forestMsg {
	t.Helper()
	msg := cmd()
	if v, ok := msg.(forestMsg); ok {
		return v
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			v := c()
			if r, ok := v.(forestMsg); ok {
				return r
			}
		}
	}
	t.Fatalf("forest missing: %T", msg)
	return forestMsg{}
}

func TestMutation_KnownConflictAndUnknownRetainDraft(t *testing.T) {
	for _, failure := range []error{ports.ErrBusy, context.DeadlineExceeded, ports.ErrConflict, core.ErrTaskNotFound, ports.NewTransactionError("write", ports.ErrConflict)} {
		m := loadedModel(fixtureNode("task", "Before", 2, nil))
		calls := 0
		m.options.Mutate = func(context.Context, mutationRequest) (*core.Task, error) { calls++; return nil, failure }
		press(m, "e")
		press(m, " changed")
		base := m.form.draft.base.Clone()
		raw := m.form.draft.fields
		deliverUI(m, press(m, "ctrl+s"))
		if m.form == nil || m.form.draft.base.Title != base.Title || m.form.draft.fields != raw || m.form.err == "" {
			t.Fatal("draft lost", failure)
		}
		if IsUnknown(failure) {
			if !m.recoveryNeeded {
				t.Fatal("unknown ignored")
			}
			press(m, "ctrl+s")
			if calls != 1 {
				t.Fatal("unknown replay")
			}
		} else if errors.Is(failure, ports.ErrConflict) || errors.Is(failure, core.ErrTaskNotFound) {
			if !m.form.conflict || m.form.prompt != promptConflict || m.form.confirm {
				t.Fatal("unsafe conflict choice")
			}
			press(m, "enter")
			deliverUI(m, press(m, "ctrl+s"))
			if calls != 1 {
				t.Fatal("silent rebase")
			}
		} else {
			deliverUI(m, press(m, "ctrl+s"))
			if calls != 2 {
				t.Fatal("explicit retry unavailable")
			}
		}
	}
}

func TestMutation_ReadFailureCannotAdmitQueuedWrite(t *testing.T) {
	for _, failure := range []error{ports.ErrStorage, ports.NewTransactionError("read", context.Canceled)} {
		m := loadedModel()
		m.options.Mutate = func(context.Context, mutationRequest) (*core.Task, error) {
			t.Fatal("write after failed read")
			return nil, nil
		}
		press(m, "a")
		press(m, "Draft")
		read := m.requestRefresh()
		press(m, "ctrl+s")
		msg := read().(forestMsg)
		msg.err = failure
		deliverUI(m, func() tea.Msg { return msg })
		if m.saving || m.form == nil || !m.stale {
			t.Fatal("pending write not abandoned")
		}
	}
}

func TestMutation_NoOpAndToggle(t *testing.T) {
	n := fixtureNode("task", "Task", 2, nil)
	m := loadedModel(n)
	calls := 0
	m.options.Mutate = func(_ context.Context, r mutationRequest) (*core.Task, error) {
		calls++
		if r.kind != mutationComplete || r.task.Base == &n.Task {
			t.Fatal("toggle not detached")
		}
		return &n.Task, nil
	}
	press(m, "e")
	deliverUI(m, press(m, "ctrl+s"))
	if calls != 0 || m.form != nil || m.notice != "No changes" {
		t.Fatal("no-op wrote")
	}
	cmd := press(m, "x")
	press(m, "x")
	deliverUI(m, cmd)
	if calls != 1 {
		t.Fatal("duplicate toggle")
	}
	n.Task.Status = core.StatusDone
	m.options.Mutate = func(_ context.Context, r mutationRequest) (*core.Task, error) {
		if r.kind != mutationReopen {
			t.Fatal("wrong toggle")
		}
		return &n.Task, nil
	}
	deliverUI(m, formKey(m, tea.KeySpace))
}

func TestMutation_MissingCreateParentRetainsRetryableDraft(t *testing.T) {
	m := loadedModel()
	m.options.Mutate = func(context.Context, mutationRequest) (*core.Task, error) { return nil, core.ErrTaskNotFound }
	press(m, "a")
	press(m, "Child")
	setFormField(m, fieldParent, "missing")
	deliverUI(m, press(m, "ctrl+s"))
	if m.form.conflict || m.form.field != fieldParent || m.form.draft.base != nil {
		t.Fatal("missing parent treated as missing edit target")
	}
}

func TestMutation_ConflictReloadFailureAndStaleReplies(t *testing.T) {
	m := loadedModel(fixtureNode("id", "Original", 2, nil))
	m.options.Mutate = func(context.Context, mutationRequest) (*core.Task, error) { return nil, ports.ErrConflict }
	press(m, "e")
	press(m, "Mine")
	deliverUI(m, press(m, "ctrl+s"))
	m.options.Load = func(context.Context) ([]*core.TaskNode, error) { return nil, ports.ErrStorage }
	press(m, "tab")
	deliverUI(m, press(m, "enter"))
	if m.form.reloading || !m.form.conflict || m.form.draft.fields[fieldTitle] != "OriginalMine" {
		t.Fatal("reload failure lost draft")
	}
	before := m.operation
	m.Update(mutationMsg{owner: m.owner + 1, operation: before, err: ports.NewTransactionError("old", ports.ErrStorage)})
	if m.recoveryNeeded {
		t.Fatal("old owner froze new state")
	}
}

func TestMutation_HistoryDrainBeforeWrite(t *testing.T) {
	m := loadedModel(fixtureNode("id", "Task", 2, nil))
	m.options.History = func(ctx context.Context, _ string) ([]ports.TaskEvent, error) { return nil, ctx.Err() }
	m.options.Mutate = func(_ context.Context, r mutationRequest) (*core.Task, error) {
		if r.kind != mutationEdit {
			t.Fatal("wrong edit")
		}
		return &core.Task{ID: "id"}, nil
	}
	press(m, "e")
	press(m, " changed")
	// Opening the form starts the pending history read after the seam is installed.
	m.history.pending = true
	read := m.dispatchHistory()
	if read == nil {
		t.Fatal("no history read")
	}
	press(m, "ctrl+s")
	ctx := m.readContext
	if ctx.Err() != context.Canceled {
		t.Fatal("history not canceled")
	}
	deliverUI(m, read)
	if m.saving || m.form != nil || m.awaitingRead {
		t.Fatal("write or refresh not settled")
	}
}

func TestMutation_DisabledWriteHints(t *testing.T) {
	m := loadedModel(fixtureNode("id", "Task", 2, nil))
	m.options.Mutate = func(context.Context, mutationRequest) (*core.Task, error) { return &core.Task{ID: "id"}, nil }
	cmd := press(m, "x")
	if !strings.Contains(m.View(), "Saving… · Ctrl+C") || strings.Contains(m.View(), "Refreshing") {
		t.Fatal("write mislabeled as refresh")
	}
	completion := findMutation(t, cmd)
	_, read := m.Update(completion)
	msg := findForest(t, read)
	msg.err = ports.ErrStorage
	m.Update(msg)
	footer := strings.Split(m.View(), "\n")[m.height-1]
	if !strings.Contains(footer, "writes paused") || strings.Contains(footer, "a new") {
		t.Fatal("disabled writes not explained", footer)
	}
}
