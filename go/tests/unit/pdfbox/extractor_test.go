package pdfbox_test

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/extractor"
	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/model"
	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, err)
	defer doc.Close()

	texts, err := extractor.ExtractTextChunks(doc, 0)
	assert.NoError(t, err)
	t.Logf("Extracted %d text chunks from page 1", len(texts))
}

func TestPageCount(t *testing.T) {
	pdf := findSamplePDF(t)
	doc, err := model.Open(pdf, "")
	assert.NoError(t, err)
	defer doc.Close()

	assert.Greater(t, doc.PageCount(), 0)
}
