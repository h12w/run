package run

import (
	"context"
	"sync"
)

// Group combines multiple concurrent tasks into one
type Group struct {
	ctx    context.Context
	cancel func()
	pool   GroupPool

	logFunc func(info *LogInfo)

	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

// GroupPool is an interface for a goroutine pool used by Group
type GroupPool interface {
	Go(ctx context.Context, fn func()) error
}

// GroupOption is used to specify an option for Group
type GroupOption func(*Group)

// Pool specifies the goroutine pool for a group, if not set, a dummy
// implementation is used (always starting new goroutines)
func Pool(p *GoroutinePool) GroupOption {
	return func(g *Group) {
		g.pool = p
	}
}

func Log(logFunc func(info *LogInfo)) GroupOption {
	return func(g *Group) {
		g.logFunc = logFunc
	}
}

// TODO:
func Recover(yes bool) GroupOption {
	return func(g *Group) {
		// g.recoverFunc = ......
	}
}

// NewGroup creates a new Group
func NewGroup(ctx context.Context, options ...GroupOption) *Group {
	ctx, cancel := context.WithCancel(ctx)
	g := &Group{
		ctx:    ctx,
		cancel: cancel,
		pool:   dummyPool{},
	}
	for _, opt := range options {
		opt(g)
	}
	return g
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

		if g.logFunc != nil {
			g.logFunc(&LogInfo{
				Runner: runner,
				Event:  Start,
			})
		}

		err := runner.Run(g.ctx)
		if err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
		if g.logFunc != nil {
			g.logFunc(&LogInfo{
				Runner: runner,
				Event:  Exit,
				Err:    err,
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
