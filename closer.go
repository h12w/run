package run

import (
	"context"
	"errors"
	"io"
)

// Closer wraps a Runner into a Closer, whose Close method will cancel the
// runner and wait for its exit and return its error
func Closer(runner Runner) io.Closer {
	return newCloser(runner, false)
}

// WaitCloser wraps a Runner into a Closer, whose Close method will wait for
// the runner to exit and return its error
func WaitCloser(runner Runner) io.Closer {
	return newCloser(runner, true)
}

// ErrClosed is returned when the Closer is already closed
var ErrClosed = errors.New("run.Closer: already closed")

type closer struct {
	cancel  func()
	errChan chan error
	wait    bool
}

func newCloser(runner Runner, wait bool) io.Closer {
	ctx, cancel := context.WithCancel(context.Background())
	c := &closer{
		cancel:  cancel,
		errChan: make(chan error, 1),
		wait:    wait,
	}
	go func() {
		if err := runner.Run(ctx); err != nil {
			c.errChan <- err
		}
		// close errChan so that subsequent calls to Close will not block
		close(c.errChan)
	}()
	return c
}

// Close cancels the internal goroutine for runner, waits for its exit and
// returns the error returned by the runner. Subsequent calls to Close will
// return ErrClosed
func (c *closer) Close() error {
	if !c.wait {
		c.cancel()
	}
	err, ok := <-c.errChan
	if !ok {
		return ErrClosed
	}
	return err
}
