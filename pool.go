package run

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrDispatchTimeout is returned when the context is cancelled when waiting for
// a task to be dispatched
var ErrDispatchTimeout = errors.New("failed to dispatch the goroutine due to timeout")

// Pool is an interface for a goroutine pool
type Pool interface {
	Go(ctx context.Context, fn func()) error
}

// dummyPool provides a dummy implementation satisifying the Pool interface
type dummyPool struct{}

// Go starts fn in s new goroutine and always returns nil
func (dummyPool) Go(ctx context.Context, fn func()) error {
	go fn()
	return nil
}

// GoroutinePool provides a goroutine pool, see the documentation for method Go
// for more information
type GoroutinePool struct {
	fnChan  chan func()
	idle    time.Duration
	maxChan chan struct{}

	closeOnce sync.Once
	quitChan  chan struct{}
	wg        sync.WaitGroup
}

// PoolOption is used to specify an option for GoroutinePool
type PoolOption func(p *GoroutinePool)

// IdleTime returns the option to specify the idle time before a goroutine exits
// and releases the resource if no task is submitted, if not specified, the
// default idle time is 1s
func IdleTime(idleTime time.Duration) PoolOption {
	if idleTime <= 0 {
		panic("idle time should always be positive")
	}
	return func(p *GoroutinePool) {
		p.idle = idleTime
	}
}

// Max returns the option to specifiy the maximum number of goroutines that a
// GoroutinePool can hold, if not specified, there is no upper limit
func Max(n int) PoolOption {
	return func(p *GoroutinePool) {
		p.maxChan = make(chan struct{}, n)
	}
}

// NewPool creates a new GoroutinePool based on the options provided
func NewPool(options ...PoolOption) *GoroutinePool {
	p := &GoroutinePool{
		fnChan:   make(chan func()),
		quitChan: make(chan struct{}),
		idle:     time.Second,
	}
	for _, opt := range options {
		opt(p)
	}
	return p
}

// Go tries to dispatch function fn onto its own goroutine.
// It returns nil if fn is successfully dispatched.
// It returns ErrClosed if the pool is already closed.
// It returns ErrDispatchTimeout if the context is cancelled when waiting for an
// idle goroutine to be available.
//
// A gouroutine will stay idle and be reused for a period specified by IdleTime
// option (default 1s).
//
// If Max option is specified, there will be a maximum limit on the goroutines
// that the pool can hold in total, and Go will block and wait for an idle
// goroutine is available. Otherwise, there is no limit on the goroutine number.
func (p *GoroutinePool) Go(ctx context.Context, fn func()) error {
	select {
	case <-p.quitChan:
		return ErrClosed
	default:
	}

	for {
		select {
		case p.fnChan <- fn:
			return nil
		default:
			if err := p.startGoroutine(ctx, fn); err != nil {
				return err
			}
			select {
			case p.fnChan <- fn:
				return nil
			case <-time.After(time.Millisecond):
			}
		}
	}
}

func (p *GoroutinePool) startGoroutine(ctx context.Context, fn func()) error {
	if p.maxChan != nil {
		select {
		case p.maxChan <- struct{}{}:
		case <-ctx.Done():
			return ErrDispatchTimeout
		}
	}
	p.wg.Add(1)
	started := &sync.WaitGroup{}
	started.Add(1)
	go func() {
		defer p.wg.Done()
		if p.maxChan != nil {
			defer func() {
				<-p.maxChan
			}()
		}
		started.Done()
		for {
			select {
			case fn := <-p.fnChan:
				fn()
			case <-time.After(p.idle):
				return
			case <-p.quitChan:
				return
			}
		}
	}()
	started.Wait()
	return nil
}

// Close stops the pool from accepting new tasks, waits for existing tasks
// complete and return nil. All subsequent calls will return ErrClosed
func (p *GoroutinePool) Close() error {
	first := false
	p.closeOnce.Do(func() {
		first = true
		close(p.quitChan)
		p.wg.Wait()
	})
	if !first {
		return ErrClosed
	}
	return nil
}
