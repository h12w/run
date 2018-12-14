package run

import (
	"context"
	"errors"
	"testing"
)

func TestGroupGoExactlyOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	group := NewGroup(ctx, nil)

	n := 10
	counts := make([]int, n)
	for i := 0; i < n; i++ {
		i := i
		err := group.Go(RunnerFunc(func(ctx context.Context) error {
			counts[i]++
			return nil
		}))
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := group.Wait(); err != nil {
		t.Fatal(err)
	}
	for _, count := range counts {
		if count != 1 {
			t.Fatalf("expect each fn executed once but got %d times", count)
		}
	}
}

func TestGroupGoTaskError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	group := NewGroup(ctx, nil)
	errRun := errors.New("err run")
	if err := group.Go(RunnerFunc(func(ctx context.Context) error {
		return errRun
	})); err != nil {
		t.Fatal(err)
	}
	if err := group.Wait(); err != errRun {
		t.Fatalf("expect error %v got %v", errRun, err)
	}
	if err := group.Go(RunnerFunc(func(ctx context.Context) error {
		return nil
	})); err != errRun {
		t.Fatalf("expect error %v got %v", errRun, err)
	}
}
