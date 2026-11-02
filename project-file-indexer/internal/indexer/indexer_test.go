package indexer

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestIndexAndDuplicates(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"a.txt": "same", "nested/b.txt": "same", "unique.txt": "different", "empty.txt": ""} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := Index(context.Background(), root, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Fatalf("got %d entries", len(entries))
	}
	for i, entry := range entries {
		if entry.Err != nil {
			t.Fatal(entry.Err)
		}
		if i > 0 && entries[i-1].Path > entry.Path {
			t.Fatal("entries not sorted")
		}
		data, err := os.ReadFile(entry.Path)
		if err != nil {
			t.Fatal(err)
		}
		if entry.SHA256 != fmt.Sprintf("%x", sha256.Sum256(data)) || entry.Size != int64(len(data)) {
			t.Fatalf("incorrect hash or size: %+v", entry)
		}
	}
	groups := Duplicates(entries)
	if len(groups) != 1 || len(groups[0]) != 2 {
		t.Fatalf("unexpected duplicates: %v", groups)
	}
	if groups[0][0] != filepath.Join(root, "a.txt") || groups[0][1] != filepath.Join(root, "nested/b.txt") {
		t.Fatalf("wrong duplicate paths: %v", groups)
	}
}

func TestErrorsAndCancellation(t *testing.T) {
	if _, err := Index(context.Background(), t.TempDir(), 0); err == nil {
		t.Fatal("accepted zero workers")
	}
	if _, err := Index(context.Background(), filepath.Join(t.TempDir(), "missing"), 2); err == nil {
		t.Fatal("missing root succeeded")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Index(ctx, t.TempDir(), 2); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	if groups := Duplicates([]Entry{{Path: "bad", SHA256: "abc", Err: errors.New("read failed")}, {Path: "ok", SHA256: "abc"}}); len(groups) != 0 {
		t.Fatal("failed file included in duplicates")
	}
}

func TestHashReadFailureAndCancellation(t *testing.T) {
	if entry := hash(context.Background(), filepath.Join(t.TempDir(), "missing")); entry.Err == nil {
		t.Fatal("missing file hashed successfully")
	}
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if entry := hash(ctx, path); !errors.Is(entry.Err, context.Canceled) {
		t.Fatalf("got %v", entry.Err)
	}
}
