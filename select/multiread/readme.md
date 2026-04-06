# Multiplexing Reads (select)

## Overview

This exercise shows how to multiplex reads from multiple channels, falling back to a timeout if no data is received promptly.

## Concepts

### Channel Multiplexing
If you have multiple workers or upstream services sending data over different channels, you want to handle whichever one finishes first. A simple `select` block natively handles this pattern by providing a `case` for each distinct channel. Whichever channel receives data first executes its block.

### Timeouts with time.After
One of the most powerful paradigms of the Go `select` statement is pairing it with `time.After()`.
`time.After(duration)` actually returns a channel (specifically `<-chan time.Time`) that fires a single value when the specified duration has passed.

```go
select {
case res := <-fastChannel:
    fmt.Println(res)
case res := <-slowChannel:
    fmt.Println(res)
case <-time.After(2 * time.Second):
    fmt.Println("Timeout! Moving on.")
}
```

This ensures we don't hang forever if an upstream API fails or a goroutine stalls, adding resilience to the application via predictable bounding of operational times.

## Running the Exercise

```bash
go run .
```
