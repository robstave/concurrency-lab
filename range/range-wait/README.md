# Channel Ranging (Wait)

## Overview

This subdirectory provides an exercise on ranging over a buffered channel populated concurrently by a background goroutine.

## Concepts

In Go, it's very common to have a producer-consumer setup:
1. **Producer goroutine**: generates data and sends it down a channel.
2. **Main goroutine**: consumes data off the channel and prints/processes it.

### Buffered Channels
By default, channels are *unbuffered*—meaning a sender will block until the receiver is ready to receive.
If you know ahead of time that you need to send 10 items rapidly, you can create a buffered channel `make(chan int, 10)`. The sender can load all 10 items into the channel without blocking, freeing it up to end and `close(ch)` faster.

### Putting it together
Here, a background goroutine fills a buffered channel with values and, importantly, closes it once done. The main function seamlessly retrieves the data with a continuous `range` loop that ends cleanly when the buffer is empty and the channel was marked closed.

## Running the Exercise

```bash
go run .
```
