//go:build unix

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func processContext() (context.Context, func()) {
	pipe := make(chan os.Signal, 1)
	signal.Notify(pipe, syscall.SIGPIPE)
	ctx, stop := interruptContext(os.Interrupt, syscall.SIGTERM)
	return ctx, func() { stop(); signal.Stop(pipe) }
}
