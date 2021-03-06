package gopool

import (
	"bytes"
	"context"
	"errors"
	"strings"
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
	wantLogs := []string{
		"h12.io/run/gopool.TestGroupLog.func2 starts",
		"h12.io/run/gopool.TestGroupLog.func2 exits",
	}
	logs := strings.Split(strings.TrimSpace(w.String()), "\n")
	if len(logs) != len(wantLogs) {
		t.Fatal("wrong number of log lines")
	}
	for i := range logs {
		if !strings.HasSuffix(logs[i], wantLogs[i]) {
			t.Fatalf("expect %s got %s", wantLogs[i], logs[i])
		}
	}
}

func TestGroupRecoverPanic(t *testing.T) {
	t.Parallel()

	group := NewGroup(context.Background(), Recover(true))
	if err := group.Go(Func(func(context.Context) error {
		panic("test panic")
	})); err != nil {
		t.Fatal(err)
	}
	if err := group.Wait(); err == nil {
		t.Fatal("expect captured panic error but got nil")
	}

}
