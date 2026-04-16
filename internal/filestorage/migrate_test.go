package filestorage

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestMigrateAttachmentsCopiesBlobsByStoragePath(t *testing.T) {
	t.Parallel()

	source, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocal source: %v", err)
	}
	target, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocal target: %v", err)
	}
	if _, err := source.Put(context.Background(), "farm-1/a.pdf", strings.NewReader("alpha"), 10); err != nil {
		t.Fatalf("source.Put a: %v", err)
	}
	if _, err := source.Put(context.Background(), "farm-1/b.pdf", strings.NewReader("beta"), 10); err != nil {
		t.Fatalf("source.Put b: %v", err)
	}

	summary, err := MigrateAttachments(context.Background(), source, target, []MigrationAttachment{
		{ID: 1, StoragePath: "farm-1/a.pdf", FileName: "a.pdf"},
		{ID: 2, StoragePath: "farm-1/b.pdf", FileName: "b.pdf"},
	}, false)
	if err != nil {
		t.Fatalf("MigrateAttachments: %v", err)
	}
	if summary.Scanned != 2 || summary.Copied != 2 || summary.Failed != 0 {
		t.Fatalf("summary = %+v", summary)
	}

	rc, err := target.Open(context.Background(), "farm-1/a.pdf")
	if err != nil {
		t.Fatalf("target.Open: %v", err)
	}
	defer rc.Close()
	body, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(body) != "alpha" {
		t.Fatalf("body = %q, want alpha", string(body))
	}
}

func TestMigrateAttachmentsDryRun(t *testing.T) {
	t.Parallel()

	source, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocal source: %v", err)
	}
	target, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocal target: %v", err)
	}
	if _, err := source.Put(context.Background(), "farm-1/a.pdf", strings.NewReader("alpha"), 10); err != nil {
		t.Fatalf("source.Put: %v", err)
	}

	summary, err := MigrateAttachments(context.Background(), source, target, []MigrationAttachment{
		{ID: 1, StoragePath: "farm-1/a.pdf", FileName: "a.pdf"},
	}, true)
	if err != nil {
		t.Fatalf("MigrateAttachments: %v", err)
	}
	if summary.Copied != 1 || summary.Failed != 0 {
		t.Fatalf("summary = %+v", summary)
	}
	if _, err := target.Open(context.Background(), "farm-1/a.pdf"); err == nil {
		t.Fatal("target blob exists after dry run, want absent")
	}
}
