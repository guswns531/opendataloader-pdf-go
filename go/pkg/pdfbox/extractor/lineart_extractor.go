// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// This package provides functionality equivalent to Apache PDFBox 3.0.4
// (https://pdfbox.apache.org/), implemented using pdfcpu.

package extractor

import (
	"math"

	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/model"
)

type ExtractedLineArt struct {
	X            float64
	Y            float64
	Width        float64
	Height       float64
	IsHorizontal bool
	IsVertical   bool
	LineWidth    float64
	Page         int
}

type pathSegment struct {
	x1 float64
	y1 float64
	x2 float64
	y2 float64
}

type rectPath struct {
	x float64
	y float64
	w float64
	h float64
}

func ExtractLineArts(doc *model.PDDocument, pageIdx int) ([]*ExtractedLineArt, error) {
	page, err := doc.GetPage(pageIdx)
	if err != nil {
		return nil, err
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
	var out []*ExtractedLineArt
	var path []pathSegment
	var rects []rectPath
	var currentX float64
	var currentY float64
	var haveCurrent bool

	flushPath := func() {
		for _, seg := range path {
			dx := math.Abs(seg.x2 - seg.x1)
			dy := math.Abs(seg.y2 - seg.y1)
			if dy < 2.0 && dx > 5.0 {
				out = append(out, &ExtractedLineArt{
					X:            math.Min(seg.x1, seg.x2),
					Y:            math.Min(seg.y1, seg.y2),
					Width:        dx,
					Height:       dy,
					IsHorizontal: true,
					LineWidth:    gs.lineWidth,
					Page:         page.Number,
				})
			} else if dx < 2.0 && dy > 5.0 {
				out = append(out, &ExtractedLineArt{
					X:          math.Min(seg.x1, seg.x2),
					Y:          math.Min(seg.y1, seg.y2),
					Width:      dx,
					Height:     dy,
					IsVertical: true,
					LineWidth:  gs.lineWidth,
					Page:       page.Number,
				})
			}
		}
		for _, rect := range rects {
			if rect.w > rect.h && rect.w > 5.0 {
				out = append(out, &ExtractedLineArt{
					X:            rect.x,
					Y:            rect.y,
					Width:        rect.w,
					Height:       rect.h,
					IsHorizontal: true,
					LineWidth:    gs.lineWidth,
					Page:         page.Number,
				})
			} else if rect.h > rect.w && rect.h > 5.0 {
				out = append(out, &ExtractedLineArt{
					X:          rect.x,
					Y:          rect.y,
					Width:      rect.w,
					Height:     rect.h,
					IsVertical: true,
					LineWidth:  gs.lineWidth,
					Page:       page.Number,
				})
			}
		}
		path = nil
		rects = nil
		haveCurrent = false
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
		case "w":
			if len(operands) == 1 {
				if v, ok := parseFloatToken(operands[0]); ok {
					gs.lineWidth = v
				}
			}
		case "m":
			if values, ok := parseFloats(operands); ok && len(values) == 2 {
				currentX, currentY = gs.ctm.transform(values[0], values[1])
				haveCurrent = true
			}
		case "l":
			if values, ok := parseFloats(operands); ok && len(values) == 2 && haveCurrent {
				x, y := gs.ctm.transform(values[0], values[1])
				path = append(path, pathSegment{x1: currentX, y1: currentY, x2: x, y2: y})
				currentX, currentY = x, y
			}
		case "re":
			if values, ok := parseFloats(operands); ok && len(values) == 4 {
				x0, y0 := gs.ctm.transform(values[0], values[1])
				x1, y1 := gs.ctm.transform(values[0]+values[2], values[1]+values[3])
				rects = append(rects, rectPath{
					x: math.Min(x0, x1),
					y: math.Min(y0, y1),
					w: math.Abs(x1 - x0),
					h: math.Abs(y1 - y0),
				})
			}
		case "S", "s":
			flushPath()
		case "n", "f", "F", "f*", "B", "B*", "b", "b*":
			path = nil
			rects = nil
			haveCurrent = false
		default:
		}

		operands = operands[:0]
	}

	if out == nil {
		return []*ExtractedLineArt{}, nil
	}
	return out, nil
}
