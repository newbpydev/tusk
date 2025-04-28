// Package worker provides utilities for concurrent task processing
package worker

import (
	"context"
	"sync"
)

// Task represents a unit of work to be processed by the worker pool
type Task func() error

// Pool represents a pool of workers that process tasks concurrently
type Pool struct {
	tasks   chan Task
	wg      sync.WaitGroup
	results chan error
	ctx     context.Context
	cancel  context.CancelFunc
	size    int
	started bool
	closed  bool
	mu      sync.Mutex // Protects started and closed flags
}

// NewPool creates a new worker pool with the specified number of workers
func NewPool(size int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		tasks:   make(chan Task, size*10), // Buffer size is 10x the worker count
		results: make(chan error, size*10),
		ctx:     ctx,
		cancel:  cancel,
		size:    size,
	}
}

// Start starts the worker pool
func (p *Pool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started || p.closed {
		return
	}

	p.started = true

	// Start the workers
	for i := 0; i < p.size; i++ {
		go p.worker()
	}
}

// worker is the main goroutine that processes tasks from the queue
func (p *Pool) worker() {
	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				// Channel closed, worker exits
				return
			}

			// Execute the task and send the result
			err := task()

			select {
			case p.results <- err:
				// Result successfully sent
			case <-p.ctx.Done():
				// Context canceled, exit worker
				return
			}

			p.wg.Done()

		case <-p.ctx.Done():
			// Context canceled, exit worker
			return
		}
	}
}

// Submit adds a task to the pool
// If the pool hasn't been started, it starts it automatically
func (p *Pool) Submit(task Task) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}

	if !p.started {
		p.started = true
		p.mu.Unlock()
		p.Start()
	} else {
		p.mu.Unlock()
	}

	p.wg.Add(1)

	select {
	case p.tasks <- task:
		// Task submitted successfully
	case <-p.ctx.Done():
		// Context canceled, don't submit
		p.wg.Done()
	}
}

// CollectResults collects results from tasks in a non-blocking way
// handler is called for each result
func (p *Pool) CollectResults(handler func(error)) {
	go func() {
		for {
			select {
			case err, ok := <-p.results:
				if !ok {
					// Results channel closed
					return
				}
				if handler != nil {
					handler(err)
				}
			case <-p.ctx.Done():
				return
			}
		}
	}()
}

// Wait blocks until all submitted tasks are completed
func (p *Pool) Wait() {
	p.wg.Wait()
}

// Stop stops the worker pool gracefully, waiting for all tasks to complete
func (p *Pool) Stop() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()

	// Wait for all tasks to complete
	p.Wait()

	// Signal all workers to exit
	p.cancel()

	// Close channels
	close(p.tasks)
	close(p.results)
}

// NewPoolWithContext creates a pool that is tied to a context
// The pool will be stopped when the context is canceled
func NewPoolWithContext(ctx context.Context, size int) *Pool {
	childCtx, cancel := context.WithCancel(ctx)
	pool := &Pool{
		tasks:   make(chan Task, size*10),
		results: make(chan error, size*10),
		ctx:     childCtx,
		cancel:  cancel,
		size:    size,
	}

	// Monitor parent context for cancellation
	go func() {
		select {
		case <-ctx.Done():
			pool.Stop()
		case <-childCtx.Done():
			// Pool was stopped by another call
		}
	}()

	return pool
}
