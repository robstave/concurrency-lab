# Error Groups (errgroup)

## Overview

This directory demonstrates using `golang.org/x/sync/errgroup` to run a group of goroutines collectively, propagate any errors that occur, and cancel the entire group via a shared context.

## Concepts

When you fan out work to multiple goroutines using a standard `sync.WaitGroup`, handling errors and fast failure can become messy. You typically have to set up your own error channels, mutex-protected variables, and pass context cancellation manually to stop remaining workers if one fails.

**`errgroup.Group`** streamlines this pattern.

### Features:
1. **Simplified waiting**: Similar to a WaitGroup, but easier. You call `g.Go(func() error { ... })` and then `err := g.Wait()`.
2. **Error propagation**: `g.Wait()` returns the *first* non-nil error returned by any of the goroutines.
3. **Context cancellation (Fail Fast)**: By creating the group with `errgroup.WithContext(ctx)`, the shared context is automatically canceled as soon as *any* goroutine returns an error. The other goroutines can watch `ctx.Done()` and cleanly stop working, preventing wasted resources.

### When to use:
- Running multiple independent sub-tasks that make up a single HTTP request (e.g., fetching user profile, settings, and friends list concurrently). Wait for all to finish, but fail the whole request immediately if one database call fails.

## Running the Exercise

```bash
go run .
```
