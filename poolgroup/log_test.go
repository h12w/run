package poolgroup

import (
	"context"
	"strings"
	"testing"
)

type namedRunner struct {
	name string
}

func (namedRunner) Run(context.Context) error { return nil }

func (r namedRunner) Name() string { return r.name }

type unnamedRunner struct {
}

func (unnamedRunner) Run(context.Context) error { return nil }

func testRunnerFunc(context.Context) error { return nil }

func TestName(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name string
		r    Runner
		want string
	}{
		{
			name: "named runner",
			r:    namedRunner{name: "test runner"},
			want: "test runner",
		},
		{
			name: "unnamed runner",
			r:    unnamedRunner{},
			want: "h12.io/run/pool.unnamedRunner",
		},
		{
			name: "unnamed runner pointer",
			r:    &unnamedRunner{},
			want: "h12.io/run/pool.unnamedRunner",
		},
		{
			name: "runner func",
			r:    Func(testRunnerFunc),
			want: "h12.io/run/pool.testRunnerFunc",
		},
		{
			name: "runner func by method",
			r:    Func(unnamedRunner{}.Run),
			want: "h12.io/run/pool.unnamedRunner.Run-fm",
		},
		{
			name: "nil",
			r:    nil,
			want: "nil",
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			name := logName(tc.r)
			if !strings.HasSuffix(name, tc.want) {
				t.Fatalf("expect %s got %v", tc.want, name)
			}
		})
	}
}
