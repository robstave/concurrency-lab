package debounce

import (
	"context"
	"testing"
	"time"
)

func receive[T any](t *testing.T, ch <-chan T) (T, bool) {
	t.Helper()
	select {
	case value, ok := <-ch:
		return value, ok
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not respond")
		var zero T
		return zero, false
	}
}

func TestQuietPeriodResetsAndEmitsLatest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	input := make(chan string)
	const quiet = 100 * time.Millisecond
	output, err := Stream(ctx, input, quiet)
	if err != nil {
		t.Fatal(err)
	}
	input <- "first"
	time.Sleep(20 * time.Millisecond)
	start := time.Now()
	input <- "latest"
	value, ok := receive(t, output)
	if !ok || value != "latest" {
		t.Fatalf("got %q, open=%v", value, ok)
	}
	if time.Since(start) < quiet-5*time.Millisecond {
		t.Fatal("emitted before the reset quiet period elapsed")
	}
	close(input)
	if _, ok := receive(t, output); ok {
		t.Fatal("duplicate output")
	}
}

func TestCloseFlushesPending(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	input := make(chan int)
	output, _ := Stream(ctx, input, time.Hour)
	input <- 1
	input <- 2
	close(input)
	if value, ok := receive(t, output); !ok || value != 2 {
		t.Fatalf("got %d, open=%v", value, ok)
	}
	if _, ok := receive(t, output); ok {
		t.Fatal("output should close after flush")
	}
}

func TestCancellationUnblocksPendingFlush(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	input := make(chan int)
	output, _ := Stream(ctx, input, time.Hour)
	input <- 42
	close(input) // No output consumer: cancellation must unblock the flush.
	cancel()
	if _, ok := receive(t, output); ok {
		// At most one pending value may race with cancellation.
		if _, ok := receive(t, output); ok {
			t.Fatal("multiple values after cancellation")
		}
	}
}

func TestSlowConsumerReceivesReplacement(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	input := make(chan int)
	output, _ := Stream(ctx, input, time.Millisecond)
	input <- 1
	// Allow the quiet period to expire without an output consumer.
	time.Sleep(20 * time.Millisecond)
	select {
	case input <- 2:
	case <-time.After(2 * time.Second):
		t.Fatal("slow consumer prevented input replacement")
	}
	close(input)
	if value, ok := receive(t, output); !ok || value != 2 {
		t.Fatalf("got %d, open=%v", value, ok)
	}
	if _, ok := receive(t, output); ok {
		t.Fatal("unexpected extra output")
	}
}

func TestEmptyAndInvalid(t *testing.T) {
	input := make(chan int)
	if _, err := Stream(context.Background(), input, 0); err == nil {
		t.Fatal("accepted zero quiet period")
	}
	close(input)
	output, _ := Stream(context.Background(), input, time.Second)
	if _, ok := receive(t, output); ok {
		t.Fatal("empty stream emitted a value")
	}
}
