Atomic.Value configuration snapshots
===================================

This example demonstrates the idiomatic way to share configuration between goroutines using `sync/atomic.Value`.
One goroutine periodically publishes a brand‑new immutable config snapshot, while another goroutine reads the
current snapshot at a fixed interval without any locks.

Why atomic.Value
----------------
- Lock‑free reads: Readers call `Load()` to get a consistent pointer to the latest value. No mutexes, no contention.
- Snapshot semantics: Writers create a new value (copy‑on‑write) and `Store()` it. Readers never observe a partially
	written structure.
- Memory ordering: `Store` has release semantics and `Load` has acquire semantics, ensuring a happens‑before edge from
	writer to reader for the stored value and everything it references.

Key rules and pitfalls
----------------------
- First store sets the type: After the first `Store(x)`, all subsequent `Store` calls must store the exact same concrete
	type as `x` (or it panics). Use a single struct type for your snapshot and stick to it.
- No nil stores: `Store(nil)` panics. Initialize once with a non‑nil value before concurrent reads begin.
- Don’t mutate in place: Treat the stored value as immutable. If you mutate the struct a reader loaded, you’ll introduce
	data races. Always build a new struct (snapshot) and `Store` that.
- Zero value behavior: Before the first `Store`, `Load()` returns `nil`. Ensure you perform an initial `Store` or guard
	reads accordingly.

What this example does
----------------------
- Defines a `Config` struct with fields: `Version`, `Threshold`, and `FeatureEnabled`.
- Initializes an `atomic.Value` with version `1.0`.
- Updater goroutine:
	- Runs once per second, three times total.
	- Loads the current config, builds a new config (bumps version to `1.1`, `1.2`, `1.3`, tweaks other fields), and `Store`s it.
	- After publishing three updates, signals `done`.
- Reader goroutine:
	- Ticks every 500ms via `time.NewTicker`.
	- On each tick, `Load()`s and prints the current version.
	- Exits when it receives `done`.
- `WaitGroup` waits for both goroutines, then prints the final config snapshot.

Run it
------
From the `atomic` module folder:

```
go run ./value
```

Example output
--------------
Output timing will vary slightly, but you’ll see interleaved reader ticks and updater publications, then shutdown:

```
reader: current version=1.0
reader: current version=1.0
updater: published config version=1.1 (threshold=15 feature=true)
reader: current version=1.1
reader: current version=1.1
updater: published config version=1.2 (threshold=20 feature=false)
reader: current version=1.2
reader: current version=1.2
updater: published config version=1.3 (threshold=25 feature=true)
reader: current version=1.3
reader: done signal received, exiting
final config: version=1.3 threshold=25 feature=true
```

When to use this pattern
------------------------
- Hot‑path readers with occasional writes (feature flags, routing tables, pricing/config snapshots).
- Broadcasting updates safely to many goroutines without per‑read locking.

Alternatives
------------
- `sync.RWMutex`: Simpler to reason about when you must mutate in place, but adds read‑side lock/unlock overhead.
- Channels: Great for streaming updates to workers, but readers that start late need the latest snapshot; atomic.Value
	gives instant access to “the latest”. You can combine: channel for update notifications + atomic.Value for payload.

Exercises
---------
- Add more reader goroutines to see how reads scale without contention.
- Swap the `ticker` for `time.After` in a loop, or use a context for cancellation.
- Extend `Config` with nested structures and confirm readers still see consistent snapshots.

