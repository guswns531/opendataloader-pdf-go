package nativepdf

import "github.com/guswns531/opendataloader-pdf-go/internal/core"

// UnavailableIngestor is the default placeholder for real PDF parsing until a
// native backend is implemented behind the nativepdf seam.
type UnavailableIngestor struct{}

// NewUnavailableIngestor returns the current native PDF placeholder.
func NewUnavailableIngestor() *UnavailableIngestor {
	return &UnavailableIngestor{}
}

// Name identifies the placeholder backend.
func (i *UnavailableIngestor) Name() string {
	return "nativepdf-unavailable"
}

// Ingest reports that the real native backend is not implemented yet.
func (i *UnavailableIngestor) Ingest(_ *core.ProcessingContext, _ core.Source) (*core.Document, error) {
	return nil, ErrBackendUnavailable
}
