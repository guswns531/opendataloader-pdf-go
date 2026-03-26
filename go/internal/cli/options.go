// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/spf13/cobra"
)

type CLIOptions struct {
	OutputDir           string
	Password            string
	Format              []string
	Quiet               bool
	ContentSafetyOff    []string
	Sanitize            bool
	KeepLineBreaks      bool
	ReplaceInvalidChars string
	UseStructTree       bool
	TableMethod         string
	ReadingOrder        string
	MarkdownPageSep     string
	TextPageSep         string
	HTMLPageSep         string
	ImageOutput         string
	ImageFormat         string
	ImageDir            string
	Pages               string
	IncludeHeaderFooter bool
	DetectStrikethrough bool
	Hybrid              string
	HybridMode          string
	HybridURL           string
	HybridTimeout       int
	HybridFallback      bool
	HybridOCR           string
	ExportOptions       bool

	MarkdownReport         bool
	HTMLReport             bool
	PDFReport              bool
	MarkdownWithHTMLReport bool
	MarkdownWithImages     bool
	NoJSONReport           bool
}

func AddFlags(cmd *cobra.Command, opts *CLIOptions) {
	if opts == nil {
		return
	}

	defaults := api.DefaultConfig()
	flags := cmd.Flags()
	flags.StringVarP(&opts.OutputDir, "output-dir", "o", "", "Directory where output files are written. Default: input file directory")
	flags.StringVarP(&opts.Password, "password", "p", "", "Password for encrypted PDF files")
	flags.StringSliceVarP(&opts.Format, "format", "f", append([]string(nil), defaults.Formats...), "Output formats (comma-separated). Values: json, text, html, pdf, markdown, markdown-with-html, markdown-with-images. Default: json")
	flags.BoolVarP(&opts.Quiet, "quiet", "q", defaults.Quiet, "Suppress console logging output")
	flags.StringSliceVar(&opts.ContentSafetyOff, "content-safety-off", nil, "Disable content safety filters. Values: all, hidden-text, off-page, tiny, hidden-ocg")
	flags.BoolVar(&opts.Sanitize, "sanitize", defaults.Sanitize, "Enable sensitive data sanitization. Replaces emails, phone numbers, IPs, credit cards, and URLs with placeholders")
	flags.BoolVar(&opts.KeepLineBreaks, "keep-line-breaks", defaults.KeepLineBreaks, "Preserve original line breaks in extracted text")
	flags.StringVar(&opts.ReplaceInvalidChars, "replace-invalid-chars", defaults.ReplaceInvalidChars, "Replacement character for invalid/unrecognized characters. Default: space")
	flags.BoolVar(&opts.UseStructTree, "use-struct-tree", defaults.UseStructTree, "Use PDF structure tree (tagged PDF) for reading order and semantic structure")
	flags.StringVar(&opts.TableMethod, "table-method", defaults.TableMethod, "Table detection method. Values: default (border-based), cluster (border + cluster). Default: default")
	flags.StringVar(&opts.ReadingOrder, "reading-order", defaults.ReadingOrder, "Reading order algorithm. Values: off, xycut. Default: xycut")
	flags.StringVar(&opts.MarkdownPageSep, "markdown-page-separator", defaults.MarkdownPageSeparator, "Separator between pages in Markdown output. Use %page-number% for page numbers. Default: none")
	flags.StringVar(&opts.TextPageSep, "text-page-separator", defaults.TextPageSeparator, "Separator between pages in text output. Use %page-number% for page numbers. Default: none")
	flags.StringVar(&opts.HTMLPageSep, "html-page-separator", defaults.HTMLPageSeparator, "Separator between pages in HTML output. Use %page-number% for page numbers. Default: none")
	flags.StringVar(&opts.ImageOutput, "image-output", defaults.ImageOutput, "Image output mode. Values: off (no images), embedded (Base64 data URIs), external (file references). Default: external")
	flags.StringVar(&opts.ImageFormat, "image-format", defaults.ImageFormat, "Output format for extracted images. Values: png, jpeg. Default: png")
	flags.StringVar(&opts.ImageDir, "image-dir", defaults.ImageDir, "Directory for extracted images")
	flags.StringVar(&opts.Pages, "pages", defaults.Pages, "Pages to extract (e.g., \"1,3,5-7\"). Default: all pages")
	flags.BoolVar(&opts.IncludeHeaderFooter, "include-header-footer", defaults.IncludeHeaderFooter, "Include page headers and footers in output")
	flags.BoolVar(&opts.DetectStrikethrough, "detect-strikethrough", defaults.DetectStrikethrough, "Detect strikethrough text and wrap with ~~ in Markdown output (experimental)")
	flags.StringVar(&opts.Hybrid, "hybrid", defaults.Hybrid, "Hybrid backend for AI processing. Values: off (default), docling, docling-fast, hancom")
	flags.StringVar(&opts.HybridMode, "hybrid-mode", defaults.HybridMode, "Hybrid triage mode. Values: auto (default, dynamic triage), full (skip triage, all pages to backend)")
	flags.StringVar(&opts.HybridURL, "hybrid-url", defaults.HybridURL, "Hybrid backend server URL (overrides default)")
	flags.IntVar(&opts.HybridTimeout, "hybrid-timeout", defaults.HybridTimeout, "Hybrid backend request timeout in milliseconds. Default: 30000")
	flags.BoolVar(&opts.HybridFallback, "hybrid-fallback", defaults.HybridFallback, "Opt in to Java fallback on hybrid backend error (default: disabled)")
	flags.StringVar(&opts.HybridOCR, "hybrid-ocr", "", "[Deprecated] OCR settings are now configured on the hybrid server (--ocr-lang, --force-ocr)")
	flags.BoolVar(&opts.ExportOptions, "export-options", false, "Export machine-readable option metadata")

	flags.BoolVar(&opts.PDFReport, "pdf", false, "Deprecated legacy format flag")
	flags.BoolVar(&opts.MarkdownReport, "markdown", false, "Deprecated legacy format flag")
	flags.BoolVar(&opts.HTMLReport, "html", false, "Deprecated legacy format flag")
	flags.BoolVar(&opts.MarkdownWithHTMLReport, "markdown-with-html", false, "Deprecated legacy format flag")
	flags.BoolVar(&opts.MarkdownWithImages, "markdown-with-images", false, "Deprecated legacy format flag")
	flags.BoolVar(&opts.NoJSONReport, "no-json", false, "Deprecated legacy format flag")

	_ = flags.MarkHidden("export-options")
	_ = flags.MarkHidden("hybrid-ocr")
	_ = flags.MarkHidden("pdf")
	_ = flags.MarkHidden("markdown")
	_ = flags.MarkHidden("html")
	_ = flags.MarkHidden("markdown-with-html")
	_ = flags.MarkHidden("markdown-with-images")
	_ = flags.MarkHidden("no-json")
}

func (o *CLIOptions) ToConfig() (*api.Config, error) {
	return o.toConfig()
}

func (o *CLIOptions) BuildConfig(inputArgs []string) (*api.Config, error) {
	cfg, err := o.toConfig()
	if err != nil {
		return nil, err
	}
	if cfg.OutputDir == "" && len(inputArgs) > 0 {
		cfg.OutputDir = defaultOutputDir(inputArgs[0])
	}
	return cfg, nil
}

func (o *CLIOptions) toConfig() (*api.Config, error) {
	cfg := api.DefaultConfig()
	if o == nil {
		return cfg, nil
	}

	cfg.OutputDir = strings.TrimSpace(o.OutputDir)
	cfg.Password = o.Password
	cfg.Quiet = o.Quiet
	cfg.ContentSafetyOff = normalizeUniqueStrings(o.ContentSafetyOff)
	cfg.Sanitize = o.Sanitize
	cfg.KeepLineBreaks = o.KeepLineBreaks
	if o.ReplaceInvalidChars != "" {
		cfg.ReplaceInvalidChars = o.ReplaceInvalidChars
	}
	cfg.UseStructTree = o.UseStructTree
	cfg.MarkdownPageSeparator = o.MarkdownPageSep
	cfg.TextPageSeparator = o.TextPageSep
	cfg.HTMLPageSeparator = o.HTMLPageSep
	cfg.ImageDir = o.ImageDir
	cfg.IncludeHeaderFooter = o.IncludeHeaderFooter
	cfg.DetectStrikethrough = o.DetectStrikethrough
	cfg.HybridURL = strings.TrimSpace(o.HybridURL)
	if o.HybridTimeout != 0 {
		cfg.HybridTimeout = o.HybridTimeout
	}
	cfg.HybridFallback = o.HybridFallback

	formats, err := o.normalizeFormats(cfg.Formats)
	if err != nil {
		return nil, err
	}
	cfg.Formats = formats

	cfg.TableMethod, err = api.NormalizeTableMethod(o.TableMethod)
	if err != nil {
		return nil, err
	}
	cfg.ReadingOrder, err = api.NormalizeReadingOrder(o.ReadingOrder)
	if err != nil {
		return nil, err
	}
	cfg.ImageOutput, err = api.NormalizeImageOutput(o.ImageOutput)
	if err != nil {
		return nil, err
	}
	cfg.ImageFormat, err = api.NormalizeImageFormat(o.ImageFormat)
	if err != nil {
		return nil, err
	}
	cfg.Hybrid, err = api.NormalizeHybrid(o.Hybrid)
	if err != nil {
		return nil, err
	}
	cfg.HybridMode, err = api.NormalizeHybridMode(o.HybridMode)
	if err != nil {
		return nil, err
	}
	if cfg.HybridTimeout <= 0 {
		return nil, fmt.Errorf("invalid timeout value %d. Must be a positive integer", cfg.HybridTimeout)
	}
	if err := validateContentSafetyFlags(cfg.ContentSafetyOff); err != nil {
		return nil, err
	}
	if err := cfg.SetPages(o.Pages); err != nil {
		return nil, err
	}
	if o.HybridOCR != "" {
		fmt.Fprintln(os.Stderr, "Warning: --hybrid-ocr is deprecated. Configure OCR settings on the hybrid server instead (--ocr-lang, --force-ocr).")
	}

	return cfg, nil
}

func (o *CLIOptions) normalizeFormats(defaultFormats []string) ([]string, error) {
	formats := append([]string(nil), o.Format...)
	if len(formats) == 0 {
		formats = append(formats, defaultFormats...)
	}
	if o.PDFReport {
		formats = append(formats, api.FormatPDF)
	}
	if o.MarkdownReport {
		formats = append(formats, api.FormatMarkdown)
	}
	if o.HTMLReport {
		formats = append(formats, api.FormatHTML)
	}
	if o.MarkdownWithHTMLReport {
		formats = append(formats, api.FormatMarkdownWithHTML)
	}
	if o.MarkdownWithImages {
		formats = append(formats, api.FormatMarkdownWithImages)
	}

	formats = normalizeUniqueStrings(formats)
	if o.NoJSONReport {
		formats = slices.DeleteFunc(formats, func(format string) bool {
			return format == api.FormatJSON
		})
	}
	for _, format := range formats {
		if !isSupportedFormat(format) {
			return nil, fmt.Errorf("unsupported format %q", format)
		}
	}
	return formats, nil
}

func validateContentSafetyFlags(flags []string) error {
	valid := map[string]struct{}{
		api.ContentSafetyAll:        {},
		api.ContentSafetyHiddenText: {},
		api.ContentSafetyOffPage:    {},
		api.ContentSafetyTiny:       {},
		api.ContentSafetyHiddenOCG:  {},
	}
	for _, value := range flags {
		if value == "sensitive-data" {
			fmt.Fprintln(os.Stderr, "Warning: '--content-safety-off sensitive-data' is deprecated and has no effect. Sensitive data sanitization is now opt-in. Use '--sanitize' to enable masking.")
			continue
		}
		if _, ok := valid[value]; !ok {
			return fmt.Errorf("unsupported value %q for --content-safety-off", value)
		}
	}
	return nil
}

func normalizeUniqueStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			trimmed := strings.ToLower(strings.TrimSpace(part))
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			out = append(out, trimmed)
		}
	}
	return out
}

func isSupportedFormat(format string) bool {
	switch format {
	case api.FormatJSON, api.FormatText, api.FormatHTML, api.FormatPDF, api.FormatMarkdown, api.FormatMarkdownWithHTML, api.FormatMarkdownWithImages:
		return true
	default:
		return false
	}
}

func defaultOutputDir(inputPath string) string {
	if strings.TrimSpace(inputPath) == "" {
		return ""
	}
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		absPath = inputPath
	}
	info, statErr := os.Stat(absPath)
	if statErr == nil && info.IsDir() {
		return absPath
	}
	return filepath.Dir(absPath)
}
