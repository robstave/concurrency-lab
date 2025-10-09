package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()
	g, gctx := errgroup.WithContext(ctx)

	// worker 1: sleeps and returns nil
	g.Go(func() error {
		select {
		case <-time.After(500 * time.Millisecond):
			fmt.Println("worker1 done")
			return nil
		case <-gctx.Done():
			fmt.Println("worker1 cancelled")
			return gctx.Err()
		}

	})

	// worker 2: returns an error quickly
	g.Go(func() error {
		select {
		case <-time.After(200 * time.Millisecond):
			return errors.New("worker2 failed")
		case <-gctx.Done():
			fmt.Println("worker3 cancelled")
			return gctx.Err()
		}
	})

	// worker 3: long-running, should be cancelled when error occurs
	g.Go(func() error {
		select {
		case <-time.After(2 * time.Second):
			fmt.Println("worker3 done")
			return nil
		case <-gctx.Done():
			fmt.Println("worker3 cancelled")
			return gctx.Err()
		}
	})

	if err := g.Wait(); err != nil {
		fmt.Println("errgroup finished with error:", err)
		return
	}

	fmt.Println("errgroup finished successfully")
}
