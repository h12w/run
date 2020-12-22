package gopool

import "context"

type fakeRunner struct {
	errRun error
	cntRun int
}

func (r *fakeRunner) Run(ctx context.Context) error {
	defer func() {
		r.cntRun++
	}()
	return r.errRun
}
