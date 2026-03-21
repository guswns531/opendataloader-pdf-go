package pdfbridge

import (
	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/pdftext"
)

// Ingestor is the temporary bridge for real `.pdf` inputs until the native
// parser backend is ready. It deliberately wraps the current pdftext-based path
// so selector logic can stop importing pdftext directly.
type Ingestor struct {
	delegate core.Ingestor
}

// New returns the current temporary PDF bridge.
func New() *Ingestor {
	return &Ingestor{delegate: pdftext.New()}
}

// Name identifies the bridge rather than the underlying implementation.
func (i *Ingestor) Name() string {
	return "pdf-bridge"
}

// Ingest delegates to the current bridge backend.
func (i *Ingestor) Ingest(ctx *core.ProcessingContext, source core.Source) (*core.Document, error) {
	return i.delegate.Ingest(ctx, source)
}
