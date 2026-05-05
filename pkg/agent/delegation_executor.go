package agent

import (
	"context"
	"sync"
)

type asyncDelegationExecutor struct {
	ctx    context.Context
	cancel context.CancelFunc
	sem    chan struct{}

	mu     sync.Mutex
	closed bool
	wg     sync.WaitGroup
}

func newAsyncDelegationExecutor(maxConcurrent int) *asyncDelegationExecutor {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &asyncDelegationExecutor{
		ctx:    ctx,
		cancel: cancel,
		sem:    make(chan struct{}, maxConcurrent),
	}
}

func (e *asyncDelegationExecutor) Submit(ctx context.Context, run func(context.Context)) error {
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
	case e.sem <- struct{}{}:
	default:
		return ErrAgentDelegationExecutorFull
	}
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		defer func() { <-e.sem }()
		run(e.ctx)
	}()
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
	}
	e.mu.Unlock()
	e.wg.Wait()
}
