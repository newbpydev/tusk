// ghdeadline bounds maintainer GitHub operations; it is not linked into Tusk.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

func run(parent context.Context, timeout string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	limit := 5 * time.Minute
	if timeout != "" {
		var err error
		limit, err = time.ParseDuration(timeout)
		if err != nil || limit <= 0 || limit > 5*time.Minute {
			fmt.Fprintln(stderr, "GitHub deadline: GH_REQUEST_TIMEOUT must be positive and at most 5m")
			return 2
		}
	}
	ctx, cancel := context.WithTimeout(parent, limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr
	// Bound waiting for inherited pipes as well as the owned gh process.
	cmd.WaitDelay = time.Second
	err := cmd.Run()
	if ctx.Err() != nil {
		fmt.Fprintf(stderr, "GitHub deadline: %v; reconcile remote state before another write\n", ctx.Err())
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return 124
		}
		return 130
	}
	if err == nil {
		return 0
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() >= 0 {
		return exit.ExitCode()
	}
	fmt.Fprintf(stderr, "GitHub deadline: %v\n", err)
	return 127
}

func commandMain(args []string) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return run(ctx, os.Getenv("GH_REQUEST_TIMEOUT"), args, os.Stdin, os.Stdout, os.Stderr)
}

func main() { os.Exit(commandMain(os.Args[1:])) }
