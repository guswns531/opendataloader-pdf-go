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

func TestResolveUsesPDFRuntimeForPDF(t *testing.T) {
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
	if got.Name() != "pdf-runtime" {
		t.Fatalf("Resolve() selected %q, want pdf-runtime", got.Name())
	}
}

func TestResolveWrapsNativePDFAndBridgeInRuntime(t *testing.T) {
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
	if got.Name() != "pdf-runtime" {
		t.Fatalf("Resolve() selected %q, want pdf-runtime", got.Name())
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

func TestNewCanOptIntoNativeSkeletonPDFBackend(t *testing.T) {
	t.Setenv("OPENDATALOADER_GO_PDF_BACKEND", "native-skeleton")

	ingestor, err := New().Resolve("sample.pdf", false)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	document, err := ingestor.Ingest(nil, core.Source{
		Name:   "sample.pdf",
		Reader: strings.NewReader("%PDF-1.7\n1 0 obj << /Type /Page /MediaBox [0 0 612 792] >> endobj\n(Env Native Skeleton)\n"),
	})
	if err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	if len(document.Pages) != 1 {
		t.Fatalf("len(document.Pages) = %d, want 1", len(document.Pages))
	}
	if len(document.Pages[0].Artifacts) != 1 {
		t.Fatalf("len(document.Pages[0].Artifacts) = %d, want 1", len(document.Pages[0].Artifacts))
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
