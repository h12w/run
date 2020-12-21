package poolgroup

import (
	"context"
	"runtime"
	"sync"
)

// Group combines multiple concurrent tasks into one
type Group struct {
	ctx    context.Context
	cancel func()
	pool   GroupPool

	logFunc func(info *LogInfo)
	recover bool

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

// Log specifies the logging function for a group, if not set, the LogInfo is
// not generated
func Log(logFunc func(info *LogInfo)) GroupOption {
	return func(g *Group) {
		g.logFunc = logFunc
	}
}

// Recover specifies if a panic in the runner goroutine should be recovered or
// not, if not set, the default behavior is recovered.
func Recover(yes bool) GroupOption {
	return func(g *Group) {
		g.recover = yes
	}
}

// NewGroup creates a new Group
func NewGroup(ctx context.Context, options ...GroupOption) *Group {
	ctx, cancel := context.WithCancel(ctx)
	g := &Group{
		ctx:     ctx,
		cancel:  cancel,
		pool:    dummyPool{},
		recover: false,
	}
	for _, opt := range options {
		opt(g)
	}
	return g
}

// Go runs the given runner in the internal goroutine pool.
// It returns nil when the goroutine is dispatched successfully.
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
		if g.logFunc != nil {
			g.logFunc(&LogInfo{
				Runner: runner,
				Event:  Start,
			})
		}

		var err error
		defer func() {
			if g.recover {
				if r := recover(); r != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					err = &PanicError{Err: r, Stack: buf}
				}
			}
			if err != nil {
				g.setErrOnce(err)
				g.cancel()
			}
			if g.logFunc != nil {
				g.logFunc(&LogInfo{
					Runner: runner,
					Event:  Exit,
					Err:    err,
				})
			}
			g.wg.Done()
		}()

		err = runner.Run(g.ctx)
	})
	if err == ErrDispatchTimeout {
		g.wg.Done()
	}
	return err
}

// Cancel cancels the group
func (g *Group) Cancel() {
	g.cancel()
}

func (g *Group) setErrOnce(err error) {
	g.errOnce.Do(func() {
		g.err = err
	})
}

// Wait waits for all goroutines exit and returns the first returned error
func (g *Group) Wait() error {
	g.wg.Wait()
	g.cancel() // still cancel the context if all goroutines exit returning no errors
	return g.err
}
