package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var buf bytes.Buffer
	exitCode := run([]string{"tusk", "--version"}, &buf)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	output := buf.String()
	if !strings.Contains(output, Version) {
		t.Fatalf("expected version %q in output, got %q", Version, output)
	}
}

func TestRunHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "flag -h", args: []string{"tusk", "-h"}},
		{name: "flag --help", args: []string{"tusk", "--help"}},
		{name: "subcommand help", args: []string{"tusk", "help"}},
		{name: "no arguments", args: []string{"tusk"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			exitCode := run(tc.args, &buf)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d", exitCode)
			}

			output := buf.String()
			if !strings.Contains(output, "Tusk - Zero-friction terminal task management system") {
				t.Fatalf("expected help banner in output, got %q", output)
			}
		})
	}
}
