package nativepdf

import (
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestExtractTextFragmentsFromStreamTracksFontAndLineMovement(t *testing.T) {
	stream := []byte("BT /F1 20 Tf 100 700 Td (Hello) Tj 0 -24 Td (World) Tj ET")

	fragments := extractTextFragmentsFromStream(stream, nil)
	if got, want := len(fragments), 2; got != want {
		t.Fatalf("len(fragments) = %d, want %d", got, want)
	}
	if fragments[0].fontSize != 20 {
		t.Fatalf("fragments[0].fontSize = %v, want 20", fragments[0].fontSize)
	}
	if got, want := fragments[0].bounds.Left, 100.0; got != want {
		t.Fatalf("fragments[0].bounds.Left = %v, want %v", got, want)
	}
	if got, want := fragments[0].bounds.Bottom, 700.0; got != want {
		t.Fatalf("fragments[0].bounds.Bottom = %v, want %v", got, want)
	}
	if got, want := fragments[1].bounds.Bottom, 676.0; got != want {
		t.Fatalf("fragments[1].bounds.Bottom = %v, want %v", got, want)
	}
	if fragments[1].fontSize != 20 {
		t.Fatalf("fragments[1].fontSize = %v, want 20", fragments[1].fontSize)
	}
}

func TestExtractTextFragmentsFromStreamAppliesGraphicsMatrixToTmCoordinates(t *testing.T) {
	stream := []byte("q 2 0 0 2 10 20 cm BT /F1 10 Tf 0 0 Tm (A) Tj ET Q")

	fragments := extractTextFragmentsFromStream(stream, nil)
	if got, want := len(fragments), 1; got != want {
		t.Fatalf("len(fragments) = %d, want %d", got, want)
	}
	if got, want := fragments[0].bounds.Left, 10.0; got != want {
		t.Fatalf("fragments[0].bounds.Left = %v, want %v", got, want)
	}
	if got, want := fragments[0].bounds.Bottom, 20.0; got != want {
		t.Fatalf("fragments[0].bounds.Bottom = %v, want %v", got, want)
	}
	if got := fragments[0].bounds.Right; got <= 20.0 || got >= 26.0 {
		t.Fatalf("fragments[0].bounds.Right = %v, want within (20, 26)", got)
	}
	if got, want := fragments[0].bounds.Top, 40.0; got != want {
		t.Fatalf("fragments[0].bounds.Top = %v, want %v", got, want)
	}
}

func TestExtractTextFragmentsFromStreamTracksMarkedContentID(t *testing.T) {
	stream := []byte("BT /F1 12 Tf /P << /MCID 7 >> BDC 10 20 Td (Hello) Tj EMC /Artifact BMC 0 -14 Td (World) Tj EMC ET")

	fragments := extractTextFragmentsFromStream(stream, nil)
	if got, want := len(fragments), 2; got != want {
		t.Fatalf("len(fragments) = %d, want %d", got, want)
	}
	if fragments[0].markedContentID == nil || *fragments[0].markedContentID != 7 {
		t.Fatalf("fragments[0].markedContentID = %#v, want 7", fragments[0].markedContentID)
	}
	if fragments[1].markedContentID != nil {
		t.Fatalf("fragments[1].markedContentID = %#v, want nil", fragments[1].markedContentID)
	}
}

func TestSkeletonLoaderUsesPositionedTextArtifactsForSimpleDigitalPDF(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Page /MediaBox [0 0 612 792] >> endobj
2 0 obj << /Length 56 >> stream
BT /F1 20 Tf 100 700 Td (Hello) Tj 0 -24 Td (World) Tj ET
endstream
`

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "sample.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	if got, want := handle.PageCount(), 1; got != want {
		t.Fatalf("handle.PageCount() = %d, want %d", got, want)
	}
	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}
	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeText: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 2; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}
	if got, want := artifacts[0].Text, "Hello"; got != want {
		t.Fatalf("artifacts[0].Text = %q, want %q", got, want)
	}
	if got, want := artifacts[0].Style.FontSize, 20.0; got != want {
		t.Fatalf("artifacts[0].Style.FontSize = %v, want %v", got, want)
	}
	if got, want := artifacts[0].Bounds.Left, 100.0; got != want {
		t.Fatalf("artifacts[0].Bounds.Left = %v, want %v", got, want)
	}
	if got, want := artifacts[1].Bounds.Bottom, 676.0; got != want {
		t.Fatalf("artifacts[1].Bounds.Bottom = %v, want %v", got, want)
	}
}

func TestNativePDFIngestorKeepsPositionedArtifactsVisibleInDocument(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Page /MediaBox [0 0 612 792] >> endobj
2 0 obj << /Length 56 >> stream
BT /F1 20 Tf 100 700 Td (Hello) Tj 0 -24 Td (World) Tj ET
endstream
`

	ingestor := NewIngestor(NewSkeletonLoader())
	document, err := ingestor.Ingest(nil, core.Source{
		Name:   "sample.pdf",
		Reader: strings.NewReader(pdf),
	})
	if err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}
	if got, want := len(document.Pages), 1; got != want {
		t.Fatalf("len(document.Pages) = %d, want %d", got, want)
	}
	if got, want := len(document.Pages[0].Artifacts), 2; got != want {
		t.Fatalf("len(document.Pages[0].Artifacts) = %d, want %d", got, want)
	}
	if document.Pages[0].Artifacts[0].Kind != model.ArtifactKindText {
		t.Fatalf("artifact kind = %q, want text", document.Pages[0].Artifacts[0].Kind)
	}
}

func TestSkeletonLoaderResolvesBaseFontFromPageResources(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 300 300] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >> endobj
4 0 obj << /Type /Font /Subtype /Type1 /BaseFont /MyResolvedFont >> endobj
5 0 obj << /Length 39 >> stream
BT /F1 18 Tf 10 20 Td (Hello) Tj ET
endstream
endobj
`

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "sample.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}
	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeText: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 1; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}
	if got, want := artifacts[0].Style.Font, "MyResolvedFont"; got != want {
		t.Fatalf("artifacts[0].Style.Font = %q, want %q", got, want)
	}
	if got, want := artifacts[0].Bounds.Left, 10.0; got != want {
		t.Fatalf("artifacts[0].Bounds.Left = %v, want %v", got, want)
	}
}

func TestEstimateTextWidthUsesGlyphCategories(t *testing.T) {
	wide := estimateTextWidth("WWW", 10, "Helvetica")
	narrow := estimateTextWidth("iii", 10, "Helvetica")
	if wide <= narrow {
		t.Fatalf("wide width = %v, narrow width = %v, want wide > narrow", wide, narrow)
	}

	courierWide := estimateTextWidth("WWW", 10, "Courier")
	courierNarrow := estimateTextWidth("iii", 10, "Courier")
	if diff := courierWide - courierNarrow; diff < -0.01 || diff > 0.01 {
		t.Fatalf("courier widths differ too much: wide=%v narrow=%v", courierWide, courierNarrow)
	}
}

func TestEstimateTokenAdvanceAccountsForTJKerning(t *testing.T) {
	token := contentToken{
		kind: contentTokenArray,
		items: []contentToken{
			{kind: contentTokenLiteralString, text: "AB"},
			{kind: contentTokenNumber, number: -500},
			{kind: contentTokenLiteralString, text: "CD"},
		},
	}

	got := estimateTokenAdvance(token, 10, "Helvetica", nil, fontMetrics{})
	plain := estimateTextWidth("ABCD", 10, "Helvetica")
	if got <= plain {
		t.Fatalf("advance with TJ kerning = %v, plain width = %v, want kerning to widen advance", got, plain)
	}
}

func TestEstimateTokenAdvanceUsesSimpleFontWidthTable(t *testing.T) {
	metrics := fontMetrics{
		widths: map[int]float64{
			int('W'): 900,
			int('i'): 250,
		},
		codeBytes: 1,
	}

	wide := estimateTokenAdvance(contentToken{kind: contentTokenLiteralString, text: "WWW"}, 10, "Helvetica", nil, metrics)
	narrow := estimateTokenAdvance(contentToken{kind: contentTokenLiteralString, text: "iii"}, 10, "Helvetica", nil, metrics)
	if wide <= narrow {
		t.Fatalf("width-table advance = %v, narrow advance = %v, want width table to distinguish glyph widths", wide, narrow)
	}
	if got, want := wide, 27.0; got != want {
		t.Fatalf("wide advance = %v, want %v", got, want)
	}
}

func TestEstimateTokenAdvanceUsesStandardBase14FallbackWidths(t *testing.T) {
	metrics, ok := fontMetricsFromDict(nil, pdfDict{
		"Subtype":  pdfName("Type1"),
		"BaseFont": pdfName("Helvetica"),
	})
	if !ok {
		t.Fatalf("fontMetricsFromDict() reported no metrics")
	}

	wide := estimateTokenAdvance(contentToken{kind: contentTokenLiteralString, text: "WWW"}, 10, "Helvetica", nil, metrics)
	narrow := estimateTokenAdvance(contentToken{kind: contentTokenLiteralString, text: "iii"}, 10, "Helvetica", nil, metrics)
	if wide <= narrow {
		t.Fatalf("base14 fallback advance = %v, narrow advance = %v, want fallback widths to distinguish glyph widths", wide, narrow)
	}
	if got, want := wide, 28.32; got != want {
		t.Fatalf("wide advance = %v, want %v", got, want)
	}
}

func TestEstimateTokenAdvanceUsesCIDWidthTableForHexText(t *testing.T) {
	metrics := fontMetrics{
		widths: map[int]float64{
			1: 900,
			2: 300,
		},
		defaultWidth: 1000,
		codeBytes:    2,
	}

	wide := estimateTokenAdvance(contentToken{kind: contentTokenHexString, text: "00010001"}, 10, "Type0Font", nil, metrics)
	narrow := estimateTokenAdvance(contentToken{kind: contentTokenHexString, text: "00020002"}, 10, "Type0Font", nil, metrics)
	if wide <= narrow {
		t.Fatalf("cid width-table advance = %v, narrow advance = %v, want width table to distinguish CIDs", wide, narrow)
	}
	if got, want := narrow, 6.0; got != want {
		t.Fatalf("narrow advance = %v, want %v", got, want)
	}
}

func TestSkeletonLoaderUsesFontWidthsForPositionedBounds(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 300 300] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >> endobj
4 0 obj << /Type /Font /Subtype /Type1 /BaseFont /MetricFont /FirstChar 87 /LastChar 88 /Widths [900 250] >> endobj
5 0 obj << /Length 39 >> stream
BT /F1 10 Tf 10 20 Td (WX) Tj ET
endstream
endobj
`

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "sample.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}
	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeText: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 1; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}
	if got, want := artifacts[0].Bounds.Right, 21.5; got != want {
		t.Fatalf("artifacts[0].Bounds.Right = %v, want %v", got, want)
	}
}
