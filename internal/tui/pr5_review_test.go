package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func TestPR5_StaleFormSaveExplainsPause(t *testing.T) {
	m := loadedModel()
	press(m, "a")
	press(m, "Draft retained")
	m.stale = true
	formKey(m, tea.KeyCtrlS)
	if !strings.Contains(m.form.err, "Ctrl+R") || !strings.Contains(m.View(), "Ctrl+R") || m.saving {
		t.Fatalf("save pause hidden: error=%q saving=%v", m.form.err, m.saving)
	}
	if m.form.draft.fields[fieldTitle] != "Draft retained" {
		t.Fatal("draft lost")
	}
}

func TestPR5_DeleteAbandonedReadCanRefresh(t *testing.T) {
	for _, key := range []tea.KeyMsg{{Type: tea.KeyRunes, Runes: []rune("r")}, {Type: tea.KeyCtrlR}} {
		m, _ := deleteModel(true)
		m.options.Delete = func(context.Context, ports.DeleteTaskCommand) (ports.DeleteResult, error) {
			t.Fatal("replayed delete")
			return ports.DeleteResult{}, nil
		}
		deliverUI(m, press(m, "d"))
		press(m, "tab")
		formKey(m, tea.KeySpace)
		press(m, "tab")
		read := m.requestRefresh()
		press(m, "enter")
		reply := read().(forestMsg)
		reply.err = ports.ErrStorage
		m.Update(reply)
		if !m.stale || m.pendingMutation != nil || m.saving {
			t.Fatal("failed read did not abandon delete")
		}
		_, cmd := m.Update(key)
		if cmd == nil {
			t.Fatal("delete dialog cannot refresh")
		}
		if m.confirmation.preview != nil || m.confirmation.recursive || m.confirmation.field != 0 {
			t.Fatal("refresh retained old consent")
		}
		deliverUI(m, cmd)
		if m.stale || m.confirmation.preview == nil || m.confirmation.recursive {
			t.Fatal("fresh preview did not renew consent")
		}
	}
}

func TestPR5_FailedTogglePausesWritesUntilReadback(t *testing.T) {
	m := loadedModel(fixtureNode("task", "Task", core.PriorityMedium, nil))
	calls := 0
	m.options.Mutate = func(context.Context, mutationRequest) (*core.Task, error) { calls++; return nil, ports.ErrConflict }
	deliverUI(m, press(m, "x"))
	if !m.stale || m.canWrite() {
		t.Fatal("failed action still allows stale writes")
	}
	deliverUI(m, press(m, "x"))
	if calls != 1 {
		t.Fatal("stale toggle replayed")
	}
	deliverUI(m, press(m, "r"))
	if m.stale {
		t.Fatal("successful refresh left writes paused")
	}
}
