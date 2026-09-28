package main

import (
	"context"
	"io"
	"os"

	"github.com/newbpydev/tusk/internal/cli"
)

// Version defines the current application release version.
const Version = "0.2.0-reboot"

func run(args []string, stdout io.Writer) int {
	return cli.Run(context.Background(), args[1:], cli.Options{Stdout: stdout, Stderr: os.Stderr, Version: Version, Getenv: os.Getenv, OpenService: openService})
}

func main() {
	ctx, stop := processContext()
	code := cli.Run(ctx, os.Args[1:], cli.Options{Stdout: os.Stdout, Stderr: os.Stderr, Version: Version, Getenv: os.Getenv, OpenService: openService})
	stop()
	os.Exit(code)
}
