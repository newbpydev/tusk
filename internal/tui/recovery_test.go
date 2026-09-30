package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func TestRecovery_UnknownCreateNeverReplays(t *testing.T) {
	m := loadedModel()
	writes, reloads := 0, 0
	m.options.Mutate = func(context.Context, mutationRequest) (*core.Task, error) {
		writes++
		return nil, ports.NewTransactionError("commit", ports.ErrBusy)
	}
	m.options.Recover = func(context.Context, string) (recoverySnapshot, error) {
		reloads++
		return recoverySnapshot{forest: []*core.TaskNode{fixtureNode("external", "Same title", 2, nil)}}, nil
	}
	press(m, "a")
	press(m, "Same title")
	deliverUI(m, press(m, "ctrl+s"))
	if !m.recoveryNeeded || m.form == nil || m.recovery == nil {
		t.Fatal("uncertain draft lost")
	}
	raw := m.form.draft.fields
	oldOwner := m.owner
	cmd := press(m, "r")
	press(m, "r")
	press(m, "ctrl+s")
	deliverUI(m, cmd)
	if reloads != 1 || writes != 1 || m.owner == oldOwner || !m.recoveryNeeded || m.recovery.state != loaded || m.form.draft.fields != raw || m.recovery.target.id != "" {
		t.Fatal("guessed creation identity, replayed, or failed to freeze")
	}
	press(m, "a")
	press(m, "ctrl+s")
	if writes != 1 {
		t.Fatal("write before acknowledgment")
	}
	press(m, "tab")
	deliverUI(m, press(m, "enter"))
	if m.recoveryNeeded || m.form != nil || m.recovery != nil || writes != 1 {
		t.Fatal("acknowledgment replayed intent")
	}
	if !strings.Contains(m.View(), "q quit") {
		t.Fatal("acknowledgment hides quit hint")
	}
	press(m, "a")
	if m.form == nil || m.form.draft.fields[fieldTitle] != "" {
		t.Fatal("old draft reused")
	}
}

func TestRecovery_OldOwnerMessagesIgnored(t *testing.T) {
	m := loadedModel(fixtureNode("task", "Original", 2, nil))
	unknown := ports.NewTransactionError("read", context.Canceled)
	read := m.requestRefresh()
	msg := read().(forestMsg)
	msg.err = errors.Join(ports.ErrConflict, &unknown)
	m.Update(msg)
	oldOwner := m.owner
	m.options.Recover = func(context.Context, string) (recoverySnapshot, error) {
		return recoverySnapshot{forest: []*core.TaskNode{fixtureNode("fresh", "Fresh", 2, nil)}}, nil
	}
	cmd := press(m, "r")
	operation := m.operation
	m.Update(forestMsg{owner: oldOwner, operation: operation, err: unknown})
	m.Update(historyMsg{owner: oldOwner, operation: operation, err: unknown})
	m.Update(mutationMsg{owner: oldOwner, operation: operation, err: unknown})
	if !m.busy || m.operation != operation {
		t.Fatal("old completion stole recovery slot")
	}
	deliverUI(m, cmd)
	if m.selectedTask().ID != "fresh" || m.recovery.state != loaded || !m.recoveryNeeded {
		t.Fatal("late owner contaminated readback")
	}
}

func TestRecovery_FailedReadbackCannotUnfreeze(t *testing.T) {
	m := loadedModel()
	m.freezeWrites()
	calls := 0
	m.options.Recover = func(context.Context, string) (recoverySnapshot, error) {
		calls++
		return recoverySnapshot{}, ports.ErrStorage
	}
	deliverUI(m, press(m, "r"))
	press(m, "tab")
	press(m, "enter")
	if !m.recoveryNeeded || m.recovery.state != loadFailed || calls != 1 {
		t.Fatal("failed readback acknowledged")
	}
}

func TestRecovery_OwnsFocusAboveEveryModal(t *testing.T) {
	for _, key := range []string{"?", "/", "f"} {
		m := loadedModel()
		press(m, key)
		m.freezeWrites()
		m.finish(nil)
		footer := strings.Split(m.View(), "\n")[m.height-1]
		if !strings.Contains(footer, "r reload") || m.helpOpen || m.searching || m.filters != nil {
			t.Fatal("old modal retained recovery focus", key, footer)
		}
	}
}

func TestRecovery_ComparisonIncludesRawAndObservedFields(t *testing.T) {
	n := fixtureNode("id", "Before", 2, nil)
	m := loadedModel(n)
	press(m, "e")
	setFormField(m, fieldDue, "tomorrow")
	setFormField(m, fieldTags, "raw-tag")
	setFormField(m, fieldParent, "raw-parent")
	m.options.Mutate = func(context.Context, mutationRequest) (*core.Task, error) {
		return nil, ports.NewTransactionError("edit", ports.ErrConflict)
	}
	m.options.ParseDue = func(string, time.Time, *time.Location) (time.Time, error) { return time.Now(), nil }
	deliverUI(m, press(m, "ctrl+s"))
	current := n.Task.Clone()
	current.Tags = []core.Tag{"stored-tag"}
	current.Priority = core.PriorityUrgent
	m.options.Recover = func(context.Context, string) (recoverySnapshot, error) {
		return recoverySnapshot{forest: []*core.TaskNode{{Task: current}}, history: []ports.TaskEvent{{TaskID: "id", Sequence: 7, Kind: ports.EventMetadata}}}, nil
	}
	deliverUI(m, press(m, "r"))
	lines, _, _ := m.recoveryContent(68, 16)
	all := strings.Join(lines, "\n")
	for _, want := range []string{"tomorrow", "raw-tag", "raw-parent", "stored-tag", "urgent", "#7"} {
		if !strings.Contains(all, want) {
			t.Fatal("comparison omitted field", want)
		}
	}
}

func TestRecovery_InitialUnknownStartsRefreshTimerAfterReadback(t *testing.T) {
	o := testOptions()
	o.Load = func(context.Context) ([]*core.TaskNode, error) {
		return nil, ports.NewTransactionError("initial", ports.ErrBusy)
	}
	o.Recover = func(context.Context, string) (recoverySnapshot, error) { return recoverySnapshot{}, nil }
	m := New(o)
	deliverUI(m, m.Init())
	if m.timerStarted {
		t.Fatal("initial failure started refresh")
	}
	deliverUI(m, press(m, "r"))
	if !m.timerStarted {
		t.Fatal("recovered session never refreshes")
	}
}

func TestRecovery_ErrorVisibleAboveLongDraft(t *testing.T) {
	m := loadedModel()
	press(m, "a")
	press(m, "Draft")
	setFormField(m, fieldNotes, strings.Repeat("notes\n", 60))
	m.options.Mutate = func(context.Context, mutationRequest) (*core.Task, error) {
		return nil, ports.NewTransactionError("create", ports.ErrBusy)
	}
	m.options.Recover = func(context.Context, string) (recoverySnapshot, error) {
		return recoverySnapshot{}, errors.Join(errRetireFailed, ports.ErrStorage)
	}
	deliverUI(m, press(m, "ctrl+s"))
	deliverUI(m, press(m, "r"))
	if !strings.Contains(m.View(), "Storage could not close") {
		t.Fatal("blocking error hidden below draft")
	}
}

func TestRecovery_ReadbackStateIsClearWithoutRuntimeTerms(t *testing.T) {
	m := loadedModel()
	m.freezeWrites()
	m.finish(nil)
	if strings.Contains(m.View(), "storage owner") {
		t.Fatal("runtime ownership leaked into app flow")
	}
	m.options.Recover = func(context.Context, string) (recoverySnapshot, error) { return recoverySnapshot{}, nil }
	deliverUI(m, press(m, "r"))
	if !strings.Contains(m.View(), "Saved data reloaded") {
		t.Fatal("successful readback has no visible state change")
	}
}

func TestRecovery_OldRenderReleasesLaneWithoutPublishing(t *testing.T) {
	n := fixtureNode("id", "Task", 2, nil)
	m := loadedModel(n)
	n.Task.Description = "Old queued render"
	m.finish(nil)
	old := notesMsg{key: m.markdown.wanted, token: m.markdown.token, lines: []string{"OLD OUTPUT"}}
	fresh := n.Task.Clone()
	fresh.Description = "Fresh notes"
	m.freezeWrites()
	m.options.Recover = func(context.Context, string) (recoverySnapshot, error) {
		return recoverySnapshot{forest: []*core.TaskNode{{Task: fresh}}}, nil
	}
	deliverUI(m, press(m, "r"))
	_, cmd := m.Update(old)
	deliverUI(m, cmd)
	if strings.Contains(strings.Join(m.markdown.lines, "\n"), "OLD OUTPUT") || m.markdown.wanted.generation == old.key.generation || m.markdown.active {
		t.Fatal("old renderer contaminated fresh state or held lane")
	}
}

func TestRecovery_NamesUncertainTaskActionBeforeReload(t *testing.T) {
	for _, kind := range []mutationKind{mutationComplete, mutationReopen, mutationDelete, mutationEdit} {
		n := fixtureNode("full-target-id", "Affected task", 2, nil)
		m := loadedModel(n)
		base := n.Task.Clone()
		m.uncertain = &mutationRequest{kind: kind, task: ports.TaskCommand{ID: base.ID, Base: &base}, update: ports.UpdateTaskCommand{ID: base.ID, Base: &base}, deletion: ports.DeleteTaskCommand{ID: base.ID, Expected: &ports.DeletePreview{Target: base, IDs: []string{base.ID}}}}
		m.freezeWrites()
		lines, _, _ := m.recoveryContent(68, 16)
		content := strings.Join(lines, "\n")
		if !strings.Contains(content, "Affected task") || !strings.Contains(content, "full-target-id") || !strings.Contains(content, "Attempted action:") {
			t.Fatal("uncertain action lacks context", kind, content)
		}
	}
}
