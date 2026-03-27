package processors

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/markdown"
)

func TestParsePageRangePreservesOrderAndDuplicates(t *testing.T) {
	got, explicit, err := parsePageRange("3,1,5-7,6", 10)
	if err != nil {
		t.Fatalf("parsePageRange returned error: %v", err)
	}
	if !explicit {
		t.Fatal("expected explicit selection")
	}
	want := []int{3, 1, 5, 6, 7, 6}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestParsePageRangeFiltersOutOfBoundsPages(t *testing.T) {
	got, explicit, err := parsePageRange("1,11", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !explicit {
		t.Fatal("expected explicit selection")
	}
	want := []int{1}
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestParsePageRangeFiltersOutOfBoundsRanges(t *testing.T) {
	got, explicit, err := parsePageRange("2-12", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !explicit {
		t.Fatal("expected explicit selection")
	}
	want := []int{2, 3, 4, 5, 6, 7, 8, 9, 10}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestResolveImageDirUsesConfiguredDirectory(t *testing.T) {
	cfg := api.DefaultConfig()
	cfg.ImageDir = "/tmp/images"
	cfg.ImageOutput = api.ImageOutputExternal

	got := resolveImageDir("/tmp/sample.pdf", cfg)
	if got != "/tmp/images" {
		t.Fatalf("expected configured image dir, got %q", got)
	}
}

func TestResolveImageDirDefaultsToOutputFolder(t *testing.T) {
	cfg := api.DefaultConfig()
	cfg.OutputDir = "/tmp/out"
	cfg.ImageOutput = api.ImageOutputExternal

	got := resolveImageDir("/docs/sample.pdf", cfg)
	want := filepath.Join("/tmp/out", "sample_images")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestProcessJavaDocumentPreservesBibliographyBoundarySpaceInFixture(t *testing.T) {
	pdf := filepath.Clean("../../../samples/pdf/1901.03003.pdf")
	if _, err := os.Stat(pdf); err != nil {
		t.Skip("fixture not available")
	}
	cfg := api.DefaultConfig()
	cfg.Pages = "15"
	cfg.ImageOutput = api.ImageOutputOff

	ctx := containers.NewProcessorContext()
	processor := NewDocumentProcessor()
	doc, err := processor.loadDocument(pdf, cfg, ctx)
	if err != nil {
		t.Fatalf("loadDocument() error = %v", err)
	}

	processed, err := processor.processJavaDocument(doc, cfg, ctx)
	if err != nil {
		t.Fatalf("processJavaDocument() error = %v", err)
	}

	output, err := markdown.NewMarkdownGenerator(cfg, false, false).Generate(processed)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(output, "Word spotting in the wild. In Proceedings of European Conference on Computer") {
		t.Fatalf("expected bibliography output to preserve boundary space, got %q", output)
	}
	if strings.Contains(output, "Word spotting in the wild. InProceedings") {
		t.Fatalf("unexpected joined bibliography boundary in output: %q", output)
	}
}

func TestProcessJavaDocumentSuppressesIntrawordSyntheticSpacesInFixture(t *testing.T) {
	pdf := filepath.Clean("../../../samples/pdf/1901.03003.pdf")
	if _, err := os.Stat(pdf); err != nil {
		t.Skip("fixture not available")
	}
	cfg := api.DefaultConfig()
	cfg.Pages = "1"
	cfg.ImageOutput = api.ImageOutputOff

	ctx := containers.NewProcessorContext()
	processor := NewDocumentProcessor()
	doc, err := processor.loadDocument(pdf, cfg, ctx)
	if err != nil {
		t.Fatalf("loadDocument() error = %v", err)
	}

	processed, err := processor.processJavaDocument(doc, cfg, ctx)
	if err != nil {
		t.Fatalf("processJavaDocument() error = %v", err)
	}

	output, err := markdown.NewMarkdownGenerator(cfg, false, false).Generate(processed)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(output, "we thus propose a multi-object rectiﬁed attention network") {
		t.Fatalf("expected fixture output to preserve boundary space before rectiﬁed, got %q", output)
	}
	if strings.Contains(output, "multi-objectrectiﬁed") {
		t.Fatalf("unexpected joined rectiﬁed boundary in output: %q", output)
	}
	if strings.Contains(output, "m ulti-") || strings.Contains(output, "multi- object") || strings.Contains(output, "objectr ectiﬁed") || strings.Contains(output, "a ttention") {
		t.Fatalf("unexpected intra-word synthetic spacing in output: %q", output)
	}
}
