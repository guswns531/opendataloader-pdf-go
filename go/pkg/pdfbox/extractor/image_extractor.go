// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// This package provides functionality equivalent to Apache PDFBox 3.0.4
// (https://pdfbox.apache.org/), implemented using pdfcpu.

package extractor

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

type ExtractedImage struct {
	X            float64
	Y            float64
	Width        float64
	Height       float64
	Data         []byte
	ExternalPath string
	Format       string
	Page         int
}

func ExtractImages(doc *model.PDDocument, pageIdx int, outputDir string) ([]*ExtractedImage, error) {
	page, err := doc.GetPage(pageIdx)
	if err != nil {
		return nil, err
	}

	imagesByObj, err := pdfcpu.ExtractPageImages(doc.Context, page.Number, false)
	if err != nil {
		return nil, err
	}

	resourceImages := map[string]*ExtractedImage{}
	objNrs := make([]int, 0, len(imagesByObj))
	for objNr := range imagesByObj {
		objNrs = append(objNrs, objNr)
	}
	sort.Ints(objNrs)

	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return nil, err
		}
	}

	for _, objNr := range objNrs {
		img := imagesByObj[objNr]
		var data []byte
		if img.Reader != nil {
			data, err = io.ReadAll(img.Reader)
			if err != nil {
				return nil, err
			}
		}

		entry := &ExtractedImage{
			Data:   data,
			Format: normalizeImageFormat(img.FileType),
			Page:   page.Number,
		}

		if outputDir != "" && len(data) > 0 {
			ext := entry.Format
			if ext == "" {
				ext = "bin"
			}
			outPath := filepath.Join(outputDir, fmt.Sprintf("page_%03d_image_%s_%d.%s", page.Number, sanitizeResourceName(img.Name), objNr, ext))
			if err := os.WriteFile(outPath, data, 0o644); err != nil {
				return nil, err
			}
			entry.ExternalPath = outPath
		}

		resourceImages["/"+img.Name] = entry
	}

	content, err := extractPageContentBytes(doc, pageIdx)
	if err != nil {
		return nil, err
	}
	tokens, err := tokenizeContent(content)
	if err != nil {
		return nil, err
	}

	gs := defaultGraphicsState()
	var stack []graphicsState
	var operands []streamToken
	var out []*ExtractedImage

	appendImage := func(name string) {
		img, ok := resourceImages[name]
		if !ok {
			return
		}
		x0, y0 := gs.ctm.transform(0, 0)
		x1, y1 := gs.ctm.transform(1, 0)
		x2, y2 := gs.ctm.transform(0, 1)
		x3, y3 := gs.ctm.transform(1, 1)
		minX := math.Min(math.Min(x0, x1), math.Min(x2, x3))
		maxX := math.Max(math.Max(x0, x1), math.Max(x2, x3))
		minY := math.Min(math.Min(y0, y1), math.Min(y2, y3))
		maxY := math.Max(math.Max(y0, y1), math.Max(y2, y3))

		copied := *img
		copied.X = minX
		copied.Y = minY
		copied.Width = maxX - minX
		copied.Height = maxY - minY
		out = append(out, &copied)
	}

	for _, tok := range tokens {
		if tok.kind != "word" {
			operands = append(operands, tok)
			continue
		}

		switch tok.value {
		case "q":
			stack = append(stack, gs)
		case "Q":
			if len(stack) > 0 {
				gs = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}
		case "cm":
			if m, ok := matrixFromOperands(operands); ok {
				gs.ctm = gs.ctm.multiply(m)
			}
		case "Do":
			if len(operands) == 1 && operands[0].kind == "name" {
				appendImage(operands[0].value)
			}
		default:
		}

		operands = operands[:0]
	}

	if out == nil {
		return []*ExtractedImage{}, nil
	}
	return out, nil
}

func normalizeImageFormat(v string) string {
	v = strings.ToLower(strings.TrimPrefix(v, "."))
	switch v {
	case "jpg":
		return "jpeg"
	default:
		return v
	}
}

func sanitizeResourceName(v string) string {
	v = strings.TrimPrefix(v, "/")
	if v == "" {
		return "image"
	}
	var b strings.Builder
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}
