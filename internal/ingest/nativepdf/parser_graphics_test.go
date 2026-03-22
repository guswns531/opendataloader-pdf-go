package nativepdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestGraphicsExtractionEmitsLineAndPathArtifacts(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Page /MediaBox [0 0 200 200] >> endobj
2 0 obj << /Length 88 >> stream
q
10 20 m 110 20 l S
15 30 m 15 120 l S
40 40 60 50 re S
endstream
`

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "graphics.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}

	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeLine: true, IncludePath: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 3; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}

	var lineCount, pathCount int
	for _, artifact := range artifacts {
		switch artifact.Kind {
		case model.ArtifactKindLine:
			lineCount++
		case model.ArtifactKindPath:
			pathCount++
		}
	}
	if lineCount != 2 {
		t.Fatalf("lineCount = %d, want 2", lineCount)
	}
	if pathCount != 1 {
		t.Fatalf("pathCount = %d, want 1", pathCount)
	}

	tableCandidates, err := page.TableCandidates(nil)
	if err != nil {
		t.Fatalf("page.TableCandidates() error = %v", err)
	}
	if got, want := len(tableCandidates.HorizontalLines), 1; got != want {
		t.Fatalf("len(horizontal lines) = %d, want %d", got, want)
	}
	if got, want := len(tableCandidates.VerticalLines), 1; got != want {
		t.Fatalf("len(vertical lines) = %d, want %d", got, want)
	}
	if got, want := len(tableCandidates.Rectangles), 1; got != want {
		t.Fatalf("len(rectangles) = %d, want %d", got, want)
	}
}

func TestGraphicsExtractionAppliesMatrixAndGraphicsStateStack(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Contents 4 0 R >> endobj
4 0 obj << /Length 76 >> stream
q
2 0 0 2 10 20 cm
0 0 m 10 0 l S
Q
0 0 m 0 10 l S
endstream
endobj
`

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "matrix.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}

	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeLine: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 2; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}
	if got, want := artifacts[0].Bounds.Left, 9.5; got != want {
		t.Fatalf("first line left = %v, want %v", got, want)
	}
	if got, want := artifacts[0].Bounds.Right, 30.5; got != want {
		t.Fatalf("first line right = %v, want %v", got, want)
	}
	if got, want := artifacts[1].Bounds.Left, -0.5; got != want {
		t.Fatalf("second line left = %v, want %v", got, want)
	}
	if got, want := artifacts[1].Bounds.Top, 10.5; got != want {
		t.Fatalf("second line top = %v, want %v", got, want)
	}
}

func TestGraphicsExtractionEmitsImageXObjectArtifact(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /XObject << /Im0 5 0 R >> >> /Contents 4 0 R >> endobj
4 0 obj << /Length 28 >> stream
q
20 0 0 10 30 40 cm
/Im0 Do
Q
endstream
endobj
5 0 obj << /Subtype /Image /Width 2 /Height 1 >> stream

endstream
endobj
`

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "image.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}

	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeImage: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 1; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}
	if artifacts[0].Kind != model.ArtifactKindImage {
		t.Fatalf("artifact kind = %q, want image", artifacts[0].Kind)
	}
	if got, want := artifacts[0].Bounds.Left, 30.0; got != want {
		t.Fatalf("image bounds left = %v, want %v", got, want)
	}
	if got, want := artifacts[0].Bounds.Right, 70.0; got != want {
		t.Fatalf("image bounds right = %v, want %v", got, want)
	}
	if got, want := artifacts[0].Bounds.Top, 50.0; got != want {
		t.Fatalf("image bounds top = %v, want %v", got, want)
	}
}

func TestGraphicsExtractionTracksMarkedContentID(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Page /MediaBox [0 0 200 200] >> endobj
2 0 obj << /Length 46 >> stream
/Figure << /MCID 9 >> BDC
10 20 m 30 20 l S
EMC
endstream
`

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "graphics-mcid.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}

	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeLine: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 1; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}
	if artifacts[0].MarkedContentID == nil || *artifacts[0].MarkedContentID != 9 {
		t.Fatalf("artifacts[0].MarkedContentID = %#v, want 9", artifacts[0].MarkedContentID)
	}
}

func TestGraphicsExtractionEmitsCurvePathArtifact(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Page /MediaBox [0 0 200 200] >> endobj
2 0 obj << /Length 80 >> stream
0 0 m
10 20 20 20 30 0 c
40 0 50 10 60 10 v
70 0 80 10 90 0 y
S
endstream
endobj
`

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "curve.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}

	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludePath: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 1; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}
	if artifacts[0].Kind != model.ArtifactKindPath {
		t.Fatalf("artifact kind = %q, want path", artifacts[0].Kind)
	}
	if got, want := artifacts[0].Bounds.Left, 0.0; got != want {
		t.Fatalf("path bounds left = %v, want %v", got, want)
	}
	if got, want := artifacts[0].Bounds.Right, 90.0; got != want {
		t.Fatalf("path bounds right = %v, want %v", got, want)
	}
	if got, want := artifacts[0].Bounds.Top, 20.0; got != want {
		t.Fatalf("path bounds top = %v, want %v", got, want)
	}
}

func TestGraphicsExtractionCapturesImagePayloadAndFormat(t *testing.T) {
	const pdf = "%PDF-1.7\n" +
		"1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n" +
		"2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj\n" +
		"3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /XObject << /Im0 5 0 R >> >> /Contents 4 0 R >> endobj\n" +
		"4 0 obj << /Length 28 >> stream\n" +
		"q\n" +
		"20 0 0 10 30 40 cm\n" +
		"/Im0 Do\n" +
		"Q\n" +
		"endstream\n" +
		"endobj\n" +
		"5 0 obj << /Subtype /Image /Filter /DCTDecode /ColorSpace /DeviceRGB /BitsPerComponent 8 /Width 2 /Height 1 >> stream\n" +
		"\xFF\xD8\xFF\xE0JPEGDATA\xFF\xD9\n" +
		"endstream\n" +
		"endobj\n"

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "image.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}

	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeImage: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 1; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}
	if artifacts[0].Kind != model.ArtifactKindImage {
		t.Fatalf("artifact kind = %q, want image", artifacts[0].Kind)
	}
	if artifacts[0].Format != model.ImageFormatJPEG {
		t.Fatalf("artifact format = %q, want jpeg", artifacts[0].Format)
	}
	if got, want := artifacts[0].ColorSpace, "DeviceRGB"; got != want {
		t.Fatalf("artifact color space = %q, want %q", got, want)
	}
	if got, want := artifacts[0].BitsPerComponent, 8; got != want {
		t.Fatalf("artifact bits per component = %d, want %d", got, want)
	}
	if got, want := len(artifacts[0].Filters), 1; got != want {
		t.Fatalf("artifact filters len = %d, want %d", got, want)
	}
	if got, want := artifacts[0].Filters[0], "DCTDecode"; got != want {
		t.Fatalf("artifact filter = %q, want %q", got, want)
	}
	if got, want := len(artifacts[0].Data) > 0, true; got != want {
		t.Fatalf("artifact data present = %v, want %v", got, want)
	}
	if got, want := artifacts[0].Bounds.Left, 30.0; got != want {
		t.Fatalf("image bounds left = %v, want %v", got, want)
	}
}

func TestGraphicsExtractionResolvesReferencedImageMetadata(t *testing.T) {
	const pdf = "%PDF-1.7\n" +
		"1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n" +
		"2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj\n" +
		"3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /XObject << /Im0 5 0 R >> >> /Contents 4 0 R >> endobj\n" +
		"4 0 obj << /Length 28 >> stream\n" +
		"q\n" +
		"20 0 0 10 30 40 cm\n" +
		"/Im0 Do\n" +
		"Q\n" +
		"endstream\n" +
		"endobj\n" +
		"5 0 obj << /Subtype /Image /Filter 6 0 R /ColorSpace 7 0 R /BitsPerComponent 8 0 R /Width 2 /Height 1 >> stream\n" +
		"\xFF\xD8\xFF\xE0JPEGDATA\xFF\xD9\n" +
		"endstream\n" +
		"endobj\n" +
		"6 0 obj [/DCTDecode] endobj\n" +
		"7 0 obj [/ICCBased 9 0 R] endobj\n" +
		"8 0 obj 8 endobj\n" +
		"9 0 obj << /N 3 >> stream\n" +
		"icc\n" +
		"endstream\n" +
		"endobj\n"

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "image-metadata.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}

	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeImage: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 1; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}
	if got, want := artifacts[0].Format, model.ImageFormatJPEG; got != want {
		t.Fatalf("artifact format = %q, want %q", got, want)
	}
	if got, want := artifacts[0].ColorSpace, "ICCBased"; got != want {
		t.Fatalf("artifact color space = %q, want %q", got, want)
	}
	if got, want := artifacts[0].BitsPerComponent, 8; got != want {
		t.Fatalf("artifact bits per component = %d, want %d", got, want)
	}
	if got, want := len(artifacts[0].Filters), 1; got != want {
		t.Fatalf("artifact filters len = %d, want %d", got, want)
	}
	if got, want := artifacts[0].Filters[0], "DCTDecode"; got != want {
		t.Fatalf("artifact filter = %q, want %q", got, want)
	}
}

func TestGraphicsExtractionCanonicalizesImageFilterAliasMetadata(t *testing.T) {
	const pdf = "%PDF-1.7\n" +
		"1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n" +
		"2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj\n" +
		"3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /XObject << /Im0 5 0 R >> >> /Contents 4 0 R >> endobj\n" +
		"4 0 obj << /Length 28 >> stream\n" +
		"q\n" +
		"20 0 0 10 30 40 cm\n" +
		"/Im0 Do\n" +
		"Q\n" +
		"endstream\n" +
		"endobj\n" +
		"5 0 obj << /Subtype /Image /Filter /DCT /ColorSpace /DeviceRGB /BitsPerComponent 8 /Width 2 /Height 1 >> stream\n" +
		"\xFF\xD8\xFF\xE0JPEGDATA\xFF\xD9\n" +
		"endstream\n" +
		"endobj\n"

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "image-alias.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}

	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeImage: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 1; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}
	if got, want := artifacts[0].Format, model.ImageFormatJPEG; got != want {
		t.Fatalf("artifact format = %q, want %q", got, want)
	}
	if got, want := len(artifacts[0].Filters), 1; got != want {
		t.Fatalf("artifact filters len = %d, want %d", got, want)
	}
	if got, want := artifacts[0].Filters[0], "DCTDecode"; got != want {
		t.Fatalf("artifact filter = %q, want %q", got, want)
	}
}

func TestGraphicsExtractionDecodesASCII85FlateImagePayload(t *testing.T) {
	encoded := ascii85EncodeForTest(t, zlibEncodeForTest(t, []byte{
		0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A,
		'P', 'N', 'G', 'D', 'A', 'T', 'A',
	}))
	pdf := "%PDF-1.7\n" +
		"1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n" +
		"2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj\n" +
		"3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /XObject << /Im0 5 0 R >> >> /Contents 4 0 R >> endobj\n" +
		"4 0 obj << /Length 28 >> stream\n" +
		"q\n" +
		"20 0 0 10 30 40 cm\n" +
		"/Im0 Do\n" +
		"Q\n" +
		"endstream\n" +
		"endobj\n" +
		"5 0 obj << /Subtype /Image /Filter [/ASCII85Decode /FlateDecode] /ColorSpace /DeviceRGB /BitsPerComponent 8 /Width 2 /Height 1 >> stream\n" +
		string(encoded) + "\n" +
		"endstream\n" +
		"endobj\n"

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "image-chain.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}

	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeImage: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 1; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}
	if got, want := artifacts[0].Format, model.ImageFormatPNG; got != want {
		t.Fatalf("artifact format = %q, want %q", got, want)
	}
	if len(artifacts[0].Data) < 8 {
		t.Fatalf("artifact data len = %d, want at least 8", len(artifacts[0].Data))
	}
	if !bytes.Equal(artifacts[0].Data[:8], []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}) {
		t.Fatalf("artifact data missing PNG signature: %v", artifacts[0].Data)
	}
	if got, want := len(artifacts[0].Filters), 2; got != want {
		t.Fatalf("artifact filters len = %d, want %d", got, want)
	}
}

func TestGraphicsExtractionFillOnlyPathsDoNotEmitStrokeLines(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Contents 4 0 R >> endobj
4 0 obj << /Length 48 >> stream
10 20 30 40 re f
60 60 m
80 60 l
80 90 l
f
endstream
endobj
`

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "fill-only.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}

	lineArtifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeLine: true})
	if err != nil {
		t.Fatalf("page.Artifacts(lines) error = %v", err)
	}
	if got := len(lineArtifacts); got != 0 {
		t.Fatalf("len(lineArtifacts) = %d, want 0", got)
	}

	pathArtifacts, err := page.Artifacts(nil, ArtifactOptions{IncludePath: true})
	if err != nil {
		t.Fatalf("page.Artifacts(paths) error = %v", err)
	}
	if got, want := len(pathArtifacts), 2; got != want {
		t.Fatalf("len(pathArtifacts) = %d, want %d", got, want)
	}

	tableCandidates, err := page.TableCandidates(nil)
	if err != nil {
		t.Fatalf("page.TableCandidates() error = %v", err)
	}
	if got := len(tableCandidates.HorizontalLines) + len(tableCandidates.VerticalLines); got != 0 {
		t.Fatalf("stroke candidate lines = %d, want 0", got)
	}
	if got, want := len(tableCandidates.Rectangles), 1; got != want {
		t.Fatalf("len(rectangles) = %d, want %d", got, want)
	}
}

func TestGraphicsExtractionCloseAndStrokeAddsClosingSegment(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Contents 4 0 R >> endobj
4 0 obj << /Length 34 >> stream
10 10 m
40 10 l
40 50 l
s
endstream
endobj
`

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "close-stroke.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}

	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeLine: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts), 3; got != want {
		t.Fatalf("len(artifacts) = %d, want %d", got, want)
	}
	if got, want := artifacts[2].Bounds.Left, 9.5; got != want {
		t.Fatalf("closing segment left = %v, want %v", got, want)
	}
	if got, want := artifacts[2].Bounds.Top, 50.5; got != want {
		t.Fatalf("closing segment top = %v, want %v", got, want)
	}

	tableCandidates, err := page.TableCandidates(nil)
	if err != nil {
		t.Fatalf("page.TableCandidates() error = %v", err)
	}
	if got, want := len(tableCandidates.HorizontalLines), 1; got != want {
		t.Fatalf("len(horizontal lines) = %d, want %d", got, want)
	}
	if got, want := len(tableCandidates.VerticalLines), 1; got != want {
		t.Fatalf("len(vertical lines) = %d, want %d", got, want)
	}
}

func TestGraphicsExtractionNoOpPathTerminatorsDiscardPendingGeometry(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Contents 4 0 R >> endobj
4 0 obj << /Length 46 >> stream
10 20 30 40 re W n
50 50 m
90 50 l
h
n
endstream
endobj
`

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "noop-path.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}

	artifacts, err := page.Artifacts(nil, ArtifactOptions{IncludeLine: true, IncludePath: true})
	if err != nil {
		t.Fatalf("page.Artifacts() error = %v", err)
	}
	if got := len(artifacts); got != 0 {
		t.Fatalf("len(artifacts) = %d, want 0", got)
	}

	tableCandidates, err := page.TableCandidates(nil)
	if err != nil {
		t.Fatalf("page.TableCandidates() error = %v", err)
	}
	if got := len(tableCandidates.HorizontalLines) + len(tableCandidates.VerticalLines) + len(tableCandidates.Rectangles); got != 0 {
		t.Fatalf("table candidate count = %d, want 0", got)
	}
}
