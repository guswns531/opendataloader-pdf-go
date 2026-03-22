package nativepdf

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestSkeletonLoaderReadsSampleLoremPDFArtifacts(t *testing.T) {
	loader := NewSkeletonLoader()
	document, err := loader.OpenPath(nil, filepath.Clean("../../../samples/pdf/lorem.pdf"), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenPath() error = %v", err)
	}
	defer document.Close()

	page, err := document.Page(0)
	if err != nil {
		t.Fatalf("Page(0) error = %v", err)
	}
	t.Logf("page metadata: %+v", page.Metadata())
	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeText: true})
	if err != nil {
		t.Fatalf("Artifacts() error = %v", err)
	}
	t.Logf("artifact count: %d", len(artifacts))
	if len(artifacts) == 0 {
		t.Fatalf("len(artifacts) = 0, want non-zero")
	}
	textCount := 0
	for _, artifact := range artifacts {
		if artifact != nil && artifact.Kind == model.ArtifactKindText {
			textCount++
			if strings.Contains(artifact.Text, "Lorem") {
				return
			}
		}
	}
	if textCount == 0 {
		t.Fatalf("len(text artifacts) = 0, want non-zero")
	}

	ingestor := NewIngestor(loader)
	doc, err := ingestor.Ingest(nil, core.Source{
		Path: filepath.Clean("../../../samples/pdf/lorem.pdf"),
		Name: "lorem.pdf",
	})
	if err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	if len(doc.Pages) == 0 || len(doc.Pages[0].Artifacts) == 0 {
		t.Fatalf("len(doc.Pages[0].Artifacts) = %d, want non-zero", len(doc.Pages[0].Artifacts))
	}
}
