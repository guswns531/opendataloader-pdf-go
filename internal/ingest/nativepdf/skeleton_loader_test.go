package nativepdf

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkeletonLoaderOpenReaderBuildsHandleShell(t *testing.T) {
	loader := NewSkeletonLoader()

	handle, err := loader.OpenReader(context.Background(), "sample.pdf", strings.NewReader("%PDF-1.7"), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	if handle.Metadata().FileName != "sample.pdf" {
		t.Fatalf("handle.Metadata().FileName = %q, want sample.pdf", handle.Metadata().FileName)
	}
	if handle.PageCount() != 0 {
		t.Fatalf("handle.PageCount() = %d, want 0", handle.PageCount())
	}
	if _, err := handle.Page(0); err == nil {
		t.Fatal("handle.Page(0) error = nil, want error")
	}
}

func TestSkeletonLoaderOpenPathBuildsHandleShell(t *testing.T) {
	loader := NewSkeletonLoader()
	dir := t.TempDir()
	path := filepath.Join(dir, "from-disk.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.7"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	handle, err := loader.OpenPath(context.Background(), path, OpenOptions{})
	if err != nil {
		t.Fatalf("OpenPath() error = %v", err)
	}
	defer handle.Close()

	if handle.Metadata().FileName != "from-disk.pdf" {
		t.Fatalf("handle.Metadata().FileName = %q, want from-disk.pdf", handle.Metadata().FileName)
	}
}
