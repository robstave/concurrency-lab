# Project: Token Bucket Limiters

## Overview

This self-contained project demonstrates a **Token-Bucket rate limiter**, a fundamental algorithm used in API design to regulate traffic.

## Concepts

Rate limiting restricts how many actions can be performed within a given timeframe (e.g., 5 network requests per second). 

### The Token Bucket Algorithm
The simplest mental model for rate-limiting is a bucket filled with tokens:
- **Bucket Capacity**: Defines the maximum "burst" size (how many tokens can accumulate before they overflow and are lost).
- **Refill Rate**: A background process continually drops new tokens into the bucket at a steady, fixed rate.
- **Consumption**: Every time an action occurs (like handling a web request), one token is removed from the bucket.
- **Throttling**: If a request comes in and the bucket is empty, the action blocks or is rejected until a new token is added.

### Implementation in Go
While you can implement this manually using timers and buffered channels, Go provides an industry-standard implementation in the `golang.org/x/time/rate` package. 
It yields highly performant, lock-safe, and precise token distribution tracking, commonly utilized for managing high-volume concurrency pressure and enforcing SLA terms.

## Running the Project

```bash
go run .
```
