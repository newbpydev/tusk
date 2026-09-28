package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func TestSanitize_Controls(t *testing.T) {
	raw := "界 e\u0301 👩‍💻 \x1b]52;c;payload\a\r\n\t\x00\x7f\u0085\u2028\u2029\u202e\u2066"
	got := sanitize(raw)
	for _, r := range got {
		if r < 32 || r >= 127 && r <= 159 || r == 0x202e || r == 0x2066 {
			t.Fatalf("active control %U in %q", r, got)
		}
	}
	if !strings.Contains(got, "界 e\u0301 👩‍💻") || !strings.Contains(got, `\u001b`) || !strings.Contains(got, `\n`) {
		t.Fatal(got)
	}
}
func TestFormat_Plain(t *testing.T) {
	task := jsonFixture()
	task.Title = "safe\n\x1b]52;bad\a"
	before := task
	f := newFormatter(TerminalFacts{}, time.UTC, func(string) string { return "" })
	data, err := f.format([]core.Task{task})
	if err != nil {
		t.Fatal(err)
	}
	want := "ID\tSTATUS\tPRIORITY\tPROGRESS\tDUE\tTITLE\nopaque-complete-identifier\ttodo\tmedium\t0%\t-\tsafe\\n\\u001b]52;bad\\u0007\n"
	if string(data) != want || !reflect.DeepEqual(task, before) {
		t.Fatalf("%q", data)
	}
	data, err = f.format([]core.Task(nil))
	if err != nil || string(data) != "ID\tSTATUS\tPRIORITY\tPROGRESS\tDUE\tTITLE\n" {
		t.Fatalf("%q %v", data, err)
	}
	data, err = f.format([]*core.TaskNode(nil))
	if err != nil || string(data) != "No tasks.\n" {
		t.Fatalf("%q %v", data, err)
	}
	data, err = f.format([]ports.TaskEvent(nil))
	if err != nil || string(data) != "SEQUENCE\tOCCURRED_AT\tKIND\tCHANGED_FIELDS\n" {
		t.Fatalf("%q %v", data, err)
	}
}
func TestFormat_CellWidths(t *testing.T) {
	for _, width := range []int{1, 20, 40, 80, 120, 200} {
		task := jsonFixture()
		task.ID = strings.Repeat("id", 30)
		task.Title = "界 e\u0301 👩‍💻 " + strings.Repeat("long", 70)
		task.DueDate = &task.CreatedAt
		f := newFormatter(TerminalFacts{Out: true, Width: width}, time.UTC, func(string) string { return "" })
		node := &core.TaskNode{Task: task, Depth: 1, Children: []*core.TaskNode{{Task: task, Depth: 2}}}
		for _, value := range []any{&task, []core.Task{task}, []*core.TaskNode{node}, ports.TaskStats{}, ports.DeleteResult{ID: task.ID, DeletedIDs: []string{task.ID}, DeletedCount: 1, Deleted: true}, []ports.TaskEvent{{Sequence: 1, OccurredAt: task.CreatedAt, Kind: ports.EventCreate, ChangedFields: []string{"title"}}}} {
			data, err := f.format(value)
			if err != nil {
				t.Fatal(err)
			}
			for _, line := range strings.Split(string(data), "\n") {
				if n := ansi.StringWidth(line); n > width {
					t.Fatalf("width %d got %d: %q", width, n, line)
				}
			}
			if bytes.Contains(data, []byte{0xef, 0xbf, 0xbd}) {
				t.Fatal("broken UTF8")
			}
			again, err := f.format(value)
			if err != nil || !bytes.Equal(data, again) {
				t.Fatal("nondeterministic")
			}
		}
	}
}
func TestTerminal_Capabilities(t *testing.T) {
	for _, tty := range []bool{false, true} {
		for _, noColor := range []string{"", "1"} {
			for _, term := range []string{"xterm-kitty", "dumb"} {
				task := jsonFixture()
				f := newFormatter(TerminalFacts{Out: tty, Width: 120}, time.UTC, func(k string) string {
					if k == "NO_COLOR" {
						return noColor
					}
					return term
				})
				out, err := f.format(&task)
				if err != nil {
					t.Fatal(err)
				}
				colored := bytes.Contains(out, []byte{27})
				if colored != (tty && noColor == "" && term != "dumb") {
					t.Fatalf("color %v tty %v no %q term %q", colored, tty, noColor, term)
				}
			}
		}
	}
	for _, facts := range []TerminalFacts{{Out: true, Width: -1}, {Out: true, Width: 0}, {Out: true, Width: 4, WidthError: io.ErrClosedPipe}} {
		f := newFormatter(facts, nil, nil)
		want := 1
		if facts.WidthError != nil {
			want = 80
		}
		if f.width != want {
			t.Fatalf("width %d", f.width)
		}
	}
}
func TestFormat_OtherValues(t *testing.T) {
	f := newFormatter(TerminalFacts{}, time.UTC, nil)
	data, err := f.format(ports.TaskStats{Total: 3, Done: 1, CompletionPercent: 33, CompletedLast7Days: 1})
	if err != nil || !strings.Contains(string(data), "Completed last 7 days\t1\n") {
		t.Fatalf("%q %v", data, err)
	}
	data, err = f.format(ports.DeleteResult{DeletedIDs: []string{"z", "a"}, DeletedCount: 2})
	if err != nil || string(data) != "Deleted 2 task(s): a, z\n" {
		t.Fatalf("%q %v", data, err)
	}
	task := jsonFixture()
	data, err = f.format([]*core.TaskNode{{Task: task, Depth: 1, Children: []*core.TaskNode{{Task: task, Depth: 2}, {Task: task, Depth: 2}}}})
	if err != nil || !strings.Contains(string(data), "|--") || !strings.Contains(string(data), "`--") {
		t.Fatalf("%q %v", data, err)
	}
	for _, v := range []any{42, (*core.Task)(nil), []*core.TaskNode{nil}} {
		if _, err = f.format(v); err == nil {
			t.Fatalf("accepted %T", v)
		}
	}
}
func TestFormat_Parallel(t *testing.T) {
	for n := range 10 {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			t.Parallel()
			f := newFormatter(TerminalFacts{Out: true, Width: 40 + n}, time.UTC, nil)
			if _, err := f.format([]core.Task{jsonFixture()}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestFormat_KittyPreview(t *testing.T) {
	if os.Getenv("TUSK_FORMAT_PREVIEW") != "1" {
		t.Skip("explicit visual fixture")
	}
	task := jsonFixture()
	task.ID = "01987654-1234-7000-8000-0123456789ab"
	task.Title = "Review Unicode: 界 é 👩‍💻"
	f := newFormatter(TerminalFacts{Out: true, Width: 100}, time.UTC, nil)
	data, err := f.format([]core.Task{task})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Print(string(data))
}

func TestFormat_NoTerminalProbe(t *testing.T) {
	calls := 0
	inv := newInvocation(Options{Terminal: func() TerminalFacts { calls++; return TerminalFacts{Out: true, Width: 120} }})
	if _, err := inv.formatResult([]core.Task{jsonFixture()}, true); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatal("JSON sampled terminal")
	}
	for range 2 {
		if _, err := inv.formatResult([]core.Task{jsonFixture()}, false); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Fatalf("terminal samples %d", calls)
	}
}
func TestFormat_TreeContinuations(t *testing.T) {
	task := jsonFixture()
	task.ID = "id"
	f := newFormatter(TerminalFacts{Out: true, Width: 20}, time.UTC, nil)
	data, err := f.format([]*core.TaskNode{{Task: task, Depth: 1}})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) < 2 || !strings.HasPrefix(lines[1], "    ") {
		t.Fatalf("missing continuation indentation: %q", data)
	}
}

func TestFormat_HeaderAlignment(t *testing.T) {
	f := newFormatter(TerminalFacts{Out: true, Width: 120}, time.UTC, nil)
	data, err := f.format([]core.Task{jsonFixture()})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(ansi.Strip(string(data)), "\n")
	if strings.Index(lines[0], "PROGRESS") != strings.Index(lines[1], "0%") {
		t.Fatalf("misaligned progress column: %q", data)
	}
}

func TestFormat_HistoryNarrowRecords(t *testing.T) {
	f := newFormatter(TerminalFacts{Out: true, Width: 40}, time.UTC, nil)
	data, err := f.format([]ports.TaskEvent{{Sequence: 1, OccurredAt: jsonFixture().CreatedAt, Kind: ports.EventMetadata, ChangedFields: []string{"description", "title"}}})
	if err != nil || !strings.Contains(string(data), "Sequence: 1\n") || !strings.Contains(string(data), "Kind: metadata\n") {
		t.Fatalf("%q %v", data, err)
	}
}
