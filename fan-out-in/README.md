# Fan-Out / Fan-In Pattern

## Overview

This exercise illustrates the Fan-out / Fan-in concurrency pattern, allowing you to distribute jobs across multiple goroutines (workers) and collect their results back into a single pipeline.

## Concepts

### Fan-Out
Multiple functions (or goroutines) are mapped to read from the same channel until that channel is closed.
- **Why?** It distributes the workload among a pool of workers, maximizing CPU/IO usage.
- **Example**: If you have 100 images to download, you can fan out the workload by sending 100 URLs down a jobs channel, and starting 10 worker goroutines to read from it.

### Fan-In
A function reads from multiple input channels and multiplexes all of them onto a single output channel.
- **Why?** It allows a downstream consumer to process results from many workers without needing to know how many workers exist or polling multiple channels.
- **Mechanism**: Usually achieved by having each worker send its result to a shared `results` channel, and a separate goroutine waiting (via `sync.WaitGroup`) for all workers to finish before closing the `results` channel.

### Benefits of the pattern:
- Creates bounded concurrency: you can control exactly how many workers run simultaneously (e.g., preventing opening too many file descriptors).
- Clean separation of concerns: Generator -> Workers -> Consumer.

## Running the Exercise

```bash
go run .
```
