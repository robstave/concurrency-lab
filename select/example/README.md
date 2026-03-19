# Select statements (Basics)

## Overview

This example demonstrates non-blocking channel operations in Go using the `select` statement paired with a `default` case.

## Concepts

### The select keyword
The `select` statement lets a goroutine wait on multiple communication operations. 
It blocks until one of its `case` statements (which must be a complete channel operation) can proceed.

### Non-blocking operations
If you try to read from an unbuffered channel and no data is ready, your goroutine will block. However, sometimes you want to "attempt" a read, and if nothing is there, immediately do something else without waiting.
This is achieved by adding a `default:` case to a `select` block.
- If data is available immediately: the `case val := <-ch:` branch executes.
- If data is NOT available immediately: the `default:` branch executes. 
- There is no waiting.

```go
select {
case val := <-ch:
    fmt.Println("Received:", val)
default:
    fmt.Println("Nothing received, moving on!")
}
```

This pattern is fundamental for implementing polling or busy-wait loops safely.

## Running the Exercise

```bash
go run .
```
