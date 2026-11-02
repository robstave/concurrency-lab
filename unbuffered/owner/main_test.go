package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestConcurrentClientsAndShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	counter := NewCounter(ctx)
	var clients sync.WaitGroup
	for range 10 {
		clients.Add(1)
		go func() {
			defer clients.Done()
			for range 20 {
				if _, err := counter.Add(ctx, 1); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	clients.Wait()
	if value, err := counter.Add(ctx, 0); err != nil || value != 200 {
		t.Fatalf("value=%d error=%v", value, err)
	}
	canceled, stop := context.WithCancel(ctx)
	stop()
	if _, err := counter.Add(canceled, 100); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled call: %v", err)
	}
	if value, err := counter.Add(ctx, 0); err != nil || value != 200 {
		t.Fatalf("canceled request changed counter: value=%d error=%v", value, err)
	}
	cancel()
	select {
	case <-counter.done:
	case <-time.After(2 * time.Second):
		t.Fatal("owner did not stop")
	}
	if _, err := counter.Add(context.Background(), 1); err == nil {
		t.Fatal("stopped owner accepted a request")
	}
}

func TestAbandonedReplyDoesNotBlockOwner(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	counter := NewCounter(ctx)
	requestCtx, abandon := context.WithCancel(ctx)
	select {
	case counter.requests <- request{ctx: requestCtx, delta: 1, reply: make(chan int)}:
	case <-ctx.Done():
		t.Fatal("request not accepted")
	}
	// Nobody receives the reply. Canceling this request must free the owner.
	abandon()
	if _, err := counter.Add(ctx, 0); err != nil {
		t.Fatalf("owner stuck behind abandoned reply: %v", err)
	}
	cancel()
	<-counter.done
}
