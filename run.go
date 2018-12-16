// Package run provides graceful goroutine orchestration.
package run

import (
	"context"
	"fmt"
)

// Runner defines the Run method to be exeucuted within a goroutine
type Runner interface {
	Run(context.Context) error
}

// PanicError represents recovered panic info
type PanicError struct {
	Err   interface{}
	Stack []byte
}

// Error satisifies error interface
func (e *PanicError) Error() string {
	return fmt.Sprintf("%v\n%s", e.Err, e.Stack)
}

// The Func type is an adapter to allow the use of ordinary functions as
// runners. If f is a function with the appropriate signature, Func(f) is
// a Runner that calls f.
type Func func(context.Context) error

// Run calls f(ctx)
func (f Func) Run(ctx context.Context) error { return f(ctx) }
