package tui

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

type tuiMeasurement struct {
	Name        string   `json:"name"`
	Run         int      `json:"run"`
	Budget      bool     `json:"preparation_budget"`
	Warmups     []int64  `json:"warmup_ns"`
	Samples     []int64  `json:"samples_ns"`
	Allocations []uint64 `json:"allocations"`
	Bytes       []uint64 `json:"allocated_bytes"`
	P50         int64    `json:"p50_ns"`
	P95         int64    `json:"p95_ns"`
	P99         int64    `json:"p99_ns"`
	Max         int64    `json:"max_ns"`
	Misses      int      `json:"samples_over_16_7ms"`
	Output      int      `json:"consumed_output_size"`
	Passed      bool     `json:"passed"`
}

func summarizeTUI(r *tuiMeasurement) {
	r.Passed = len(r.Warmups) == 5 && len(r.Samples) == 100 && r.Output > 0
	for _, n := range r.Warmups {
		if n <= 0 {
			r.Passed = false
		}
	}
	r.Misses = 0
	ordered := slices.Clone(r.Samples)
	slices.Sort(ordered)
	if len(ordered) == 0 {
		r.Passed = false
		return
	}
	r.P95 = ordered[(len(ordered)*95-1)/100]
	r.P50 = ordered[(len(ordered)*50-1)/100]
	r.P99 = ordered[(len(ordered)*99-1)/100]
	r.Max = ordered[len(ordered)-1]
	for _, n := range r.Samples {
		if n <= 0 {
			r.Passed = false
		}
		if n >= 16_700_000 {
			r.Misses++
		}
	}
	if r.Budget {
		r.Passed = r.Passed && r.P95 < 16_700_000 && r.P99 < 33_300_000 && r.Max < 50_000_000
	}
}

func measureTUI(name string, run int, budget bool, work func() int) tuiMeasurement {
	r := tuiMeasurement{Name: name, Run: run, Budget: budget}
	for i := range 105 {
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		start := time.Now()
		n := work()
		elapsed := time.Since(start).Nanoseconds()
		runtime.ReadMemStats(&after)
		r.Output = n
		if i < 5 {
			r.Warmups = append(r.Warmups, elapsed)
		} else {
			r.Samples = append(r.Samples, elapsed)
			r.Allocations = append(r.Allocations, after.Mallocs-before.Mallocs)
			r.Bytes = append(r.Bytes, after.TotalAlloc-before.TotalAlloc)
		}
		if n <= 0 {
			r.Output = 0
			break
		}
	}
	summarizeTUI(&r)
	return r
}

func benchmarkForest(count int) []*core.TaskNode {
	now := testOptions().Now()
	var roots []*core.TaskNode
	var parent *core.TaskNode
	for i := range count {
		n := fixtureNode(fmt.Sprintf("01990000-0000-7000-8000-%012x", i+1), fmt.Sprintf("Task %05d · 界 café 👩‍💻", i), core.Priority(i%4+1), nil)
		n.Task.UpdatedAt = n.Task.CreatedAt
		if i%11 == 0 {
			switch (i / 11) % 4 {
			case 0:
				due := now
				n.Task.DueDate = &due
			case 1:
				due := now.Add(48 * time.Hour)
				n.Task.DueDate = &due
			case 3:
				n.Task.Status = core.StatusDone
				n.Task.Progress = 100
			}
			roots = append(roots, n)
		} else {
			n.Task.ParentID = &parent.Task.ID
			parent.Children = append(parent.Children, n)
		}
		parent = n
	}
	return roots
}

func benchmarkModel(count, width, height, noteBytes, events int) *Model {
	o := testOptions()
	o.Profile = termenv.TrueColor
	forest := benchmarkForest(count)
	o.Load = func(context.Context) ([]*core.TaskNode, error) { return forest, nil }
	history := make([]ports.TaskEvent, events)
	for i := range history {
		history[i] = ports.TaskEvent{Sequence: int64(i + 1), Kind: ports.EventMetadata, OccurredAt: o.Now(), ChangedFields: []string{"title", "priority"}}
	}
	o.History = func(context.Context, string) ([]ports.TaskEvent, error) { return history, nil }
	m := sizedModel(o)
	m.width = width
	m.height = height
	deliverUI(m, m.Init())
	if task := m.selectedTask(); task != nil {
		const line = "Note **bold** 界 and plain text.\n"
		task.Description = strings.Repeat(line, noteBytes/len(line)+1)[:noteBytes]
		// End on ASCII so the fixture remains valid UTF-8 at its exact byte budget.
		for !strings.HasSuffix(task.Description, "\n") {
			task.Description = task.Description[:len(task.Description)-1]
		}
		task.Description += strings.Repeat(" ", noteBytes-len(task.Description))
		_, cmd := m.finish(nil)
		deliverUI(m, cmd)
	}
	return m
}

func TestTUIMeasurements(t *testing.T) {
	output := os.Getenv("TUSK_TUI_BENCH_OUTPUT")
	if output == "" {
		t.Skip("retained measurements run only through make bench-tui")
	}
	manifest := map[string]string{"go": runtime.Version(), "os": runtime.GOOS, "arch": runtime.GOARCH, "gomaxprocs": strconv.Itoa(runtime.GOMAXPROCS(0)), "started_utc": time.Now().UTC().Format(time.RFC3339Nano), "policy": "3 runs; 5 retained warmups + 100 samples per case; no trimming; preparation p95/p99/max <16.7/33.3/50 ms", "fixture": "depth 10; mixed groups/Unicode; 32 KiB note; 100 history events; TrueColor; service and Markdown excluded from synchronous preparation"}
	for key, path := range map[string]string{"cpu": "/proc/cpuinfo", "kernel": "/proc/sys/kernel/osrelease", "power_profile": "/sys/firmware/acpi/platform_profile", "governor": "/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor"} {
		data, err := os.ReadFile(path)
		if err != nil {
			manifest[key] = "unavailable"
		} else {
			manifest[key] = strings.TrimSpace(string(data))
		}
	}
	for key, args := range map[string][]string{"revision": {"git", "rev-parse", "HEAD"}, "source_diff": {"git", "diff", "--binary", "HEAD"}, "filesystem": {"stat", "-f", "-c", "%T", "."}} {
		data, err := exec.Command(args[0], args[1:]...).Output()
		if err != nil {
			t.Fatal(err)
		}
		if key == "source_diff" {
			manifest["diff_sha256"] = fmt.Sprintf("%x", sha256.Sum256(data))
		} else {
			manifest[key] = strings.TrimSpace(string(data))
		}
	}
	binary := os.Getenv("TUSK_RELEASE_BINARY")
	if binary == "" {
		binary = "../../bin/tusk"
	}
	data, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	manifest["binary_sha256"] = fmt.Sprintf("%x", sha256.Sum256(data))
	list := exec.Command("git", "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	list.Dir = "../.."
	paths, err := list.Output()
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, path := range strings.Split(string(paths), "\x00") {
		if strings.HasSuffix(path, ".go") || path == "Makefile" || path == "go.mod" || path == "go.sum" {
			data, err := os.ReadFile(filepath.Join("../..", path))
			if err != nil {
				t.Fatal(err)
			}
			sources[path] = fmt.Sprintf("%x", sha256.Sum256(data))
		}
	}
	var cases []tuiMeasurement
	for run := 1; run <= 3; run++ {
		for _, count := range []int{0, 100, 1000} {
			for _, size := range [][2]int{{80, 24}, {120, 40}, {200, 60}} {
				m := benchmarkModel(count, size[0], size[1], 32*1024, 100)
				name := fmt.Sprintf("%d-tasks/%dx%d", count, size[0], size[1])
				cases = append(cases, measureTUI("prepare/"+name, run, true, func() int { m.finish(nil); return len(m.View()) }))
				cases = append(cases, measureTUI("view/"+name, run, false, func() int { return len(m.View()) }))
			}
		}
		for _, size := range [][2]int{{80, 24}, {120, 40}, {200, 60}} {
			m := benchmarkModel(1000, size[0], size[1], 32*1024, 100)
			m.beginForm(false)
			m.focusForm(fieldDue)
			m.beginCalendar()
			cases = append(cases, measureTUI(fmt.Sprintf("calendar/1000-tasks/%dx%d", size[0], size[1]), run, true, func() int {
				m.Update(tea.KeyMsg{Type: tea.KeyRight})
				return len(m.View())
			}))
		}
		for _, size := range []int{32 * 1024, 1024 * 1024} {
			const line = "# Notes\nA **bold** line and a link [text](https://example.test).\n"
			source := strings.Repeat(line, size/len(line)+1)
			source = source[:size]
			cases = append(cases, measureTUI(fmt.Sprintf("markdown/%d-bytes", size), run, false, func() int {
				lines, _ := renderNoteText(context.Background(), RenderMarkdown, source, 66, termenv.TrueColor)
				return len(lines)
			}))
		}
		_, external, session := mutationFixture(t, false)
		for i := range 100 {
			if _, err := external.CreateTask(context.Background(), ports.CreateTaskCommand{Title: fmt.Sprintf("Refresh %d", i)}); err != nil {
				t.Fatal(err)
			}
		}
		cases = append(cases, measureTUI("disk-refresh/100-tasks", run, false, func() int {
			nodes, err := session.Load(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			return len(nodes)
		}))
		session.Close(nil)
		stress := benchmarkModel(10000, 120, 40, 1024*1024, 1000)
		cases = append(cases, measureTUI("stress/10000-tasks-1MiB-note-1000-events", run, false, func() int { stress.finish(nil); return len(stress.View()) }))
		stress.beginForm(true)
		cases = append(cases, measureTUI("stress/read-only-form-1MiB", run, false, func() int { stress.finish(nil); return len(stress.View()) }))
	}
	report := struct {
		Manifest map[string]string `json:"manifest"`
		Sources  map[string]string `json:"source_sha256"`
		Cases    []tuiMeasurement  `json:"cases"`
		Passed   bool              `json:"passed"`
	}{manifest, sources, cases, true}
	for _, r := range cases {
		if !r.Passed {
			report.Passed = false
			t.Logf("miss: %s run %d p95=%d p99=%d max=%d", r.Name, r.Run, r.P95, r.P99, r.Max)
		}
	}
	data, err = json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(output, append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	if !report.Passed {
		t.Fatal("TUI preparation budget failed; all samples retained")
	}
}

// One keystroke through the full Update path (finish: projection plus frame
// preparation) at the budgeted 1000-task scale.
func BenchmarkUpdate_Keystroke1000Tasks(b *testing.B) {
	m := benchmarkModel(1000, 120, 40, 32*1024, 100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			m.Update(tea.KeyMsg{Type: tea.KeyDown})
		} else {
			m.Update(tea.KeyMsg{Type: tea.KeyUp})
		}
	}
}

// Parent-picker typing over the repo's 10000-task stress forest: one query
// keystroke per iteration through Update (match plus windowed modal render).
func BenchmarkForm_ParentPickerTyping10000Tasks(b *testing.B) {
	m := benchmarkModel(10000, 120, 40, 1024, 100)
	m.beginForm(false)
	m.focusForm(fieldParent)
	m.beginParentPicker()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%8 == 7 {
			m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		} else {
			m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		}
	}
}

// Search dismissal at the budgeted 1000-task scale: one Esc per iteration
// through Update, which must project the forest exactly once (finish's
// rebuild) while restoring the pre-search selection.
func BenchmarkSearch_EscRestore1000Tasks(b *testing.B) {
	m := benchmarkModel(1000, 120, 40, 32*1024, 100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.beginSearch()
		m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	}
}

func TestWorkflow_StressCommandsStayBounded(t *testing.T) {
	if testing.Short() {
		t.Skip("large TUI stress fixture")
	}
	m := benchmarkModel(10000, 120, 40, 1024*1024, 1000)
	held := m.requestRefresh()
	op := m.operation
	tick := m.tickToken
	var pending tea.Cmd
	for i := range 60 {
		_, cmd := m.Update(tea.WindowSizeMsg{Width: 80 + i%3*40, Height: 24 + i%3*16})
		if cmd != nil && pending == nil {
			pending = cmd
		}
		m.requestRefresh()
		if m.operation != op || m.tickToken != tick {
			t.Fatal("refresh/resize accumulated reads or timers")
		}
	}
	deliverUI(m, pending)
	deliverUI(m, held)
	if m.busy || m.markdown.active || m.markdown.pending || m.refreshPending || m.history.pending {
		t.Fatal("commands did not settle")
	}
	if len(m.rows) != 10000 || m.tickToken != tick {
		t.Fatal("stress truncated tasks or accumulated timers")
	}
}

// Execute the actual cancellable waits, rather than only counting prepared
// commands. Every timer worker must return after the search is dismissed.
func TestWorkflow_StressTimerCommandsDrain(t *testing.T) {
	m := loadedModel(fixtureNode("one", "One", core.PriorityMedium, nil))
	m.options.Wait = Wait
	const count = 64
	returned := make(chan tea.Msg, count)
	press(m, "/")
	for range count {
		cmd := press(m, "a")
		go func() { returned <- cmd() }()
	}
	press(m, "esc")
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for range count {
		select {
		case msg := <-returned:
			m.Update(msg)
		case <-deadline.C:
			t.Fatal("search timer workers did not drain")
		}
	}
	if m.searchCancel != nil || m.searching || m.filter.SearchTerm != "" {
		t.Fatal("search timer retained live state after settling")
	}
}
