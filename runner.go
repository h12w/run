package run

import "context"

// Runner defines the Run method to be exeucuted within a goroutine
type Runner interface {
	Run(context.Context) error
}

// The Func type is an adapter to allow the use of ordinary functions as
// runners. If f is a function with the appropriate signature, Func(f) is
// a Runner that calls f.
type Func func(context.Context) error

// Run calls f(ctx)
func (f Func) Run(ctx context.Context) error { return f(ctx) }
