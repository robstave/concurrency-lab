package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"example.com/project-file-indexer/internal/indexer"
)

func main() {
	root := flag.String("root", ".", "directory to index (read only)")
	workers := flag.Int("workers", runtime.NumCPU(), "maximum concurrent file reads")
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	entries, err := indexer.Index(ctx, *root, *workers)
	failed := err != nil
	for _, entry := range entries {
		if entry.Err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", entry.Path, entry.Err)
			failed = true
			continue
		}
		fmt.Printf("%s  %8d  %s\n", entry.SHA256, entry.Size, entry.Path)
	}
	groups := indexer.Duplicates(entries)
	for i, paths := range groups {
		fmt.Printf("duplicate group %d:\n", i+1)
		for _, path := range paths {
			fmt.Println("  ", path)
		}
	}
	fmt.Printf("indexed %d files; %d duplicate groups\n", len(entries), len(groups))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if failed {
		os.Exit(1)
	}
}
