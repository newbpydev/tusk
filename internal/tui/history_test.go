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

func TestHistory_FailuresRetryAndUnknownPrecedeFreshness(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{{"known", ports.ErrStorage}, {"missing", core.ErrTaskNotFound}, {"unknown", ports.NewTransactionError("read", errors.New("private"))}, {"panic", errRuntime}} {
		t.Run(tc.name, func(t *testing.T) {
			m := loadedModel(fixtureNode("one", "One", 1, nil))
			m.options.History = func(context.Context, string) ([]ports.TaskEvent, error) { return nil, tc.err }
			m.historyRefresh = true
			_, cmd := m.Update(struct{}{})
			msg := cmd().(historyMsg)
			old := msg
			old.owner++
			m.Update(old)
			if !m.busy {
				t.Fatal("wrong owner released active operation")
			}
			if tc.name == "unknown" {
				m.history.token++
			}
			_, next := m.Update(msg)
			switch tc.name {
			case "known":
				if m.busy || m.history.state != loadFailed || next != nil || !strings.Contains(strings.Join(m.detailLines(44), "\n"), "Could not load history") {
					t.Fatal("known failure not visible or retry storm")
				}
				m.options.History = func(context.Context, string) ([]ports.TaskEvent, error) { return nil, nil }
				deliverUI(m, press(m, "r"))
				if m.history.state != loaded || !strings.Contains(strings.Join(m.detailLines(44), "\n"), "No history") {
					t.Fatal("read retry did not recover")
				}
			case "missing":
				if next == nil || !m.busy {
					t.Fatal("missing task did not request forest")
				}
			case "unknown":
				if m.busy || !m.recoveryNeeded || next != nil {
					t.Fatal("obsolete unknown history did not freeze owner")
				}
			case "panic":
				if _, ok := next().(tea.QuitMsg); !ok {
					t.Fatal("runtime error did not quit")
				}
			}
		})
	}
}

func TestHistory_RefreshTakesPriorityAndCoalesces(t *testing.T) {
	m := loadedModel(fixtureNode("one", "One", 1, nil))
	calls := 0
	m.options.History = func(context.Context, string) ([]ports.TaskEvent, error) { calls++; return nil, nil }
	m.historyRefresh = true
	_, history := m.Update(struct{}{})
	for range 30 {
		if press(m, "r") != nil {
			t.Fatal("overlapping service read")
		}
	}
	_, next := m.Update(history())
	if calls != 1 {
		t.Fatal("duplicate history request")
	}
	msg := next()
	if _, ok := msg.(forestMsg); !ok {
		t.Fatalf("forest did not precede queued history: %T", msg)
	}
	_, next = m.Update(msg)
	deliverUI(m, next)
	if m.busy || calls != 2 || m.history.pending || m.refreshPending {
		t.Fatal("refresh/history queue not bounded")
	}
}

func TestSession_HistoryUsesOwnedService(t *testing.T) {
	s := NewSession(context.Background(), func(context.Context) (ports.TaskService, func() error, error) { return nil, nil, ports.ErrStorage })
	if _, err := s.History(context.Background(), "one"); !errors.Is(err, ports.ErrStorage) {
		t.Fatal(err)
	}
	s.Close(nil)
}
