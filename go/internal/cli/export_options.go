// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
)

type OptionDefinition struct {
	Name        string      `json:"name"`
	ShortName   string      `json:"shortName,omitempty"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Default     interface{} `json:"default,omitempty"`
	Choices     []string    `json:"choices,omitempty"`
	Multiple    bool        `json:"multiple,omitempty"`
}

type cliOptionDefinition struct {
	OptionDefinition
	Exported bool
	Hidden   bool
}

type exportedOptions struct {
	Options []OptionDefinition `json:"options"`
}

var optionDefinitions = []cliOptionDefinition{
	{OptionDefinition: OptionDefinition{Name: "output-dir", ShortName: "o", Type: "string", Description: "Directory where output files are written. Default: input file directory"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "password", ShortName: "p", Type: "string", Description: "Password for encrypted PDF files"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "format", ShortName: "f", Type: "[]string", Default: []string{api.FormatJSON}, Choices: []string{api.FormatJSON, api.FormatText, api.FormatHTML, api.FormatPDF, api.FormatMarkdown, api.FormatMarkdownWithHTML, api.FormatMarkdownWithImages}, Multiple: true, Description: "Output formats (comma-separated). Values: json, text, html, pdf, markdown, markdown-with-html, markdown-with-images. Default: json"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "quiet", ShortName: "q", Type: "boolean", Default: false, Description: "Suppress console logging output"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "content-safety-off", Type: "[]string", Choices: []string{api.ContentSafetyAll, api.ContentSafetyHiddenText, api.ContentSafetyOffPage, api.ContentSafetyTiny, api.ContentSafetyHiddenOCG}, Multiple: true, Description: "Disable content safety filters. Values: all, hidden-text, off-page, tiny, hidden-ocg"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "sanitize", Type: "boolean", Default: false, Description: "Enable sensitive data sanitization. Replaces emails, phone numbers, IPs, credit cards, and URLs with placeholders"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "keep-line-breaks", Type: "boolean", Default: false, Description: "Preserve original line breaks in extracted text"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "replace-invalid-chars", Type: "string", Default: " ", Description: "Replacement character for invalid/unrecognized characters. Default: space"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "use-struct-tree", Type: "boolean", Default: false, Description: "Use PDF structure tree (tagged PDF) for reading order and semantic structure"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "table-method", Type: "string", Default: api.TableMethodDefault, Choices: []string{api.TableMethodDefault, api.TableMethodCluster}, Description: "Table detection method. Values: default (border-based), cluster (border + cluster). Default: default"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "reading-order", Type: "string", Default: api.ReadingOrderXYCut, Choices: []string{api.ReadingOrderOff, api.ReadingOrderXYCut}, Description: "Reading order algorithm. Values: off, xycut. Default: xycut"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "markdown-page-separator", Type: "string", Description: "Separator between pages in Markdown output. Use %page-number% for page numbers. Default: none"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "text-page-separator", Type: "string", Description: "Separator between pages in text output. Use %page-number% for page numbers. Default: none"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "html-page-separator", Type: "string", Description: "Separator between pages in HTML output. Use %page-number% for page numbers. Default: none"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "image-output", Type: "string", Default: api.ImageOutputExternal, Choices: []string{api.ImageOutputOff, api.ImageOutputEmbedded, api.ImageOutputExternal}, Description: "Image output mode. Values: off (no images), embedded (Base64 data URIs), external (file references). Default: external"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "image-format", Type: "string", Default: api.ImageFormatPNG, Choices: []string{api.ImageFormatPNG, api.ImageFormatJPEG}, Description: "Output format for extracted images. Values: png, jpeg. Default: png"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "image-dir", Type: "string", Description: "Directory for extracted images"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "pages", Type: "string", Description: "Pages to extract (e.g., \"1,3,5-7\"). Default: all pages"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "include-header-footer", Type: "boolean", Default: false, Description: "Include page headers and footers in output"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "detect-strikethrough", Type: "boolean", Default: false, Description: "Detect strikethrough text and wrap with ~~ in Markdown output (experimental)"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "hybrid", Type: "string", Default: api.HybridOff, Choices: []string{api.HybridOff, api.HybridDocling, api.HybridDoclingFast, api.HybridHancom}, Description: "Hybrid backend for AI processing. Values: off (default), docling, docling-fast, hancom"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "hybrid-mode", Type: "string", Default: api.HybridModeAuto, Choices: []string{api.HybridModeAuto, api.HybridModeFull}, Description: "Hybrid triage mode. Values: auto (default, dynamic triage), full (skip triage, all pages to backend)"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "hybrid-url", Type: "string", Description: "Hybrid backend server URL (overrides default)"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "hybrid-timeout", Type: "integer", Default: 30000, Description: "Hybrid backend request timeout in milliseconds. Default: 30000"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "hybrid-fallback", Type: "boolean", Default: false, Description: "Opt in to Java fallback on hybrid backend error (default: disabled)"}, Exported: true},
	{OptionDefinition: OptionDefinition{Name: "export-options", Type: "boolean"}, Hidden: true},
	{OptionDefinition: OptionDefinition{Name: "hybrid-ocr", Type: "string", Description: "[Deprecated] OCR settings are now configured on the hybrid server (--ocr-lang, --force-ocr)"}, Hidden: true},
	{OptionDefinition: OptionDefinition{Name: "pdf", Type: "boolean"}, Hidden: true},
	{OptionDefinition: OptionDefinition{Name: "markdown", Type: "boolean"}, Hidden: true},
	{OptionDefinition: OptionDefinition{Name: "html", Type: "boolean"}, Hidden: true},
	{OptionDefinition: OptionDefinition{Name: "markdown-with-html", Type: "boolean"}, Hidden: true},
	{OptionDefinition: OptionDefinition{Name: "markdown-with-images", Type: "boolean"}, Hidden: true},
	{OptionDefinition: OptionDefinition{Name: "no-json", Type: "boolean"}, Hidden: true},
}

func exportedOptionDefinitions() []OptionDefinition {
	exported := make([]OptionDefinition, 0, len(optionDefinitions))
	for _, definition := range optionDefinitions {
		if definition.Exported {
			exported = append(exported, definition.OptionDefinition)
		}
	}
	return exported
}

func ExportOptionsJSON() error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(exportedOptions{Options: exportedOptionDefinitions()}); err != nil {
		return err
	}
	return nil
}

func ExportOptionsAndExit() {
	if err := ExportOptionsJSON(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
