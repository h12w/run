package run

import (
	"context"
	"sync"
)

// A Group is an error group that cancels the context on error, without all those options
//
// modified from https://golang.org/x/sync/errgroup
type Group struct {
	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup

	errOnce sync.Once
	err     error
}

// NewGroup creates a new GroupGroup
func NewGroup(ctx context.Context) *Group {
	internalCtx, cancel := context.WithCancel(ctx)
	return &Group{ctx: internalCtx, cancel: cancel}
}

// Cancel cancels the group
func (g *Group) Cancel() {
	g.cancel()
}

// Wait waits for all goroutines exit and returns the first returned error
func (g *Group) Wait() error {
	g.wg.Wait()
	g.cancel()
	return g.err
}

// Go runs the given runner in a goroutine
func (g *Group) Go(runner Runner) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := runner.Run(g.ctx); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				g.cancel()
			})
		}
	}()
}
