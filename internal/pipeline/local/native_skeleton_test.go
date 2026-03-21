package local

import (
	"bytes"
	"compress/zlib"
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

const inlineSkeletonPDF = `%PDF-1.7
1 0 obj << /Type /Page /MediaBox [0 0 612 792] >> endobj
(Hello Native Pipeline)
`
