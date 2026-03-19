# Debounce Pattern in Go

## Overview

This exercise implements a "debouncer" using Go channels and timed token emission.

## Concepts

**Debouncing** is a pattern often used in UI development or event handling to ensure that a function or event is not called too frequently. If an event fires rapidly in succession, a debouncer will ignore all but the last event, waiting until a specified quiet period has elapsed before acting.

In Go, we can implement debouncing elegantly using background goroutines, channels, and `time.After` or `time.Timer`.

### How it works:
1. Inputs arrive on an intake channel.
2. A goroutine monitors the intake channel via a `select` statement.
3. Each time an input arrives, a timer is reset.
4. If the timer ticks (via `time.After`), it means no new inputs arrived during the timeout window. The last received input is then emitted or processed.

### Real-world use cases:
- Save-as-you-type features (don't save to DB on every keystroke, wait until the user pauses).
- Auto-complete search boxes.
- Handling rapid webhook events where only the final system state matters.

## Running the Exercise

```bash
go run .
```
