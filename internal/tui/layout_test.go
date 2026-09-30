package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/core"
)

func assertFrameBounds(t *testing.T, frame string, w, h int) {
	t.Helper()
	if w <= 0 || h <= 0 {
		if frame != "" {
			t.Fatalf("zero-sized terminal: %q", frame)
		}
		return
	}
	lines := strings.Split(frame, "\n")
	if len(lines) != h {
		t.Fatalf("got %d rows, want %d", len(lines), h)
	}
	for i, line := range lines {
		if width := ansi.StringWidth(line); width != w {
			t.Fatalf("row %d has %d cells, want %d: %q", i, width, w, line)
		}
	}
}

func TestResize_CellBoundsAndDraftRetention(t *testing.T) {
	m := New(testOptions())
	m.notes.SetValue("Draft 界 é 👩‍💻\nkeep this")
	want := m.notes.Value()
	op := m.operation
	for _, size := range [][2]int{{80, 24}, {120, 40}, {200, 60}, {0, 0}, {1, 1}, {79, 24}, {80, 23}, {-9, -4}, {120, 40}} {
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		assertFrameBounds(t, m.View(), max(0, size[0]), max(0, size[1]))
		if m.notes.Value() != want || m.operation != op {
			t.Fatal("resize changed draft or operation")
		}
		if size[0] >= 80 && size[1] >= 24 {
			rows := strings.Split(ansi.Strip(m.View()), "\n")
			if !strings.Contains(rows[1], "> Tasks ") {
				t.Fatal("panel label must be separated from its border")
			}
			if !strings.Contains(rows[0], "TUSK") || !strings.Contains(rows[len(rows)-1], "q") {
				t.Fatal("missing fixed header/help")
			}
			lw := size[0] * 2 / 5
			if ansi.Cut(rows[1], lw-1, lw+1) != "╮╭" || ansi.Cut(rows[len(rows)-2], lw-1, lw+1) != "╯╰" {
				t.Fatalf("panel bounds drift at width %d", size[0])
			}
		}
		before := snapshot(m)
		frame := m.View()
		for range 100 {
			if m.View() != frame {
				t.Fatal("impure frame")
			}
		}
		if snapshot(m) != before {
			t.Fatal("View mutated state")
		}
	}
}

func TestResize_StateAndUnicodeBounds(t *testing.T) {
	for _, profile := range []termenv.Profile{termenv.Ascii, termenv.ANSI, termenv.ANSI256, termenv.TrueColor} {
		for _, count := range []int{0, 1, 1000} {
			o := testOptions()
			o.Profile = profile
			m := New(o)
			for i := range count {
				m.forest = append(m.forest, &core.TaskNode{Depth: 10, Task: core.Task{ID: fmt.Sprint(i), Title: "界 é 👩‍💻 " + strings.Repeat("長", 100) + "\x1b]52;c;secret\a", Description: "notes\n\ttwo\r\u202e", Status: core.StatusTodo, Priority: core.PriorityHigh, Progress: 25}})
			}
			before := snapshot(m.forest)
			for _, state := range []struct {
				load         loadState
				busy, saving bool
			}{{loading, true, false}, {loaded, false, false}, {loadFailed, false, false}, {loaded, true, false}, {loaded, true, true}} {
				m.state, m.busy, m.saving = state.load, state.busy, state.saving
				m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
				assertFrameBounds(t, m.View(), 80, 24)
				rows := strings.Split(ansi.Strip(m.View()), "\n")
				if ansi.Cut(rows[1], 31, 33) != "╮╭" || ansi.Cut(rows[22], 31, 33) != "╯╰" {
					t.Fatalf("panel geometry shifted in state %+v", state)
				}
				if strings.Contains(m.View(), "\x1b]52") || strings.Contains(m.View(), "\u202e") {
					t.Fatal("terminal injection")
				}
				if profile == termenv.Ascii && strings.Contains(m.View(), "\x1b") {
					t.Fatal("NO_COLOR emitted SGR")
				}
			}
			if snapshot(m.forest) != before {
				t.Fatal("display mutated raw task data")
			}
		}
	}
}

func TestOverlay_FocusAndClip(t *testing.T) {
	m := New(testOptions())
	listFrame := m.View()
	m.Update(tea.KeyMsg{Type: tea.KeyTab})
	before := m.View()
	if before == listFrame || !strings.Contains(before, "> Task details") {
		t.Fatal("focus must be visible without color")
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if cmd != nil || !strings.Contains(m.View(), "Keyboard shortcuts") {
		t.Fatal("help did not open")
	}
	for _, k := range []tea.KeyMsg{{Type: tea.KeyTab}, {Type: tea.KeyShiftTab}, {Type: tea.KeyRunes, Runes: []rune{'a'}}} {
		_, cmd = m.Update(k)
		if cmd != nil || !strings.Contains(m.View(), "Keyboard shortcuts") {
			t.Fatal("modal did not trap input")
		}
		assertFrameBounds(t, m.View(), 80, 24)
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil || m.View() != before {
		t.Fatal("q must dismiss help and restore focus")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m.Update(tea.WindowSizeMsg{Width: 79, Height: 23})
	if !strings.Contains(m.View(), "Resize to 80×24; Ctrl+C quits") {
		t.Fatal("missing resize hint")
	}
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if !strings.Contains(m.View(), "Keyboard shortcuts") {
		t.Fatal("resize lost modal")
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil || m.exitErr != context.Canceled {
		t.Fatal("modal trapped emergency exit")
	}
}

func TestOverlay_GraphemeBoundaries(t *testing.T) {
	for _, tc := range []struct {
		s           string
		left, right int
		want        string
	}{
		{"a界b", 0, 2, "a "}, {"a界b", 2, 4, " b"}, {"👩‍💻Z", 0, 1, " "}, {"é界", 0, 1, "é"},
		{"abc", 2, 2, ""}, {"\x1b[31m界\x1b[0mZ", 1, 3, " Z"},
	} {
		if got := ansi.Strip(cellSlice(tc.s, tc.left, tc.right)); got != tc.want {
			t.Errorf("%q [%d:%d] = %q want %q", tc.s, tc.left, tc.right, got, tc.want)
		}
	}
	m := New(testOptions())
	l := measure(80, 24)
	rows := make([]string, 24)
	for i := range rows {
		rows[i] = strings.Repeat("界", 40)
	}
	content := make([]string, 100)
	for i := range content {
		content[i] = fmt.Sprintf("Field %d: 👩‍💻 é 界", i)
	}
	rows = m.overlay(rows, l, "Edit", content, 85, "Save · Cancel")
	assertFrameBounds(t, strings.Join(rows, "\n"), 80, 24)
	if !strings.Contains(rows[l.modal.y+l.modal.height-3], "Save · Cancel") {
		t.Fatal("scroll hid buttons")
	}
	if !strings.Contains(strings.Join(rows, "\n"), "Field 85") {
		t.Fatal("content did not scroll")
	}
}

func TestOverlay_HelpScrollAndPureViews(t *testing.T) {
	m := New(testOptions())
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	for _, k := range []tea.KeyMsg{{Type: tea.KeyDown}, {Type: tea.KeyUp}, {Type: tea.KeyPgDown}, {Type: tea.KeyPgUp}, {Type: tea.KeyEnd}, {Type: tea.KeyHome}, {Type: tea.KeyRunes, Runes: []rune{'j'}}} {
		m.Update(k)
		assertFrameBounds(t, m.View(), 80, 24)
		before := snapshot(m)
		for range 100 {
			m.View()
		}
		if snapshot(m) != before {
			t.Fatal("help View mutated state")
		}
		if !strings.Contains(m.View(), "Ctrl+C exit") {
			t.Fatal("scroll hid help controls")
		}
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.helpOpen {
		t.Fatal("Esc did not dismiss")
	}
}

func TestTerminal_DeepTreeKeepsTitlesAndRawText(t *testing.T) {
	m := New(testOptions())
	m.state = loaded
	root := &core.TaskNode{Task: core.Task{Title: "Root", Status: core.StatusTodo, Priority: core.PriorityHigh}}
	m.forest = []*core.TaskNode{root}
	last := root
	for range 10 {
		child := &core.TaskNode{Task: core.Task{Title: "界 é", Status: core.StatusTodo}}
		last.Children = []*core.TaskNode{child}
		last = child
	}
	before := snapshot(m.forest)
	m.rebuildRows()
	lines := m.listLines(12, 40)
	if len(lines) != 26 || !strings.Contains(lines[len(lines)-2], "…") || !strings.Contains(lines[len(lines)-2], "界") {
		t.Fatalf("deep title/context lost: %q", lines)
	}
	if !strings.Contains(lines[len(lines)-1], "Todo") {
		t.Fatal("deep metadata lost to indentation")
	}
	for _, line := range lines[3:] {
		if ansi.StringWidth(line) != 12 {
			t.Fatalf("deep line overflow: %q", line)
		}
	}
	if snapshot(m.forest) != before {
		t.Fatal("display mutated source")
	}
	m.notes.SetValue("prepared editor cache")
	m.notes.Focus()
	m.prepareFrame()
	before = snapshot(m)
	m.View()
	if snapshot(m) != before {
		t.Fatal("View mutated editor cache")
	}
}
