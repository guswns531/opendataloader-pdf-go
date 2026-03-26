/*
 * Copyright 2025-2026 Hancom Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package integration_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/markdown"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/processors"
	_ "github.com/opendataloader-project/opendataloader-pdf-go/internal/processors"
)

func writeTestPDF(t *testing.T) string {
	t.Helper()

	content := "BT /F1 18 Tf 72 400 Td (Hello PDFBox from integration test) Tj ET\n"
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

	pdfPath := filepath.Join(t.TempDir(), "integration_text.pdf")
	require.NoError(t, os.WriteFile(pdfPath, buf.Bytes(), 0o644))
	return pdfPath
}

func TestEndToEnd(t *testing.T) {
	samplesDir := filepath.Join("..", "..", "..", "samples")

	var pdfPath string
	preferredSamples := []string{
		filepath.Join(samplesDir, "pdf", "lorem.pdf"),
		filepath.Join(samplesDir, "pdf", "1901.03003.pdf"),
		filepath.Join(samplesDir, "pdf", "2408.02509v1.pdf"),
	}
	for _, candidate := range preferredSamples {
		if _, err := os.Stat(candidate); err == nil {
			pdfPath = candidate
			break
		}
	}

	if pdfPath == "" {
		err := filepath.WalkDir(samplesDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.EqualFold(filepath.Ext(path), ".pdf") {
				return nil
			}
			pdfPath = path
			return fs.SkipAll
		})
		require.NoError(t, err)
	}

	if pdfPath == "" {
		t.Skip("no sample PDFs found")
	}

	outputDir := t.TempDir()
	cfg := api.DefaultConfig()
	cfg.OutputDir = outputDir
	cfg.Formats = []string{api.FormatMarkdown, api.FormatJSON, api.FormatHTML}
	cfg.ContentSafetyOff = []string{"all"}

	fixturePDF := writeTestPDF(t)
	require.NoError(t, api.ProcessFile(fixturePDF, cfg))

	doc, err := processors.NewDocumentProcessor().Process(fixturePDF, cfg)
	require.NoError(t, err)

	textChunkCount := 0
	nonEmptyChunkIDs := 0
	for _, page := range doc.Pages {
		textChunkCount += len(page.Chunks)
		for _, chunk := range page.Chunks {
			if chunk != nil && chunk.ID != "" {
				nonEmptyChunkIDs++
			}
		}
	}
	t.Logf("extracted text chunk count: %d", textChunkCount)
	assert.Greater(t, nonEmptyChunkIDs, 0)

	markdownOutput, err := markdown.NewMarkdownGenerator(cfg, false, false).Generate(doc)
	require.NoError(t, err)
	assert.NotEmpty(t, strings.TrimSpace(markdownOutput))

	base := strings.TrimSuffix(filepath.Base(fixturePDF), filepath.Ext(fixturePDF))
	mdPath := filepath.Join(outputDir, base+".md")
	assert.FileExists(t, mdPath)
	assert.FileExists(t, filepath.Join(outputDir, base+".json"))
	assert.FileExists(t, filepath.Join(outputDir, base+".html"))

	markdownBytes, err := os.ReadFile(mdPath)
	require.NoError(t, err)
	assert.NotEmpty(t, strings.TrimSpace(string(markdownBytes)))

	jsonPath := filepath.Join(outputDir, base+".json")
	jsonBytes, err := os.ReadFile(jsonPath)
	require.NoError(t, err)

	schemaPath := filepath.Join("..", "..", "..", "schema.json")
	schemaBytes, err := os.ReadFile(schemaPath)
	require.NoError(t, err)

	schemaResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schemaBytes),
		gojsonschema.NewBytesLoader(jsonBytes),
	)
	require.NoError(t, err)
	assert.Truef(t, schemaResult.Valid(), "schema errors: %v", schemaResult.Errors())

	sampleOutputDir := t.TempDir()
	sampleCfg := api.DefaultConfig()
	sampleCfg.OutputDir = sampleOutputDir
	sampleCfg.Formats = []string{api.FormatMarkdown}
	require.NoError(t, api.ProcessFile(pdfPath, sampleCfg))
}

func TestEndToEndWithContent(t *testing.T) {
	samplesDir := filepath.Join("..", "..", "..", "samples")

	var pdfPath string
	err := filepath.WalkDir(samplesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.EqualFold(filepath.Ext(path), ".pdf") {
			return nil
		}
		pdfPath = path
		return fs.SkipAll
	})
	require.NoError(t, err)
	if pdfPath == "" {
		t.Skip("no sample PDFs found")
	}

	outputDir := t.TempDir()
	cfg := api.DefaultConfig()
	cfg.OutputDir = outputDir
	cfg.Formats = []string{api.FormatMarkdown}
	cfg.ContentSafetyOff = []string{"all"}

	require.NoError(t, api.ProcessFile(pdfPath, cfg))

	base := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	markdownPath := filepath.Join(outputDir, base+".md")
	assert.FileExists(t, markdownPath)

	markdownBytes, err := os.ReadFile(markdownPath)
	require.NoError(t, err)
	markdownContent := strings.TrimSpace(string(markdownBytes))
	t.Logf("sample markdown length for %s: %d", filepath.Base(pdfPath), len(markdownContent))
	assert.NotEmpty(t, markdownContent, "markdown output should not be empty for real PDF")
}
