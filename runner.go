package run

import "context"

// Runner defines the Run method to be exeucuted within a goroutine
type Runner interface {
	Run(context.Context) error
}

// The RunnerFunc type is an adapter to allow the use of ordinary functions as
// runners. If f is a function with the appropriate signature, RunnerFunc(f) is
// a Runner that calls f.
type RunnerFunc func(context.Context) error

// Run calls f(ctx)
func (f RunnerFunc) Run(ctx context.Context) error { return f(ctx) }
