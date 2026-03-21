package local

import (
	"bytes"
	"compress/zlib"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/nativepdf"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestPipelineBuildsParagraphFromNativeSkeletonPDF(t *testing.T) {
	pipeline := New(nativepdf.NewIngestor(nativepdf.NewSkeletonLoader()))
	ctx := core.NewProcessingContext(nil, core.ProcessingOptions{})

	document, err := pipeline.Run(ctx, core.Source{
		Name:   "inline.pdf",
		Reader: strings.NewReader(inlineSkeletonPDF),
	}, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(document.Pages) != 1 {
		t.Fatalf("len(document.Pages) = %d, want 1", len(document.Pages))
	}
	if len(document.Kids) != 1 {
		t.Fatalf("len(document.Kids) = %d, want 1", len(document.Kids))
	}
	para, ok := document.Kids[0].(*model.Paragraph)
	if !ok {
		t.Fatalf("document.Kids[0] type = %T, want *model.Paragraph", document.Kids[0])
	}
	if !strings.Contains(para.Content, "Hello Native Pipeline") {
		t.Fatalf("paragraph content = %q", para.Content)
	}
}

func TestPipelineBuildsParagraphFromCompressedNativeSkeletonPDF(t *testing.T) {
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write([]byte("BT (Compressed Native Pipeline) Tj ET")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	pdf := "%PDF-1.7\n" +
		"1 0 obj << /Type /Page /MediaBox [0 0 612 792] >> endobj\n" +
		"2 0 obj << /Filter /FlateDecode >>stream\n" +
		compressed.String() + "\nendstream\n"

	pipeline := New(nativepdf.NewIngestor(nativepdf.NewSkeletonLoader()))
	ctx := core.NewProcessingContext(nil, core.ProcessingOptions{})

	document, err := pipeline.Run(ctx, core.Source{
		Name:   "compressed.pdf",
		Reader: strings.NewReader(pdf),
	}, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(document.Kids) != 1 {
		t.Fatalf("len(document.Kids) = %d, want 1", len(document.Kids))
	}
	para, ok := document.Kids[0].(*model.Paragraph)
	if !ok {
		t.Fatalf("document.Kids[0] type = %T, want *model.Paragraph", document.Kids[0])
	}
	if !strings.Contains(para.Content, "Compressed Native Pipeline") {
		t.Fatalf("paragraph content = %q", para.Content)
	}
}

func TestPipelineBuildsParagraphFromSampleLoremPDFWithNativeSkeleton(t *testing.T) {
	pipeline := New(nativepdf.NewIngestor(nativepdf.NewSkeletonLoader()))
	ctx := core.NewProcessingContext(nil, core.ProcessingOptions{})

	document, err := pipeline.Run(ctx, core.Source{
		Path: filepath.Clean("../../../samples/pdf/lorem.pdf"),
		Name: "lorem.pdf",
	}, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(document.Pages) != 1 {
		t.Fatalf("len(document.Pages) = %d, want 1", len(document.Pages))
	}
	if len(document.Kids) == 0 {
		t.Fatal("len(document.Kids) = 0, want non-zero")
	}
	para, ok := document.Kids[0].(*model.Paragraph)
	if !ok {
		t.Fatalf("document.Kids[0] type = %T, want *model.Paragraph", document.Kids[0])
	}
	if !strings.Contains(para.Content, "Lorem") {
		t.Fatalf("paragraph content = %q, want Lorem...", para.Content)
	}
}

func TestPipelineOrdersTwoColumnNativeRawFixtureByColumns(t *testing.T) {
	pipeline := New(nativepdf.NewIngestor(nativepdf.NewFixtureLoader()))
	ctx := core.NewProcessingContext(nil, core.ProcessingOptions{})

	document, err := pipeline.Run(ctx, core.Source{
		Path: filepath.Clean("../../../testdata/fixtures/native/two_column_reading_order.raw.json"),
		Name: "two_column_reading_order.raw.json",
	}, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(document.Kids) != 4 {
		t.Fatalf("len(document.Kids) = %d, want 4", len(document.Kids))
	}

	got := make([]string, 0, len(document.Kids))
	for _, element := range document.Kids {
		para, ok := element.(*model.Paragraph)
		if !ok {
			t.Fatalf("element type = %T, want *model.Paragraph", element)
		}
		got = append(got, para.Content)
	}
	want := []string{
		"Left column first line.",
		"Left column second line.",
		"Right column first line.",
		"Right column second line.",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("reading order = %v, want %v", got, want)
	}

	sortedGot := append([]string(nil), got...)
	sort.Strings(sortedGot)
	sort.Strings(want)
	if strings.Join(sortedGot, "\n") != strings.Join(want, "\n") {
		t.Fatalf("paragraph contents = %v, want %v", got, want)
	}
}

const inlineSkeletonPDF = `%PDF-1.7
1 0 obj << /Type /Page /MediaBox [0 0 612 792] >> endobj
(Hello Native Pipeline)
`
