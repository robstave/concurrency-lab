package ratelimit

import (
	"context"
	"time"
)

// Bucket is a token bucket with an explicit refill rate and a configurable burst capacity.
// - Refill at one token every interval
// - Burst allows up to 'burst' tokens to accumulate
// - Wait(ctx) consumes 1 token or returns ctx.Err()
//
// Example: NewBucket(5*time.Millisecond, 3) -> ~1 token/5ms, up to 3 tokens can accumulate.
//
// This is similar to the project crawler's limiter, but exposes burst directly for clarity.
// It also optionally pre-fills the bucket to its burst capacity on start (typical for TBs).

type Bucket struct {
	ch     chan struct{}
	stopCh chan struct{}
}

// NewBucket creates a bucket that refills 1 token per 'interval' with buffer 'burst'.
func NewBucket(interval time.Duration, burst int) *Bucket {
	if burst <= 0 {
		burst = 1
	}
	ch := make(chan struct{}, burst)
	stop := make(chan struct{})

	// Pre-fill to allow immediate burst.
	for i := 0; i < burst; i++ {
		ch <- struct{}{}
	}

	// Refill loop: every interval try to add a token; drop if full.
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				select {
				case ch <- struct{}{}:
					// added
				default:
					// full -> drop token
				}
			case <-stop:
				return
			}
		}
	}()

	return &Bucket{
		ch:     ch,
		stopCh: stop,
	}
}

// Wait blocks until a token is available or ctx is done.
func (b *Bucket) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.ch:
		return nil
	}
}

// Stop stops the refill loop.
func (b *Bucket) Stop() {
	close(b.stopCh)
}
