//go:build ignore

// This executable belongs only to API boundary tests. It never invokes gh.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func main() {
	shell, script := os.Getenv("FIXTURE_GH_BASH"), os.Getenv("FIXTURE_GH_SCRIPT")
	if shell == "" || script == "" {
		fmt.Fprintln(os.Stderr, "owned gh fixture requires FIXTURE_GH_BASH and FIXTURE_GH_SCRIPT")
		os.Exit(127)
	}
	// Deadline controls stall this native child without leaving a Bash grandchild.
	if os.Getenv("FIXTURE_GH_STALL") == "1" {
		time.Sleep(time.Minute)
		return
	}
	cmd := exec.Command(shell, append([]string{script}, os.Args[1:]...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			code := exit.ExitCode()
			if status, ok := exit.Sys().(syscall.WaitStatus); ok && status.Signaled() {
				code = 128 + int(status.Signal())
			}
			if code < 0 {
				code = 1
			}
			os.Exit(code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(127)
	}
}
