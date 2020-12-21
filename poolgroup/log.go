package poolgroup

import (
	"fmt"
	"reflect"
	"runtime"
)

// LogInfo is a logging event of a runner
type LogInfo struct {
	Runner Runner
	Event  Event
	Err    error
}

// Event enum of a runner
type Event int

// Event constants
const (
	Start Event = iota // runner starts
	Exit               // runner exits
)

// String representation of int enum
func (e Event) String() string {
	switch e {
	case Start:
		return "start"
	case Exit:
		return "exit"
	}
	return ""
}

// RunnerName returns a meaningful name of the runner for logging
func (li *LogInfo) RunnerName() string {
	return logName(li.Runner)
}

// LogInfo provides a default string representation of the LogInfo
func (li *LogInfo) String() string {
	errMsg := ""
	if li.Err != nil {
		errMsg = ", err=" + li.Err.Error()
	}
	return fmt.Sprintf("%s %vs", li.RunnerName(), li.Event) + errMsg
}

type namer interface {
	Name() string
}

// logName tried to get a meaningful name of a variable for logging purpose,
// meant to be used for logging the name of a runner.
//
// If it provides a Name() string method, it is returned.
// If it is a function (e.g. run.RunnerFunc), the full name of the function is
// returned.
// Otherwise, the full name of the concrete type is returned.
func logName(runner interface{}) string {
	if runner == nil {
		return "nil"
	}
	if n, ok := runner.(namer); ok {
		return n.Name()
	}
	typ := reflect.TypeOf(runner)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(runner).Pointer()).Name()
	}
	return typ.PkgPath() + "." + typ.Name()
}
