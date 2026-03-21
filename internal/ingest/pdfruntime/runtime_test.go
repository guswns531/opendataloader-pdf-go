package pdfruntime

import (
	"errors"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/nativepdf"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestIngestFallsBackWhenNativeBackendUnavailable(t *testing.T) {
	want := model.NewDocument(model.DocumentMetadata{FileName: "fallback.pdf"})
	runtime := New(
		stubIngestor{name: "native", err: nativepdf.ErrBackendUnavailable},
		stubIngestor{name: "bridge", document: want},
	)

	got, err := runtime.Ingest(nil, core.Source{Name: "sample.pdf"})
	if err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	if got != want {
		t.Fatalf("Ingest() returned unexpected fallback document")
	}
}

func TestIngestUsesPrimaryWhenAvailable(t *testing.T) {
	want := model.NewDocument(model.DocumentMetadata{FileName: "native.pdf"})
	runtime := New(
		stubIngestor{name: "native", document: want},
		stubIngestor{name: "bridge"},
	)

	got, err := runtime.Ingest(nil, core.Source{Name: "sample.pdf"})
	if err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	if got != want {
		t.Fatalf("Ingest() returned unexpected primary document")
	}
}

func TestIngestPropagatesUnexpectedPrimaryError(t *testing.T) {
	runtime := New(
		stubIngestor{name: "native", err: errors.New("boom")},
		stubIngestor{name: "bridge"},
	)

	_, err := runtime.Ingest(nil, core.Source{Name: "sample.pdf"})
	if err == nil {
		t.Fatal("Ingest() error = nil, want error")
	}
	if err.Error() != "boom" {
		t.Fatalf("Ingest() error = %v, want boom", err)
	}
}

func TestIngestErrorsWithoutFallback(t *testing.T) {
	runtime := New(
		stubIngestor{name: "native", err: nativepdf.ErrBackendUnavailable},
		nil,
	)

	_, err := runtime.Ingest(nil, core.Source{Name: "sample.pdf"})
	if err == nil {
		t.Fatal("Ingest() error = nil, want error")
	}
}

type stubIngestor struct {
	name     string
	document *core.Document
	err      error
}

func (s stubIngestor) Name() string {
	return s.name
}

func (s stubIngestor) Ingest(_ *core.ProcessingContext, _ core.Source) (*core.Document, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.document, nil
}
