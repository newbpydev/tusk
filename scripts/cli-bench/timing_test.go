package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestTimingDiagnostic_ParentIsolation(t *testing.T) {
	for _, tc := range []struct {
		mode            string
		gc, preallocate bool
	}{
		{"baseline", false, false}, {"parent-gc", true, false}, {"preallocated-pipe", false, true},
		{"parent-gc-and-preallocated-pipe", true, true}, {"child-gc-200", false, false},
		{"child-gc-400", false, false}, {"child-trace", false, false},
		{"child-procs-2", false, false}, {"child-procs-6", false, false},
	} {
		gc, preallocate := timingParentSetup(tc.mode)
		if gc != tc.gc || preallocate != tc.preallocate {
			t.Errorf("%s: GC %v preallocate %v", tc.mode, gc, preallocate)
		}
	}
}

// Diagnostic only. Compare child CPU time to launch-through-drain wall time,
// and test whether parent validation garbage/output allocation causes delays.
// These samples never replace the fixed reference report.
func timingParentSetup(mode string) (gc, preallocate bool) {
	return mode == "parent-gc" || mode == "parent-gc-and-preallocated-pipe",
		mode == "preallocated-pipe" || mode == "parent-gc-and-preallocated-pipe"
}

func timingDiagnostic(t *testing.T, binary, dir string, manifest map[string]string) {
	t.Helper()
	path := filepath.Join(dir, "timing.db")
	if _, err := seed(path, 1000, 128); err != nil {
		t.Fatal(err)
	}
	type sample struct {
		Mode                          string
		Index                         int
		Warmup                        bool
		WallNS, UserNS, SystemNS      int64
		ParentGCBefore, ParentGCAfter uint32
		ChildDiagnostic               string
	}
	var samples []sample
	args := []string{"list", "--all", "--json"}
	for _, mode := range []string{"baseline", "parent-gc", "preallocated-pipe", "parent-gc-and-preallocated-pipe", "child-gc-200", "child-gc-400", "child-trace", "child-procs-2", "child-procs-6"} {
		parentGC, preallocate := timingParentSetup(mode)
		var maxWall, maxCPU int64
		violations := 0
		for i := range 105 {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			cmd := exec.CommandContext(ctx, binary, args...)
			cmd.Env = processEnv(dir, path)
			childGC := strings.TrimPrefix(mode, "child-gc-")
			if childGC != mode {
				cmd.Env = append(cmd.Env, "GOGC="+childGC)
			}
			if mode == "child-trace" {
				cmd.Env = append(cmd.Env, "GODEBUG=gctrace=1")
			}
			childProcs := strings.TrimPrefix(mode, "child-procs-")
			if childProcs != mode {
				cmd.Env = append(cmd.Env, "GOMAXPROCS="+childProcs)
			}
			var out, stderr bytes.Buffer
			if preallocate {
				out.Grow(512 << 10)
			}
			cmd.Stdout, cmd.Stderr = &out, &stderr
			if parentGC {
				runtime.GC()
			}
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			start := time.Now()
			err := cmd.Run()
			elapsed := time.Since(start).Nanoseconds()
			runtime.ReadMemStats(&after)
			cancel()
			if err != nil || stderr.Len() != 0 && mode != "child-trace" {
				t.Fatalf("child: %v %s", err, stderr.String())
			}
			if err := validateOutput(args, out.Bytes(), 1000, 20); err != nil {
				t.Fatal(err)
			}
			s := sample{mode, i, i < 5, elapsed, cmd.ProcessState.UserTime().Nanoseconds(), cmd.ProcessState.SystemTime().Nanoseconds(), before.NumGC, after.NumGC, stderr.String()}
			samples = append(samples, s)
			if i >= 5 {
				maxWall = max(maxWall, s.WallNS)
				maxCPU = max(maxCPU, s.UserNS+s.SystemNS)
				if s.WallNS >= int64(15*time.Millisecond) {
					violations++
				}
			}
		}
		t.Logf("%s: max wall %.3fms; max child CPU %.3fms; violations %d/100", mode, float64(maxWall)/1e6, float64(maxCPU)/1e6, violations)
	}
	data, err := json.MarshalIndent(struct {
		Manifest map[string]string
		Samples  []sample
	}{manifest, samples}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("TUSK_CLI_CONDITIONS_OUTPUT"), append(data, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
}
