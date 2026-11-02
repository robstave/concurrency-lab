package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Source struct {
	Name  string
	Delay time.Duration
	Fail  bool
}

type Result struct {
	Source string
	Value  string
	Err    error
}

// FirstSuccess returns the first successful response and joins every worker.
// A fast error does not win. Every wait and send can be canceled.
func FirstSuccess(parent context.Context, sources []Source) (Result, error) {
	if len(sources) == 0 {
		return Result{}, fmt.Errorf("no search sources")
	}
	ctx, cancel := context.WithCancel(parent)
	results := make(chan Result) // Many senders, one receiver, no buffered leftovers.
	var workers sync.WaitGroup
	defer func() {
		cancel() // Release losing workers before waiting for them.
		workers.Wait()
	}()
	for _, source := range sources {
		workers.Add(1)
		go func() {
			defer workers.Done()
			timer := time.NewTimer(source.Delay)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				fmt.Printf("%s: search canceled\n", source.Name)
				return
			case <-timer.C:
			}
			result := Result{Source: source.Name, Value: "found the answer"}
			if source.Fail {
				result.Err = fmt.Errorf("%s unavailable", source.Name)
			}
			// Without this cancellation case, a loser could remain blocked
			// forever after the caller accepts a successful response.
			select {
			case results <- result:
			case <-ctx.Done():
				fmt.Printf("%s: abandoned result canceled\n", source.Name)
			}
		}()
	}
	var failures []error
	for range sources {
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		case result := <-results:
			if result.Err == nil {
				return result, nil
			}
			fmt.Println("Ignoring failed response:", result.Err)
			failures = append(failures, result.Err)
		}
	}
	return Result{}, fmt.Errorf("all searches failed: %w", errors.Join(failures...))
}

func main() {
	fmt.Println("=== Unbuffered Channels: First Successful Result Wins ===")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err := FirstSuccess(ctx, []Source{
		{Name: "fast-but-failing", Delay: 20 * time.Millisecond, Fail: true},
		{Name: "winning-search", Delay: 80 * time.Millisecond},
		{Name: "slow-search", Delay: 400 * time.Millisecond},
	})
	if err != nil {
		fmt.Println("Search failed:", err)
		return
	}
	fmt.Printf("Winner: %s: %s\n", result.Source, result.Value)
	fmt.Println("Main: all search workers have stopped")
}
