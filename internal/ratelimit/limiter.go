package ratelimit

import (
	"context"
	"time"
)

// TokenBucket is a simple leaky-bucket style limiter.
// rate tokens per window are produced; Wait consumes 1 token.

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
	go func() {
		t := time.NewTicker(window / time.Duration(rate))
		defer t.Stop()
		for {
			select {
			case <-t.C:
				select {
				case ch <- struct{}{}:
				default:
				}
			case <-stop:
				return
			}
		}
	}()

	return &TokenBucket{ch: ch, stopCh: stop}
}

func (tb *TokenBucket) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-tb.ch:
		return nil
	}
}

func (tb *TokenBucket) Stop() { close(tb.stopCh) }
