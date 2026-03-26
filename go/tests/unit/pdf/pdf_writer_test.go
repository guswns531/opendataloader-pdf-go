package pdf_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	_ "unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	internalpdf "github.com/opendataloader-project/opendataloader-pdf-go/internal/pdf"
)

//go:linkname getColor github.com/opendataloader-project/opendataloader-pdf-go/internal/pdf.getColor
func getColor(objType entities.ObjectType) [3]float64

func TestGetColor(t *testing.T) {
	assert.Equal(t, [3]float64{0, 0, 1}, getColor(entities.ObjectTypeHeading))
	assert.Equal(t, [3]float64{0, 0, 1}, getColor(entities.ObjectTypeHeaderFooter))
	assert.Equal(t, [3]float64{0, 1, 0}, getColor(entities.ObjectTypeList))
	assert.Equal(t, [3]float64{0, 1, 1}, getColor(entities.ObjectTypeParagraph))
	assert.Equal(t, [3]float64{1, 0, 0}, getColor(entities.ObjectTypeImage))
	assert.Equal(t, [3]float64{1, 0, 1}, getColor(entities.ObjectTypeTable))
	assert.Equal(t, [3]float64{1, 1, 0}, getColor(entities.ObjectTypeCaption))
	assert.Equal(t, [3]float64{0.9, 0.9, 0.9}, getColor(entities.ObjectTypeUnknown))
}

func TestPDFLayerValues(t *testing.T) {
	assert.Equal(t, internalpdf.PDFLayer("content"), internalpdf.PDFLayerContent)
	assert.Equal(t, internalpdf.PDFLayer("table cells"), internalpdf.PDFLayerTableCells)
	assert.Equal(t, internalpdf.PDFLayer("list items"), internalpdf.PDFLayerListItems)
	assert.Equal(t, internalpdf.PDFLayer("table content"), internalpdf.PDFLayerTableContent)
	assert.Equal(t, internalpdf.PDFLayer("list content"), internalpdf.PDFLayerListContent)
	assert.Equal(t, internalpdf.PDFLayer("text blocks content"), internalpdf.PDFLayerTextBlockContent)
	assert.Equal(t, internalpdf.PDFLayer("header and footer content"), internalpdf.PDFLayerHeaderFooterContent)
}

func TestPDFWriterUpdatePDFCreatesAnnotatedOutput(t *testing.T) {
	inputPath := writeTestPDF(t)
	outputPath := filepath.Join(t.TempDir(), "annotated.pdf")
	writer := internalpdf.NewPDFWriter()

	doc := &entities.Document{
		Pages: []*entities.Page{
			{
				PageMetadata: entities.PageMetadata{Number: 0, Width: 612, Height: 792},
				Elements: []entities.IObject{
					&entities.SemanticParagraph{
						BaseObject: entities.BaseObject{
							BBox: entities.BoundingBox{X: 72, Y: 380, Width: 220, Height: 24, Page: 0},
						},
						Lines: []*entities.TextLine{
							{
								BaseObject: entities.BaseObject{
									BBox: entities.BoundingBox{X: 72, Y: 380, Width: 220, Height: 24, Page: 0},
								},
								Chunks: []*entities.TextChunk{
									{
										BaseObject: entities.BaseObject{
											BBox: entities.BoundingBox{X: 72, Y: 380, Width: 220, Height: 24, Page: 0},
										},
										Text: "Hello annotated PDF",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	err := writer.UpdatePDF(inputPath, "", outputPath, doc)
	require.NoError(t, err)

	info, err := os.Stat(outputPath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func writeTestPDF(t *testing.T) string {
	t.Helper()

	content := "BT /F1 18 Tf 72 400 Td (Hello PDFWriter) Tj ET\n"
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

	pdfPath := filepath.Join(t.TempDir(), "input.pdf")
	if err := os.WriteFile(pdfPath, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write test pdf: %v", err)
	}
	return pdfPath
}
