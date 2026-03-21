package selector

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/fixture"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/nativepdf"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/pdftext"
)

// Selector isolates runtime ingestion routing so the CLI and future wrappers
// do not hard-code the current temporary backend choices.
type Selector struct {
	SemanticFixture core.Ingestor
	NativeFixture   core.Ingestor
	PDF             core.Ingestor
}

// New returns a selector with the current default runtime paths.
func New() Selector {
	return Selector{
		SemanticFixture: fixture.New(),
		NativeFixture:   nativepdf.NewIngestor(nativepdf.NewFixtureLoader()),
		PDF:             TemporaryPDFIngestor(),
	}
}

// Resolve chooses the current ingestor for an input path.
func (s Selector) Resolve(path string, useFixture bool) (core.Ingestor, error) {
	path = strings.TrimSpace(path)
	switch {
	case strings.HasSuffix(strings.ToLower(path), ".raw.json"):
		if s.NativeFixture == nil {
			return nil, fmt.Errorf("native raw fixture ingestor is not configured")
		}
		return s.NativeFixture, nil
	case useFixture, strings.EqualFold(filepath.Ext(path), ".json"):
		if s.SemanticFixture == nil {
			return nil, fmt.Errorf("semantic fixture ingestor is not configured")
		}
		return s.SemanticFixture, nil
	case strings.EqualFold(filepath.Ext(path), ".pdf"):
		if s.PDF == nil {
			return nil, fmt.Errorf("pdf ingestor is not configured")
		}
		return s.PDF, nil
	default:
		return nil, fmt.Errorf("unsupported input type for %q", path)
	}
}

// TemporaryPDFIngestor returns the current bridge implementation for `.pdf`
// inputs. This must be replaced by a real native parser backend.
func TemporaryPDFIngestor() core.Ingestor {
	return pdftext.New()
}
