package pipeline

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
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
	Concurrency   int
	PerRequestTO  time.Duration
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

	results := make(chan Result, cfg.Concurrency)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(cfg.Concurrency) // bounded parallelism safeguard

	// Track errors to trigger cancellation
	errCh := make(chan error, cfg.Concurrency)
	g.Go(func() error {
		defer close(errCh)
		return nil
	})

	// Workers
	for i := 0; i < cfg.Concurrency; i++ {

		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case job, ok := <-jobs:
					if !ok {
						return nil
					}

					// Rate limit gate (blocks until a token is available or ctx canceled)
					if err := lim.Wait(ctx); err != nil {
						push(results, Result{URL: job.URL, Err: err})
						return err
					}

					start := time.Now()
					status, n, err := fetchWithRetry(ctx, client, job.URL, cfg.PerRequestTO)
					lat := time.Since(start)

					metrics.ObserveRequest(job.URL, status, lat, n, err == nil)
					push(results, Result{URL: job.URL, Status: status, Bytes: n, Latency: lat, Err: err})

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

	// Cancellation supervisor: if too many errors, cancel group
	g.Go(func() error {
		var errs int
		for err := range errCh {
			if err != nil {
				errs++
				if errs >= cfg.ErrorCancelAt {
					return errors.New("error threshold reached; canceling")
				}
			}
		}
		return nil
	})

	// Closer
	go func() {
		_ = g.Wait() // ignore returned error here; we still close results
		close(results)
	}()

	return results, nil
}

func push(ch chan<- Result, r Result) {
	select {
	case ch <- r:
	default:
		// if caller is slow at draining, don't block indefinitely: best-effort
		select {
		case ch <- r:
		case <-time.After(100 * time.Millisecond):
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
	start := time.Now()
	resp, err := client.Do(req)
	lat := time.Since(start)
	_ = lat // used in metrics at higher level
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	n, err := io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, n, err
}
