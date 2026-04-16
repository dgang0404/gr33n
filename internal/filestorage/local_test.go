package filestorage

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestLocalPutOpenDelete(t *testing.T) {
	t.Parallel()

	store, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}

	written, err := store.Put(context.Background(), "farm-1/receipt.pdf", strings.NewReader("hello"), 10)
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if written != 5 {
		t.Fatalf("written = %d, want 5", written)
	}

	rc, err := store.Open(context.Background(), "farm-1/receipt.pdf")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer rc.Close()

	body, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(body) != "hello" {
		t.Fatalf("body = %q, want hello", string(body))
	}

	if err := store.Delete(context.Background(), "farm-1/receipt.pdf"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Open(context.Background(), "farm-1/receipt.pdf"); err == nil {
		t.Fatal("Open after Delete succeeded, want error")
	}
}

func TestLocalPutRejectsOversizedFile(t *testing.T) {
	t.Parallel()

	store, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}
	if _, err := store.Put(context.Background(), "farm-1/too-big.pdf", strings.NewReader("toolarge"), 4); err == nil {
		t.Fatal("Put oversized file succeeded, want error")
	}
}
