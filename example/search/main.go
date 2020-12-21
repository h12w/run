package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"h12.io/run/poolgroup"
)

type GoogleSearch struct {
	Query  string
	Search Search
	Result Result
}

type Result string

var (
	Web   = fakeSearch("web")
	Image = fakeSearch("image")
	Video = fakeSearch("video")
)

type Search func(ctx context.Context, query string) (Result, error)

func fakeSearch(kind string) Search {
	return func(ctx context.Context, query string) (Result, error) {
		// a real implementation will cancel and return when ctx is cancelled
		return Result(fmt.Sprintf("%s result for %q", kind, query)), nil
	}
}

func (s *GoogleSearch) Run(ctx context.Context) error {
	result, err := s.Search(ctx, s.Query)
	if err != nil {
		return err
	}
	s.Result = result
	return nil
}

func main() {
	// the goroutine pool
	pool := poolgroup.NewGoroutinePool(
		poolgroup.Max(8),                // the pool contains maximum 8 goroutines
		poolgroup.IdleTime(time.Minute), // a goroutine will stay in idle for maximum 1 minute before exiting
	)

	// the run group
	// the goroutine pool might have longer lifespan than the group
	group := poolgroup.NewGroup(
		context.Background(),    // a context that can cancel the whole group
		poolgroup.Pool(pool),    // the goroutine pool used by the group
		poolgroup.Recover(true), // recover from panic and returns the PanicError
		poolgroup.Log(func(info *poolgroup.LogInfo) { // a log function for all starts/stops
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
}
