package main

import (
	"context"
	"runtime"
	"time"
)

// threadConfirmation keeps the synchronous reader pinned until cancellation has
// completed. CancelSynchronousIo can race entry into ReadFile (no pending IO), so
// cancellation is retried while awaiting the reader, never treated as a join.
func threadConfirmation(ctx context.Context, setup func() (read func() (byte, error), cancel func() error, release func(), err error)) (bool, error) {
	type admission struct {
		cancel func() error
		err    error
	}
	type result struct {
		yes bool
		err error
	}
	ready := make(chan admission)
	done := make(chan result, 1)
	releaseThread := make(chan struct{})
	joined := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		defer func() { runtime.UnlockOSThread(); close(joined) }()
		read, cancel, release, err := setup()
		ready <- admission{cancel, err}
		if err != nil {
			return
		}
		yes, err := readConfirmation(ctx, read)
		done <- result{yes, err}
		<-releaseThread
		release()
	}()
	admitted := <-ready
	if admitted.err != nil {
		<-joined
		return false, admitted.err
	}
	defer func() { close(releaseThread); <-joined }()
	select {
	case answer := <-done:
		if err := ctx.Err(); err != nil {
			return false, err
		}
		return answer.yes, answer.err
	case <-ctx.Done():
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			_ = admitted.cancel()
			select {
			case <-done:
				return false, ctx.Err()
			case <-ticker.C:
			}
		}
	}
}
