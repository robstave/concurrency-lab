# Unbuffered Pipeline: Backpressure and Early Exit

## Concept

Connect a number producer, a squaring stage, and a slow consumer with unbuffered
channels. A slow consumer makes the squaring stage wait, which eventually makes
the producer wait too. This is backpressure propagating upstream.

## Running

Requires Go 1.24+. From this directory:

```sh
go run main.go
go test -timeout 20s main.go main_test.go
```

The program runs two scenarios:

1. Consume all five squares: **1, 4, 9, 16, 25**. Closing the producer's channel
   lets the squaring stage finish and close its own output.
2. Generate up to ten numbers, but stop after **1, 4, 9**. Cancel the shared context
   so upstream stages exit even though the consumer no longer receives.

## Watch the Handoffs

`Producer: offering` appears before the send; `Producer: handed off` appears after
a receiver accepts that number. The consumer simulates 100ms of processing per
value. Watch upstream progress pause while it works. Concurrent log lines can
interleave, so their exact order is not guaranteed.

Unbuffered channels hold no queued values, but stages can still hold local work:
the squaring stage can accept the next number after its previous output is
received, even while the consumer processes that output. A handoff acknowledges
receipt, not completion of downstream processing.

## Shutdown Rules

Each stage closes only the output channel it owns. The consumer never closes an
upstream channel. Both receives and sends select on the shared context, including
the squaring stage's potentially blocked output send. Main cancels on early exit
and waits for both stage goroutines with a `WaitGroup`; it does not rely on sleep
to guess whether shutdown finished.

Removing cancellation from either blocked send can strand a goroutine when the
consumer stops. Merely breaking the consumer loop does not notify upstream.
Cancellation can race with a ready handoff, so a stage may do a little more work
before stopping; the example guarantees all stages are joined before returning.
