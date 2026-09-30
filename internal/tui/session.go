package tui

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

const errRuntime core.Error = "interactive session failed"

type Result struct{ HadCommittedChanges, OutcomeUnknown bool }
type Factory func(context.Context) (ports.TaskService, func() error, error)

// Session owns admission and outcomes independently of Bubble Tea message
// delivery. Its mutex never protects or accesses Model state.
type Session struct {
	ctx             context.Context
	cancel          context.CancelFunc
	open            Factory
	mu              sync.Mutex
	active, stopped bool
	done            chan struct{}
	service         ports.TaskService
	closeOwner      func() error
	result          Result
	closeErr        error
	closeOnce       sync.Once
	cleanupDelay    time.Duration
	frozen          bool
}

func NewSession(ctx context.Context, open Factory) *Session {
	ctx, cancel := context.WithCancel(ctx)
	return &Session{ctx: ctx, cancel: cancel, open: open, cleanupDelay: 10 * time.Second}
}

func IsUnknown(err error) bool {
	var value ports.TransactionError
	var pointer *ports.TransactionError
	return errors.As(err, &value) || errors.As(err, &pointer)
}

func (s *Session) Call(ctx context.Context, mutation bool, work func(context.Context, ports.TaskService) (any, error)) (any, error) {
	return s.call(ctx, mutation, false, work)
}

func (s *Session) call(ctx context.Context, mutation, reopen bool, work func(context.Context, ports.TaskService) (any, error)) (value any, err error) {
	s.mu.Lock()
	if s.stopped || s.ctx.Err() != nil || ctx.Err() != nil {
		s.mu.Unlock()
		return nil, context.Canceled
	}
	if s.active {
		s.mu.Unlock()
		return nil, ports.ErrTransactionInUse
	}
	if s.closeErr != nil {
		s.mu.Unlock()
		return nil, s.closeErr
	}
	if s.frozen && !reopen {
		s.mu.Unlock()
		return nil, ports.ErrReadOnly
	}
	s.active = true
	s.done = make(chan struct{})
	s.mu.Unlock()
	defer func() {
		if recover() != nil {
			err = errRuntime
			if mutation {
				err = ports.NewTransactionError("tui", err)
			}
		}
		s.mu.Lock()
		if IsUnknown(err) {
			s.result.OutcomeUnknown = true
			s.frozen = true
		}
		if mutation && err == nil {
			s.result.HadCommittedChanges = true
		}
		s.active = false
		close(s.done)
		s.mu.Unlock()
	}()
	opctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	stop := context.AfterFunc(ctx, cancel)
	defer func() { stop(); cancel() }()
	if reopen {
		if s.closeOwner != nil {
			closeOwner := s.closeOwner
			s.closeOwner = nil
			s.service = nil
			if err = closeSafely(closeOwner); err != nil {
				s.closeErr = errors.Join(errRetireFailed, err)
				return nil, s.closeErr
			}
		}
	}
	if s.service == nil {
		service, closeOwner, openErr := s.open(opctx)
		if openErr != nil {
			return nil, openErr
		}
		if service == nil || closeOwner == nil {
			if closeOwner != nil {
				return nil, errors.Join(ports.ErrInvalidServiceOptions, closeSafely(closeOwner))
			}
			return nil, ports.ErrInvalidServiceOptions
		}
		s.service, s.closeOwner = service, closeOwner
	}
	value, err = work(opctx, s.service)
	if mutation && err == nil && value == nil {
		err = ports.NewTransactionError("tui", ports.ErrInvalidRecord)
	}
	if reopen && err == nil {
		s.mu.Lock()
		s.frozen = false
		s.mu.Unlock()
	}
	return value, err
}

func (s *Session) Load(ctx context.Context) ([]*core.TaskNode, error) {
	value, err := s.Call(ctx, false, func(ctx context.Context, svc ports.TaskService) (any, error) { return svc.GetTaskTree(ctx, "") })
	if err != nil {
		return nil, err
	}
	return value.([]*core.TaskNode), nil
}

func (s *Session) Reopen(ctx context.Context) ([]*core.TaskNode, error) {
	value, err := s.call(ctx, false, true, func(ctx context.Context, svc ports.TaskService) (any, error) { return svc.GetTaskTree(ctx, "") })
	if err != nil {
		return nil, err
	}
	return value.([]*core.TaskNode), nil
}

func (s *Session) Close(progress func()) (Result, error) {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.stopped = true
		s.cancel()
		done := s.done
		s.mu.Unlock()
		if done != nil {
			timer := time.NewTimer(s.cleanupDelay)
			select {
			case <-done:
			case <-timer.C:
				if progress != nil {
					progress()
				}
				<-done
			}
			timer.Stop()
		}
		if s.closeOwner != nil {
			s.closeErr = closeSafely(s.closeOwner)
			s.closeOwner = nil
		}
		if IsUnknown(s.closeErr) {
			s.result.OutcomeUnknown = true
		}
	})
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.result, s.closeErr
}

func closeSafely(closeOwner func() error) (err error) {
	defer func() {
		if recover() != nil {
			err = errRuntime
		}
	}()
	return closeOwner()
}
