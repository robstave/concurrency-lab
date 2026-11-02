# Project: Trailing-edge debouncer

Keep the latest event in a burst and emit it after a quiet period. This complements
the token-based throttle in `../debounce`, which drops events immediately.

Requires Go 1.24+. From this directory:

```sh
go run ./cmd/debounce-demo
go run ./cmd/debounce-demo -quiet 250ms
go test ./...
```

The demo simulates typing every 40ms. With the default 150ms quiet period, only
`goroutine` is emitted. Smaller quiet periods can emit intermediate values.

`internal/debounce.Stream[T]` uses one goroutine, a resettable timer, and nil
channels to enable timer and output cases only when needed. A new input replaces
the pending value and resets the timer. Closing input flushes the pending value
immediately, even before its quiet period ends. Closing an empty input emits nothing.

Output is unbuffered. If the consumer is slow, newer input can replace the value
waiting to be emitted. Cancel the context when abandoning output, so a pending
send or close-time flush cannot leave a goroutine blocked. Cancellation normally
discards pending input; a send already ready at cancellation can still win the select.
The caller owns and closes input; the debouncer closes output. Ctrl+C cancels the demo.
