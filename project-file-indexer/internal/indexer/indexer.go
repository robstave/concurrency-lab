// Package indexer walks a directory and hashes regular files with bounded workers.
package indexer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type Entry struct {
	Path   string
	Size   int64
	SHA256 string
	Err    error
}

// Index skips symbolic links and special files. File read failures are returned
// on individual entries; traversal failures and cancellation return a top-level
// error along with partial entries. Queue capacity equals the worker count.
func Index(ctx context.Context, root string, workers int) ([]Entry, error) {
	if workers <= 0 {
		return nil, fmt.Errorf("workers must be positive")
	}
	jobs := make(chan string, workers)
	results := make(chan Entry, workers)
	walkDone := make(chan error, 1)
	go func() {
		defer close(jobs)
		walkDone <- filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil {
				return err
			}
			if d.IsDir() || d.Type()&os.ModeSymlink != 0 {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			select {
			case jobs <- path:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for {
				if ctx.Err() != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case path, ok := <-jobs:
					if !ok {
						return
					}
					entry := hash(ctx, path)
					select {
					case results <- entry:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
	go func() { wg.Wait(); close(results) }()
	var entries []Entry
	for entry := range results {
		entries = append(entries, entry)
	}
	walkErr := <-walkDone
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	if ctx.Err() != nil {
		return entries, ctx.Err()
	}
	return entries, walkErr
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

func hash(ctx context.Context, path string) Entry {
	entry := Entry{Path: path}
	f, err := os.Open(path)
	if err != nil {
		entry.Err = err
		return entry
	}
	defer f.Close()
	h := sha256.New()
	entry.Size, entry.Err = io.Copy(h, contextReader{ctx: ctx, r: f})
	if entry.Err == nil {
		entry.SHA256 = fmt.Sprintf("%x", h.Sum(nil))
	}
	return entry
}

// Duplicates groups successfully read files by digest, with deterministic order.
func Duplicates(entries []Entry) [][]string {
	byHash := make(map[string][]string)
	for _, entry := range entries {
		if entry.Err == nil && entry.SHA256 != "" {
			byHash[entry.SHA256] = append(byHash[entry.SHA256], entry.Path)
		}
	}
	var groups [][]string
	for _, paths := range byHash {
		if len(paths) > 1 {
			sort.Strings(paths)
			groups = append(groups, paths)
		}
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i][0] < groups[j][0] })
	return groups
}
