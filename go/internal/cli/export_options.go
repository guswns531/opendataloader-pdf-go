// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package cli

import (
	"encoding/json"
	"os"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
)

type OptionDefinition struct {
	Name        string      `json:"name"`
	ShortName   string      `json:"shortName,omitempty"`
	Description string      `json:"description,omitempty"`
	Type        string      `json:"type"`
	Default     interface{} `json:"default,omitempty"`
	Choices     []string    `json:"choices,omitempty"`
	Multiple    bool        `json:"multiple"`
}

var exportableOptionDefinitions = []OptionDefinition{
	{Name: "output-dir", ShortName: "o", Description: "Directory where output files are written. Default: input file directory", Type: "string", Multiple: false},
	{Name: "password", ShortName: "p", Description: "Password for encrypted PDF files", Type: "string", Multiple: false},
	{Name: "format", ShortName: "f", Description: "Output formats (comma-separated). Values: json, text, html, pdf, markdown, markdown-with-html, markdown-with-images", Type: "[]string", Choices: []string{api.FormatJSON, api.FormatText, api.FormatHTML, api.FormatPDF, api.FormatMarkdown, api.FormatMarkdownWithHTML, api.FormatMarkdownWithImages}, Multiple: true},
	{Name: "quiet", ShortName: "q", Description: "Suppress console logging output", Type: "boolean", Default: false, Multiple: false},
	{Name: "content-safety-off", Description: "Disable content safety filters. Values: all, hidden-text, off-page, tiny, hidden-ocg", Type: "[]string", Choices: []string{"all", "hidden-text", "off-page", "tiny", "hidden-ocg"}, Multiple: true},
	{Name: "sanitize", Description: "Enable sensitive data sanitization", Type: "boolean", Default: false, Multiple: false},
	{Name: "keep-line-breaks", Description: "Preserve original line breaks in extracted text", Type: "boolean", Default: false, Multiple: false},
	{Name: "replace-invalid-chars", Description: "Replacement string for invalid or unrecognized characters", Type: "string", Default: " ", Multiple: false},
	{Name: "use-struct-tree", Description: "Use PDF structure tree for reading order and semantic structure", Type: "boolean", Default: false, Multiple: false},
	{Name: "table-method", Description: "Table detection method. Values: default, cluster", Type: "string", Default: api.TableMethodDefault, Choices: []string{api.TableMethodDefault, api.TableMethodCluster}, Multiple: false},
	{Name: "reading-order", Description: "Reading order algorithm. Values: off, xycut", Type: "string", Default: api.ReadingOrderXYCut, Choices: []string{api.ReadingOrderOff, api.ReadingOrderXYCut}, Multiple: false},
	{Name: "markdown-page-separator", Description: "Separator between pages in Markdown output", Type: "string", Multiple: false},
	{Name: "text-page-separator", Description: "Separator between pages in text output", Type: "string", Multiple: false},
	{Name: "html-page-separator", Description: "Separator between pages in HTML output", Type: "string", Multiple: false},
	{Name: "image-output", Description: "Image output mode. Values: off, embedded, external", Type: "string", Default: api.ImageOutputExternal, Choices: []string{api.ImageOutputOff, api.ImageOutputEmbedded, api.ImageOutputExternal}, Multiple: false},
	{Name: "image-format", Description: "Output format for extracted images. Values: png, jpeg", Type: "string", Default: api.ImageFormatPNG, Choices: []string{api.ImageFormatPNG, api.ImageFormatJPEG}, Multiple: false},
	{Name: "image-dir", Description: "Directory for extracted images", Type: "string", Multiple: false},
	{Name: "pages", Description: "Pages to extract (e.g. \"1,3,5-7\")", Type: "string", Multiple: false},
	{Name: "include-header-footer", Description: "Include page headers and footers in output", Type: "boolean", Default: false, Multiple: false},
	{Name: "detect-strikethrough", Description: "Detect strikethrough text for Markdown output", Type: "boolean", Default: false, Multiple: false},
	{Name: "hybrid", Description: "Hybrid backend. Values: off, docling-fast", Type: "string", Default: api.HybridOff, Choices: []string{api.HybridOff, api.HybridDoclingFast}, Multiple: false},
	{Name: "hybrid-mode", Description: "Hybrid triage mode. Values: auto, full", Type: "string", Default: api.HybridModeAuto, Choices: []string{api.HybridModeAuto, api.HybridModeFull}, Multiple: false},
	{Name: "hybrid-url", Description: "Hybrid backend server URL", Type: "string", Multiple: false},
	{Name: "hybrid-timeout", Description: "Hybrid backend request timeout in milliseconds", Type: "integer", Default: 30000, Multiple: false},
	{Name: "hybrid-fallback", Description: "Opt in to Java fallback on hybrid backend error", Type: "boolean", Default: false, Multiple: false},
}

func ExportOptionsJSON() error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(exportableOptionDefinitions)
}
