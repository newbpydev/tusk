package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func deleteModel(children bool) (*Model, *ports.DeletePreview) {
	n := fixtureNode("root", "Delete target", 2, nil)
	ids := []string{"root"}
	if children {
		n.Children = []*core.TaskNode{fixtureNode("child", "Child", 1, nil)}
		ids = append(ids, "child")
	}
	m := loadedModel(n)
	preview := &ports.DeletePreview{Target: n.Task.Clone(), IDs: ids}
	m.options.Preview = func(context.Context, string) (ports.DeletePreview, error) { return *preview, nil }
	return m, preview
}

func TestDelete_DefaultCancelAndSeparateRecursiveIntent(t *testing.T) {
	for _, parent := range []bool{false, true} {
		m, preview := deleteModel(parent)
		calls := 0
		m.options.Delete = func(_ context.Context, c ports.DeleteTaskCommand) (ports.DeleteResult, error) {
			calls++
			if c.Force || c.Recursive != parent || c.Expected == nil || c.Expected.Target.ID != "root" || len(c.Expected.IDs) != len(preview.IDs) {
				t.Fatal("wrong consent", c)
			}
			return ports.DeleteResult{ID: "root", Deleted: true, DeletedCount: len(c.Expected.IDs), DeletedIDs: c.Expected.IDs}, nil
		}
		deliverUI(m, press(m, "d"))
		if m.confirmation == nil || m.confirmation.field != 0 || m.confirmation.recursive {
			t.Fatal("unsafe initial consent")
		}
		for _, key := range []string{"d", "y", "q"} {
			press(m, key)
		}
		press(m, "enter")
		if calls != 0 || m.confirmation != nil {
			t.Fatal("initial Enter deleted")
		}
		deliverUI(m, press(m, "d"))
		press(m, "tab")
		if parent {
			press(m, "tab")
			deliverUI(m, press(m, "enter"))
			if calls != 0 {
				t.Fatal("recursive deletion without checkbox")
			}
			formKey(m, tea.KeyShiftTab)
			formKey(m, tea.KeySpace)
			press(m, "tab")
		}
		cmd := press(m, "enter")
		press(m, "enter")
		press(m, "esc")
		deliverUI(m, cmd)
		if calls != 1 || m.confirmation != nil {
			t.Fatal("duplicate or incomplete deletion")
		}
	}
}

func TestDelete_ChangedMembershipRequiresNewConsent(t *testing.T) {
	m, preview := deleteModel(true)
	calls := 0
	m.options.Delete = func(_ context.Context, c ports.DeleteTaskCommand) (ports.DeleteResult, error) {
		calls++
		if c.Expected.Target.Title != "Delete target" {
			t.Fatal("preview aliased external state")
		}
		preview.IDs = append(preview.IDs, "new-child")
		return ports.DeleteResult{}, ports.ErrConflict
	}
	deliverUI(m, press(m, "d"))
	preview.Target.Title = "Changed externally"
	press(m, "tab")
	formKey(m, tea.KeySpace)
	press(m, "tab")
	deliverUI(m, press(m, "enter"))
	if calls != 1 || m.confirmation == nil || m.confirmation.preview == nil || len(m.confirmation.preview.IDs) != 3 || m.confirmation.field != 0 || m.confirmation.recursive {
		t.Fatal("old consent carried forward")
	}
	press(m, "enter")
	if calls != 1 {
		t.Fatal("replayed delete")
	}
}

func TestDelete_DepartedPreviewCannotArmNewDialog(t *testing.T) {
	m, _ := deleteModel(false)
	first := press(m, "d")
	oldID := m.confirmation.id
	press(m, "esc")
	press(m, "d")
	if m.confirmation.id == oldID {
		t.Fatal("modal ID reused")
	}
	// Resource release starts the latest request; the old payload cannot arm it.
	msg := findPreview(t, first)
	_, next := m.Update(msg)
	if m.confirmation.preview != nil {
		t.Fatal("old preview armed new modal")
	}
	deliverUI(m, next)
	if m.confirmation.preview == nil {
		t.Fatal("latest preview never dispatched")
	}
}

func findPreview(t *testing.T, cmd tea.Cmd) previewMsg {
	t.Helper()
	if cmd == nil {
		t.Fatal("no preview command")
	}
	msg := cmd()
	if p, ok := msg.(previewMsg); ok {
		return p
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c == nil {
				continue
			}
			v := c()
			if p, ok := v.(previewMsg); ok {
				return p
			}
			if b, ok := v.(tea.BatchMsg); ok {
				return findPreview(t, tea.Batch(b...))
			}
		}
	}
	t.Fatalf("missing preview: %T", msg)
	return previewMsg{}
}

func TestDelete_LongConsentScrollAndCanceledReadRefresh(t *testing.T) {
	m, preview := deleteModel(true)
	preview.Target.Title = strings.Repeat("界", 255)
	deliverUI(m, press(m, "d"))
	press(m, "tab")
	if m.confirmation.scroll == 0 || !strings.Contains(m.View(), "Delete entire subtree") {
		t.Fatal("recursive checkbox hidden under long title")
	}
	press(m, "home")
	if m.confirmation.scroll != 0 {
		t.Fatal("cannot review full title")
	}
	press(m, "esc")
	cmd := press(m, "d")
	m.requestRefresh()
	press(m, "esc")
	_, next := m.Update(findPreview(t, cmd))
	if !m.busy || next == nil {
		t.Fatal("queued refresh stranded after closed preview")
	}
	deliverUI(m, next)
}

func TestDelete_PreviewFailureOrMissingNeverWrites(t *testing.T) {
	for _, failure := range []error{ports.ErrBusy, core.ErrTaskNotFound} {
		m, _ := deleteModel(false)
		m.options.Preview = func(context.Context, string) (ports.DeletePreview, error) { return ports.DeletePreview{}, failure }
		m.options.Delete = func(context.Context, ports.DeleteTaskCommand) (ports.DeleteResult, error) {
			t.Fatal("delete without preview")
			return ports.DeleteResult{}, nil
		}
		deliverUI(m, press(m, "d"))
		press(m, "tab")
		deliverUI(m, press(m, "enter"))
		if failure == core.ErrTaskNotFound && m.confirmation != nil {
			t.Fatal("missing target dialog retained")
		}
	}
}

func TestDelete_UnknownAndWrongIncarnationNeverArmConsent(t *testing.T) {
	m, preview := deleteModel(false)
	preview.Target.CreatedAt = preview.Target.CreatedAt.Add(time.Hour)
	deliverUI(m, press(m, "d"))
	if m.confirmation != nil {
		t.Fatal("replacement incarnation armed consent")
	}
	m, _ = deleteModel(false)
	unknown := ports.NewTransactionError("preview", ports.ErrConflict)
	m.options.Preview = func(context.Context, string) (ports.DeletePreview, error) { return ports.DeletePreview{}, &unknown }
	cmd := press(m, "d")
	press(m, "esc")
	deliverUI(m, cmd)
	if !m.recoveryNeeded || m.busy {
		t.Fatal("departed preview uncertainty ignored")
	}
}

func TestDelete_KnownFailureRefreshesPreviewBeforeRetry(t *testing.T) {
	for _, failure := range []error{ports.ErrBusy, core.ErrTaskNotFound} {
		m, _ := deleteModel(false)
		reads := 0
		original := m.options.Preview
		m.options.Preview = func(ctx context.Context, id string) (ports.DeletePreview, error) { reads++; return original(ctx, id) }
		m.options.Delete = func(context.Context, ports.DeleteTaskCommand) (ports.DeleteResult, error) {
			return ports.DeleteResult{}, failure
		}
		deliverUI(m, press(m, "d"))
		press(m, "tab")
		deliverUI(m, press(m, "enter"))
		if failure == ports.ErrBusy {
			if reads != 2 || m.confirmation == nil || m.confirmation.field != 0 {
				t.Fatal("known error reused preview")
			}
		} else if m.confirmation != nil {
			t.Fatal("missing target still armed")
		}
	}
}

func TestDelete_ReadableCountsAndScrollHint(t *testing.T) {
	for _, parent := range []bool{false, true} {
		m, _ := deleteModel(parent)
		deliverUI(m, press(m, "d"))
		want := "1 task total · 0 descendants"
		if parent {
			want = "2 tasks total · 1 descendant"
		}
		if !strings.Contains(m.View(), want) || !strings.Contains(m.View(), "↑↓ scroll") {
			t.Fatal("counts or scroll affordance unclear", m.View())
		}
	}
}
