# Unbuffered Channel with Timeout Example

This demonstrates using `select` with unbuffered channels to implement timeouts - showing both the power and the pitfalls.

## The Scenario

- **Worker**: Takes 2 seconds to process a task
- **Main**: Only waits 1 second before timing out
- **Result**: Timeout occurs, but with consequences!

## How Timeouts Work with Select

### Sending with Timeout
```go
select {
case w.taskChan <- task:
    // Successfully sent
case <-time.After(1 * time.Second):
    // Timeout - worker wasn't ready
}
```

### Receiving with Timeout
```go
select {
case result := <-w.doneChan:
    // Got result
case <-time.After(1 * time.Second):
    // Timeout - worker took too long
}
```

## What This Example Shows

### The Power
- Main doesn't hang forever waiting for a slow worker
- Can set SLAs and fail fast
- Better than blocking indefinitely

### The Problem: Goroutine Leak
When we timeout, the worker is still running! It will:
1. Finish processing (after 2 seconds)
2. Try to send the result on `doneChan`
3. **Block forever** because nobody is listening anymore

This orphaned goroutine is a **goroutine leak** - it stays in memory forever.

## The Real Solution

In production code, you need proper cancellation:

### Option 1: Context Cancellation
```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()

select {
case result := <-w.doneChan:
    return result, nil
case <-ctx.Done():
    // Signal worker to stop via another channel
    return "", ctx.Err()
}
```

### Option 2: Done Channel
Pass a done channel to the worker so it can check if it should abort:
```go
select {
case <-doneChan:
    return // abort processing
default:
    // continue working
}
```

### Option 3: Buffered Channels
Use a buffered channel for results so the worker can send and exit even if nobody receives:
```go
doneChan: make(chan string, 1)  // buffer of 1
```

## Key Takeaways

1. **Select enables timeouts** - This is where channels shine vs function calls
2. **Timeouts alone aren't enough** - You need cancellation mechanisms
3. **Goroutine leaks are real** - Always think about how goroutines exit
4. **Unbuffered channels + timeout = potential deadlock** - Worker gets stuck

## When to Use This Pattern

Timeouts with channels are powerful when:
- You need to enforce SLAs
- You're calling unreliable services
- You want to fail fast rather than hang
- **AND** you have proper cancellation mechanisms in place

## Running

```bash
go run main.go
```

You'll see the timeout occur, and note the warning about the orphaned goroutine. This demonstrates why proper cancellation is critical in real systems.
