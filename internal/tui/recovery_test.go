package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	m := sizedModel(o)
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

func findRecovery(t *testing.T, cmd tea.Cmd) recoveryMsg {
	t.Helper()
	if cmd == nil {
		t.Fatal("no recovery command")
	}
	msg := cmd()
	if r, ok := msg.(recoveryMsg); ok {
		return r
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c == nil {
				continue
			}
			v := c()
			if r, ok := v.(recoveryMsg); ok {
				return r
			}
			if b, ok := v.(tea.BatchMsg); ok {
				return findRecovery(t, tea.Batch(b...))
			}
		}
	}
	t.Fatalf("missing recovery message: %T", msg)
	return recoveryMsg{}
}

func TestRecovery_SavedNestedSubtaskDetailRendersParentDueAndCompletion(t *testing.T) {
	due := time.Date(2026, 10, 1, 9, 30, 0, 0, time.UTC)
	completed := time.Date(2026, 9, 28, 18, 5, 0, 0, time.UTC)
	grand := fixtureNode("grand", "Stored subtask", core.PriorityHigh, &due)
	grand.Task.Status = core.StatusDone
	grand.Task.Progress = 100
	grand.Task.CompletedAt = &completed
	grand.Task.Description = "Stored body"
	middle := "middle"
	grand.Task.ParentID = &middle
	root := fixtureNode("root", "Root", core.PriorityMedium, nil, fixtureNode("middle", "Middle", core.PriorityMedium, nil, grand))
	m := loadedModel(root)
	press(m, "down")
	press(m, "down")
	if m.selectedTask().ID != "grand" {
		t.Fatal("fixture: nested subtask not selected", m.selectedTask().ID)
	}
	events := []ports.TaskEvent{
		{TaskID: "grand", Sequence: 4, Kind: ports.EventStatus, OccurredAt: completed},
		{TaskID: "grand", Sequence: 5, Kind: ports.EventProgress, OccurredAt: completed},
	}
	m.options.Recover = func(context.Context, string) (recoverySnapshot, error) {
		return recoverySnapshot{forest: []*core.TaskNode{root}, history: events}, nil
	}
	m.freezeWrites()
	if m.recovery.target.id != "grand" {
		t.Fatal("recovery lost the nested target", m.recovery.target)
	}
	deliverUI(m, press(m, "r"))
	r := m.recovery
	if r.state != loaded || r.observed == nil || r.observed.ID != "grand" {
		t.Fatal("readback did not observe the nested subtask", r)
	}
	lines, _, _ := m.recoveryContent(68, 16)
	all := strings.Join(lines, "\n")
	for _, want := range []string{
		"Parent: middle",
		"Due: " + due.Format(time.RFC3339Nano),
		"Completed: " + completed.Format(time.RFC3339Nano),
		"Status: done · Progress: 100%",
		"2 stored activity events",
		"#5 progress",
	} {
		if !strings.Contains(all, want) {
			t.Fatal("saved detail omitted", want, all)
		}
	}
}

func TestRecovery_ReplacedIncarnationIsFlagged(t *testing.T) {
	n := fixtureNode("id", "Replacement body", core.PriorityMedium, nil)
	m := loadedModel(n)
	replacement := n.Task.Clone()
	replacement.CreatedAt = replacement.CreatedAt.Add(time.Hour)
	m.options.Recover = func(context.Context, string) (recoverySnapshot, error) {
		return recoverySnapshot{forest: []*core.TaskNode{{Task: replacement}}}, nil
	}
	m.freezeWrites()
	deliverUI(m, press(m, "r"))
	lines, _, _ := m.recoveryContent(68, 16)
	if !strings.Contains(strings.Join(lines, "\n"), "A different task now uses this ID.") {
		t.Fatal("replacement incarnation not flagged")
	}
}

func TestRecovery_KeysScrollMoveFocusAndEnterReloads(t *testing.T) {
	n := fixtureNode("id", "Task", core.PriorityMedium, nil)
	m := loadedModel(n)
	m.options.Recover = func(context.Context, string) (recoverySnapshot, error) {
		return recoverySnapshot{forest: []*core.TaskNode{n}, history: []ports.TaskEvent{
			{TaskID: "id", Sequence: 1, Kind: ports.EventCreate, OccurredAt: n.Task.CreatedAt},
			{TaskID: "id", Sequence: 2, Kind: ports.EventStatus, OccurredAt: n.Task.CreatedAt},
			{TaskID: "id", Sequence: 3, Kind: ports.EventProgress, OccurredAt: n.Task.CreatedAt},
		}}, nil
	}
	m.freezeWrites()
	r := m.recovery
	if r.field != 0 || r.state != loadFailed {
		t.Fatal("fixture: reload is not the default focus", r.field, r.state)
	}
	cmd := press(m, "enter")
	findRecovery(t, cmd)
	if !m.busy || r.state != loading {
		t.Fatal("enter on reload did not dispatch the readback")
	}
	deliverUI(m, cmd)
	if r.state != loaded || r.field != 0 {
		t.Fatal("readback result ignored", r.state, r.field)
	}

	area := measure(m.width, m.height).modal.height - 6
	page := max(1, area)
	lines, _, _ := m.recoveryContent(measure(m.width, m.height).modal.width-6, area)
	maxScroll := max(0, len(lines)-area)
	if maxScroll == 0 {
		t.Fatal("fixture: saved detail fits without scrolling")
	}
	press(m, "down")
	if r.scroll != 1 {
		t.Fatal("down did not scroll one line", r.scroll)
	}
	press(m, "up")
	press(m, "up")
	if r.scroll != 0 {
		t.Fatal("up did not reverse and clamp at the top", r.scroll)
	}
	press(m, "pgdown")
	if r.scroll != min(page, maxScroll) {
		t.Fatal("pgdown did not scroll one modal page", r.scroll)
	}
	press(m, "pgup")
	if r.scroll != 0 {
		t.Fatal("pgup did not scroll back one page", r.scroll)
	}
	press(m, "end")
	if r.scroll != maxScroll {
		t.Fatal("end did not land on the last reviewable line", r.scroll, maxScroll)
	}
	press(m, "home")
	if r.scroll != 0 {
		t.Fatal("home did not return to the top", r.scroll)
	}

	formKey(m, tea.KeyShiftTab)
	if r.field != 2 {
		t.Fatal("shift+tab did not step back to quit", r.field)
	}
	formKey(m, tea.KeyShiftTab)
	if r.field != 1 {
		t.Fatal("shift+tab did not reach acknowledge", r.field)
	}
	formKey(m, tea.KeyShiftTab)
	if r.field != 0 {
		t.Fatal("backward wrap did not return to reload", r.field)
	}

	blocked := loadedModel(n)
	blocked.options.Recover = func(context.Context, string) (recoverySnapshot, error) {
		return recoverySnapshot{}, errors.Join(errRetireFailed, ports.ErrStorage)
	}
	blocked.freezeWrites()
	deliverUI(blocked, press(blocked, "r"))
	if !blocked.recovery.blocked {
		t.Fatal("fixture: storage close failure not blocking")
	}
	formKey(blocked, tea.KeyShiftTab)
	press(blocked, "tab")
	if blocked.recovery.field != 0 {
		t.Fatal("blocked readback offered focus changes", blocked.recovery.field)
	}
}

func TestRecovery_UnavailableReloadFreezesWrites(t *testing.T) {
	m := loadedModel(fixtureNode("id", "Task", core.PriorityMedium, nil))
	m.freezeWrites()
	if cmd := press(m, "r"); cmd != nil {
		t.Fatal("dispatched a reload without a recovery seam")
	}
	if !strings.Contains(m.View(), "Reload is unavailable") {
		t.Fatal("CLI guidance missing")
	}
	press(m, "a")
	press(m, "e")
	if m.form != nil || m.canWrite() {
		t.Fatal("write admission unfrozen without a readback")
	}
}
