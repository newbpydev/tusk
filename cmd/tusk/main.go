package main

import (
	"context"
	"io"
	"os"

	"github.com/newbpydev/tusk/internal/cli"
)

func run(args []string, stdout io.Writer) int {
	return cli.Run(context.Background(), args[1:], cli.Options{Stdout: stdout, Stderr: os.Stderr, Version: Version, Getenv: os.Getenv, OpenService: openService, Terminal: terminalFacts, Confirm: confirm, RunTUI: tuiRunner(os.Stdin, stdout, os.Stderr, os.Getenv)})
}

func main() {
	ctx, stop := processContext()
	code := cli.Run(ctx, os.Args[1:], cli.Options{Stdout: os.Stdout, Stderr: os.Stderr, Version: Version, Getenv: os.Getenv, OpenService: openService, Terminal: terminalFacts, Confirm: confirm, RunTUI: tuiRunner(os.Stdin, os.Stdout, os.Stderr, os.Getenv)})
	stop()
	os.Exit(code)
}
