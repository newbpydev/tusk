package tui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func TestDelete_DiskLeafSubtreeConflictAndNeighbor(t *testing.T) {
	m, svc, session := mutationFixture(t, false)
	m.options.Preview = session.Preview
	m.options.Delete = session.Delete
	m.options.Recover = session.Recover
	ctx := context.Background()
	root, err := svc.CreateTask(ctx, ports.CreateTaskCommand{Title: "Root", Priority: core.PriorityHigh})
	if err != nil {
		t.Fatal(err)
	}
	child, err := svc.CreateTask(ctx, ports.CreateTaskCommand{Title: "Child", ParentID: &root.ID})
	if err != nil {
		t.Fatal(err)
	}
	other, err := svc.CreateTask(ctx, ports.CreateTaskCommand{Title: "Other"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CompleteTask(ctx, ports.TaskCommand{ID: child.ID})
	if err != nil {
		t.Fatal(err)
	}
	deliverUI(m, m.requestRefresh())
	selectTask(t, m, child.ID)
	deliverUI(m, press(m, "d"))
	press(m, "tab")
	deliverUI(m, press(m, "enter"))
	got, _ := svc.GetTask(ctx, root.ID)
	if got.Progress != 0 {
		t.Fatal("leaf deletion did not roll up parent")
	}
	if _, err = svc.GetTaskHistory(ctx, child.ID); !errors.Is(err, core.ErrTaskNotFound) {
		t.Fatal("deleted task history still accessible")
	}
	child, err = svc.CreateTask(ctx, ports.CreateTaskCommand{Title: "Second child", ParentID: &root.ID})
	if err != nil {
		t.Fatal(err)
	}
	deliverUI(m, m.requestRefresh())
	selectTask(t, m, root.ID)
	deliverUI(m, press(m, "d"))
	added, err := svc.CreateTask(ctx, ports.CreateTaskCommand{Title: "Arrived after preview", ParentID: &root.ID})
	if err != nil {
		t.Fatal(err)
	}
	press(m, "tab")
	formKey(m, tea.KeySpace)
	press(m, "tab")
	deliverUI(m, press(m, "enter"))
	if m.confirmation == nil || m.confirmation.recursive || len(m.confirmation.preview.IDs) != 3 || m.confirmation.field != 0 {
		t.Fatal("membership conflict carried consent")
	}
	for _, id := range []string{root.ID, child.ID, added.ID} {
		if _, err = svc.GetTask(ctx, id); err != nil {
			t.Fatal("partial conflict deletion", id, err)
		}
	}
	press(m, "tab")
	formKey(m, tea.KeySpace)
	press(m, "tab")
	deliverUI(m, press(m, "enter"))
	tasks, err := svc.ListTasks(ctx, ports.TaskQuery{All: true})
	if err != nil || len(tasks) != 1 || tasks[0].ID != other.ID || m.selectedTask().ID != other.ID || m.notice != "Deleted" {
		t.Fatal("wrong deletion/selection", tasks, err)
	}
	for _, id := range []string{root.ID, child.ID, added.ID} {
		if _, err = svc.GetTaskHistory(ctx, id); !errors.Is(err, core.ErrTaskNotFound) {
			t.Fatal("subtree history remains")
		}
	}
}

func TestRecovery_DiskUnknownDeletionObservesAbsence(t *testing.T) {
	m, svc, session := mutationFixture(t, false)
	m.options.Preview = session.Preview
	m.options.Recover = session.Recover
	ctx := context.Background()
	task, err := svc.CreateTask(ctx, ports.CreateTaskCommand{Title: "Uncertain delete"})
	if err != nil {
		t.Fatal(err)
	}
	m.options.Delete = func(ctx context.Context, c ports.DeleteTaskCommand) (ports.DeleteResult, error) {
		_, err := session.Delete(ctx, c)
		if err != nil {
			return ports.DeleteResult{}, err
		}
		// The UI sees an injected unknown outcome after a real acknowledged commit.
		return ports.DeleteResult{}, ports.NewTransactionError("injected", ports.ErrStorage)
	}
	deliverUI(m, m.requestRefresh())
	deliverUI(m, press(m, "d"))
	press(m, "tab")
	deliverUI(m, press(m, "enter"))
	if !m.recoveryNeeded || m.uncertain.kind != mutationDelete || m.uncertain.deletion.ID != task.ID {
		t.Fatal("unknown intent lost")
	}
	deliverUI(m, press(m, "r"))
	if m.recovery.observed != nil || m.recovery.state != loaded || !m.recoveryNeeded {
		t.Fatal("absent target not read back")
	}
	press(m, "tab")
	deliverUI(m, press(m, "enter"))
	if m.confirmation != nil || m.recoveryNeeded || len(m.forest) != 0 {
		t.Fatal("delete intent replayed")
	}
}

func TestRecovery_DiskUnknownCreateNeverInfersIdentity(t *testing.T) {
	m, external, session := mutationFixture(t, false)
	m.options.Recover = session.Recover
	calls := 0
	m.options.Mutate = func(ctx context.Context, r mutationRequest) (*core.Task, error) {
		calls++
		_, err := session.Call(ctx, true, func(ctx context.Context, svc ports.TaskService) (any, error) {
			if _, err := svc.CreateTask(ctx, r.create); err != nil {
				return nil, err
			}
			return nil, ports.NewTransactionError("injected-after-commit", ports.ErrBusy)
		})
		return nil, err
	}
	press(m, "a")
	press(m, "Identical title")
	deliverUI(m, press(m, "ctrl+s"))
	if _, err := external.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "Identical title"}); err != nil {
		t.Fatal(err)
	}
	deliverUI(m, press(m, "r"))
	if len(m.forest) != 2 || m.recovery.target.id != "" || m.recovery.observed != nil || !m.recoveryNeeded {
		t.Fatal("guessed or deduplicated uncertain creation")
	}
	press(m, "tab")
	deliverUI(m, press(m, "enter"))
	if calls != 1 {
		t.Fatal("automatic replay")
	}
	m.options.Mutate = session.Mutate
	press(m, "a")
	press(m, "Independent action after review")
	deliverUI(m, press(m, "ctrl+s"))
	if len(m.forest) != 3 || m.form != nil {
		t.Fatal("fresh owner could not accept an independent action")
	}
	result, _ := session.Close(nil)
	if !result.OutcomeUnknown || !result.HadCommittedChanges {
		t.Fatal("unknown or later acknowledged receipt lost", result)
	}
}
