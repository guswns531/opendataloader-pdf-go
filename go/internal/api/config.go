// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package api

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	ReadingOrderOff   = "off"
	ReadingOrderXYCut = "xycut"

	PageNumberString = "%page-number%"

	HybridModeAuto = "auto"
	HybridModeFull = "full"

	HybridOff         = "off"
	HybridDocling     = "docling"
	HybridDoclingFast = "docling-fast"
	HybridHancom      = "hancom"
	HybridAzure       = "azure"
	HybridGoogle      = "google"

	TableMethodDefault = "default"
	TableMethodCluster = "cluster"

	ImageFormatPNG  = "png"
	ImageFormatJPEG = "jpeg"

	ImageOutputOff      = "off"
	ImageOutputEmbedded = "embedded"
	ImageOutputExternal = "external"

	ContentSafetyAll        = "all"
	ContentSafetyHiddenText = "hidden-text"
	ContentSafetyOffPage    = "off-page"
	ContentSafetyTiny       = "tiny"
	ContentSafetyHiddenOCG  = "hidden-ocg"

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

	cachedPageNumbers []int
}

func DefaultConfig() *Config {
	return &Config{
		Formats:             []string{FormatJSON},
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

var (
	readingOrderOptions = map[string]struct{}{
		ReadingOrderOff:   {},
		ReadingOrderXYCut: {},
	}
	tableMethodOptions = map[string]struct{}{
		TableMethodDefault: {},
		TableMethodCluster: {},
	}
	imageFormatOptions = map[string]struct{}{
		ImageFormatPNG:  {},
		ImageFormatJPEG: {},
	}
	imageOutputOptions = map[string]struct{}{
		ImageOutputOff:      {},
		ImageOutputEmbedded: {},
		ImageOutputExternal: {},
	}
	hybridOptions = map[string]struct{}{
		HybridOff:         {},
		HybridDocling:     {},
		HybridDoclingFast: {},
		HybridHancom:      {},
	}
	hybridModeOptions = map[string]struct{}{
		HybridModeAuto: {},
		HybridModeFull: {},
	}
)

func NormalizeReadingOrder(readingOrder string) (string, error) {
	return normalizeOption(readingOrder, ReadingOrderXYCut, readingOrderOptions, "reading order")
}

func NormalizeTableMethod(tableMethod string) (string, error) {
	return normalizeOption(tableMethod, TableMethodDefault, tableMethodOptions, "table method")
}

func NormalizeImageOutput(imageOutput string) (string, error) {
	return normalizeOption(imageOutput, ImageOutputExternal, imageOutputOptions, "image output mode")
}

func NormalizeImageFormat(imageFormat string) (string, error) {
	return normalizeOption(imageFormat, ImageFormatPNG, imageFormatOptions, "image format")
}

func NormalizeHybrid(hybrid string) (string, error) {
	value, err := normalizeOption(hybrid, HybridOff, hybridOptions, "hybrid backend")
	if err != nil {
		return "", err
	}
	if value == HybridDocling {
		return HybridDoclingFast, nil
	}
	return value, nil
}

func NormalizeHybridMode(mode string) (string, error) {
	return normalizeOption(mode, HybridModeAuto, hybridModeOptions, "hybrid mode")
}

func IsValidReadingOrder(readingOrder string) bool {
	_, ok := readingOrderOptions[strings.ToLower(strings.TrimSpace(readingOrder))]
	return ok
}

func IsValidTableMethod(tableMethod string) bool {
	_, ok := tableMethodOptions[strings.ToLower(strings.TrimSpace(tableMethod))]
	return ok
}

func IsValidImageOutput(imageOutput string) bool {
	_, ok := imageOutputOptions[strings.ToLower(strings.TrimSpace(imageOutput))]
	return ok
}

func IsValidImageFormat(imageFormat string) bool {
	_, ok := imageFormatOptions[strings.ToLower(strings.TrimSpace(imageFormat))]
	return ok
}

func IsValidHybrid(hybrid string) bool {
	_, ok := hybridOptions[strings.ToLower(strings.TrimSpace(hybrid))]
	return ok
}

func IsValidHybridMode(mode string) bool {
	_, ok := hybridModeOptions[strings.ToLower(strings.TrimSpace(mode))]
	return ok
}

func ParsePageRanges(pages string) ([]int, error) {
	pages = strings.TrimSpace(pages)
	if pages == "" {
		return nil, nil
	}

	result := make([]int, 0)
	for _, part := range strings.Split(pages, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return nil, fmt.Errorf("invalid page range format: %q. Expected format: 1,3,5-7", pages)
		}

		if strings.Contains(trimmed, "-") {
			rangePages, err := parsePageRangePart(trimmed, pages)
			if err != nil {
				return nil, err
			}
			result = append(result, rangePages...)
			continue
		}

		pageNum, err := parsePositivePage(trimmed, pages)
		if err != nil {
			return nil, err
		}
		result = append(result, pageNum)
	}

	return result, nil
}

func (c *Config) SetPages(pages string) error {
	pageNumbers, err := ParsePageRanges(pages)
	if err != nil {
		return err
	}
	c.Pages = pages
	c.cachedPageNumbers = append([]int(nil), pageNumbers...)
	return nil
}

func (c *Config) PageNumbers() []int {
	if c == nil || len(c.cachedPageNumbers) == 0 {
		return []int{}
	}
	return append([]int(nil), c.cachedPageNumbers...)
}

func normalizeOption(value, defaultValue string, options map[string]struct{}, label string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}
	normalized := strings.ToLower(strings.TrimSpace(value))
	if _, ok := options[normalized]; !ok {
		return "", fmt.Errorf("unsupported %s %q", label, value)
	}
	return normalized, nil
}

func parsePageRangePart(value, fullInput string) ([]int, error) {
	bounds := strings.Split(value, "-")
	if len(bounds) != 2 || strings.TrimSpace(bounds[0]) == "" || strings.TrimSpace(bounds[1]) == "" {
		return nil, fmt.Errorf("invalid page range format: %q. Expected format: 1,3,5-7", fullInput)
	}

	start, err := parsePositivePage(bounds[0], fullInput)
	if err != nil {
		return nil, err
	}
	end, err := parsePositivePage(bounds[1], fullInput)
	if err != nil {
		return nil, err
	}
	if start > end {
		return nil, fmt.Errorf("invalid page range %q: start page cannot be greater than end page", value)
	}

	result := make([]int, 0, end-start+1)
	for pageNum := start; pageNum <= end; pageNum++ {
		result = append(result, pageNum)
	}
	return result, nil
}

func parsePositivePage(value, fullInput string) (int, error) {
	pageNum, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("invalid page range format: %q. Expected format: 1,3,5-7", fullInput)
	}
	if pageNum < 1 {
		return 0, fmt.Errorf("page numbers must be positive: %q", fullInput)
	}
	return pageNum, nil
}
