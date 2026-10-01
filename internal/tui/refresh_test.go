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

func TestRefresh_CoalescesAndKeepsOneTimer(t *testing.T) {
	o := testOptions()
	calls := 0
	o.Load = func(context.Context) ([]*core.TaskNode, error) {
		calls++
		return []*core.TaskNode{fixtureNode("one", "One", core.PriorityLow, nil)}, nil
	}
	m := sizedModel(o)
	first := m.Init()
	for range 50 {
		if press(m, "r") != nil {
			t.Fatal("parallel initial read")
		}
	}
	if calls != 0 {
		t.Fatal("I/O in Update")
	}
	_, cmd := m.Update(first())
	if !m.busy || calls != 1 {
		t.Fatal("pending refresh not admitted")
	}
	batch := cmd().(tea.BatchMsg)
	if len(batch) != 2 {
		t.Fatalf("want timer and single read, got %d", len(batch))
	}
	var tick tickMsg
	for _, c := range batch {
		switch msg := c().(type) {
		case forestMsg:
			m.Update(msg)
		case tickMsg:
			tick = msg
		default:
			t.Fatalf("unexpected %T", msg)
		}
	}
	if m.busy || calls != 2 || m.now != o.Now() {
		t.Fatal("refresh did not settle")
	}
	_, cmd = m.Update(tick)
	if cmd == nil || !m.busy {
		t.Fatal("tick did not refresh")
	}
	if _, duplicate := m.Update(tick); duplicate != nil {
		t.Fatal("duplicate timer chain")
	}
	for range 50 {
		if press(m, "r") != nil {
			t.Fatal("parallel periodic read")
		}
	}
	batch = cmd().(tea.BatchMsg)
	for _, c := range batch {
		if msg, ok := c().(forestMsg); ok {
			_, cmd = m.Update(msg)
		}
	}
	if calls != 3 || cmd == nil {
		t.Fatal("queued refresh lost")
	}
	m.Update(cmd())
	if calls != 4 || m.busy || m.refreshPending {
		t.Fatal("queue was not bounded")
	}
}

func TestRefresh_FailureKeepsSnapshotAndModal(t *testing.T) {
	m := loadedModel(fixtureNode("one", "One", core.PriorityLow, nil))
	m.helpOpen = true
	m.options.Load = func(context.Context) ([]*core.TaskNode, error) { return nil, errors.New("private database path") }
	cmd := m.requestRefresh()
	m.Update(cmd())
	if !m.stale || m.state != loaded || len(m.rows) != 1 || !m.helpOpen || !strings.Contains(m.View(), "Stale") || strings.Contains(m.View(), "private database") {
		t.Fatal("failed refresh discarded snapshot or modal")
	}
	m.options.Load = func(context.Context) ([]*core.TaskNode, error) {
		return []*core.TaskNode{fixtureNode("two", "Two", core.PriorityLow, nil)}, nil
	}
	cmd = m.requestRefresh()
	m.Update(cmd())
	if m.stale || m.rows[0].node.Task.ID != "two" || !m.helpOpen {
		t.Fatal("successful refresh did not recover")
	}
}

func TestRefresh_RejectsOldPayloadButRetiresMatchingOperation(t *testing.T) {
	m := loadedModel(fixtureNode("one", "One", core.PriorityLow, nil))
	cmd := m.requestRefresh()
	reply := cmd().(forestMsg)
	m.generation++
	_, next := m.Update(reply)
	if m.busy || next != nil || m.selectedTask().ID != "one" {
		t.Fatal("stale payload retained busy slot")
	}
	cmd = m.requestRefresh()
	reply = cmd().(forestMsg)
	old := reply
	old.owner++
	m.Update(old)
	if !m.busy {
		t.Fatal("old owner retired active operation")
	}
	m.generation++
	reply.err = ports.NewTransactionError("read", errors.New("cleanup"))
	m.refreshPending = true
	m.Update(reply)
	if m.busy || !m.recoveryNeeded || m.refreshPending || m.requestRefresh() != nil {
		t.Fatal("unknown stale outcome lost or more reads admitted")
	}
}

func TestRefresh_CanceledTickDoesNotRestart(t *testing.T) {
	m := loadedModel()
	token := m.tickToken
	_, cmd := m.Update(tickMsg{token: token, err: context.Canceled, now: time.Now()})
	if cmd != nil || m.busy {
		t.Fatal("canceled timer restarted work")
	}
}
