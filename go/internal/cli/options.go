// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package cli

import (
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/spf13/cobra"
)

type CLIOptions struct {
	OutputDir             string
	Password              string
	Format                []string
	Quiet                 bool
	ContentSafetyOff      []string
	Sanitize              bool
	KeepLineBreaks        bool
	ReplaceInvalidChars   string
	UseStructTree         bool
	TableMethod           string
	ReadingOrder          string
	DetectStrikethrough   bool
	MarkdownPageSeparator string
	TextPageSeparator     string
	HTMLPageSeparator     string
	ImageOutput           string
	ImageFormat           string
	ImageDir              string
	Pages                 string
	IncludeHeaderFooter   bool
	Hybrid                string
	HybridMode            string
	HybridURL             string
	HybridTimeout         int
	HybridFallback        bool
	ExportOptions         bool
}

func AddFlags(cmd *cobra.Command, opts *CLIOptions) {
	flags := cmd.Flags()

	flags.StringVarP(&opts.OutputDir, "output-dir", "o", "", "Directory where output files are written. Default: input file directory")
	flags.StringVarP(&opts.Password, "password", "p", "", "Password for encrypted PDF files")
	flags.StringSliceVarP(&opts.Format, "format", "f", nil, "Output formats (comma-separated). Values: json, text, html, pdf, markdown, markdown-with-html, markdown-with-images")
	flags.BoolVarP(&opts.Quiet, "quiet", "q", false, "Suppress console logging output")
	flags.StringSliceVar(&opts.ContentSafetyOff, "content-safety-off", nil, "Disable content safety filters. Values: all, hidden-text, off-page, tiny, hidden-ocg")
	flags.BoolVar(&opts.Sanitize, "sanitize", false, "Enable sensitive data sanitization")
	flags.BoolVar(&opts.KeepLineBreaks, "keep-line-breaks", false, "Preserve original line breaks in extracted text")
	flags.StringVar(&opts.ReplaceInvalidChars, "replace-invalid-chars", " ", "Replacement string for invalid or unrecognized characters")
	flags.BoolVar(&opts.UseStructTree, "use-struct-tree", false, "Use PDF structure tree for reading order and semantic structure")
	flags.StringVar(&opts.TableMethod, "table-method", api.TableMethodDefault, "Table detection method. Values: default, cluster")
	flags.StringVar(&opts.ReadingOrder, "reading-order", api.ReadingOrderXYCut, "Reading order algorithm. Values: off, xycut")
	flags.StringVar(&opts.MarkdownPageSeparator, "markdown-page-separator", "", "Separator between pages in Markdown output")
	flags.StringVar(&opts.TextPageSeparator, "text-page-separator", "", "Separator between pages in text output")
	flags.StringVar(&opts.HTMLPageSeparator, "html-page-separator", "", "Separator between pages in HTML output")
	flags.StringVar(&opts.ImageOutput, "image-output", api.ImageOutputExternal, "Image output mode. Values: off, embedded, external")
	flags.StringVar(&opts.ImageFormat, "image-format", api.ImageFormatPNG, "Output format for extracted images. Values: png, jpeg")
	flags.StringVar(&opts.ImageDir, "image-dir", "", "Directory for extracted images")
	flags.StringVar(&opts.Pages, "pages", "", "Pages to extract (e.g. \"1,3,5-7\")")
	flags.BoolVar(&opts.IncludeHeaderFooter, "include-header-footer", false, "Include page headers and footers in output")
	flags.BoolVar(&opts.DetectStrikethrough, "detect-strikethrough", false, "Detect strikethrough text for Markdown output")
	flags.StringVar(&opts.Hybrid, "hybrid", api.HybridOff, "Hybrid backend. Values: off, docling-fast")
	flags.StringVar(&opts.HybridMode, "hybrid-mode", api.HybridModeAuto, "Hybrid triage mode. Values: auto, full")
	flags.StringVar(&opts.HybridURL, "hybrid-url", "", "Hybrid backend server URL")
	flags.IntVar(&opts.HybridTimeout, "hybrid-timeout", 30000, "Hybrid backend request timeout in milliseconds")
	flags.BoolVar(&opts.HybridFallback, "hybrid-fallback", false, "Opt in to Java fallback on hybrid backend error")
	flags.BoolVar(&opts.ExportOptions, "export-options", false, "Export CLI option definitions as JSON")
}

func (o *CLIOptions) ToConfig() *api.Config {
	cfg := api.DefaultConfig()
	cfg.OutputDir = o.OutputDir
	cfg.Password = o.Password
	if len(o.Format) > 0 {
		cfg.Formats = append([]string(nil), o.Format...)
	}
	cfg.Quiet = o.Quiet
	if len(o.ContentSafetyOff) > 0 {
		cfg.ContentSafetyOff = append([]string(nil), o.ContentSafetyOff...)
	}
	cfg.Sanitize = o.Sanitize
	cfg.KeepLineBreaks = o.KeepLineBreaks
	cfg.ReplaceInvalidChars = o.ReplaceInvalidChars
	cfg.UseStructTree = o.UseStructTree
	cfg.TableMethod = o.TableMethod
	cfg.ReadingOrder = o.ReadingOrder
	cfg.DetectStrikethrough = o.DetectStrikethrough
	cfg.MarkdownPageSeparator = o.MarkdownPageSeparator
	cfg.TextPageSeparator = o.TextPageSeparator
	cfg.HTMLPageSeparator = o.HTMLPageSeparator
	cfg.ImageOutput = o.ImageOutput
	cfg.ImageFormat = o.ImageFormat
	cfg.ImageDir = o.ImageDir
	cfg.Pages = o.Pages
	cfg.IncludeHeaderFooter = o.IncludeHeaderFooter
	cfg.Hybrid = o.Hybrid
	cfg.HybridMode = o.HybridMode
	cfg.HybridURL = o.HybridURL
	cfg.HybridTimeout = o.HybridTimeout
	cfg.HybridFallback = o.HybridFallback
	return cfg
}
