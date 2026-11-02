package main

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestPipelineDrainsAndCloses(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var stages sync.WaitGroup
	output := square(ctx, &stages, produce(ctx, &stages, 5))
	var got []int
	for value := range output {
		got = append(got, value)
	}
	stages.Wait()
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
	if !reflect.DeepEqual(got, []int{1, 4, 9, 16, 25}) {
		t.Fatalf("got %v", got)
	}
}

func TestEarlyExitJoinsBlockedStages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var stages sync.WaitGroup
	output := square(ctx, &stages, produce(ctx, &stages, 100))
	select {
	case value := <-output:
		if value != 1 {
			t.Fatalf("got %d", value)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("first result did not arrive")
	}
	// Abandon output without draining it, just as an early-exiting consumer does.
	cancel()
	done := make(chan struct{})
	go func() { stages.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("a stage remained blocked after cancellation")
	}
}
