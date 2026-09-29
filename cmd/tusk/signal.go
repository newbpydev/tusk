package main

import (
	"context"
	"os"
	"os/signal"
)

// interruptContext restores default handling before publishing cancellation.
// A later signal can terminate a process whose graceful cleanup has stalled.
func interruptContext(signals ...os.Signal) (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	incoming := make(chan os.Signal, 1)
	signal.Notify(incoming, signals...)
	done := make(chan struct{})
	go func() {
		defer close(done)
		select {
		case <-incoming:
		case <-ctx.Done():
		}
		signal.Stop(incoming)
		cancel()
	}()
	return ctx, func() {
		cancel()
		<-done
	}
}
