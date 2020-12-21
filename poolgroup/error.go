package poolgroup

import "fmt"

// PanicError represents recovered panic info
type PanicError struct {
	Err   interface{}
	Stack []byte
}

// Error satisifies error interface
func (e *PanicError) Error() string {
	return fmt.Sprintf("%v\n%s", e.Err, e.Stack)
}
