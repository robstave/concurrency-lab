# Using WaitGroups

## Overview

This subdirectory contains a foundational exercise demonstrating basic `sync.WaitGroup` usage to wait for a pool of goroutines to finish executing.

## Concepts

### When to use WaitGroups?
If you spawn a goroutine `go doWork()`, the main function continues executing instantly. If the main function hits the end of the file, the program will terminate natively *without* waiting for `doWork()` to finish.
To prevent the program from ending prematurely, we need a mechanism to track pending goroutines and actively wait for them.

### `sync.WaitGroup`
A WaitGroup acts as a thread-safe counter.
1. Form the WaitGroup: `var wg sync.WaitGroup`
2. **`wg.Add(n)`**: Called *before* launching goroutines, it increments the internal counter by `n`. (e.g., if you are launching 5 workers, call `wg.Add(5)`).
3. **`wg.Done()`**: Called by each worker right as it finishes (often paired with `defer wg.Done()`), it decrements the counter by 1.
4. **`wg.Wait()`**: Called by the main function, it blocks execution line-progression until the internal counter hits zero.

### Important rules
- Pass pointers, not values! If you send a waitgroup to a function, you must pass it as `*sync.WaitGroup` so that the worker decrements the *same* counter the main function is observing, rather than a copied version.

## Running the Exercise

```bash
go run .
```
