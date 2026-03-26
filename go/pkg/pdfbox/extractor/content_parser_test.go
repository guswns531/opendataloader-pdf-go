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

func TestNormalizePDFStringDecodesWinAnsiAndLigature(t *testing.T) {
	assert.Equal(t, "Euro: €", normalizePDFString([]byte("Euro: \x80")))
	assert.Equal(t, "Rectified", normalizePDFString([]byte("Recti\x02ed")))
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
