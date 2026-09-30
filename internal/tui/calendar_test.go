package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/service/dateparse"
)

func openDueCalendar(t *testing.T, m *Model, value string) {
	t.Helper()
	press(m, "a")
	press(m, "Calendar task")
	for range 3 {
		press(m, "tab")
	}
	if value != "" {
		press(m, value)
	}
	formKey(m, tea.KeyCtrlP)
	if !strings.Contains(ansi.Strip(m.View()), "Choose due date") {
		t.Fatal("due field did not open a calendar")
	}
}

func TestCalendar_ChooseAndCancelPreserveDraft(t *testing.T) {
	m := loadedModel()
	m.options.ParseDue = dateparse.ParseDue
	openDueCalendar(t, m, "2026-10-15T17:30:00.123456789Z")
	before := m.form.draft.fields
	for _, k := range []string{"right", "pgdown", "q", "d", "ctrl+s", "tab"} {
		press(m, k)
	}
	if m.form.draft.fields != before || m.form.saving || m.helpOpen {
		t.Fatal("calendar keys changed/submitted the draft or leaked to browsing")
	}
	press(m, "esc")
	if m.form.field != fieldDue || m.form.draft.fields != before || m.form.prompt != promptNone {
		t.Fatal("calendar cancellation changed the draft or focus")
	}
	for _, hint := range []string{"Due date · e.g. 2026-10-15 or tomorrow", "Ctrl+P calendar", "Ctrl+U clear", "UTC"} {
		if !strings.Contains(ansi.Strip(m.View()), hint) {
			t.Error("missing due-field guidance", hint)
		}
	}
	formKey(m, tea.KeyCtrlP)
	press(m, "right")
	press(m, "enter")
	if m.form.draft.fields[fieldDue] != "2026-10-16" || m.form.inputs[fieldDue].Value() != "2026-10-16" || m.form.field != fieldDue || m.form.saving {
		t.Fatal("calendar selection did not return an ISO day to the editor")
	}
	formKey(m, tea.KeyCtrlU)
	if m.form.draft.fields[fieldDue] != "" {
		t.Fatal("clear selected day")
	}
}

func TestCalendar_CivilNavigation(t *testing.T) {
	for _, tc := range []struct{ start, keys, want string }{
		{"2024-01-31", "pgdown", "2024-02-29"},
		{"2025-01-31", "pgdown", "2025-02-28"},
		{"2026-03-31", "pgup", "2026-02-28"},
		{"2026-12-31", "right", "2027-01-01"},
		{"2026-01-01", "left", "2025-12-31"},
		{"2026-09-29", "up down down", "2026-10-06"},
		{"2026-10-15", "t", "2026-09-29"},
		{"tomorrow", "", "2026-09-30"},
		{"invalid", "", "2026-09-29"},
		{"", "", "2026-09-29"},
		{"0001-01-01", "left up pgup", "0001-01-01"},
		{"9999-12-31", "right down pgdown", "9999-12-31"},
	} {
		t.Run(tc.start+"/"+tc.keys, func(t *testing.T) {
			m := loadedModel()
			m.options.ParseDue = dateparse.ParseDue
			openDueCalendar(t, m, tc.start)
			for _, key := range strings.Fields(tc.keys) {
				press(m, key)
			}
			press(m, "enter")
			if got := m.form.draft.fields[fieldDue]; got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCalendar_LocalTodayAndDST(t *testing.T) {
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ value, keys, want string }{
		{"", "", "2026-09-29"},
		{"2026-09-30T01:30:00Z", "", "2026-09-29"},
		{"2026-03-08", "left right right", "2026-03-09"},
		{"2026-11-01", "left right right", "2026-11-02"},
	} {
		m := loadedModel()
		m.options.Location, m.options.ParseDue = zone, dateparse.ParseDue
		m.options.Now = func() time.Time { return time.Date(2026, 9, 30, 1, 0, 0, 0, time.UTC) }
		openDueCalendar(t, m, tc.value)
		for _, key := range strings.Fields(tc.keys) {
			press(m, key)
		}
		press(m, "enter")
		if got := m.form.draft.fields[fieldDue]; got != tc.want {
			t.Fatalf("%s: got %s, want %s", tc.value, got, tc.want)
		}
	}
}

func TestCalendar_LayoutPurityAndResize(t *testing.T) {
	for _, profile := range []termenv.Profile{termenv.Ascii, termenv.TrueColor} {
		m := loadedModel()
		m.options.Profile = profile
		m.renderer.SetColorProfile(profile)
		m.options.ParseDue = dateparse.ParseDue
		openDueCalendar(t, m, "2026-08-31") // six rows in a Monday-first grid
		for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 24}, {Width: 120, Height: 40}, {Width: 200, Height: 60}} {
			m.Update(size)
			frame, before := m.View(), snapshot(m)
			for range 3 {
				if m.View() != frame || snapshot(m) != before {
					t.Fatal("calendar View mutated state")
				}
			}
			plain := ansi.Strip(frame)
			for _, hint := range []string{"August 2026", "Mon", "Sun", "[31]", "2026-08-31", "PgUp/PgDn", "Enter", "Esc"} {
				if !strings.Contains(plain, hint) {
					t.Errorf("missing visible calendar content %q at %+v", hint, size)
				}
			}
			if len(strings.Split(frame, "\n")) != size.Height {
				t.Fatal("wrong frame height")
			}
			for _, line := range strings.Split(frame, "\n") {
				if ansi.StringWidth(line) != size.Width {
					t.Fatal("wrong frame width")
				}
			}
			if profile == termenv.Ascii && strings.ContainsRune(frame, '\x1b') {
				t.Fatal("plain calendar emitted escapes")
			}
		}
		m.Update(tea.WindowSizeMsg{Width: 79, Height: 23})
		for _, k := range []string{"right", "enter", "esc", "ctrl+s"} {
			press(m, k)
		}
		m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		if !strings.Contains(ansi.Strip(m.View()), "[31]") {
			t.Fatal("hidden calendar accepted input or resize lost selection")
		}
		press(m, "enter")
		if m.form.draft.fields[fieldDue] != "2026-08-31" {
			t.Fatal("resize changed selection")
		}
	}
}

func TestCalendar_RealStorageSaveAndUnchangedEdit(t *testing.T) {
	m, external, session := mutationFixture(t, false)
	defer session.Close(nil)
	openDueCalendar(t, m, "2026-10-15")
	press(m, "right")
	press(m, "enter")
	deliverUI(m, press(m, "ctrl+s"))
	if m.form != nil || m.selectedTask() == nil {
		t.Fatal("calendar date did not save")
	}
	task, err := external.GetTask(context.Background(), m.selectedTask().ID)
	if err != nil || task.DueDate == nil || task.DueDate.Format(time.RFC3339Nano) != "2026-10-16T23:59:59.999999999Z" {
		t.Fatal("selected day did not use existing end-of-day semantics", task, err)
	}
	press(m, "e")
	setFormField(m, fieldTitle, "Title only")
	m.focusForm(fieldDue)
	before := m.form.draft.fields[fieldDue]
	formKey(m, tea.KeyCtrlP)
	press(m, "right")
	press(m, "esc")
	if m.form.draft.fields[fieldDue] != before {
		t.Fatal("cancel replaced stored timestamp")
	}
	deliverUI(m, press(m, "ctrl+s"))
	updated, err := external.GetTask(context.Background(), task.ID)
	if err != nil || !updated.DueDate.Equal(*task.DueDate) || updated.Title != "Title only" {
		t.Fatal("unchanged due date was reinterpreted", updated, err)
	}
}
