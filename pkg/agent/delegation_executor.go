package agent

import (
	"context"
	"sync"
)

type asyncDelegationTask func(context.Context)

type asyncDelegationExecutor struct {
	ctx    context.Context
	cancel context.CancelFunc
	queue  chan asyncDelegationTask

	mu     sync.Mutex
	closed bool
	wg     sync.WaitGroup
}

func newAsyncDelegationExecutor(maxConcurrent, maxQueued int) *asyncDelegationExecutor {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	if maxQueued <= 0 {
		maxQueued = maxConcurrent * 16
	}
	ctx, cancel := context.WithCancel(context.Background())
	e := &asyncDelegationExecutor{
		ctx:    ctx,
		cancel: cancel,
		queue:  make(chan asyncDelegationTask, maxQueued),
	}
	e.wg.Add(maxConcurrent)
	for range maxConcurrent {
		go e.worker()
	}
	return e
}

func (e *asyncDelegationExecutor) Submit(ctx context.Context, run asyncDelegationTask) error {
	if e == nil {
		return ErrAgentDelegationExecutorClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return ErrAgentDelegationExecutorClosed
	}
	select {
	case e.queue <- run:
	default:
		return ErrAgentDelegationExecutorFull
	}
	return nil
}

func (e *asyncDelegationExecutor) Close() {
	if e == nil {
		return
	}
	e.mu.Lock()
	if !e.closed {
		e.closed = true
		e.cancel()
		close(e.queue)
	}
	e.mu.Unlock()
	e.wg.Wait()
}

func (e *asyncDelegationExecutor) worker() {
	defer e.wg.Done()
	for run := range e.queue {
		if run != nil {
			run(e.ctx)
		}
	}
}
