# Project: Bucket Limiters + Workers

## Overview

This directory provides an advanced self-contained project that connects a Token-Bucket rate limiter directly into a concurrent worker pool.

## Concepts

### The Goal
Building on the prior project, simply throttling a main execution loop is not enough. In a microservices or distributed job processing architecture, you typically have dozens of independent worker goroutines pulling from a common work queue simultaneously.

### The Challenge
If you have 10 workers independently making HTTP requests to an external API (which only allows 5 requests per second), how do you ensure the 10 workers don't collectively exceed the limit?

### The Solution: Shared Limiter
We create a single `rate.Limiter` and pass it to every worker goroutine. 
Inside the worker's processing loop, before they perform the restricted action (e.g., the HTTP call), they call `limiter.Wait(ctx)`.
This call is thread-safe. If multiple workers request a token simultaneously, the rate limiter queues them and ensures the global frequency limit is respected without any of the workers "clipping" or overstepping. This provides graceful wait times across the entire worker fleet.

## Running the Project

```bash
go run .
```
