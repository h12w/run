h12.io/run: graceful goroutine orchestration
============================================

[![GoDoc](https://godoc.org/h12.io/run?status.svg)](https://godoc.org/h12.io/run)

### Overview

While Go provides goroutines, channels and selects as first-class citizens to
support concurrent programming, it is not trivial to combine these elements
to address important concerns of goroutine orchestration, e.g. error handling,
panic recovery, goroutine leak prevention, goroutine reuse, goroutine throttle
and logging.

The package provides a mini-framework to address those cross-cutting concerns.

### Quick start

```bash
go get -u h12.io/run/gopoolgroup
```

Here is an example illustrating the usage of the goroutine pool and the group.
The task is described in the "Google Search 2.0" page from [this slide](https://talks.golang.org/2012/concurrency.slide#46).

```go
// the goroutine pool
pool := gopool.NewGoroutinePool(
	gopool.Max(8),                // the pool contains maximum 8 goroutines
	gopool.IdleTime(time.Minute), // a goroutine will stay in idle for maximum 1 minute before exiting
)

// the group
// the goroutine pool might have longer lifespan than the group
group := gopool.NewGroup(
	context.Background(), // a context that can cancel the whole group
	gopool.Pool(pool),       // the goroutine pool used by the group
	gopool.Recover(true),    // recover from panic and returns the PanicError
	gopool.Log(func(info *gopool.LogInfo) { // a log function for all starts/stops
		log.Print(info)
	}),
)

searches := []*GoogleSearch{
	{Search: Web, Query: "golang"},
	{Search: Image, Query: "golang"},
	{Search: Video, Query: "golang"},
}
for _, search := range searches {
	// start searching in parallel
	if err := group.Go(search); err != nil {
		log.Fatal(err)
	}
}

// wait for all searches stop
if err := group.Wait(); err != nil {
	log.Fatal(err)
}

for _, search := range searches {
	fmt.Println(search.Result)
}
```

See the full example [here](example/search/main.go).

### Design

The package is built around the concept of a runner.

```go
type Runner interface {
	Run(context.Context) error
}
```

Correct implementation of a runner should satisfy the following conditions:

* blocks when the work is on going
* returns when all work is done, an error occurred or context is cancelled

With goroutine pool and group in the package, the user does not need to use
the go statement explicitly, but only needs to implement their objects
satisfying the Runner interface.

A Group is useful when multiple concurrent sub-tasks needed to be combined as
a single task (the task failed when one of them failed, every sub-task should be
cancelled when the task is cancelled).

A Pool is useful when there are many short-lived goroutines.

A group can be built upon a pool, not vice versa.
