package main

import (
	"context"
	"os"
	"os/signal"
)

// interruptContext restores prior signal handling before publishing cancellation.
// With the normal process disposition, a signal delivered after cancellation
// can force exit even during progressing cleanup. Rapid signals may coalesce;
// this is an escalation policy, not an exact signal-count guarantee.
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
