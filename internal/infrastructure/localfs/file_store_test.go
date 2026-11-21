package localfs

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"uuid"
)

func TestFileStore_StoreOpenDelete(t *testing.T) {
	ctx := context.Background()
	store := NewFileStore(t.TempDir())
	fileID := uuid.NewV7()

	if err := store.Store(ctx, fileID, bytes.NewReader([]byte("hello world"))); err != nil {
		t.Fatalf("Store: %v", err)
	}

	r, err := store.Open(ctx, fileID)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading content: %v", err)
	}
	if string(b) != "hello world" {
		t.Errorf("content = %q, want %q", b, "hello world")
	}

	if err := store.Delete(ctx, fileID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := store.Open(ctx, fileID); err == nil {
		t.Error("Open after Delete = nil, want an error")
	}
}

func TestFileStore_DeleteMissingIsNotAnError(t *testing.T) {
	ctx := context.Background()
	store := NewFileStore(t.TempDir())

	if err := store.Delete(ctx, uuid.NewV7()); err != nil {
		t.Errorf("Delete on missing file = %v, want nil", err)
	}
}

func TestFileStore_StoreCreatesDirectory(t *testing.T) {
	ctx := context.Background()
	dir := filepath.Join(t.TempDir(), "nested", "files")
	store := NewFileStore(dir)
	fileID := uuid.NewV7()

	if err := store.Store(ctx, fileID, bytes.NewReader([]byte("data"))); err != nil {
		t.Fatalf("Store: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, fileID.String())); err != nil {
		t.Errorf("stat stored file: %v", err)
	}
}

func TestFileStore_StoreOverwritesExistingContent(t *testing.T) {
	ctx := context.Background()
	store := NewFileStore(t.TempDir())
	fileID := uuid.NewV7()

	if err := store.Store(ctx, fileID, bytes.NewReader([]byte("first"))); err != nil {
		t.Fatal(err)
	}
	if err := store.Store(ctx, fileID, bytes.NewReader([]byte("second"))); err != nil {
		t.Fatal(err)
	}

	r, err := store.Open(ctx, fileID)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "second" {
		t.Errorf("content = %q, want %q", b, "second")
	}
}
