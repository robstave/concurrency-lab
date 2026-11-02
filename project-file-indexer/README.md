# Project: Concurrent file indexer

Walk a directory, compute SHA-256 hashes with bounded workers, and report groups
of files with matching content. Files are read only; duplicates are never deleted.

Requires Go 1.24+. From this directory:

```sh
go run ./cmd/indexer -root . -workers 4
go test ./...
```

Use `-root ..` to scan the entire lab. The default root is the current directory;
the default worker count is the logical CPU count. Dot directories are included.
Symbolic links and special files are skipped. Output includes each path, byte
count, digest, and sorted duplicate groups. Empty files can form a duplicate group.

The walker feeds a channel with capacity equal to the worker count. Each worker
opens one file at a time and streams its bytes into a hash. A bounded results
channel pushes back on workers. One collector sorts the index after all workers
exit. File data is streamed, but the final index and duplicate groups live in memory.

Ctrl+C cancels walking, queued work, and reads between chunks, then waits for the
pipeline to exit. Cancellation cannot interrupt an OS file read already in progress.
Traversal errors stop discovery; already queued files are still processed. Individual
file read errors are reported per entry and excluded from duplicate groups. The CLI
returns a nonzero exit status for errors or cancellation and prints partial results.

Hashes describe the bytes read, not a filesystem snapshot: files changing during a
scan can produce inconsistent results. Hash matches are candidate duplicates; verify
contents before taking any destructive action.
