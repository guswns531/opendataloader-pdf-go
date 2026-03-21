package nativepdf

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestBuildDocumentFromHandleAssignsDefaultsAndArtifactIDs(t *testing.T) {
	handle := &stubDocumentHandle{
		metadata: model.DocumentMetadata{},
		pages: []PageHandle{
			&stubPageHandle{
				metadata: model.PageMetadata{},
				artifacts: []*model.RawArtifact{
					{
						Kind:     model.ArtifactKindText,
						Text:     "hello",
						Sequence: 7,
						Bounds:   model.Box{Left: 1, Bottom: 2, Right: 3, Top: 4},
					},
				},
			},
		},
	}

	document, err := BuildDocumentFromHandle(handle, core.Source{Name: "sample.pdf"})
	if err != nil {
		t.Fatalf("BuildDocumentFromHandle() error = %v", err)
	}

	if document.Metadata.FileName != "sample.pdf" {
		t.Fatalf("document.Metadata.FileName = %q, want sample.pdf", document.Metadata.FileName)
	}
	if document.Metadata.PageCount != 1 {
		t.Fatalf("document.Metadata.PageCount = %d, want 1", document.Metadata.PageCount)
	}
	if len(document.Pages) != 1 {
		t.Fatalf("len(document.Pages) = %d, want 1", len(document.Pages))
	}

	page := document.Pages[0]
	if page.Metadata.Index != 0 {
		t.Fatalf("page.Metadata.Index = %d, want 0", page.Metadata.Index)
	}
	if page.Metadata.Number != 1 {
		t.Fatalf("page.Metadata.Number = %d, want 1", page.Metadata.Number)
	}
	if len(page.Artifacts) != 1 {
		t.Fatalf("len(page.Artifacts) = %d, want 1", len(page.Artifacts))
	}

	artifact := page.Artifacts[0]
	if artifact.ID == 0 {
		t.Fatalf("artifact.ID = 0, want non-zero")
	}
	if artifact.PageIndex != 0 {
		t.Fatalf("artifact.PageIndex = %d, want 0", artifact.PageIndex)
	}
	if artifact.PageNumber != 1 {
		t.Fatalf("artifact.PageNumber = %d, want 1", artifact.PageNumber)
	}
	if artifact.Sequence != 7 {
		t.Fatalf("artifact.Sequence = %d, want 7", artifact.Sequence)
	}
}

func TestIngestorUsesReaderAndPropagatesOptions(t *testing.T) {
	loader := &stubLoader{
		handle: &stubDocumentHandle{
			metadata: model.DocumentMetadata{},
			pages: []PageHandle{
				&stubPageHandle{
					metadata: model.PageMetadata{Number: 3},
				},
			},
		},
	}

	ingestor := NewIngestor(loader)
	ctx := core.NewProcessingContext(nil, core.ProcessingOptions{
		Strict: true,
		Extras: map[string]any{
			"password": "secret",
			"pages":    []int{2, 3},
		},
	})

	document, err := ingestor.Ingest(ctx, core.Source{
		Name:   "reader.pdf",
		Reader: strings.NewReader("%PDF"),
	})
	if err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}

	if !loader.usedReader {
		t.Fatalf("expected loader.OpenReader to be used")
	}
	if loader.lastName != "reader.pdf" {
		t.Fatalf("loader.lastName = %q, want reader.pdf", loader.lastName)
	}
	if !loader.lastOptions.Strict {
		t.Fatalf("loader.lastOptions.Strict = false, want true")
	}
	if loader.lastOptions.Password != "secret" {
		t.Fatalf("loader.lastOptions.Password = %q, want secret", loader.lastOptions.Password)
	}
	if len(loader.lastOptions.Pages) != 2 || loader.lastOptions.Pages[0] != 2 || loader.lastOptions.Pages[1] != 3 {
		t.Fatalf("loader.lastOptions.Pages = %#v, want []int{2,3}", loader.lastOptions.Pages)
	}
	if document.Metadata.FileName != "reader.pdf" {
		t.Fatalf("document.Metadata.FileName = %q, want reader.pdf", document.Metadata.FileName)
	}
	if document.Pages[0].Metadata.Number != 3 {
		t.Fatalf("document.Pages[0].Metadata.Number = %d, want 3", document.Pages[0].Metadata.Number)
	}
}

func TestIngestorUsesPathWhenReaderMissing(t *testing.T) {
	loader := &stubLoader{
		handle: &stubDocumentHandle{},
	}
	ingestor := NewIngestor(loader)

	if _, err := ingestor.Ingest(nil, core.Source{Path: "/tmp/sample.pdf"}); err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	if !loader.usedPath {
		t.Fatalf("expected loader.OpenPath to be used")
	}
	if loader.lastPath != "/tmp/sample.pdf" {
		t.Fatalf("loader.lastPath = %q, want /tmp/sample.pdf", loader.lastPath)
	}
}

func TestIngestorRejectsMissingSource(t *testing.T) {
	ingestor := NewIngestor(&stubLoader{})

	_, err := ingestor.Ingest(nil, core.Source{})
	if err == nil {
		t.Fatal("Ingest() error = nil, want error")
	}
}

type stubLoader struct {
	handle      DocumentHandle
	err         error
	usedPath    bool
	usedReader  bool
	lastPath    string
	lastName    string
	lastOptions OpenOptions
}

func (s *stubLoader) OpenPath(_ context.Context, path string, opts OpenOptions) (DocumentHandle, error) {
	s.usedPath = true
	s.lastPath = path
	s.lastOptions = opts
	if s.err != nil {
		return nil, s.err
	}
	return s.handle, nil
}

func (s *stubLoader) OpenReader(_ context.Context, name string, r io.Reader, opts OpenOptions) (DocumentHandle, error) {
	s.usedReader = true
	s.lastName = name
	s.lastOptions = opts
	if r == nil {
		return nil, errors.New("nil reader")
	}
	if s.err != nil {
		return nil, s.err
	}
	return s.handle, nil
}

type stubDocumentHandle struct {
	metadata model.DocumentMetadata
	pages    []PageHandle
}

func (s *stubDocumentHandle) Metadata() model.DocumentMetadata {
	return s.metadata
}

func (s *stubDocumentHandle) PageCount() int {
	return len(s.pages)
}

func (s *stubDocumentHandle) Page(pageIndex int) (PageHandle, error) {
	if pageIndex < 0 || pageIndex >= len(s.pages) {
		return nil, errors.New("page out of range")
	}
	return s.pages[pageIndex], nil
}

func (s *stubDocumentHandle) Close() error {
	return nil
}

type stubPageHandle struct {
	metadata  model.PageMetadata
	artifacts []*model.RawArtifact
}

func (s *stubPageHandle) Metadata() model.PageMetadata {
	return s.metadata
}

func (s *stubPageHandle) Artifacts(_ context.Context, _ ArtifactOptions) ([]*model.RawArtifact, error) {
	return s.artifacts, nil
}

func (s *stubPageHandle) TableCandidates(_ context.Context) (*TableCandidateSet, error) {
	return nil, nil
}

func (s *stubPageHandle) StructTree(_ context.Context) (*StructNode, error) {
	return nil, nil
}
