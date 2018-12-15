package run

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestGroupGoExactlyOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	group := NewGroup(ctx)

	n := 10
	counts := make([]int, n)
	for i := 0; i < n; i++ {
		i := i
		err := group.Go(Func(func(ctx context.Context) error {
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
	group := NewGroup(ctx)
	errRun := errors.New("err run")
	if err := group.Go(Func(func(ctx context.Context) error {
		return errRun
	})); err != nil {
		t.Fatal(err)
	}
	if err := group.Wait(); err != errRun {
		t.Fatalf("expect error %v got %v", errRun, err)
	}
	if err := group.Go(Func(func(ctx context.Context) error {
		return nil
	})); err != errRun {
		t.Fatalf("expect error %v got %v", errRun, err)
	}
}

func TestGroupPool(t *testing.T) {
	t.Parallel()
	pool := NewGoroutinePool()
	group := NewGroup(context.Background(), Pool(pool))
	cnt := 0
	if err := group.Go(Func(func(context.Context) error {
		cnt++
		return nil
	})); err != nil {
		t.Fatal(err)
	}
	if err := group.Wait(); err != nil {
		t.Fatal(err)
	}
	if cnt != 1 {
		t.Fatalf("expect run exactly 1 time but got %d", cnt)
	}
}

func TestGroupLog(t *testing.T) {
	t.Parallel()
	w := &bytes.Buffer{}
	group := NewGroup(context.Background(), Log(func(info *LogInfo) {
		w.WriteString(info.String())
		w.WriteByte('\n')
	}))
	if err := group.Go(Func(func(context.Context) error {
		return nil
	})); err != nil {
		t.Fatal(err)
	}
	if err := group.Wait(); err != nil {
		t.Fatal(err)
	}
	wantLog := "h12.io/run.TestGroupLog.func2 starts\n" +
		"h12.io/run.TestGroupLog.func2 exits\n"
	if log := w.String(); log != wantLog {
		t.Fatalf("expect %s got %s", wantLog, log)
	}
}
