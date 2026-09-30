package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func TestWorkflow_EditConflictReadbackAndQuit(t *testing.T) {
	m, external, session := mutationFixture(t, false)
	m.options.Preview, m.options.Delete, m.options.Recover = session.Preview, session.Delete, session.Recover
	ctx := context.Background()
	key := func(k string) { deliverUI(m, press(m, k)) }
	key("a")
	key("Project")
	key("ctrl+s")
	root := m.selectedTask().Clone()
	key("a")
	key("Research")
	setFormField(m, fieldParent, root.ID)
	setFormField(m, fieldNotes, "Meeting\n界 details")
	key("ctrl+s")
	child := m.rows[1].node.Task.Clone()
	key("/")
	key("Research")
	key("enter")
	if len(m.rows) != 2 || !m.rows[0].context {
		t.Fatal("search lost ancestor context")
	}
	selectTask(t, m, child.ID)
	key("e")
	setFormField(m, fieldDue, "tomorrow")
	setFormField(m, fieldTitle, "My research")
	title := "External research"
	if _, err := external.UpdateTask(ctx, ports.UpdateTaskCommand{ID: child.ID, Base: &child, Title: &title}); err != nil {
		t.Fatal(err)
	}
	key("ctrl+s")
	if m.form == nil || !m.form.conflict || m.form.draft.fields[fieldTitle] != "My research" {
		t.Fatal("conflict did not preserve draft")
	}
	key("tab")
	key("enter")
	if m.form.conflict || m.form.draft.fields[fieldTitle] != title {
		t.Fatal("explicit reload failed")
	}
	setFormField(m, fieldProgress, "60")
	key("ctrl+s")
	key("esc")
	selectTask(t, m, child.ID)
	key("x")
	if got, err := external.GetTask(ctx, child.ID); err != nil || got.Status != core.StatusDone {
		t.Fatal("complete", got, err)
	}
	key("x")
	key("e")
	setFormField(m, fieldParent, "")
	key("ctrl+s")
	if got, err := external.GetTask(ctx, root.ID); err != nil || got.Progress != 0 {
		t.Fatal("move rollup", got, err)
	}
	selectTask(t, m, child.ID)
	key("d")
	if m.confirmation.field != 0 {
		t.Fatal("delete not default Cancel")
	}
	key("tab")
	key("enter")
	if _, err := external.GetTask(ctx, child.ID); !errors.Is(err, core.ErrTaskNotFound) {
		t.Fatal("delete did not persist", err)
	}
	if len(m.forest) != 1 || m.selectedTask().ID != root.ID || m.awaitingRead {
		t.Fatal("readback/neighbor failed")
	}
	cmd := press(m, "q")
	if cmd == nil {
		t.Fatal("missing quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("wrong exit")
	}
	result, err := session.Close(nil)
	if err != nil || !result.HadCommittedChanges || result.OutcomeUnknown {
		t.Fatal("wrong session receipt", result, err)
	}
}

func TestWorkflow_ToggleErrorsHaveActionableBrowseHints(t *testing.T) {
	for _, failure := range []error{ports.ErrConflict, ports.ErrBusy, core.ErrTaskNotFound, context.DeadlineExceeded, ports.ErrStorage} {
		m, external, session := mutationFixture(t, false)
		if _, err := external.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "Action"}); err != nil {
			t.Fatal(err)
		}
		deliverUI(m, m.requestRefresh())
		m.options.Mutate = func(context.Context, mutationRequest) (*core.Task, error) { return nil, failure }
		deliverUI(m, press(m, "x"))
		if strings.Contains(m.notice, "draft") || strings.Contains(m.notice, "save") {
			t.Fatalf("toggle refers to nonexistent form: %s", m.notice)
		}
		footer := strings.Split(ansi.Strip(m.View()), "\n")[23]
		for _, hint := range []string{"r refresh", "? help", "q quit"} {
			if !strings.Contains(footer, hint) {
				t.Fatalf("missing %q in %q", hint, footer)
			}
		}
		m.notice = strings.Repeat("Long external identity notice ", 6)
		m.prepareFrame()
		footer = strings.Split(ansi.Strip(m.View()), "\n")[23]
		if !strings.Contains(footer, "q quit") || !strings.Contains(footer, "…") {
			t.Fatal("long notice hides controls", footer)
		}
		session.Close(nil)
	}
}

func TestWorkflow_AllStatesKeepViewPure(t *testing.T) {
	for _, state := range []string{"loading", "empty", "loaded", "failure", "stale", "form", "filter", "delete", "recovery"} {
		t.Run(state, func(t *testing.T) {
			m := loadedModel(fixtureNode("id", "Example", core.PriorityMedium, nil))
			switch state {
			case "loading":
				m.state = loading
			case "empty":
				m.forest = nil
			case "failure":
				m.state = loadFailed
			case "stale":
				m.stale = true
			case "form":
				m.beginForm(true)
			case "filter":
				m.beginFilters()
			case "delete":
				m.confirmation = &deleteDialog{target: taskRef{id: "id"}, title: "Example", preview: &ports.DeletePreview{IDs: []string{"id"}}}
			case "recovery":
				m.freezeWrites()
			}
			m.prepareFrame()
			before := snapshot(m)
			frame := m.View()
			for range 100 {
				if m.View() != frame {
					t.Fatal("changed frame")
				}
			}
			if snapshot(m) != before {
				t.Fatal("View mutated nested state")
			}
		})
	}
}
