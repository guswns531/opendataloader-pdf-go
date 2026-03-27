package pdfbox_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/extractor"
	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findSamplePDF(t *testing.T) string {
	t.Helper()

	root := filepath.Clean("../../../samples")
	var found string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".pdf" {
			found = path
			return fs.SkipAll
		}
		return nil
	})
	if found == "" {
		t.Skip("no sample PDFs found in samples/")
	}
	return found
}

func TestExtractTextChunks(t *testing.T) {
	pdf := findSamplePDF(t)
	doc, err := model.Open(pdf, "")
	require.NoError(t, err)
	defer doc.Close()

	texts, err := extractor.ExtractTextChunks(doc, 0)
	require.NoError(t, err)
	assert.Greater(t, len(texts), 0)
	t.Logf("Extracted %d text chunks from page 1", len(texts))
}

func TestExtractImages(t *testing.T) {
	pdf := findSamplePDF(t)
	doc, err := model.Open(pdf, "")
	require.NoError(t, err)
	defer doc.Close()

	images, err := extractor.ExtractImages(doc, 0, "")
	require.NoError(t, err)
	assert.NotNil(t, images)
	t.Logf("Extracted %d images from page 1", len(images))
}

func TestExtractLineArts(t *testing.T) {
	pdf := writeLineArtPDF(t)
	doc, err := model.Open(pdf, "")
	require.NoError(t, err)
	defer doc.Close()

	lines, err := extractor.ExtractLineArts(doc, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, lines)
}

func TestPageCount(t *testing.T) {
	pdf := findSamplePDF(t)
	doc, err := model.Open(pdf, "")
	require.NoError(t, err)
	defer doc.Close()

	assert.Greater(t, doc.PageCount(), 0)
}

func TestGetPageUsesZeroBasedNumber(t *testing.T) {
	pdf := writeLineArtPDF(t)
	doc, err := model.Open(pdf, "")
	require.NoError(t, err)
	defer doc.Close()

	page, err := doc.GetPage(0)
	require.NoError(t, err)
	assert.Equal(t, 0, page.Number)
}

func TestExtractTextChunksAvoidFalseOverlapForSurveyIEEEFixture(t *testing.T) {
	pdf := filepath.Clean("../../../samples/pdf/1901.03003.pdf")
	if _, err := os.Stat(pdf); err != nil {
		t.Skip("fixture not available")
	}

	doc, err := model.Open(pdf, "")
	require.NoError(t, err)
	defer doc.Close()

	chunks, err := extractor.ExtractTextChunks(doc, 15)
	require.NoError(t, err)

	surveyIdx := findChunkIndex(chunks, "tion in imagery: A survey.")
	ieeeIdx := findChunkIndex(chunks, "IEEE Trans. Pattern Anal.")
	require.NotEqual(t, -1, surveyIdx)
	require.NotEqual(t, -1, ieeeIdx)
	require.Equal(t, surveyIdx+1, ieeeIdx)

	surveyChunk := chunks[surveyIdx]
	ieeeChunk := chunks[ieeeIdx]
	gap := ieeeChunk.X - (surveyChunk.X + surveyChunk.Width)
	t.Logf("survey width=%.3f next x=%.3f gap=%.3f", surveyChunk.Width, ieeeChunk.X, gap)
	assert.True(t, gap > 0 || math.Abs(gap) <= 0.5, "expected no material false overlap, got gap %.3f", gap)
}

func findChunkIndex(chunks []*extractor.ExtractedText, needle string) int {
	for i, chunk := range chunks {
		if chunk != nil && strings.Contains(chunk.Text, needle) {
			return i
		}
	}
	return -1
}

func writeLineArtPDF(t *testing.T) string {
	t.Helper()

	content := "10 w 72 500 m 300 500 l S\n10 w 150 420 m 150 300 l S\n"
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>",
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

	pdfPath := filepath.Join(t.TempDir(), "line_art.pdf")
	require.NoError(t, os.WriteFile(pdfPath, buf.Bytes(), 0o644))
	return pdfPath
}
