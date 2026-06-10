package logstore

import (
	"context"
	"strings"
	"testing"
)

func TestFileStoreWriteReadDelete(t *testing.T) {
	store := NewFileStore(t.TempDir())
	ctx := context.Background()

	path, size, err := store.Write(ctx, 2001, 1001, "hello\nworld\n")
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if !strings.HasPrefix(path, "logs/") {
		t.Fatalf("Write() path = %q, want logs/ prefix", path)
	}
	if !strings.Contains(path, "job-1001/log-2001.log") {
		t.Fatalf("Write() path = %q, want job/log path", path)
	}
	if size != int64(len("hello\nworld\n")) {
		t.Fatalf("Write() size = %d, want %d", size, len("hello\nworld\n"))
	}

	content, err := store.Read(ctx, path)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if content != "hello\nworld\n" {
		t.Fatalf("Read() = %q, want original content", content)
	}

	if err := store.Delete(ctx, path); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := store.Read(ctx, path); err == nil {
		t.Fatal("Read() after Delete() error = nil, want error")
	}
}

func TestFileStoreRejectsPathTraversal(t *testing.T) {
	store := NewFileStore(t.TempDir())
	ctx := context.Background()

	if _, err := store.Read(ctx, "../secret.log"); err == nil {
		t.Fatal("Read() error = nil, want path traversal rejection")
	}
	if err := store.Delete(ctx, "../secret.log"); err == nil {
		t.Fatal("Delete() error = nil, want path traversal rejection")
	}
}
