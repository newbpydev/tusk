package tui

import (
	"context"
	"crypto/rand"
	"errors"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service"
	"github.com/newbpydev/tusk/internal/service/dateparse"
	"github.com/newbpydev/tusk/internal/storage"
)

func mutationFixture(t *testing.T, policy bool) (*Model, ports.TaskService, *Session) {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tasks.db")
	factory := func(ctx context.Context) (ports.TaskService, func() error, error) {
		repo, err := storage.Open(ctx, storage.Options{Path: path})
		if err != nil {
			return nil, nil, err
		}
		svc, err := service.NewTaskService(repo, service.Options{Clock: time.Now, NewID: func(now time.Time) (string, error) { return service.NewUUIDv7(now, rand.Reader) }, Location: time.UTC, AutoCompleteParent: policy})
		if err != nil {
			repo.Close()
			return nil, nil, err
		}
		return svc, repo.Close, nil
	}
	external, closeExternal, err := factory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeExternal() })
	session := NewSession(ctx, factory)
	t.Cleanup(func() { session.Close(nil) })
	o := testOptions()
	o.Load = session.Load
	o.History = session.History
	o.Mutate = session.Mutate
	o.ParseDue = dateparse.ParseDue
	m := New(o)
	deliverUI(m, m.Init())
	return m, external, session
}

func setFormField(m *Model, field int, value string) {
	m.form.draft.fields[field] = value
	m.loadFormField(m.form, field)
}
func selectTask(t *testing.T, m *Model, id string) {
	t.Helper()
	for i, r := range m.rows {
		if r.node.Task.ID == id {
			m.selected = i
			_, cmd := m.finish(nil)
			deliverUI(m, cmd)
			return
		}
	}
	t.Fatalf("missing row %s", id)
}

func TestMutation_DiskFormsLifecycleAndRollup(t *testing.T) {
	for _, policy := range []bool{false, true} {
		t.Run(map[bool]string{false: "default", true: "autocomplete"}[policy], func(t *testing.T) {
			m, external, session := mutationFixture(t, policy)
			ctx := context.Background()
			key := func(k string) { deliverUI(m, press(m, k)) }
			key("a")
			key("Root")
			key("ctrl+s")
			root := m.selectedTask().Clone()
			if root.Status != core.StatusTodo || root.Priority != core.PriorityMedium || root.Progress != 0 || root.ParentID != nil || root.DueDate != nil || root.Description != "" {
				t.Fatal("create defaults")
			}
			key("a")
			key("Child")
			setFormField(m, fieldNotes, "First\n第二")
			setFormField(m, fieldTags, "WORK, design")
			setFormField(m, fieldDue, "tomorrow")
			setFormField(m, fieldParent, root.ID)
			key("ctrl+s")
			tasks, err := external.ListTasks(ctx, ports.TaskQuery{All: true})
			if err != nil || len(tasks) != 2 {
				t.Fatal(tasks, err)
			}
			var child core.Task
			for _, task := range tasks {
				if task.ID != root.ID {
					child = task
				}
			}
			if child.ParentID == nil || *child.ParentID != root.ID || child.DueDate == nil || child.Description != "First\n第二" || len(child.Tags) != 2 {
				t.Fatal("optional fields", child)
			}
			selectTask(t, m, child.ID)
			key("e")
			setFormField(m, fieldProgress, "99")
			key("ctrl+s")
			got, _ := external.GetTask(ctx, root.ID)
			if got.Progress != 99 {
				t.Fatal("rollup", got)
			}
			key("x")
			got, _ = external.GetTask(ctx, root.ID)
			if got.Progress != 100 || (got.Status == core.StatusDone) != policy {
				t.Fatal("parent completion policy", got)
			}
			key("x")
			got, _ = external.GetTask(ctx, root.ID)
			wantStatus := core.StatusTodo
			if policy {
				wantStatus = core.StatusInProgress
			}
			if got.Progress != 0 || got.Status != wantStatus {
				t.Fatal("reopen rollup", got)
			}
			key("e")
			setFormField(m, fieldParent, "")
			setFormField(m, fieldDue, "")
			setFormField(m, fieldTags, "")
			key("ctrl+s")
			got, _ = external.GetTask(ctx, child.ID)
			if got.ParentID != nil || got.DueDate != nil || len(got.Tags) != 0 {
				t.Fatal("clear intent", got)
			}
			got, _ = external.GetTask(ctx, root.ID)
			if got.Progress != 0 {
				t.Fatal("last child rollup")
			}
			selectTask(t, m, root.ID)
			key("e")
			setFormField(m, fieldParent, child.ID)
			key("ctrl+s")
			selectTask(t, m, child.ID)
			key("e")
			setFormField(m, fieldParent, root.ID)
			key("ctrl+s")
			if m.form == nil || m.form.err != mutationError(core.ErrCyclicDependency) {
				t.Fatal("cycle accepted")
			}
			current, _ := external.GetTask(ctx, child.ID)
			if current.ParentID != nil {
				t.Fatal("partial move")
			}
			key("esc")
			key("tab")
			key("enter")
			selectTask(t, m, child.ID)
			key("x")
			a, _ := external.GetTask(ctx, child.ID)
			b, _ := external.GetTask(ctx, root.ID)
			if a.Status != core.StatusDone || b.Status != core.StatusDone {
				t.Fatal("subtree completion")
			}
			key("x")
			b, _ = external.GetTask(ctx, root.ID)
			if b.Status != core.StatusDone {
				t.Fatal("reopened descendant")
			}
			result, _ := session.Close(nil)
			if !result.HadCommittedChanges || result.OutcomeUnknown {
				t.Fatal("receipt", result)
			}
		})
	}
}

func TestMutation_DiskConflictReloadAndMissing(t *testing.T) {
	m, external, _ := mutationFixture(t, false)
	ctx := context.Background()
	task, err := external.CreateTask(ctx, ports.CreateTaskCommand{Title: "Original", Description: "raw\tstored"})
	if err != nil {
		t.Fatal(err)
	}
	deliverUI(m, m.requestRefresh())
	press(m, "e")
	setFormField(m, fieldTitle, "Mine")
	base := m.form.draft.base.Clone()
	externalTitle := "Theirs"
	_, err = external.UpdateTask(ctx, ports.UpdateTaskCommand{ID: task.ID, Title: &externalTitle, Base: task})
	if err != nil {
		t.Fatal(err)
	}
	deliverUI(m, m.requestRefresh())
	if m.form.draft.base.Title != base.Title || m.form.draft.fields[fieldTitle] != "Mine" {
		t.Fatal("periodic refresh stole draft")
	}
	deliverUI(m, press(m, "ctrl+s"))
	if !m.form.conflict {
		t.Fatal("missing conflict")
	}
	press(m, "enter")
	press(m, "q")
	if m.form.draft.fields[fieldTitle] != "Mine" {
		t.Fatal("quarantined draft changed")
	}
	formKey(m, tea.KeyCtrlR)
	press(m, "tab")
	deliverUI(m, press(m, "enter"))
	if m.form.conflict || m.form.draft.base.Title != "Theirs" || m.form.draft.fields[fieldTitle] != "Theirs" {
		t.Fatal("explicit reload failed")
	}
	setFormField(m, fieldTitle, "Updated")
	deliverUI(m, press(m, "ctrl+s"))
	got, _ := external.GetTask(ctx, task.ID)
	if got.Description != "raw\tstored" || got.Title != "Updated" {
		t.Fatal("untouched raw notes damaged")
	}
	press(m, "e")
	setFormField(m, fieldTitle, "Do not recreate")
	preview, _ := external.PreviewDeleteTask(ctx, task.ID)
	_, err = external.DeleteTask(ctx, ports.DeleteTaskCommand{ID: task.ID, Expected: &preview})
	if err != nil {
		t.Fatal(err)
	}
	deliverUI(m, press(m, "ctrl+s"))
	press(m, "tab")
	deliverUI(m, press(m, "enter"))
	if m.form == nil || !m.form.conflict || m.form.draft.fields[fieldTitle] != "Do not recreate" || m.form.reloading {
		t.Fatal("missing task discarded/recreated draft")
	}
	if _, err = external.GetTask(ctx, task.ID); !errors.Is(err, core.ErrTaskNotFound) {
		t.Fatal("recreated task")
	}
}

type nilMutationService struct{ sessionService }

func (nilMutationService) CreateTask(context.Context, ports.CreateTaskCommand) (*core.Task, error) {
	return nil, nil
}
func TestSession_MutationNilAndInvalidRequest(t *testing.T) {
	s := NewSession(context.Background(), func(context.Context) (ports.TaskService, func() error, error) {
		return nilMutationService{}, func() error { return nil }, nil
	})
	_, err := s.Mutate(context.Background(), mutationRequest{})
	if !errors.Is(err, ports.ErrInvalidRecord) {
		t.Fatal(err)
	}
	_, err = s.Mutate(context.Background(), mutationRequest{kind: mutationCreate})
	if !IsUnknown(err) {
		t.Fatal("nil successful mutation must freeze", err)
	}
	result, _ := s.Close(nil)
	if !result.OutcomeUnknown || result.HadCommittedChanges {
		t.Fatal(result)
	}
}

func TestMutation_DiskDepthAndFailedMoveAreAtomic(t *testing.T) {
	m, svc, _ := mutationFixture(t, false)
	ctx := context.Background()
	var parent *string
	for i := 0; i < 10; i++ {
		task, err := svc.CreateTask(ctx, ports.CreateTaskCommand{Title: "Deep task", ParentID: parent})
		if err != nil {
			t.Fatal(err)
		}
		id := task.ID
		parent = &id
	}
	task, err := svc.CreateTask(ctx, ports.CreateTaskCommand{Title: "Moving task"})
	if err != nil {
		t.Fatal(err)
	}
	deliverUI(m, m.requestRefresh())
	selectTask(t, m, task.ID)
	press(m, "e")
	setFormField(m, fieldParent, *parent)
	deliverUI(m, press(m, "ctrl+s"))
	if m.form == nil || m.form.err != mutationError(core.ErrMaxDepthExceeded) {
		t.Fatal("depth accepted")
	}
	got, _ := svc.GetTask(ctx, task.ID)
	history, _ := svc.GetTaskHistory(ctx, task.ID)
	if got.ParentID != nil || len(history) != 1 {
		t.Fatal("failed move partially persisted")
	}
	// Service value equality permits an independently restored editable value;
	// the adapter must pass the original detached Base without extra versions.
	title := "Temporary"
	_, err = svc.UpdateTask(ctx, ports.UpdateTaskCommand{ID: task.ID, Title: &title})
	if err != nil {
		t.Fatal(err)
	}
	title = task.Title
	_, err = svc.UpdateTask(ctx, ports.UpdateTaskCommand{ID: task.ID, Title: &title})
	if err != nil {
		t.Fatal(err)
	}
	setFormField(m, fieldParent, "")
	setFormField(m, fieldTitle, "Accepted after restoration")
	deliverUI(m, press(m, "ctrl+s"))
	got, _ = svc.GetTask(ctx, task.ID)
	if m.form != nil || got.Title != "Accepted after restoration" {
		t.Fatal("adapter invented a version check")
	}
}
