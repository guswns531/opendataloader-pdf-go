package extractor

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeTJTextInsertsSpaceAtThreshold(t *testing.T) {
	got := decodeTJText(streamToken{
		kind: "array",
		items: []streamToken{
			{kind: "string", value: "A"},
			{kind: "number", value: "-100"},
			{kind: "string", value: "B"},
		},
	})

	assert.Equal(t, "A B", got)
}

func TestDecodeTJTextDoesNotEmitDanglingOrDuplicateRecoveredSpaces(t *testing.T) {
	got := decodeTJText(streamToken{
		kind: "array",
		items: []streamToken{
			{kind: "number", value: "-250"},
			{kind: "string", value: "A "},
			{kind: "number", value: "-250"},
			{kind: "string", value: "B"},
			{kind: "number", value: "-250"},
		},
	})

	assert.Equal(t, "A B", got)
}

func TestNormalizePDFStringDecodesWinAnsiAndLigature(t *testing.T) {
	assert.Equal(t, "Euro: €", normalizePDFString([]byte("Euro: \x80")))
	assert.Equal(t, "Quotes: ‘’ “”— œ £", normalizePDFString([]byte("Quotes: \x91\x92 \x93\x94\x97 \x9c \xa3")))
	assert.Equal(t, "Rectified", normalizePDFString([]byte("Recti\x02ed")))
}

func TestFontDecoderWinAnsiTableDecodesExtendedBytes(t *testing.T) {
	decoder := &fontDecoder{
		hasFont:   true,
		encoding:  winAnsiTable(),
		toUnicode: nil,
	}

	assert.Equal(t, "Quotes: ‘’ “”— œ £", decoder.decode([]byte("Quotes: \x91\x92 \x93\x94\x97 \x9c \xa3")))
	assert.Equal(t, "ASCII stays ASCII", decoder.decode([]byte("ASCII stays ASCII")))
}

func TestFontDecoderWinAnsiTableMatchesRawWinAnsiFallbackForUndefinedBytes(t *testing.T) {
	raw := []byte{0x81, 0x8d, 0x8f, 0x90, 0x9d}
	decoder := &fontDecoder{
		hasFont:   true,
		encoding:  winAnsiTable(),
		toUnicode: nil,
	}

	assert.Equal(t, normalizePDFString(raw), decoder.decode(raw))
	assert.NotContains(t, decoder.decode(raw), "\uFFFD")
}

func TestParseCMapContentDecodesBFCharAndBFRange(t *testing.T) {
	cmap := `
2 beginbfchar
<01> <0041>
<02> <0042>
endbfchar
1 beginbfrange
<10> <12> <0061>
endbfrange
`
	mapping, codeLens, err := parseCMapContent(cmap)
	require.NoError(t, err)
	assert.Equal(t, []int{1}, codeLens)
	assert.Equal(t, "A", mapping[0x01])
	assert.Equal(t, "B", mapping[0x02])
	assert.Equal(t, "a", mapping[0x10])
	assert.Equal(t, "b", mapping[0x11])
	assert.Equal(t, "c", mapping[0x12])
}

func TestParseCMapContentDecodesMultiLineBFRangeArray(t *testing.T) {
	cmap := `
1 beginbfrange
<0001> <0003> [
<00E9>
<00F1>
<20AC>
]
endbfrange
`
	mapping, codeLens, err := parseCMapContent(cmap)
	require.NoError(t, err)
	assert.Equal(t, []int{2}, codeLens)
	assert.Equal(t, "é", mapping[0x0001])
	assert.Equal(t, "ñ", mapping[0x0002])
	assert.Equal(t, "€", mapping[0x0003])

	decoder := &fontDecoder{
		hasFont:   true,
		toUnicode: mapping,
		codeLens:  codeLens,
		encoding:  standardEncodingTable(),
	}
	assert.Equal(t, "éñ€", decoder.decode([]byte{0x00, 0x01, 0x00, 0x02, 0x00, 0x03}))
}

func TestDecodePDFNameDecodesHexEscapes(t *testing.T) {
	got := decodePDFName("/F#31#20A")
	if got != "/F1 A" {
		t.Fatalf("decodePDFName() = %q, want %q", got, "/F1 A")
	}
}

func TestUTF16BytesToStringDecodesSurrogatePairs(t *testing.T) {
	got := utf16BytesToString([]byte{0xD8, 0x3D, 0xDE, 0x00})
	if got != "😀" {
		t.Fatalf("utf16BytesToString() = %q, want %q", got, "😀")
	}
}

func TestExtractTextChunksPreservesTrailingSpace(t *testing.T) {
	pdf := writeTextPDF(t, "BT /F1 12 Tf 72 400 Td (Hello ) Tj (World) Tj ET\n")
	doc, err := model.Open(pdf, "")
	require.NoError(t, err)
	defer doc.Close()

	chunks, err := ExtractTextChunks(doc, 0)
	require.NoError(t, err)
	require.Len(t, chunks, 2)
	assert.Equal(t, "Hello ", chunks[0].Text)
	assert.Equal(t, "World", chunks[1].Text)
}

func TestExtractTextChunksFoldsWhitespaceOnlyBoundaryRunIntoPreviousChunk(t *testing.T) {
	pdf := writeTextPDF(t, "BT /F1 12 Tf 72 400 Td (Hello) Tj ( ) Tj (World) Tj ET\n")
	doc, err := model.Open(pdf, "")
	require.NoError(t, err)
	defer doc.Close()

	chunks, err := ExtractTextChunks(doc, 0)
	require.NoError(t, err)
	require.Len(t, chunks, 2)
	assert.Equal(t, "Hello ", chunks[0].Text)
	assert.Equal(t, "World", chunks[1].Text)
}

func TestExtractTextChunksFoldsLeadingBoundarySpaceIntoPreviousChunk(t *testing.T) {
	pdf := writeTextPDF(t, "BT /F1 12 Tf 72 400 Td (Hello) Tj ( World) Tj ET\n")
	doc, err := model.Open(pdf, "")
	require.NoError(t, err)
	defer doc.Close()

	chunks, err := ExtractTextChunks(doc, 0)
	require.NoError(t, err)
	require.Len(t, chunks, 2)
	assert.Equal(t, "Hello ", chunks[0].Text)
	assert.Equal(t, "World", chunks[1].Text)
}

func TestExtractTextChunksAvoidsDuplicatingBoundarySpaceAcrossRuns(t *testing.T) {
	pdf := writeTextPDF(t, "BT /F1 12 Tf 72 400 Td (Hello ) Tj ( World) Tj ET\n")
	doc, err := model.Open(pdf, "")
	require.NoError(t, err)
	defer doc.Close()

	chunks, err := ExtractTextChunks(doc, 0)
	require.NoError(t, err)
	require.Len(t, chunks, 2)
	assert.Equal(t, "Hello ", chunks[0].Text)
	assert.Equal(t, "World", chunks[1].Text)
}

func TestExtractTextChunksRecoversSpaceFromTJOffsets(t *testing.T) {
	pdf := writeTextPDF(t, "BT /F1 12 Tf 72 400 Td [(A)-250(Multi-Object)-250(Rectified)-250(Attention)-250(Network)] TJ ET\n")
	doc, err := model.Open(pdf, "")
	require.NoError(t, err)
	defer doc.Close()

	chunks, err := ExtractTextChunks(doc, 0)
	require.NoError(t, err)
	require.Len(t, chunks, 1)
	assert.Equal(t, "A Multi-Object Rectified Attention Network", chunks[0].Text)
}

func writeTextPDF(t *testing.T, content string) string {
	t.Helper()

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(content), content),
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")

	offsets := make([]int, 0, len(objects)+1)
	offsets = append(offsets, 0)
	for idx, obj := range objects {
		offsets = append(offsets, buf.Len())
		buf.WriteString(strconv.Itoa(idx + 1))
		buf.WriteString(" 0 obj\n")
		buf.WriteString(obj)
		buf.WriteString("\nendobj\n")
	}

	xrefOffset := buf.Len()
	buf.WriteString("xref\n")
	buf.WriteString(fmt.Sprintf("0 %d\n", len(offsets)))
	buf.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets[1:] {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}
	buf.WriteString("trailer\n")
	buf.WriteString(fmt.Sprintf("<< /Size %d /Root 1 0 R >>\n", len(offsets)))
	buf.WriteString("startxref\n")
	buf.WriteString(strconv.Itoa(xrefOffset))
	buf.WriteString("\n%%EOF\n")

	pdfPath := filepath.Join(t.TempDir(), "text.pdf")
	require.NoError(t, os.WriteFile(pdfPath, buf.Bytes(), 0o644))
	return pdfPath
}
