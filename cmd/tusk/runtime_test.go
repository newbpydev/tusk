package main

import (
	"io"
	"runtime"
	"testing"
)

func TestRuntimePolicy(t *testing.T) {
	previous := runtime.GOMAXPROCS(0)
	t.Cleanup(func() { runtime.GOMAXPROCS(previous) })
	for _, tc := range []struct {
		name, env     string
		initial, want int
	}{{"preserve runtime default", "", 2, 2}, {"explicit runtime setting", "2", 2, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GOMAXPROCS", tc.env)
			runtime.GOMAXPROCS(tc.initial)
			if code := run([]string{"tusk", "--help"}, io.Discard); code != 0 {
				t.Fatalf("help exit %d", code)
			}
			if got := runtime.GOMAXPROCS(0); got != tc.want {
				t.Fatalf("parallelism %d, want %d", got, tc.want)
			}
		})
	}
}
