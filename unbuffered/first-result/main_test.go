package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFailureDoesNotWinAndSlowLoserIsCanceled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err := FirstSuccess(ctx, []Source{
		{Name: "failure", Fail: true},
		{Name: "winner", Delay: 10 * time.Millisecond},
		{Name: "loser", Delay: time.Hour},
	})
	if err != nil || result.Source != "winner" {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	if ctx.Err() != nil {
		t.Fatal("waited for parent deadline instead of canceling loser")
	}
}

func TestAllFailedAndEmpty(t *testing.T) {
	if _, err := FirstSuccess(context.Background(), nil); err == nil {
		t.Fatal("accepted no sources")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := FirstSuccess(ctx, []Source{{Name: "a", Fail: true}, {Name: "b", Fail: true}}); err == nil {
		t.Fatal("all failures returned success")
	}
}

func TestParentCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := FirstSuccess(ctx, []Source{{Name: "slow", Delay: time.Hour}}); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestSimultaneousResultsDoNotLeaveBlockedSenders(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var sources []Source
	for range 20 {
		sources = append(sources, Source{Name: "ready"})
	}
	if _, err := FirstSuccess(ctx, sources); err != nil {
		t.Fatal(err)
	}
	if ctx.Err() != nil {
		t.Fatal("losing sends stalled shutdown")
	}
}
