package nativepdf

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestSkeletonLoaderOpenReaderBuildsHandleShell(t *testing.T) {
	loader := NewSkeletonLoader()

	handle, err := loader.OpenReader(context.Background(), "sample.pdf", strings.NewReader(twoPageShellPDF), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	if handle.Metadata().FileName != "sample.pdf" {
		t.Fatalf("handle.Metadata().FileName = %q, want sample.pdf", handle.Metadata().FileName)
	}
	if handle.PageCount() != 2 {
		t.Fatalf("handle.PageCount() = %d, want 2", handle.PageCount())
	}
	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}
	if page.Metadata().Size.Width != 612 || page.Metadata().Size.Height != 792 {
		t.Fatalf("page.Metadata().Size = %#v, want 612x792", page.Metadata().Size)
	}
}

func TestSkeletonLoaderOpenPathBuildsHandleShell(t *testing.T) {
	loader := NewSkeletonLoader()
	dir := t.TempDir()
	path := filepath.Join(dir, "from-disk.pdf")
	if err := os.WriteFile(path, []byte(onePageShellPDF), 0o644); err != nil {
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
	if handle.PageCount() != 1 {
		t.Fatalf("handle.PageCount() = %d, want 1", handle.PageCount())
	}
}

func TestSkeletonLoaderWithPagesExposesPageHandleShells(t *testing.T) {
	loader := NewSkeletonLoaderWithPages([]model.PageMetadata{
		{Number: 1, Index: 0, Size: model.PageSize{Width: 612, Height: 792}},
		{Number: 2, Index: 1, Size: model.PageSize{Width: 612, Height: 792}},
	})

	handle, err := loader.OpenReader(context.Background(), "sample.pdf", strings.NewReader("%PDF-1.7"), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	if handle.PageCount() != 2 {
		t.Fatalf("handle.PageCount() = %d, want 2", handle.PageCount())
	}
	page, err := handle.Page(1)
	if err != nil {
		t.Fatalf("handle.Page(1) error = %v", err)
	}
	if page.Metadata().Number != 2 {
		t.Fatalf("page.Metadata().Number = %d, want 2", page.Metadata().Number)
	}
}

func TestSkeletonLoaderWorksThroughNativeIngestorWithPageShells(t *testing.T) {
	loader := NewSkeletonLoaderWithPages([]model.PageMetadata{
		{Number: 1, Index: 0, Size: model.PageSize{Width: 612, Height: 792}},
	})
	ingestor := NewIngestor(loader)

	document, err := ingestor.Ingest(nil, core.Source{
		Name:   "sample.pdf",
		Reader: strings.NewReader("%PDF-1.7"),
	})
	if err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	if document.Metadata.FileName != "sample.pdf" {
		t.Fatalf("document.Metadata.FileName = %q, want sample.pdf", document.Metadata.FileName)
	}
	if len(document.Pages) != 1 {
		t.Fatalf("len(document.Pages) = %d, want 1", len(document.Pages))
	}
	if document.Pages[0].Metadata.Number != 1 {
		t.Fatalf("document.Pages[0].Metadata.Number = %d, want 1", document.Pages[0].Metadata.Number)
	}
}

const onePageShellPDF = `%PDF-1.7
1 0 obj << /Type /Page /MediaBox [0 0 612 792] >> endobj
`

const twoPageShellPDF = `%PDF-1.7
1 0 obj << /Type /Page /MediaBox [0 0 612 792] >> endobj
2 0 obj << /Type /Page /MediaBox [0 0 612 792] >> endobj
`
