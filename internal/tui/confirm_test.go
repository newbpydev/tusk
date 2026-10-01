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

func TestDelete_RetryKeyRedispatchesPreview(t *testing.T) {
	m, _ := deleteModel(false)
	reads := 0
	original := m.options.Preview
	m.options.Preview = func(ctx context.Context, id string) (ports.DeletePreview, error) {
		reads++
		if reads == 1 {
			return ports.DeletePreview{}, ports.ErrBusy
		}
		return original(ctx, id)
	}
	m.options.Delete = func(context.Context, ports.DeleteTaskCommand) (ports.DeleteResult, error) {
		t.Fatal("delete before a retried preview armed consent")
		return ports.DeleteResult{}, nil
	}
	deliverUI(m, press(m, "d"))
	d := m.confirmation
	if d == nil || d.preview != nil || d.pending || d.err == "" {
		t.Fatal("failed preview did not surface retry state", d)
	}
	press(m, "tab")
	if d.field != 1 {
		t.Fatal("fixture: focus did not move off cancel")
	}
	cmd := press(m, "r")
	if d.pending || d.field != 0 || d.recursive || d.err != "" || reads != 1 || !m.busy {
		t.Fatal("retry key did not redispatch a fresh preview", d)
	}
	deliverUI(m, cmd)
	if reads != 2 || d.pending || d.preview == nil || d.field != 0 || d.recursive || d.err != "" {
		t.Fatal("retried preview did not rearm the dialog", reads, d)
	}
	if !strings.Contains(m.View(), "1 task total") {
		t.Fatal("retried preview body not rendered")
	}
}

func TestDelete_ScrollKeysAndTabWrapResetConsent(t *testing.T) {
	m, preview := deleteModel(true)
	preview.Target.Title = strings.Repeat("界", 255)
	deliverUI(m, press(m, "d"))
	d := m.confirmation
	area := measure(m.width, m.height).modal.height - 6
	page := max(1, area)
	content, _ := m.confirmContent(measure(m.width, m.height).modal.width - 6)
	maxScroll := max(0, len(content)-area)
	if maxScroll == 0 {
		t.Fatal("fixture: long title fits without scrolling")
	}
	press(m, "down")
	if d.scroll != 1 {
		t.Fatal("down did not scroll one line", d.scroll)
	}
	press(m, "up")
	press(m, "up")
	if d.scroll != 0 {
		t.Fatal("up did not reverse and clamp at the top", d.scroll)
	}
	press(m, "pgdown")
	if d.scroll != min(page, maxScroll) {
		t.Fatal("pgdown did not scroll one modal page", d.scroll)
	}
	press(m, "pgup")
	if d.scroll != 0 {
		t.Fatal("pgup did not scroll back one page", d.scroll)
	}
	press(m, "end")
	if d.scroll != maxScroll {
		t.Fatal("end did not land on the last reviewable line", d.scroll, maxScroll)
	}
	press(m, "home")
	if d.scroll != 0 {
		t.Fatal("home did not return to the top", d.scroll)
	}

	calls := 0
	m.options.Delete = func(_ context.Context, c ports.DeleteTaskCommand) (ports.DeleteResult, error) {
		calls++
		return ports.DeleteResult{ID: "root", Deleted: true, DeletedCount: len(c.Expected.IDs), DeletedIDs: c.Expected.IDs}, nil
	}
	press(m, "tab")
	if d.field != 1 || d.scroll != maxScroll {
		t.Fatal("tab did not drop focus to the checkbox at the bottom", d.field, d.scroll)
	}
	formKey(m, tea.KeySpace)
	deliverUI(m, press(m, "enter"))
	if calls != 0 {
		t.Fatal("checkbox focus reached Delete")
	}
	press(m, "tab")
	if d.field != 2 {
		t.Fatal("fixture: Delete not focused", d.field)
	}
	press(m, "tab")
	if d.field != 0 || d.scroll != 0 {
		t.Fatal("wrapping past Delete did not reset selection and scroll", d.field, d.scroll)
	}
	deliverUI(m, press(m, "enter"))
	if calls != 0 || m.confirmation != nil {
		t.Fatal("wrapped consent deleted instead of canceling")
	}

	deliverUI(m, press(m, "d"))
	d = m.confirmation
	formKey(m, tea.KeyShiftTab)
	if d.field != 2 {
		t.Fatal("shift+tab did not step back to Delete", d.field)
	}
	formKey(m, tea.KeyShiftTab)
	formKey(m, tea.KeyShiftTab)
	if d.field != 0 || d.scroll != 0 {
		t.Fatal("backward wrap did not reset selection and scroll", d.field, d.scroll)
	}
	deliverUI(m, press(m, "enter"))
	if calls != 0 || m.confirmation != nil {
		t.Fatal("backward-wrapped consent deleted instead of canceling")
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
