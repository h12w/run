package gopool

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"
)

func newTestPoolSize(t *testing.T, n int) *GoroutinePool {
	pool := NewGoroutinePool()
	warmup(t, pool, n)
	return pool
}

func warmup(t *testing.T, pool *GoroutinePool, n int) *GoroutinePool {
	t.Helper()
	wg := &sync.WaitGroup{}
	quitChan := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		err := pool.Go(context.Background(), func() {
			defer wg.Done()
			<-quitChan // block all
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	close(quitChan) // release all
	wg.Wait()
	return pool
}

func TestPoolGoExactlyOnce(t *testing.T) {
	t.Parallel()

	pool := NewGoroutinePool()
	defer pool.Close()
	n := 10
	counts := make([]int, n)
	wg := &sync.WaitGroup{}
	for i := 0; i < n; i++ {
		i := i
		wg.Add(1)
		err := pool.Go(context.Background(), func() {
			defer wg.Done()
			counts[i]++
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
	for _, count := range counts {
		if count != 1 {
			t.Fatalf("expect each fn executed once but got %d times", count)
		}
	}
}

func TestPoolGoConcurrently(t *testing.T) {
	t.Parallel()

	doneChan := make(chan struct{})
	go func() {
		n := 10
		newTestPoolSize(t, n)
		close(doneChan)
	}()
	select {
	case <-doneChan:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("failed to run concurrently")
	}
}

func TestPoolNumGoroutines(t *testing.T) {
	numBefore := runtime.NumGoroutine()

	n := 10
	pool := newTestPoolSize(t, n)
	defer pool.Close()

	numPool := runtime.NumGoroutine()
	if num := numPool - numBefore; num != n {
		t.Fatalf("goroutines used: expect %d but got %d", n, num)
	}

	if err := pool.Close(); err != nil {
		t.Fatal(err)
	}
	numAfter := runtime.NumGoroutine()
	if numAfter != numBefore {
		t.Fatalf("expect goroutine number remains the same after pool closed, but got %d != %d", numAfter, numBefore)
	}
}

func TestPoolGoroutineReuse(t *testing.T) {
	t.Parallel()

	n := 10

	pool := NewGoroutinePool(Max(n))
	defer pool.Close()
	gidSet := make(map[uint64]bool)
	{
		wg := sync.WaitGroup{}
		quitChan := make(chan struct{})
		mu := sync.Mutex{}
		for i := 0; i < n; i++ {
			wg.Add(1)
			pool.Go(context.Background(), func() {
				defer wg.Done()
				mu.Lock()
				gidSet[getGID()] = true
				mu.Unlock()
				<-quitChan
			})
		}
		close(quitChan)
		wg.Wait()
	}

	gids := make([]uint64, 0, n)
	{
		wg := sync.WaitGroup{}
		mu := sync.Mutex{}
		for i := 0; i < n; i++ {
			wg.Add(1)
			pool.Go(context.Background(), func() {
				defer wg.Done()
				mu.Lock()
				gids = append(gids, getGID())
				mu.Unlock()
			})
		}
		wg.Wait()
	}

	notReused := 0
	for _, gid := range gids {
		if !gidSet[gid] {
			notReused++
		}
	}
	if notReused > 0 {
		t.Fatalf("expect goroutines are reused within Pool, %d not reused out of %d", notReused, len(gidSet))
	}
}

func TestPoolMultipleClose(t *testing.T) {
	t.Parallel()

	// if guarded by select default, it will be more difficult to repoduce a
	// multiple closing, but trying 1000 times will definitely do that
	for j := 0; j < 1000; j++ {
		pool := newTestPoolSize(t, 2)

		n := 10
		errChan := make(chan error, n)
		wg := &sync.WaitGroup{}
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				errChan <- pool.Close()
			}()
		}
		wg.Wait()
		close(errChan)
		successCount := 0
		errClosedCount := 0
		for err := range errChan {
			switch err {
			case nil:
				successCount++
			case ErrClosed:
				errClosedCount++
			}
		}

		if successCount != 1 {
			t.Fatalf("expect 1 success but got %d", successCount)
		}
		if errClosedCount != n-1 {
			t.Fatalf("expect %d ErrClosed but got %d", n-1, errClosedCount)
		}
	}
}

func TestPoolGoAfterClosed(t *testing.T) {
	t.Parallel()

	pool := newTestPoolSize(t, 10)
	if err := pool.Close(); err != nil {
		t.Fatal(err)
	}
	if err := pool.Go(context.Background(), func() {}); err == nil {
		t.Fatal("expect error if Go after Close but got nil")
	}
}

func TestPoolIdleTime(t *testing.T) {
	idle := time.Millisecond
	numBefore := runtime.NumGoroutine()

	n := 10
	pool := NewGoroutinePool(IdleTime(idle), Max(n))
	defer pool.Close()
	warmup(t, pool, n)

	numPool := runtime.NumGoroutine()
	if num := numPool - numBefore; num != n {
		t.Fatalf("goroutines used: expect %d but got %d", n, num)
	}

	// wait for time elapsed after idle time
	time.Sleep(5 * idle)

	numAfter := runtime.NumGoroutine()
	if numAfter != numBefore {
		t.Fatalf("expect goroutine number remains the same after idle time, but got %d != %d", numAfter, numBefore)
	}
}

func TestInvalidIdleTime(t *testing.T) {
	testcases := []struct {
		name string
		idle time.Duration
	}{
		{
			name: "minus idle time",
			idle: -time.Second,
		},
		{
			name: "zero idle time",
			idle: 0,
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("expect panic for invalid idle time")
				}
			}()
			IdleTime(tc.idle)
		})
	}
}

func TestPoolMaxSize(t *testing.T) {
	t.Parallel()

	pool := NewGoroutinePool(Max(1))
	defer pool.Close()
	quitChan := make(chan struct{})
	defer close(quitChan)
	if err := pool.Go(context.Background(), func() {
		<-quitChan
	}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := pool.Go(ctx, func() {}); err != ErrDispatchTimeout {
		t.Fatalf("expect dispatch timeout but got %v", err)
	}
}
