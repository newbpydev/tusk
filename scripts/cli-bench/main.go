// cli-bench measures fresh CLI processes. Builds and fixtures stay outside timing.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/storage"
)

type caseResult struct {
	Run         int      `json:"run"`
	Name        string   `json:"name"`
	Args        []string `json:"args"`
	Warmups     []int64  `json:"warmup_ns"`
	Samples     []int64  `json:"samples_ns"`
	LimitNS     int64    `json:"limit_ns"`
	OutputBytes int      `json:"output_bytes"`
	Min         int64    `json:"min_ns"`
	Median      int64    `json:"median_ns"`
	P90         int64    `json:"p90_ns"`
	P95         int64    `json:"p95_ns"`
	P99         int64    `json:"p99_ns"`
	Max         int64    `json:"max_ns"`
	P95LimitNS  int64    `json:"p95_limit_ns"`
	P99LimitNS  int64    `json:"p99_limit_ns"`
	MaxLimitNS  int64    `json:"max_limit_ns"`
	Violations  int      `json:"violations"`
	Passed      bool     `json:"passed"`
	Error       string   `json:"error,omitempty"`
}
type report struct {
	Manifest map[string]string `json:"manifest"`
	Cases    []caseResult      `json:"cases"`
	Passed   bool              `json:"passed"`
}
type executor func(string, []string, []string) (int64, []byte, error)

func summarize(r *caseResult) {
	r.Passed = len(r.Warmups) == 5 && len(r.Samples) == 100 && r.Error == ""
	r.Violations = 0
	r.Min, r.Median, r.P90, r.P95, r.P99, r.Max = 0, 0, 0, 0, 0, 0
	r.P95LimitNS, r.P99LimitNS, r.MaxLimitNS = 0, 0, 0
	switch r.LimitNS {
	case int64(5 * time.Millisecond):
		r.P95LimitNS, r.P99LimitNS, r.MaxLimitNS = int64(7500*time.Microsecond), int64(10*time.Millisecond), int64(15*time.Millisecond)
	case int64(15 * time.Millisecond):
		r.P95LimitNS, r.P99LimitNS, r.MaxLimitNS = int64(20*time.Millisecond), int64(30*time.Millisecond), int64(50*time.Millisecond)
	default:
		r.Passed = false
	}
	if len(r.Samples) == 0 {
		return
	}
	ordered := slices.Clone(r.Samples)
	slices.Sort(ordered)
	r.Min = ordered[0]
	// Nearest-rank empirical percentiles; the original sample order is retained.
	r.Median = ordered[(len(ordered)*50-1)/100]
	r.P90 = ordered[(len(ordered)*90-1)/100]
	r.P95 = ordered[(len(ordered)*95-1)/100]
	r.P99 = ordered[(len(ordered)*99-1)/100]
	r.Max = ordered[len(ordered)-1]
	for _, n := range r.Samples {
		if n <= 0 {
			r.Passed = false
		}
		if n <= 0 || n >= r.LimitNS {
			r.Violations++
		}
	}
	r.Passed = r.Passed && r.P90 < r.LimitNS && r.P95 < r.P95LimitNS && r.P99 < r.P99LimitNS && r.Max < r.MaxLimitNS
}
func collect(name string, limit int64, invoke func() (int64, int, error)) caseResult {
	r := caseResult{Name: name, LimitNS: limit, Warmups: []int64{}, Samples: []int64{}}
	for n := range 105 {
		duration, size, err := invoke()
		if err == nil && size <= 0 {
			err = errors.New("empty output")
		}
		if err != nil {
			r.Error = err.Error()
			break
		}
		if n == 0 {
			r.OutputBytes = size
		} else if size != r.OutputBytes {
			r.Error = "output byte count changed"
			break
		}
		if n < 5 {
			r.Warmups = append(r.Warmups, duration)
		} else {
			r.Samples = append(r.Samples, duration)
		}
	}
	summarize(&r)
	return r
}
func validateOutput(args []string, data []byte, count, events int) error {
	invalid := errors.New("incorrect output fixture")
	if len(data) == 0 {
		return invalid
	}
	if args[0] == "--help" {
		if !bytes.Contains(data, []byte("Usage:")) {
			return invalid
		}
		return nil
	}
	if args[0] == "--version" {
		if !bytes.HasPrefix(data, []byte("tusk version ")) {
			return invalid
		}
		return nil
	}
	jsonOutput := slices.Contains(args, "--json")
	if jsonOutput {
		if !json.Valid(data) {
			return invalid
		}
		if args[0] != "stats" && bytes.TrimSpace(data)[0] != '[' {
			return invalid
		}
		if !validFields(args[0], data) {
			return invalid
		}
		switch args[0] {
		case "stats":
			var value struct {
				Total *int `json:"total"`
			}
			if err := json.Unmarshal(data, &value); err != nil || value.Total == nil || *value.Total != count {
				return invalid
			}
		case "tree":
			var nodes []struct {
				Task     map[string]any  `json:"task"`
				Children json.RawMessage `json:"children"`
			}
			if err := json.Unmarshal(data, &nodes); err != nil {
				return invalid
			}
			total := bytes.Count(data, []byte(`"task":`))
			if total != count {
				return invalid
			}
		default:
			var items []json.RawMessage
			if err := json.Unmarshal(data, &items); err != nil {
				return invalid
			}
			want := count
			if args[0] == "history" {
				want = events
			}
			if len(items) != want {
				return invalid
			}
		}
	} else {
		rows := bytes.Count(data, []byte{'\n'})
		switch args[0] {
		case "list":
			if !bytes.HasPrefix(data, []byte("ID\tSTATUS\tPRIORITY\tPROGRESS\tDUE\tTITLE\n")) || rows != count+1 {
				return invalid
			}
			for _, row := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")[1:] {
				fields := strings.Split(row, "\t")
				if len(fields) != 6 || !strings.HasPrefix(fields[0], "fixture-") || fields[5] != strings.Repeat("t", 64) {
					return invalid
				}
			}
		case "tree":
			if count == 0 {
				if string(data) != "No tasks.\n" {
					return invalid
				}
			} else if rows != count {
				return invalid
			}
			if count > 0 && (bytes.Count(data, []byte("fixture-")) != count || bytes.Count(data, []byte(strings.Repeat("t", 64))) != count) {
				return invalid
			}
		case "stats":
			if !bytes.HasPrefix(data, []byte(fmt.Sprintf("Total\t%d\n", count))) || rows != 9 {
				return invalid
			}
		case "history":
			if !bytes.HasPrefix(data, []byte("SEQUENCE\tOCCURRED_AT\tKIND\tCHANGED_FIELDS\n")) || rows != events+1 {
				return invalid
			}
			for _, row := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")[1:] {
				fields := strings.Split(row, "\t")
				if len(fields) != 4 {
					return invalid
				}
				n, err := strconv.ParseInt(fields[0], 10, 64)
				if err != nil || n <= 0 || fields[2] != "metadata" || fields[3] != "description" {
					return invalid
				}
			}
		default:
			return invalid
		}
	}
	return nil
}

// Check complete wire fields as well as counts: empty objects are not fixtures.
func validFields(command string, data []byte) bool {
	if command == "stats" {
		var m map[string]json.RawMessage
		if json.Unmarshal(data, &m) != nil {
			return false
		}
		return completeFields(m, "total,by_status,done,completion_percent,overdue,completed_last_7_days")
	}
	var items []map[string]json.RawMessage
	if json.Unmarshal(data, &items) != nil {
		return false
	}
	for _, m := range items {
		switch command {
		case "tree":
			if !completeFields(m, "task,children,depth") || !validFields("list", append(append([]byte{'['}, m["task"]...), ']')) || string(m["children"]) == "null" || !validFields("tree", m["children"]) {
				return false
			}
		case "history":
			if !completeFields(m, "sequence,task_id,kind,changed_fields,occurred_at") || string(m["task_id"]) != `"fixture-000000"` || string(m["kind"]) != `"metadata"` || string(m["changed_fields"]) != `["description"]` {
				return false
			}
		default:
			if !completeFields(m, "id,title,description,status,priority,parent_id,progress,tags,due_date,created_at,updated_at,completed_at") {
				return false
			}
			var task struct {
				ID, Title, Description string
				Tags                   []string
			}
			raw, _ := json.Marshal(m)
			if json.Unmarshal(raw, &task) != nil || !strings.HasPrefix(task.ID, "fixture-") || task.Title != strings.Repeat("t", 64) || task.Description != strings.Repeat("n", 128) || !slices.Equal(task.Tags, []string{"alpha", "beta", "gamma"}) {
				return false
			}
		}
	}
	return true
}
func completeFields(m map[string]json.RawMessage, fields string) bool {
	keys := strings.Split(fields, ",")
	if len(m) != len(keys) {
		return false
	}
	for _, key := range keys {
		if _, ok := m[key]; !ok {
			return false
		}
	}
	return true
}

// Prepare the fast pipe consumer before timing. A larger output can still grow.
func prepareCapture() (out, stderr *bytes.Buffer) {
	out, stderr = new(bytes.Buffer), new(bytes.Buffer)
	out.Grow(1 << 20)
	stderr.Grow(4096)
	return out, stderr
}

func execute(binary string, args, env []string) (int64, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = env
	out, stderr := prepareCapture()
	cmd.Stdout = out
	cmd.Stderr = stderr
	// Previous output validation allocates heavily. Collect that parent-only
	// garbage before launch so it cannot pause the promised fast pipe consumer.
	// The fresh child keeps its ordinary garbage collection policy.
	runtime.GC()
	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start).Nanoseconds()
	// Run waits for process exit and both os/exec output-copy goroutines.
	if err != nil {
		return elapsed, out.Bytes(), err
	}
	if stderr.Len() != 0 {
		return elapsed, out.Bytes(), errors.New("unexpected stderr")
	}
	return elapsed, out.Bytes(), nil
}
func seed(path string, count int, notes int) (string, error) {
	repo, err := storage.Open(context.Background(), storage.Options{Path: path})
	if err != nil {
		return "", err
	}
	now := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	historyID := "fixture-000000"
	n := count
	if n == 0 && strings.Contains(path, "history") {
		n = 1
	}
	err = repo.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		for i := range n {
			status := []core.Status{core.StatusTodo, core.StatusInProgress, core.StatusBlocked, core.StatusDone}[i%4]
			if i < 9 {
				status = core.StatusTodo
			}
			task := core.Task{ID: fmt.Sprintf("fixture-%06d", i), Title: strings.Repeat("t", 64), Description: strings.Repeat("n", notes), Status: status, Priority: core.Priority(i%4 + 1), Tags: []core.Tag{"alpha", "beta", "gamma"}, CreatedAt: now, UpdatedAt: now}
			if i%10 != 0 {
				parent := fmt.Sprintf("fixture-%06d", i/10*10)
				if i < 10 {
					parent = fmt.Sprintf("fixture-%06d", i-1)
				}
				task.ParentID = &parent
			}
			if i%3 != 0 {
				due := now.Add(time.Duration(i%3-1) * 24 * time.Hour)
				task.DueDate = &due
			}
			if status == core.StatusDone {
				task.Progress = 100
				completed := now.Add(-time.Hour)
				task.CompletedAt = &completed
			}
			if err := w.Create(ctx, &task); err != nil {
				return err
			}
		}
		if count > 0 {
			for range 20 {
				if _, err := w.AppendEvent(ctx, ports.TaskEvent{TaskID: historyID, Kind: ports.EventMetadata, ChangedFields: []string{"description"}, OccurredAt: now}); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return historyID, errors.Join(err, repo.Close())
}
func processEnv(dir, path string) []string {
	return []string{"PATH=" + os.Getenv("PATH"), "HOME=" + dir, "USERPROFILE=" + dir, "TUSK_DB_PATH=" + path, "TUSK_TIMEZONE=UTC", "TZ=UTC", "TERM=dumb"}
}
func benchmark(binary, dir string, sizes []int, run executor) ([]caseResult, error) {
	var results []caseResult
	for _, bad := range []bool{false, true} {
		for _, arg := range []string{"--help", "--version"} {
			env := processEnv(dir, filepath.Join(dir, "unused.db"))
			if bad {
				env = append(env, "TUSK_TIMEZONE=not-a-zone", "TUSK_AUTO_COMPLETE_PARENT=bad", "TUSK_DB_PATH="+dir)
			}
			args := []string{arg}
			r := collect(fmt.Sprintf("%s/invalid-config=%v", arg, bad), int64(5*time.Millisecond), func() (int64, int, error) {
				d, out, err := run(binary, args, env)
				if err == nil {
					err = validateOutput(args, out, 0, 0)
				}
				return d, len(out), err
			})
			r.Args = args
			results = append(results, r)
		}
	}
	for _, count := range sizes {
		path := filepath.Join(dir, fmt.Sprintf("fixture-%d.db", count))
		historyID, err := seed(path, count, 128)
		if err != nil {
			return results, err
		}
		historyPath := path
		events := 20
		if count == 0 {
			events = 0
			historyPath = filepath.Join(dir, "empty-history.db")
			if _, err = seed(historyPath, 0, 128); err != nil {
				return results, err
			}
		}
		for _, name := range []string{"list", "tree", "stats", "history"} {
			for _, jsonOutput := range []bool{false, true} {
				args := []string{name}
				if name == "list" {
					args = append(args, "--all")
				}
				selectedPath := path
				if name == "history" {
					args = append(args, historyID)
					selectedPath = historyPath
				}
				if jsonOutput {
					args = append(args, "--json")
				}
				env := processEnv(dir, selectedPath)
				r := collect(fmt.Sprintf("%d/%s/json=%v", count, name, jsonOutput), int64(15*time.Millisecond), func() (int64, int, error) {
					d, out, err := run(binary, args, env)
					if err == nil {
						err = validateOutput(args, out, count, events)
					}
					return d, len(out), err
				})
				r.Args = args
				results = append(results, r)
			}
		}
	}
	return results, nil
}
func manifest(binary string) (map[string]string, error) {
	data, err := os.ReadFile(binary)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	m := map[string]string{"binary": binary, "sha256": fmt.Sprintf("%x", hash), "go": runtime.Version(), "os": runtime.GOOS, "arch": runtime.GOARCH, "reference_utc": "2026-09-28T12:00:00Z", "started_utc": time.Now().UTC().Format(time.RFC3339Nano), "sample_rule": "3 complete runs; each case has 5 warmups plus 100 consecutive samples; every case in every run must pass; no trimming or retries", "policy": "latency-distribution-v2"}
	m["pipe_consumer"] = "1 MiB stdout and 4 KiB stderr capacity prepared before timing; parent GC before launch; buffers may grow; child GC unchanged; launch through exit and drain timed"
	m["coordinator_gomaxprocs"] = strconv.Itoa(runtime.GOMAXPROCS(0))
	for key, path := range map[string]string{"cpu": "/proc/cpuinfo", "kernel": "/proc/sys/kernel/osrelease", "power_profile": "/sys/firmware/acpi/platform_profile", "governor": "/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor"} {
		data, err := os.ReadFile(path)
		if err != nil {
			m[key] = "unavailable"
		} else {
			m[key] = strings.TrimSpace(string(data))
		}
	}
	return m, nil
}
func run(args []string, stdout, stderr io.Writer) int {
	return runWith(args, stdout, stderr, execute, []int{0, 100, 1000})
}

func runWith(args []string, stdout, stderr io.Writer, executeProcess executor, sizes []int) int {
	flags := flag.NewFlagSet("cli-bench", flag.ContinueOnError)
	flags.SetOutput(stderr)
	binary := flags.String("binary", "bin/tusk", "Built executable")
	output := flags.String("output", "docs/verification-evidence/004/latency.json", "Raw report")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		return 2
	}
	absolute, err := filepath.Abs(*binary)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	m, err := manifest(absolute)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	dir, err := os.MkdirTemp("", "tusk-cli-bench-")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)
	fs, err := exec.Command("stat", "-f", "-c", "%T", dir).Output()
	if err == nil {
		m["filesystem"] = strings.TrimSpace(string(fs))
	} else {
		m["filesystem"] = "unavailable"
	}
	r := report{Manifest: m, Passed: false}
	initial, _ := json.MarshalIndent(r, "", "  ")
	if err = os.WriteFile(*output, append(initial, '\n'), 0600); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	completedRuns := 0
	for run := 1; run <= 3; run++ {
		runDir := filepath.Join(dir, fmt.Sprintf("run-%d", run))
		if err = os.Mkdir(runDir, 0700); err != nil {
			break
		}
		var cases []caseResult
		cases, err = benchmark(absolute, runDir, sizes, executeProcess)
		for i := range cases {
			cases[i].Run = run
		}
		r.Cases = append(r.Cases, cases...)
		if err != nil {
			break
		}
		completedRuns++
	}
	r.Passed = err == nil && completedRuns == 3
	for _, c := range r.Cases {
		r.Passed = r.Passed && c.Passed
		fmt.Fprintf(stdout, "run=%d %-38s p90=%7.3fms p95=%7.3fms p99=%7.3fms max=%7.3fms target_misses=%d passed=%v %s\n", c.Run, c.Name, float64(c.P90)/1e6, float64(c.P95)/1e6, float64(c.P99)/1e6, float64(c.Max)/1e6, c.Violations, c.Passed, c.Error)
	}
	if err != nil {
		m["error"] = err.Error()
	}
	data, _ := json.MarshalIndent(r, "", "  ")
	if err = os.WriteFile(*output, append(data, '\n'), 0600); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if !r.Passed {
		return 1
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
