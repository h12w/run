package run

import (
	"context"
	"sync"
)

// Group combines multiple concurrent tasks into one
type Group struct {
	ctx    context.Context
	cancel func()
	pool   Pool

	wg sync.WaitGroup

	errOnce sync.Once
	err     error
}

// NewGroup creates a new Group.
func NewGroup(ctx context.Context, pool Pool) *Group {
	if pool == nil {
		pool = dummyPool{}
	}
	ctx, cancel := context.WithCancel(ctx)
	return &Group{
		ctx:    ctx,
		cancel: cancel,
		pool:   pool,
	}
}

// Go runs the given runner in the internal goroutine pool.
// It returns nil when the goroutine is dispatched succesfully.
// It returns ErrDispatchTimeout if the context of the group is cancelled when
// waiting for an idle goroutine to be available.
// The first error return from a runner cancels the group, and all subsequent
// calls to Go as well as Wait will return the error
func (g *Group) Go(runner Runner) error {
	select {
	case <-g.ctx.Done():
		return g.Wait()
	default:
	}

	g.wg.Add(1)
	err := g.pool.Go(g.ctx, func() {
		defer g.wg.Done()
		if err := runner.Run(g.ctx); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	})
	if err == ErrDispatchTimeout {
		g.wg.Done()
	}
	return err
}

// Wait waits for all goroutines exit and returns the first returned error
func (g *Group) Wait() error {
	g.wg.Wait()
	return g.err
}
