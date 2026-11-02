package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"example.com/project-job-runner/internal/runner"
)

func main() {
	cfg := runner.Config{}
	flag.IntVar(&cfg.Workers, "workers", 3, "concurrent workers")
	flag.IntVar(&cfg.QueueSize, "queue", 4, "pending queue capacity")
	flag.IntVar(&cfg.MaxAttempts, "attempts", 3, "maximum attempts per job")
	flag.DurationVar(&cfg.AttemptTimeout, "timeout", 300*time.Millisecond, "deadline per attempt")
	flag.DurationVar(&cfg.RetryDelay, "retry-delay", 50*time.Millisecond, "delay between attempts")
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	var jobs []runner.Job
	for i := 1; i <= 8; i++ {
		attempts := 0 // Each job is owned by one worker, including its retries.
		jobs = append(jobs, runner.Job{ID: fmt.Sprintf("job-%d", i), Task: func(ctx context.Context) error {
			attempts++
			duration := 80 * time.Millisecond
			if i == 8 {
				duration = time.Second // Demonstrate exhausted attempt deadlines.
			}
			timer := time.NewTimer(duration)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
			}
			if i%3 == 0 && attempts == 1 {
				return fmt.Errorf("simulated transient failure")
			}
			return nil
		}})
	}
	results, err := runner.Run(ctx, jobs, cfg)
	failed := err != nil
	for _, result := range results {
		fmt.Printf("%s attempts=%d error=%v\n", result.ID, result.Attempts, result.Err)
		failed = failed || result.Err != nil
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if failed {
		os.Exit(1)
	}
}
