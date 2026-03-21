package selector

import (
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
)

func TestResolvePrefersNativeRawFixtures(t *testing.T) {
	selector := Selector{
		SemanticFixture: stubIngestor{name: "semantic"},
		NativeFixture:   stubIngestor{name: "native"},
		PDFBridge:       stubIngestor{name: "pdf-bridge"},
	}

	got, err := selector.Resolve("sample.raw.json", false)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Name() != "native" {
		t.Fatalf("Resolve() selected %q, want native", got.Name())
	}
}

func TestResolveUsesSemanticFixturesForJSON(t *testing.T) {
	selector := Selector{
		SemanticFixture: stubIngestor{name: "semantic"},
		NativeFixture:   stubIngestor{name: "native"},
		PDFBridge:       stubIngestor{name: "pdf-bridge"},
	}

	got, err := selector.Resolve("sample.json", false)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Name() != "semantic" {
		t.Fatalf("Resolve() selected %q, want semantic", got.Name())
	}
}

func TestResolveUsesTemporaryPDFBridgeForPDF(t *testing.T) {
	selector := Selector{
		SemanticFixture: stubIngestor{name: "semantic"},
		NativeFixture:   stubIngestor{name: "native"},
		PDFBridge:       stubIngestor{name: "pdf-bridge"},
	}

	got, err := selector.Resolve("sample.pdf", false)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Name() != "pdf-bridge" {
		t.Fatalf("Resolve() selected %q, want pdf-bridge", got.Name())
	}
}

func TestResolvePrefersNativePDFWhenConfigured(t *testing.T) {
	selector := Selector{
		SemanticFixture: stubIngestor{name: "semantic"},
		NativeFixture:   stubIngestor{name: "native"},
		NativePDF:       stubIngestor{name: "native-pdf"},
		PDFBridge:       stubIngestor{name: "pdf-bridge"},
	}

	got, err := selector.Resolve("sample.pdf", false)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Name() != "native-pdf" {
		t.Fatalf("Resolve() selected %q, want native-pdf", got.Name())
	}
}

func TestResolveHonorsExplicitFixtureFlag(t *testing.T) {
	selector := Selector{
		SemanticFixture: stubIngestor{name: "semantic"},
		NativeFixture:   stubIngestor{name: "native"},
		PDFBridge:       stubIngestor{name: "pdf-bridge"},
	}

	got, err := selector.Resolve("fixture.anything", true)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Name() != "semantic" {
		t.Fatalf("Resolve() selected %q, want semantic", got.Name())
	}
}

func TestResolveErrorsWhenDependencyMissing(t *testing.T) {
	selector := Selector{
		SemanticFixture: nil,
	}

	_, err := selector.Resolve("sample.json", false)
	if err == nil {
		t.Fatal("Resolve() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "semantic fixture ingestor") {
		t.Fatalf("Resolve() error = %v, want semantic fixture message", err)
	}
}

type stubIngestor struct {
	name string
}

func (s stubIngestor) Name() string {
	return s.name
}

func (s stubIngestor) Ingest(_ *core.ProcessingContext, _ core.Source) (*core.Document, error) {
	return nil, nil
}
