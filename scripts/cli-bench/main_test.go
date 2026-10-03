package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCLIBenchmark_ReportJSONKeys(t *testing.T) {
	data, err := json.Marshal(report{Cases: []caseResult{{}}})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if len(fields) != 3 || fields["manifest"] == nil || fields["cases"] == nil || fields["passed"] == nil {
		t.Fatalf("report keys: %s", data)
	}
	var cases []map[string]json.RawMessage
	if err := json.Unmarshal(fields["cases"], &cases); err != nil {
		t.Fatal(err)
	}
	want := []string{"run", "name", "args", "warmup_ns", "samples_ns", "limit_ns", "output_bytes", "min_ns", "median_ns", "p90_ns", "p95_ns", "p99_ns", "max_ns", "p95_limit_ns", "p99_limit_ns", "max_limit_ns", "violations", "passed"}
	if len(cases[0]) != len(want) {
		t.Fatalf("case keys: %s", data)
	}
	for _, key := range want {
		if cases[0][key] == nil {
			t.Errorf("missing %s", key)
		}
	}
}

func TestCLIBenchmark_RejectsInvalidAndMissingSamples(t *testing.T) {
	good := make([]int64, 100)
	for n := range good {
		good[n] = int64(time.Millisecond)
	}
	for _, tc := range []struct {
		name    string
		samples []int64
		want    bool
	}{{"good", good, true}, {"missing", good[:99], false}, {"empty", nil, false}, {"isolated target miss", append(append([]int64{}, good[:99]...), int64(5*time.Millisecond)), true}, {"negative", append(append([]int64{}, good[:99]...), -1), false}} {
		t.Run(tc.name, func(t *testing.T) {
			r := caseResult{Warmups: make([]int64, 5), Samples: tc.samples, LimitNS: int64(5 * time.Millisecond)}
			summarize(&r)
			if r.Passed != tc.want {
				t.Fatalf("%+v", r)
			}
		})
	}
}

func TestCLIBenchmark_DistributionGate(t *testing.T) {
	for _, target := range []time.Duration{5 * time.Millisecond, 15 * time.Millisecond} {
		p95, p99, maximum := 7500*time.Microsecond, 10*time.Millisecond, 15*time.Millisecond
		if target == 15*time.Millisecond {
			p95, p99, maximum = 20*time.Millisecond, 30*time.Millisecond, 50*time.Millisecond
		}
		for _, tc := range []struct {
			name  string
			count int
			slow  time.Duration
			want  bool
		}{
			{"occasional target misses", 10, target, true},
			{"p90 boundary", 11, target, false},
			{"p95 boundary", 6, p95, false},
			{"p99 boundary", 2, p99, false},
			{"maximum boundary", 1, maximum, false},
			{"retained tail below maximum", 1, maximum - 1, true},
		} {
			t.Run(fmt.Sprint(target, "/", tc.name), func(t *testing.T) {
				samples := make([]int64, 100)
				for i := range samples {
					samples[i] = int64(time.Millisecond)
				}
				for i := range tc.count {
					samples[i] = int64(tc.slow)
				}
				original := slices.Clone(samples)
				r := caseResult{Warmups: make([]int64, 5), Samples: samples, LimitNS: int64(target)}
				summarize(&r)
				if r.Passed != tc.want || r.Violations != tc.count || !slices.Equal(r.Samples, original) {
					t.Fatalf("passed=%v want=%v retained target misses=%d want=%d", r.Passed, tc.want, r.Violations, tc.count)
				}
				summarize(&r)
				if r.Violations != tc.count {
					t.Fatal("summaries accumulate stale counts")
				}
			})
		}
	}
}
func TestCLIBenchmark_Collect(t *testing.T) {
	calls := 0
	r := collect("ok", int64(5*time.Millisecond), func() (int64, int, error) { calls++; return int64(time.Millisecond), 42, nil })
	if !r.Passed || calls != 105 || len(r.Warmups) != 5 || r.OutputBytes != 42 {
		t.Fatalf("%+v calls %d", r, calls)
	}
	calls = 0
	r = collect("fail", int64(5*time.Millisecond), func() (int64, int, error) { calls++; return 1, 0, errors.New("child failure") })
	if r.Passed || r.Error == "" || calls != 1 {
		t.Fatalf("%+v", r)
	}
	calls = 0
	r = collect("bytes", int64(5*time.Millisecond), func() (int64, int, error) { calls++; return 1, calls, nil })
	if r.Passed || r.Error == "" {
		t.Fatalf("%+v", r)
	}
}
func TestCLIBenchmark_OutputValidation(t *testing.T) {
	for _, tc := range []struct {
		args          []string
		data          string
		count, events int
		valid         bool
	}{
		{[]string{"--help"}, "Usage: tusk", 0, 0, true},
		{[]string{"--version"}, "tusk version test\n", 0, 0, true},
		{[]string{"list", "--all", "--json"}, "[]\n", 0, 0, true},
		{[]string{"list", "--all", "--json"}, "{}\n", 0, 0, false},
		{[]string{"list", "--all", "--json"}, "[]\n", 100, 0, false},
		{[]string{"history", "id", "--json"}, "[]\n", 0, 0, true},
		{[]string{"tree", "--json"}, "[]\n", 0, 0, true},
		{[]string{"stats", "--json"}, `{"total":0}`, 0, 0, false},
		{[]string{"list"}, "ID\tSTATUS\tPRIORITY\tPROGRESS\tDUE\tTITLE\n", 0, 0, true},
		{[]string{"tree"}, "No tasks.\n", 0, 0, true},
		{[]string{"history", "id"}, "SEQUENCE\tOCCURRED_AT\tKIND\tCHANGED_FIELDS\n", 0, 0, true},
		{[]string{"stats"}, "Total\t0\n", 0, 0, false},
		{[]string{"stats", "--json"}, "bad", 0, 0, false},
	} {
		if err := validateOutput(tc.args, []byte(tc.data), tc.count, tc.events); (err == nil) != tc.valid {
			t.Fatalf("%q %q: %v", tc.args, tc.data, err)
		}
	}
}

func fixtureTaskJSON(n int) map[string]any {
	return map[string]any{"id": fmt.Sprintf("fixture-%06d", n), "title": strings.Repeat("t", 64), "description": strings.Repeat("n", 128), "status": "todo", "priority": 2, "parent_id": nil, "progress": 0, "tags": []string{"alpha", "beta", "gamma"}, "due_date": nil, "created_at": "2026-09-28T12:00:00Z", "updated_at": "2026-09-28T12:00:00Z", "completed_at": nil}
}

func TestCLIBenchmark_RejectsIncompleteData(t *testing.T) {
	for _, tc := range []struct{ command, data string }{
		{"list", "[{}]"}, {"history", "[{}]"}, {"tree", `[{"task":{},"children":[]}]`},
		{"stats", `{"total":1}`},
	} {
		if err := validateOutput([]string{tc.command, "--json"}, []byte(tc.data), 1, 1); err == nil {
			t.Errorf("accepted %s %s", tc.command, tc.data)
		}
	}
	for _, tc := range []struct{ command, data string }{
		{"list", "ID\tSTATUS\tPRIORITY\tPROGRESS\tDUE\tTITLE\nrow\n"},
		{"tree", "task\n"}, {"stats", "Total\t1\n"}, {"history", "SEQUENCE\tOCCURRED_AT\tKIND\tCHANGED_FIELDS\nrow\n"},
	} {
		if err := validateOutput([]string{tc.command}, []byte(tc.data), 1, 1); err == nil {
			t.Errorf("accepted human %s", tc.command)
		}
	}
}

func fakeProcess(_ string, args, env []string) (int64, []byte, error) {
	count := 0
	for _, v := range env {
		if strings.HasPrefix(v, "TUSK_DB_PATH=") {
			name := filepath.Base(strings.TrimPrefix(v, "TUSK_DB_PATH="))
			if strings.HasPrefix(name, "fixture-") {
				count, _ = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(name, "fixture-"), ".db"))
			}
		}
	}
	events := 20
	if count == 0 {
		events = 0
	}
	var out string
	switch args[0] {
	case "--help":
		out = "Usage: tusk\n"
	case "--version":
		out = "tusk version test\n"
	case "stats":
		if slices.Contains(args, "--json") {
			out = fmt.Sprintf("{\"total\":%d,\"by_status\":{\"todo\":%d,\"in-progress\":0,\"blocked\":0,\"done\":0},\"done\":0,\"completion_percent\":0,\"overdue\":0,\"completed_last_7_days\":0}\n", count, count)
		} else {
			out = fmt.Sprintf("Total\t%d\ntodo\t%d\nin-progress\t0\nblocked\t0\ndone\t0\nDone\t0\nCompletion percent\t0\nOverdue\t0\nCompleted last 7 days\t0\n", count, count)
		}
	case "tree":
		if slices.Contains(args, "--json") {
			items := make([]map[string]any, count)
			for n := range items {
				items[n] = map[string]any{"task": fixtureTaskJSON(n), "children": []any{}, "depth": 1}
			}
			b, _ := json.Marshal(items)
			out = string(b) + "\n"
		} else {
			for n := range count {
				out += fmt.Sprintf("`-- fixture-%06d [todo 0%%] %s (depth 1)\n", n, strings.Repeat("t", 64))
			}
			if count == 0 {
				out = "No tasks.\n"
			}
		}
	default:
		n := count
		if args[0] == "history" {
			n = events
		}
		if slices.Contains(args, "--json") {
			items := make([]map[string]any, n)
			for i := range items {
				items[i] = fixtureTaskJSON(i)
				if args[0] == "history" {
					items[i] = map[string]any{"sequence": i + 1, "task_id": "fixture-000000", "kind": "metadata", "changed_fields": []string{"description"}, "occurred_at": "2026-09-28T12:00:00Z"}
				}
			}
			b, _ := json.Marshal(items)
			out = string(b) + "\n"
		} else {
			out = "ID\tSTATUS\tPRIORITY\tPROGRESS\tDUE\tTITLE\n"
			if args[0] == "history" {
				out = "SEQUENCE\tOCCURRED_AT\tKIND\tCHANGED_FIELDS\n"
			}
			for i := range n {
				if args[0] == "history" {
					out += fmt.Sprintf("%d\t2026-09-28T12:00:00Z\tmetadata\tdescription\n", i+1)
				} else {
					out += fmt.Sprintf("fixture-%06d\ttodo\tmedium\t0%%\t-\t%s\n", i, strings.Repeat("t", 64))
				}
			}
		}
	}
	return int64(time.Millisecond), []byte(out), nil
}
func TestCLIBenchmark_EveryRunMustPass(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(binary, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	for failingRun := 1; failingRun <= 3; failingRun++ {
		output := filepath.Join(t.TempDir(), "report.json")
		execute := func(binary string, args, env []string) (int64, []byte, error) {
			duration, out, err := fakeProcess(binary, args, env)
			for _, value := range env {
				if strings.Contains(filepath.ToSlash(value), fmt.Sprintf("/run-%d/", failingRun)) && args[0] == "--help" {
					duration = int64(8 * time.Millisecond)
				}
			}
			return duration, out, err
		}
		if code := runWith([]string{"--binary", binary, "--output", output}, io.Discard, io.Discard, execute, nil); code != 1 {
			t.Fatalf("run %d failure returned %d", failingRun, code)
		}
		data, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		var result report
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatal(err)
		}
		if result.Passed || len(result.Cases) != 12 {
			t.Fatalf("incomplete failed report: %+v", result)
		}
		for i, c := range result.Cases {
			if c.Run != i/4+1 {
				t.Fatalf("case %d has run %d", i, c.Run)
			}
			want := c.Run != failingRun || c.Args[0] != "--help"
			if c.Passed != want || len(c.Samples) != 100 {
				t.Fatalf("run %d case %s: passed=%v, want=%v", c.Run, c.Name, c.Passed, want)
			}
		}
	}
}

func TestCLIBenchmark_ReportAndFailures(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(binary, []byte("test-binary"), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer
	if code := runWith([]string{"--binary", binary, "--output", output}, &stdout, &stderr, fakeProcess, []int{0, 10}); code != 0 {
		t.Fatalf("%d %s", code, stderr.String())
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var r report
	if err = json.Unmarshal(data, &r); err != nil || !r.Passed || len(r.Cases) != 60 || r.Manifest["sha256"] == "" {
		t.Fatalf("%+v %v", r, err)
	}
	if r.Manifest["coordinator_gomaxprocs"] != strconv.Itoa(runtime.GOMAXPROCS(0)) {
		t.Fatal("report omits the benchmark coordinator's runtime parallelism")
	}
	if code := runWith([]string{"--binary", binary, "--output", output}, io.Discard, io.Discard, func(string, []string, []string) (int64, []byte, error) { return 0, nil, errors.New("failed process") }, nil); code != 1 {
		t.Fatal(code)
	}
	for _, args := range [][]string{{"--bad"}, {"extra"}, {"--binary", filepath.Join(t.TempDir(), "missing")}, {"--binary", binary, "--output", t.TempDir()}} {
		if code := run(args, io.Discard, io.Discard); code == 0 {
			t.Fatalf("accepted %q", args)
		}
	}
	t.Run("unavailable filesystem probe", func(t *testing.T) {
		t.Setenv("PATH", "")
		if code := runWith([]string{"--binary", binary, "--output", output}, io.Discard, io.Discard, fakeProcess, nil); code != 0 {
			t.Fatal(code)
		}
	})
	t.Run("temporary directory failure", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing")
		for _, key := range []string{"TMPDIR", "TMP", "TEMP"} {
			t.Setenv(key, missing)
		}
		if code := runWith([]string{"--binary", binary, "--output", output}, io.Discard, io.Discard, fakeProcess, nil); code != 1 {
			t.Fatal(code)
		}
	})
	t.Run("final report failure", func(t *testing.T) {
		dir := t.TempDir()
		output := filepath.Join(dir, "report.json")
		calls := 0
		fake := func(b string, a, e []string) (int64, []byte, error) {
			calls++
			if calls == 1 {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatal(err)
				}
			}
			return fakeProcess(b, a, e)
		}
		if code := runWith([]string{"--binary", binary, "--output", output}, io.Discard, io.Discard, fake, nil); code != 1 {
			t.Fatal(code)
		}
	})
}
func TestCLIBenchmark_Seed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fixture.db")
	if _, err := seed(path, 1000, 128); err != nil {
		t.Fatal(err)
	}
	if _, err := seed(path, 1, 128); err == nil {
		t.Fatal("duplicate fixture accepted")
	}
	if _, err := seed(t.TempDir(), 1, 128); err == nil {
		t.Fatal("directory accepted")
	}
	if _, err := benchmark("unused", t.TempDir(), []int{0, 0}, fakeProcess); err == nil {
		t.Fatal("seed failure hidden")
	}
}
func TestCLIBenchmark_Execute(t *testing.T) {
	if mode := os.Getenv("TUSK_BENCH_EXECUTE"); mode != "" {
		switch mode {
		case "stdout":
			fmt.Fprint(os.Stdout, "output")
		case "stderr":
			fmt.Fprint(os.Stderr, "failure")
		case "exit":
			fmt.Fprint(os.Stdout, "partial output")
			os.Exit(3)
		default:
			os.Exit(4)
		}
		os.Exit(0)
	}
	for _, tc := range []struct {
		mode, output, errorText string
	}{{"stdout", "output", ""}, {"stderr", "", "unexpected stderr"}, {"exit", "partial output", "exit status 3"}} {
		t.Run(tc.mode, func(t *testing.T) {
			t.Setenv("TUSK_BENCH_EXECUTE", tc.mode)
			// Instrumented child tests retain their own coverage without diagnostics.
			t.Setenv("GOCOVERDIR", t.TempDir())
			duration, out, err := execute(os.Args[0], []string{"-test.run=^TestCLIBenchmark_Execute$"}, os.Environ())
			if duration <= 0 || string(out) != tc.output || (err == nil) != (tc.errorText == "") {
				t.Fatalf("duration %d output %q error %v", duration, out, err)
			}
			if err != nil && err.Error() != tc.errorText {
				t.Fatalf("error %v, want %s", err, tc.errorText)
			}
		})
	}
	if _, _, err := execute(filepath.Join(t.TempDir(), "missing-process"), nil, nil); err == nil {
		t.Fatal("missing process passed")
	}
}

func TestCLIBenchmark_Capture(t *testing.T) {
	out, diagnostic := prepareCapture()
	data := bytes.Repeat([]byte{'x'}, 768<<10)
	if out.Cap() < len(data) || diagnostic.Cap() < 4096 {
		t.Fatal("capture storage must be allocated before timing")
	}
	allocs := testing.AllocsPerRun(10, func() {
		out.Reset()
		if _, err := out.Write(data); err != nil {
			t.Fatal(err)
		}
	})
	if allocs != 0 || diagnostic.Len() != 0 || !bytes.Equal(out.Bytes(), data) {
		t.Fatalf("capture: allocations=%g, bytes=%d", allocs, out.Len())
	}
	// Capacity is preparation, never an output limit.
	if _, err := out.Write(data); err != nil || out.Len() != 2*len(data) {
		t.Fatalf("large output: %d %v", out.Len(), err)
	}
}
func TestCLIBenchmark_Main(t *testing.T) {
	if os.Getenv("TUSK_BENCH_MAIN") == "1" {
		os.Args = []string{"cli-bench", "--help"}
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestCLIBenchmark_Main$")
	cmd.Env = append(os.Environ(), "TUSK_BENCH_MAIN=1")
	err := cmd.Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 {
		t.Fatalf("help exit %v", err)
	}
}

func TestCLIBenchmark_RejectsNullAndMissingWarmups(t *testing.T) {
	for _, command := range []string{"list", "tree", "history"} {
		if err := validateOutput([]string{command, "--json"}, []byte("null\n"), 0, 0); err == nil {
			t.Errorf("%s accepted null array", command)
		}
	}
	samples := make([]int64, 100)
	for n := range samples {
		samples[n] = 1
	}
	r := caseResult{Samples: samples, LimitNS: 10}
	summarize(&r)
	if r.Passed {
		t.Error("missing warmups passed")
	}
	r = collect("empty", 10, func() (int64, int, error) { return 1, 0, nil })
	if r.Passed {
		t.Error("empty output passed")
	}
}
func TestCLIBenchmark_InvalidFixtures(t *testing.T) {
	for _, tc := range []struct {
		args          []string
		data          string
		count, events int
	}{{[]string{"--help"}, "bad", 0, 0}, {[]string{"--version"}, "bad", 0, 0}, {[]string{"list"}, "", 0, 0}, {[]string{"stats", "--json"}, `{"total":10}`, 0, 0}, {[]string{"tree", "--json"}, "{}", 0, 0}, {[]string{"tree", "--json"}, "[]", 1, 0}, {[]string{"list"}, "bad\n", 1, 0}, {[]string{"tree"}, "bad\n", 0, 0}, {[]string{"tree"}, "task\n", 2, 0}, {[]string{"stats"}, "Total\t2\n", 1, 0}, {[]string{"history"}, "bad\n", 0, 1}, {[]string{"unknown"}, "bad\n", 0, 0}} {
		if err := validateOutput(tc.args, []byte(tc.data), tc.count, tc.events); err == nil {
			t.Fatalf("accepted %+v", tc)
		}
	}
}
