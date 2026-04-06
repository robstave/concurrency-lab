# Unbuffered Channel with Context Cancellation

This demonstrates the **proper way** to handle timeouts and cancellation with unbuffered channels using Go's `context.Context`.

## The Fix

This is the clean version of the [cancel](../cancel) example. Instead of orphaning goroutines, we:
- Pass `context.Context` to the worker
- Worker checks context throughout its lifecycle
- Worker exits cleanly when context is cancelled

## How It Works

### 1. Worker Respects Context
```go
select {
case <-ctx.Done():
    fmt.Println("Context cancelled, shutting down")
    return
case task := <-w.taskChan:
    // process task
}
```

### 2. Context with Timeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()  // Always clean up!

select {
case result := <-w.doneChan:
    return result, nil
case <-ctx.Done():
    return "", ctx.Err()  // timeout or cancellation
}
```

### 3. Multiple Context Checks
The worker checks context at multiple points:
- **In the receive loop**: Before waiting for tasks
- **During processing**: While doing work
- **When sending results**: Before sending back to main

This ensures the worker can exit quickly when cancelled.

## Two Examples Shown

### Example 1: Timeout
- Worker takes 2 seconds
- Timeout is 1 second
- **Result**: Timeout occurs, BUT worker sees context cancellation and exits cleanly
- **No goroutine leak!**

### Example 2: Success
- Worker takes 2 seconds
- Timeout is 3 seconds
- **Result**: Task completes successfully

## Comparison with the 'cancel' Example

| Aspect | cancel (bad) | context (good) |
|--------|-------------|----------------|
| Timeout mechanism | `time.After` in select | `context.WithTimeout` |
| Worker awareness | Oblivious to timeout | Checks `ctx.Done()` |
| Goroutine leak | **Yes** - worker stuck forever | **No** - worker exits cleanly |
| Resource cleanup | Manual, error-prone | `defer cancel()` pattern |
| Idiomatic Go | No | **Yes** - standard pattern |

## Key Patterns

### Always defer cancel()
```go
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()  // Releases resources even if we return early
```

### Check context in loops
```go
for {
    select {
    case <-ctx.Done():
        return  // Exit immediately
    case work := <-workChan:
        // do work
    }
}
```

### Check context during long operations
```go
select {
case <-time.After(2 * time.Second):
    // work done
case <-ctx.Done():
    return  // abort early
}
```

## When to Use Context

Use `context.Context` when you need:
- Timeouts for operations
- Cancellation signals across goroutines
- Request-scoped values (not shown here)
- Deadline enforcement
- Graceful shutdown

**This is the standard Go pattern** for managing goroutine lifecycles.

## What This Solves

From the [cancel](../cancel) example, we had:
- ❌ Worker blocked forever trying to send result
- ❌ Goroutine leak
- ❌ No way to tell worker to stop

Now we have:
- ✅ Worker exits cleanly on cancellation
- ✅ No goroutine leaks
- ✅ Standard, idiomatic Go code
- ✅ Proper resource management

## Running

```bash
go run main.go
```

You'll see both examples: one that times out (but cleans up properly) and one that succeeds. Notice how the worker logs when it detects context cancellation.
