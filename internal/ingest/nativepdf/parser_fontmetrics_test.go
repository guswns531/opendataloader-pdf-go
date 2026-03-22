package nativepdf

import (
	"strings"
	"testing"
)

func TestFontMetricsFromDictResolvesIndirectSimpleWidths(t *testing.T) {
	graph := &objectGraph{
		objects: map[pdfRef]*indirectObject{
			{ObjectNumber: 2, Generation: 0}: {
				Ref:   pdfRef{ObjectNumber: 2, Generation: 0},
				Value: pdfArray{900.0, 250.0},
			},
		},
	}

	metrics, ok := fontMetricsFromDict(graph, pdfDict{
		"Subtype":   pdfName("Type1"),
		"FirstChar": 87,
		"Widths":    pdfRef{ObjectNumber: 2, Generation: 0},
	})
	if !ok {
		t.Fatalf("fontMetricsFromDict() reported no metrics")
	}
	if got, want := metrics.widths[87], 900.0; got != want {
		t.Fatalf("metrics.widths[87] = %v, want %v", got, want)
	}
	if got, want := metrics.widths[88], 250.0; got != want {
		t.Fatalf("metrics.widths[88] = %v, want %v", got, want)
	}
}

func TestFontMetricsFromDictUsesDescriptorFallbackWidths(t *testing.T) {
	tests := []struct {
		name string
		dict pdfDict
		want float64
	}{
		{
			name: "missing width",
			dict: pdfDict{
				"FontDescriptor": pdfDict{
					"MissingWidth": 480.0,
				},
			},
			want: 480,
		},
		{
			name: "average width",
			dict: pdfDict{
				"FontDescriptor": pdfDict{
					"AvgWidth": 520.0,
				},
			},
			want: 520,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics, ok := fontMetricsFromDict(nil, tt.dict)
			if !ok {
				t.Fatalf("fontMetricsFromDict() reported no metrics")
			}
			if got := metrics.defaultWidth; got != tt.want {
				t.Fatalf("metrics.defaultWidth = %v, want %v", got, tt.want)
			}
			if got, want := metrics.advanceForLiteralString("ABC", 10), (tt.want*3*10)/1000.0; got != want {
				t.Fatalf("advanceForLiteralString() = %v, want %v", got, want)
			}
		})
	}
}

func TestFontMetricsFromDictUsesStandardBase14Widths(t *testing.T) {
	metrics, ok := fontMetricsFromDict(nil, pdfDict{
		"Subtype":  pdfName("Type1"),
		"BaseFont": pdfName("ABCDEE+Helvetica"),
	})
	if !ok {
		t.Fatalf("fontMetricsFromDict() reported no metrics")
	}
	if got, want := metrics.widths[int('W')], 944.0; got != want {
		t.Fatalf("metrics.widths['W'] = %v, want %v", got, want)
	}
	if got, want := metrics.widths[int('i')], 222.0; got != want {
		t.Fatalf("metrics.widths['i'] = %v, want %v", got, want)
	}
	if got, want := metrics.advanceForLiteralString("Wi", 10), 11.66; got != want {
		t.Fatalf("advanceForLiteralString() = %v, want %v", got, want)
	}
}

func TestFontMetricsFromDictResolvesIndirectCIDWidths(t *testing.T) {
	graph := &objectGraph{
		objects: map[pdfRef]*indirectObject{
			{ObjectNumber: 2, Generation: 0}: {
				Ref:   pdfRef{ObjectNumber: 2, Generation: 0},
				Value: pdfArray{1, pdfArray{900.0, 300.0}},
			},
		},
	}

	metrics, ok := fontMetricsFromDict(graph, pdfDict{
		"Subtype": pdfName("Type0"),
		"DescendantFonts": pdfArray{
			pdfDict{
				"W": pdfRef{ObjectNumber: 2, Generation: 0},
			},
		},
	})
	if !ok {
		t.Fatalf("fontMetricsFromDict() reported no metrics")
	}
	if got, want := metrics.widths[1], 900.0; got != want {
		t.Fatalf("metrics.widths[1] = %v, want %v", got, want)
	}
	if got, want := metrics.widths[2], 300.0; got != want {
		t.Fatalf("metrics.widths[2] = %v, want %v", got, want)
	}
}

func TestSkeletonLoaderUsesDescriptorFallbackWidthForPositionedBounds(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 300 300] /Resources << /Font << /F1 4 0 R >> >> /Contents 6 0 R >> endobj
4 0 obj << /Type /Font /Subtype /Type1 /BaseFont /FallbackFont /FontDescriptor 5 0 R >> endobj
5 0 obj << /Type /FontDescriptor /FontName /FallbackFont /AvgWidth 500 >> endobj
6 0 obj << /Length 40 >> stream
BT /F1 10 Tf 10 20 Td (AB) Tj ET
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
	if got, want := artifacts[0].Bounds.Right, 20.0; got != want {
		t.Fatalf("artifacts[0].Bounds.Right = %v, want %v", got, want)
	}
}

func TestSkeletonLoaderUsesStandardBase14WidthsForPositionedBounds(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 300 300] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >> endobj
4 0 obj << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> endobj
5 0 obj << /Length 39 >> stream
BT /F1 10 Tf 10 20 Td (Wi) Tj ET
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
	if got, want := artifacts[0].Bounds.Right, 21.66; got != want {
		t.Fatalf("artifacts[0].Bounds.Right = %v, want %v", got, want)
	}
}
