package model

import "testing"

func TestDocumentAllocatesStableIDs(t *testing.T) {
	doc := NewDocument(DocumentMetadata{FileName: "sample.pdf"})

	if got := doc.NewNodeID(); got != 1 {
		t.Fatalf("first node ID = %d, want 1", got)
	}
	if got := doc.NewNodeID(); got != 2 {
		t.Fatalf("second node ID = %d, want 2", got)
	}
	if got := doc.NewArtifactID(); got != 1 {
		t.Fatalf("first artifact ID = %d, want 1", got)
	}
}

func TestEnsurePageCreatesAndReusesPage(t *testing.T) {
	doc := NewDocument(DocumentMetadata{FileName: "sample.pdf"})

	page := doc.EnsurePage(3)
	if page == nil {
		t.Fatal("EnsurePage returned nil")
	}
	if page.Metadata.Number != 3 {
		t.Fatalf("page number = %d, want 3", page.Metadata.Number)
	}
	if doc.Metadata.PageCount != 1 {
		t.Fatalf("page count = %d, want 1", doc.Metadata.PageCount)
	}

	again := doc.EnsurePage(3)
	if again != page {
		t.Fatal("EnsurePage should reuse existing page")
	}
}
