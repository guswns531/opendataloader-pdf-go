package pdfruntime

import (
	"errors"
	"fmt"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/nativepdf"
)

// Ingestor prefers a native PDF backend and falls back to the temporary bridge
// only when the native backend explicitly reports unavailability.
type Ingestor struct {
	Primary  core.Ingestor
	Fallback core.Ingestor
}

// New returns a PDF runtime ingestor with explicit primary and fallback paths.
func New(primary, fallback core.Ingestor) *Ingestor {
	return &Ingestor{
		Primary:  primary,
		Fallback: fallback,
	}
}

// Name identifies the orchestrating PDF runtime path.
func (i *Ingestor) Name() string {
	return "pdf-runtime"
}

// Ingest tries the native backend first and falls back only on the expected
// unimplemented error.
func (i *Ingestor) Ingest(ctx *core.ProcessingContext, source core.Source) (*core.Document, error) {
	if i == nil {
		return nil, fmt.Errorf("pdf runtime ingestor is nil")
	}
	if i.Primary != nil {
		document, err := i.Primary.Ingest(ctx, source)
		if err == nil {
			return document, nil
		}
		if !errors.Is(err, nativepdf.ErrBackendUnavailable) {
			return nil, err
		}
	}
	if i.Fallback == nil {
		return nil, fmt.Errorf("pdf runtime fallback ingestor is not configured")
	}
	return i.Fallback.Ingest(ctx, source)
}
