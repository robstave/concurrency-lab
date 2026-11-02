package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func config() Config {
	return Config{Workers: 2, QueueSize: 1, MaxAttempts: 3, AttemptTimeout: time.Second, RetryDelay: time.Millisecond}
}

func TestRetriesAndExhaustion(t *testing.T) {
	attempts := 0
	failure := errors.New("transient")
	jobs := []Job{
		{ID: "retry", Task: func(context.Context) error {
			attempts++
			if attempts < 3 {
				return failure
			}
			return nil
		}},
		{ID: "fail", Task: func(context.Context) error { return failure }},
	}
	results, err := Run(context.Background(), jobs, config())
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results", len(results))
	}
	for _, r := range results {
		if r.Attempts != 3 {
			t.Fatalf("wrong attempts: %+v", r)
		}
		if r.ID == "retry" && r.Err != nil {
			t.Fatal(r.Err)
		}
		if r.ID == "fail" && !errors.Is(r.Err, failure) {
			t.Fatalf("lost job error: %+v", r)
		}
	}
}

func TestAttemptDeadlines(t *testing.T) {
	cfg := config()
	cfg.AttemptTimeout = 5 * time.Millisecond
	results, err := Run(context.Background(), []Job{{ID: "timeout", Task: func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }}}, cfg)
	if err != nil || len(results) != 1 {
		t.Fatalf("results=%v error=%v", results, err)
	}
	if results[0].Attempts != 3 || !errors.Is(results[0].Err, context.DeadlineExceeded) {
		t.Fatalf("unexpected result: %+v", results[0])
	}
}

func TestBoundedWorkersAndShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config()
	started := make(chan struct{}, 20)
	var active atomic.Int32
	var jobs []Job
	for range 20 {
		jobs = append(jobs, Job{Task: func(ctx context.Context) error {
			active.Add(1)
			defer active.Add(-1)
			started <- struct{}{}
			<-ctx.Done()
			return ctx.Err()
		}})
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err := Run(ctx, jobs, cfg)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("got %v", err)
		}
	}()
	for range cfg.Workers {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("workers did not start")
		}
	}
	if active.Load() != int32(cfg.Workers) {
		t.Fatal("wrong active worker count")
	}
	select {
	case <-started:
		t.Fatal("worker limit exceeded")
	default:
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown stalled")
	}
	if active.Load() != 0 {
		t.Fatal("workers still running after shutdown")
	}
}

func TestCancelDuringRetryDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config()
	cfg.RetryDelay = time.Hour
	results, err := Run(ctx, []Job{{Task: func(context.Context) error { cancel(); return errors.New("failed") }}}, cfg)
	if !errors.Is(err, context.Canceled) || len(results) != 1 || results[0].Attempts != 1 {
		t.Fatalf("results=%v error=%v", results, err)
	}
}

func TestValidationAndEmpty(t *testing.T) {
	for _, cfg := range []Config{{}, {Workers: 1, QueueSize: -1, MaxAttempts: 1, AttemptTimeout: time.Second}} {
		if _, err := Run(context.Background(), nil, cfg); err == nil {
			t.Fatal("accepted invalid configuration")
		}
	}
	if _, err := Run(context.Background(), []Job{{ID: "nil"}}, config()); err == nil {
		t.Fatal("accepted nil task")
	}
	if results, err := Run(context.Background(), nil, config()); err != nil || len(results) != 0 {
		t.Fatalf("results=%v error=%v", results, err)
	}
}
