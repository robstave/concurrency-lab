
# Unbuffered Channel Example

This demonstrates the basic synchronization behavior of unbuffered channels in Go.

## Key Concepts

An unbuffered channel is created with `make(chan Type)` - no capacity specified.

### Blocking Behavior

- Send blocks until a receiver is ready
- Receive blocks until a sender is ready
- This creates a synchronization point between goroutines

### Example Flow

1. Main creates an unbuffered channel
2. Goroutine starts and attempts to send
3. Send blocks because no receiver is ready yet
4. Main reaches the receive operation
5. Both operations complete simultaneously
6. Both goroutines continue execution

### Why Unbuffered?

Unbuffered channels guarantee that:
- Data exchange happens synchronously
- Sender knows receiver got the value
- Creates a "handoff" between goroutines

This is useful for coordinating work and ensuring synchronization points.

## Running

```
go run main.go
```

You'll see the messages interleave, showing how the send and receive operations synchronize.
 
