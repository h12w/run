package run

import (
	"context"
	"sync"
)

// A SimpleGroup is an error group that cancels the context on error, without all those options
//
// modified from https://golang.org/x/sync/errgroup
type SimpleGroup struct {
	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup

	errOnce sync.Once
	err     error
}

// NewSimpleGroup creates a new GroupSimpleGroup
func NewSimpleGroup(ctx context.Context) *SimpleGroup {
	internalCtx, cancel := context.WithCancel(ctx)
	return &SimpleGroup{ctx: internalCtx, cancel: cancel}
}

// Cancel cancels the group
func (g *SimpleGroup) Cancel() {
	g.cancel()
}

// Wait waits for all goroutines exit and returns the first returned error
func (g *SimpleGroup) Wait() error {
	g.wg.Wait()
	g.cancel()
	return g.err
}

// Go runs the given runner in a goroutine
func (g *SimpleGroup) Go(runner Runner) {
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
