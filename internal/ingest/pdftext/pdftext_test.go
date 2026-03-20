package pdftext

import (
	"path/filepath"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
)

func TestIngestRejectsEmptySource(t *testing.T) {
	_, err := New().Ingest(nil, core.Source{})
	if err == nil {
		t.Fatal("expected error for empty source")
	}
}

func TestIngestSamplePDF(t *testing.T) {
	path := filepath.Clean("../../../samples/pdf/lorem.pdf")

	document, err := New().Ingest(nil, core.Source{
		Path: path,
		Name: "lorem.pdf",
	})
	if err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	if document.Metadata.FileName != "lorem.pdf" {
		t.Fatalf("file name = %q, want %q", document.Metadata.FileName, "lorem.pdf")
	}
	if len(document.Pages) == 0 {
		t.Fatal("expected at least one page")
	}
	foundText := false
	for _, page := range document.Pages {
		if page != nil && len(page.Artifacts) > 0 {
			foundText = true
			break
		}
	}
	if !foundText {
		t.Fatal("expected at least one text artifact")
	}
}
