// Package runner executes cooperative tasks with bounded concurrency and retries.
package runner

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Job struct {
	ID string
	// Task must honor ctx and be safe to retry. It runs synchronously in a worker.
	Task func(ctx context.Context) error
}

type Config struct {
	Workers        int
	QueueSize      int
	MaxAttempts    int
	AttemptTimeout time.Duration
	RetryDelay     time.Duration
}

type Result struct {
	ID       string
	Attempts int
	Err      error
}

// Run blocks until all workers exit. Cancellation stops admission, cancels active
// tasks, and abandons queued jobs. Results cover jobs actually picked up by a
// worker, in completion order. Task failures appear in Result.Err; configuration
// failures and parent cancellation are returned as the top-level error.
func Run(ctx context.Context, jobs []Job, cfg Config) ([]Result, error) {
	if cfg.Workers <= 0 || cfg.QueueSize < 0 || cfg.MaxAttempts <= 0 || cfg.AttemptTimeout <= 0 || cfg.RetryDelay < 0 {
		return nil, fmt.Errorf("require workers > 0, queue >= 0, attempts > 0, timeout > 0, retry delay >= 0")
	}
	for _, job := range jobs {
		if job.Task == nil {
			return nil, fmt.Errorf("job %q has no task", job.ID)
		}
	}
	queue := make(chan Job, cfg.QueueSize)
	completed := make(chan Result, cfg.Workers)
	producerDone := make(chan struct{})
	go func() {
		defer close(producerDone)
		defer close(queue)
		for _, job := range jobs {
			if ctx.Err() != nil {
				return
			}
			select {
			case queue <- job:
			case <-ctx.Done():
				return
			}
		}
	}()
	var wg sync.WaitGroup
	wg.Add(cfg.Workers)
	for range cfg.Workers {
		go func() {
			defer wg.Done()
			for {
				if ctx.Err() != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case job, ok := <-queue:
					if !ok {
						return
					}
					completed <- execute(ctx, job, cfg)
				}
			}
		}()
	}
	go func() { wg.Wait(); close(completed) }()
	var results []Result
	for result := range completed {
		results = append(results, result)
	}
	<-producerDone
	return results, ctx.Err()
}

func execute(ctx context.Context, job Job, cfg Config) Result {
	result := Result{ID: job.ID}
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			result.Err = err
			return result
		}
		attemptCtx, cancel := context.WithTimeout(ctx, cfg.AttemptTimeout)
		result.Attempts = attempt
		result.Err = job.Task(attemptCtx)
		if attemptCtx.Err() != nil {
			result.Err = attemptCtx.Err()
		}
		cancel()
		if result.Err == nil || attempt == cfg.MaxAttempts {
			return result
		}
		timer := time.NewTimer(cfg.RetryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			result.Err = ctx.Err()
			return result
		case <-timer.C:
		}
	}
	return result
}
