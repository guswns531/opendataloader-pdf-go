// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package api

const (
	ReadingOrderOff   = "off"
	ReadingOrderXYCut = "xycut"

	PageNumberString = "%page-number%"

	HybridModeAuto = "auto"
	HybridModeFull = "full"

	HybridOff         = "off"
	HybridDoclingFast = "docling-fast"

	TableMethodDefault = "default"
	TableMethodCluster = "cluster"

	ImageFormatPNG  = "png"
	ImageFormatJPEG = "jpeg"

	ImageOutputOff      = "off"
	ImageOutputEmbedded = "embedded"
	ImageOutputExternal = "external"

	FormatJSON               = "json"
	FormatText               = "text"
	FormatHTML               = "html"
	FormatPDF                = "pdf"
	FormatMarkdown           = "markdown"
	FormatMarkdownWithHTML   = "markdown-with-html"
	FormatMarkdownWithImages = "markdown-with-images"
)

type Config struct {
	OutputDir             string
	Password              string
	Formats               []string
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
}

func DefaultConfig() *Config {
	return &Config{
		Formats:             []string{FormatMarkdown},
		ReplaceInvalidChars: " ",
		TableMethod:         TableMethodDefault,
		ReadingOrder:        ReadingOrderXYCut,
		ImageOutput:         ImageOutputExternal,
		ImageFormat:         ImageFormatPNG,
		HybridTimeout:       30000,
		HybridMode:          HybridModeAuto,
		Hybrid:              HybridOff,
	}
}
