package pipeline

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"example.com/concurrency-lab/internal/metrics"
	"example.com/concurrency-lab/internal/ratelimit"
	"golang.org/x/sync/errgroup"
)

type Job struct{ URL string }

type Result struct {
	URL     string
	Status  int
	Bytes   int64
	Latency time.Duration
	Err     error
}

type Config struct {
	// Concurrency is the number of worker goroutines to run in parallel.
	// Higher values increase throughput up to the point you become limited by
	// external factors (network, remote server rate limits, CPU). Too high
	// can cause more contention or remote throttling.
	Concurrency int

	// PerRequestTO is the timeout applied to each individual HTTP request.
	// Every attempt (including retries) uses this as its per-attempt deadline.
	// If a request exceeds this duration, it's canceled and may be retried.
	PerRequestTO time.Duration

	// ErrorCancelAt sets a global error threshold. When workers collectively
	// report this many errors, the pipeline cancels the remaining work to
	// fail fast. Set to a large number to effectively disable early cancel.
	ErrorCancelAt int // cancel after N worker errors
}

// Run wires the pipeline. It returns a results channel that is closed when all work completes.
func Run(ctx context.Context, jobs <-chan Job, client *http.Client, lim *ratelimit.TokenBucket, cfg Config) (<-chan Result, error) {
	if cfg.Concurrency <= 0 {
		return nil, fmt.Errorf("invalid concurrency: %d", cfg.Concurrency)
	}
	if cfg.ErrorCancelAt <= 0 {
		cfg.ErrorCancelAt = 1 << 30
	} // effectively no cap unless set

	// Results channel: workers will send Result values here. We buffer it
	// with the concurrency size so that a fast producer (workers) doesn't
	// immediately block if the main goroutine is slightly slower at reading.
	// This is a small optimization and keeps backpressure reasonable.
	results := make(chan Result, cfg.Concurrency)

	// errgroup.WithContext returns a Group that runs goroutines and
	// cancels the provided context if any goroutine returns a non-nil error.
	// We use it to manage worker lifecycles and coordinated cancellation.
	g, ctx := errgroup.WithContext(ctx)

	// errCh is a small buffered channel used to count how many worker
	// errors occurred. Workers will attempt to send errors here without
	// blocking; the cancellation supervisor reads from this channel and
	// triggers a cancellation if too many errors accumulate.
	errCh := make(chan error, cfg.Concurrency)
	// We track worker completion with a WaitGroup and close errCh when all
	// workers have exited. This allows the supervisor to range over errCh
	// and finish cleanly without deadlocks or premature closure.
	var wg sync.WaitGroup
	wg.Add(cfg.Concurrency)

	// Start worker goroutines. Each worker continuously reads from the
	// jobs channel until it's closed. The outer loop creates exactly
	// cfg.Concurrency goroutines.
	for i := 0; i < cfg.Concurrency; i++ {

		g.Go(func() error {
			// Mark this worker as done when it exits.
			defer wg.Done()
			// Each worker runs until the context is canceled or the jobs
			// channel is closed and drained.
			for {
				select {
				case <-ctx.Done():
					// Group-level cancellation or parent context canceled.
					return ctx.Err()
				case job, ok := <-jobs:
					if !ok {
						// No more jobs: exit gracefully.
						return nil
					}

					// Rate limit gate: Wait will block until a token is
					// available or until ctx is canceled. This enforces the
					// global requests/second limit shared across workers.
					if err := lim.Wait(ctx); err != nil {
						// If Wait returned an error (context canceled or
						// timed out), we record a best-effort result and
						// return the error to let errgroup handle cancellation.

						// since this returns an error ( as opposed to just sending to the channel) it
						// is a  Group-level cancellation event
						push(results, Result{URL: job.URL, Err: err})
						return err
					}

					// Time the request for metrics and reporting.
					start := time.Now()
					status, n, err := fetchWithRetry(ctx, client, job.URL, cfg.PerRequestTO)
					lat := time.Since(start)

					// Record prometheus metrics for this request.
					metrics.ObserveRequest(job.URL, status, lat, n, err == nil)
					// Send the result to the results channel. push() will
					// handle slow consumers with a small timeout so workers
					// don't block forever.
					push(results, Result{URL: job.URL, Status: status, Bytes: n, Latency: lat, Err: err})

					// If there was an error, try to notify the supervisor
					// by sending to errCh. The send is non-blocking: if the
					// buffer is full, we drop the notification (best-effort).
					if err != nil {
						select {
						case errCh <- err:
						default:
						}
					}
				}
			}
		})
	}

	// Once all workers are done, close errCh so the supervisor can exit its range.
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Cancellation supervisor: this goroutine watches error notifications
	// coming from workers via errCh. If the count of non-nil errors reaches
	// cfg.ErrorCancelAt it returns an error which causes the errgroup's
	// context to be canceled and workers to stop.
	g.Go(func() error {
		var errs int
		for err := range errCh {
			if err != nil {
				errs++
				if errs >= cfg.ErrorCancelAt {
					// Returning an error from this goroutine triggers
					// cancellation of the errgroup context.
					return errors.New("error threshold reached; canceling")
				}
			}
		}
		return nil
	})

	// Closer: when the errgroup finishes (all workers exited), we close
	// the results channel so the caller's for-range loop can terminate.
	go func() {
		_ = g.Wait() // ignore returned error here; closing results is still important
		close(results)
	}()

	return results, nil
}

// push attempts to deliver a Result without stalling workers.
// Strategy:
// 1) Fast path: a non-blocking send. If the channel has room, deliver immediately.
// 2) Fallback: wait up to 100ms for space to free up; if still blocked, drop it.
//
// Trade-off: we may lose results under sustained backpressure, but workers remain
// responsive and don't block indefinitely behind a slow consumer.
func push(ch chan<- Result, r Result) {
	// Fast path: try to send immediately (non-blocking).
	select {
	case ch <- r:
		// delivered instantly
	default:
		// Channel is full right now. Give it a small grace period to clear.
		// If it doesn't, we drop this result to avoid stalling the pipeline.
		select {
		case ch <- r:
			// delivered within grace window
		case <-time.After(100 * time.Millisecond):
			// gave up: drop on purpose to keep workers moving
		}
	}
}

func fetchWithRetry(parent context.Context, client *http.Client, url string, perReq time.Duration) (int, int64, error) {
	// Up to 3 attempts with jittered backoff
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		ctx, cancel := context.WithTimeout(parent, perReq)
		status, n, err := doFetch(ctx, client, url)
		cancel()
		if err == nil && status < 500 { // treat 5xx as retryable
			return status, n, nil
		}
		lastErr = err
		// Jittered backoff: 200–600ms * 2^attempt
		base := time.Duration(200*(1<<attempt)) * time.Millisecond
		j, _ := rand.Int(rand.Reader, big.NewInt(400))
		time.Sleep(base + time.Duration(j.Int64())*time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("max retries exceeded")
	}
	return 0, 0, lastErr
}

func doFetch(ctx context.Context, client *http.Client, url string) (int, int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, 0, err
	}
	// Perform the request. Latency is measured at a higher level around
	// fetchWithRetry to include retries and waiting.
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	n, err := io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, n, err
}
