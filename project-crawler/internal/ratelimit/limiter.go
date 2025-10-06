package ratelimit

import (
	"context"
	"time"
)

// TokenBucket is a simple leaky-bucket style limiter.
// rate tokens per window are produced; Wait consumes 1 token.
//
// This file intentionally keeps the implementation small and uses
// channels + a goroutine to produce tokens. The comments below explain
// every piece so a beginner can follow along.

type TokenBucket struct {
	ch     chan struct{}
	stopCh chan struct{}
}

func NewTokenBucket(rate int, window time.Duration) *TokenBucket {
	if rate <= 0 {
		rate = 1
	}
	ch := make(chan struct{}, rate)
	stop := make(chan struct{})

	// fill loop
	// We start a goroutine that will periodically add "tokens" into the
	// buffered channel `ch`. Each token is represented by an empty struct
	// value (struct{}), which uses zero memory and is idiomatic in Go when
	// you only need a signal and don't need to carry data.
	//
	// Why a buffered channel? The buffer size (capacity) defines how many
	// tokens can accumulate when the consumer is slower than the producer.
	// That accumulated amount is the "burst" capacity.
	go func() {
		// The ticker interval determines how often we attempt to add a token.
		// If window=1s and rate=5, interval = 200ms -> approx 5 tokens/sec.
		t := time.NewTicker(window / time.Duration(rate))
		defer t.Stop() // ensure we clean up the ticker when the goroutine exits

		// The loop runs until we receive a signal on `stop` telling it to exit.
		for {
			select {
			case <-t.C:
				// On each tick, try to add a token to `ch`.
				// The inner select uses a non-blocking send:
				// - If the channel has room, we place a token in it.
				// - If the channel is full (bucket full), the default case
				//   prevents the goroutine from blocking; we simply skip adding
				//   that token. This avoids leaks where the producer waits on
				//   a slow consumer.
				select {
				case ch <- struct{}{}:
					// token added successfully
				default:
					// bucket full; drop this token
				}

			case <-stop:
				// stop signal received: exit the goroutine cleanly.
				return
			}
		}
	}()

	return &TokenBucket{ch: ch, stopCh: stop}
}

func (tb *TokenBucket) Wait(ctx context.Context) error {
	// Wait blocks until either a token is available or the provided
	// context is done (canceled or timed out).
	//
	// We use a select with two cases:
	// - <-ctx.Done(): this unblocks when the caller cancels or when a
	//   deadline/timeout on the context is reached. Returning ctx.Err()
	//   lets the caller know why we stopped waiting (e.g., context.DeadlineExceeded).
	// - <-tb.ch: this receives a token from the internal channel and
	//   allows the caller to proceed. The token itself is an empty struct{}
	//   (no data), we only care about the signal that a token was available.
	//
	// Typical usage: the caller passes a context with a per-request timeout so
	// Wait doesn't block forever if something goes wrong.
	select {
	case <-ctx.Done():
		// The context has been canceled or timed out before a token
		// became available. Return the context error so the caller can
		// inspect the reason (canceled, deadline exceeded, etc.).
		return ctx.Err()
	case <-tb.ch:
		// We successfully consumed a token. The caller may now make the
		// request. Return nil to indicate success.
		return nil
	}
}

func (tb *TokenBucket) Stop() {
	close(tb.stopCh)
}
