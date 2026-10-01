package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func hostFixture() report {
	r := report{Manifest: map[string]string{"cpu": "AMD Ryzen 5 4500U", "os": "linux", "arch": "amd64", "power_profile": "balanced"}}
	for run := 1; run <= 3; run++ {
		names := []string{"--help/invalid-config=false", "--help/invalid-config=true", "--version/invalid-config=false", "--version/invalid-config=true"}
		for _, size := range []int{0, 100, 1000} {
			for _, command := range []string{"list", "tree", "stats", "history"} {
				for _, json := range []bool{false, true} {
					names = append(names, fmt.Sprintf("%d/%s/json=%v", size, command, json))
				}
			}
		}
		for _, name := range names {
			limit := int64(15 * time.Millisecond)
			latency := int64((13 + run) * int(time.Millisecond))
			if name[0] == '-' {
				limit = int64(5 * time.Millisecond)
				latency = int64(3 * time.Millisecond)
			}
			c := caseResult{Run: run, Name: name, LimitNS: limit, Warmups: make([]int64, 5), Samples: make([]int64, 100)}
			for i := range c.Samples {
				c.Samples[i] = latency
			}
			summarize(&c)
			r.Cases = append(r.Cases, c)
		}
	}
	return r
}

func TestHostAcceptance_MeanP90WithEveryRunTailGuards(t *testing.T) {
	for _, scenario := range []string{"valid", "reference p95 miss", "p90 boundary", "p95 boundary", "maximum boundary", "missing run", "duplicate run", "wrong host", "wrong profile", "invalid sample", "child error"} {
		t.Run(scenario, func(t *testing.T) {
			r := hostFixture()
			switch scenario {
			case "p90 boundary":
				for i := range r.Cases {
					if r.Cases[i].LimitNS == int64(15*time.Millisecond) {
						for j := range r.Cases[i].Samples {
							r.Cases[i].Samples[j] = int64(18 * time.Millisecond)
						}
					}
				}
			case "reference p95 miss":
				for j := 0; j < 6; j++ {
					r.Cases[4].Samples[j] = int64(20 * time.Millisecond)
				}
			case "p95 boundary":
				for j := 0; j < 6; j++ {
					r.Cases[4].Samples[j] = int64(22 * time.Millisecond)
				}
			case "maximum boundary":
				r.Cases[4].Samples[0] = int64(50 * time.Millisecond)
			case "missing run":
				r.Cases = r.Cases[:83]
			case "duplicate run":
				r.Cases[28].Run = 1
			case "wrong host":
				r.Manifest["cpu"] = "Different CPU"
			case "wrong profile":
				r.Manifest["power_profile"] = "performance"
			case "invalid sample":
				r.Cases[4].Samples[0] = -1
			case "child error":
				r.Cases[4].Error = "failed"
			}
			a := assessHost(r)
			if a.Passed != (scenario == "valid" || scenario == "reference p95 miss") {
				t.Fatalf("acceptance=%+v", a)
			}
			if scenario == "valid" && (len(a.Cases) != 28 || a.Cases[4].MeanP90NS != float64(15*time.Millisecond)) {
				t.Fatalf("means=%+v", a)
			}
		})
	}
}

func TestHostAcceptance_ReportKeepsReferenceAssessment(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(binary, []byte("test binary"), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "report.json")
	var stdout bytes.Buffer
	code := runWith([]string{"--binary", binary, "--output", output, "--acceptance-profile", "ryzen-4500u-balanced-v1"}, &stdout, io.Discard, fakeProcess, nil)
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var r report
	if err = json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
	// This intentionally incomplete matrix passes the reference cases that ran,
	// but cannot pass host acceptance on any machine.
	if code != 1 || !r.Passed || r.HostAcceptance == nil || r.HostAcceptance.Passed || !strings.Contains(stdout.String(), "reference_passed=true host_passed=false") {
		t.Fatalf("code=%d report=%+v output=%s", code, r, stdout.String())
	}
	if runWith([]string{"--acceptance-profile", "unknown"}, io.Discard, io.Discard, fakeProcess, nil) != 2 {
		t.Fatal("unknown profile accepted")
	}
}

func TestHostAcceptance_PreservesExistingReports(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(binary, []byte("test binary"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, destination := range []string{"omitted", "existing"} {
		t.Run(destination, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "report.json")
			sentinel := []byte("retained evidence")
			if err := os.WriteFile(output, sentinel, 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{"--binary", binary, "--acceptance-profile", "ryzen-4500u-balanced-v1"}
			if destination == "existing" {
				args = append(args, "--output", output)
			}
			called := false
			var stderr bytes.Buffer
			code := runWith(args, io.Discard, &stderr, func(b string, a, e []string) (int64, []byte, error) { called = true; return fakeProcess(b, a, e) }, nil)
			after, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			if code != 2 || called || !bytes.Equal(after, sentinel) || !strings.Contains(stderr.String(), "output") {
				t.Fatalf("code=%d called=%v report=%q stderr=%q", code, called, after, stderr.String())
			}
		})
	}
}

func TestHostAcceptance_InvalidProfileDiagnostic(t *testing.T) {
	var stderr bytes.Buffer
	if code := runWith([]string{"--acceptance-profile", "typo"}, io.Discard, &stderr, fakeProcess, nil); code != 2 || !strings.Contains(stderr.String(), "typo") || !strings.Contains(stderr.String(), "reference") || !strings.Contains(stderr.String(), "ryzen-4500u-balanced-v1") {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
}

func TestHostAcceptance_MakeRequiresExplicitOutput(t *testing.T) {
	cmd := exec.Command("make", "-n", "bench-cli", "CLI_BENCH_PROFILE=ryzen-4500u-balanced-v1")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dry run: %v %s", err, out)
	}
	if strings.Contains(string(out), "--output \"docs/verification-evidence/004/latency.json\"") {
		t.Fatalf("host profile selects retained reference destination: %s", out)
	}
}

func TestHostAcceptance_FinalWriteCannotFollowReplacedPath(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "binary")
	output := filepath.Join(dir, "report.json")
	owned := filepath.Join(dir, "owned-report.json")
	if err := os.WriteFile(binary, []byte("test binary"), 0600); err != nil {
		t.Fatal(err)
	}
	sentinel := []byte("retained evidence")
	moved := false
	fake := func(b string, a, e []string) (int64, []byte, error) {
		if !moved {
			if err := os.Rename(output, owned); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(output, sentinel, 0600); err != nil {
				t.Fatal(err)
			}
			moved = true
		}
		return fakeProcess(b, a, e)
	}
	if code := runWith([]string{"--binary", binary, "--output", output, "--acceptance-profile", "ryzen-4500u-balanced-v1"}, io.Discard, io.Discard, fake, nil); code != 1 {
		t.Fatal(code)
	}
	data, err := os.ReadFile(output)
	if err != nil || !bytes.Equal(data, sentinel) {
		t.Fatalf("replacement changed: %q %v", data, err)
	}
	data, err = os.ReadFile(owned)
	if err != nil || !bytes.Contains(data, []byte("host_acceptance")) {
		t.Fatalf("owned report incomplete: %v", err)
	}
}

func TestHostAcceptance_MissingOutputDirectoryIsOperational(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(binary, []byte("test binary"), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "missing", "report.json")
	var stderr bytes.Buffer
	called := false
	fake := func(b string, a, e []string) (int64, []byte, error) { called = true; return fakeProcess(b, a, e) }
	code := runWith([]string{"--binary", binary, "--output", output, "--acceptance-profile", "ryzen-4500u-balanced-v1"}, io.Discard, &stderr, fake, nil)
	if code != 1 || called || !strings.Contains(stderr.String(), output) || strings.Contains(stderr.String(), "requires a new output file") {
		t.Fatalf("code=%d called=%v stderr=%q", code, called, stderr.String())
	}
}
