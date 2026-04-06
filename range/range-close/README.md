# Channel Ranging (Closing)

## Overview

This exercise demonstrates why a channel **must** be closed if you intend to cleanly `range` over it.

## Concepts

### Deadlocks
A deadlock happens when all goroutines in a Go program are blocked, meaning no execution can proceed. The Go runtime checks for this and will crash the application with a `fatal error: all goroutines are asleep - deadlock!`.

### ranging
The idiomatic way to consume all values from a channel until it's "done" is to use a `for value := range ch` loop.
- The loop continues to block and wait for a new value from the channel as long as the channel implies there *might* be more data coming.
- If the sender stops sending but forgets to `close(ch)`, the `range` loop block indefinitely. Since there are no other goroutines executing, the runtime panics with a deadlock.

### Solution: Close
Always `close()` channels when no more data will be sent down them. Typically, this is done by the *sender*, optionally utilizing a `defer close(ch)` if appropriate, ensuring the receiving `range` loop can exit cleanly.

## Running the Exercise

```bash
go run .
```
