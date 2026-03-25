// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package processors

import (
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

func lineFontSize(line *entities.TextLine) float64 {
	if line == nil || len(line.Chunks) == 0 {
		return 0
	}
	size := 0.0
	for _, chunk := range line.Chunks {
		if chunk != nil && chunk.FontStyle.FontSize > size {
			size = chunk.FontStyle.FontSize
		}
	}
	return size
}

func lineFontWeight(line *entities.TextLine) float64 {
	if line == nil || len(line.Chunks) == 0 {
		return 400
	}
	for _, chunk := range line.Chunks {
		if chunk != nil && chunk.FontStyle.Bold {
			return 700
		}
	}
	return 400
}

func lineIsBold(line *entities.TextLine) bool {
	for _, chunk := range line.Chunks {
		if chunk != nil && chunk.FontStyle.Bold {
			return true
		}
	}
	return false
}

func mergeBoxes(boxes ...entities.BoundingBox) entities.BoundingBox {
	if len(boxes) == 0 {
		return entities.BoundingBox{}
	}
	out := boxes[0]
	for _, b := range boxes[1:] {
		out = unionBox(out, b)
	}
	return out
}

func unionBox(a, b entities.BoundingBox) entities.BoundingBox {
	if a.Width == 0 && a.Height == 0 {
		return b
	}
	if b.Width == 0 && b.Height == 0 {
		return a
	}
	left := minFloat(a.X, b.X)
	bottom := minFloat(a.Y, b.Y)
	right := maxFloat(a.X+a.Width, b.X+b.Width)
	top := maxFloat(a.Y+a.Height, b.Y+b.Height)
	return entities.BoundingBox{X: left, Y: bottom, Width: right - left, Height: top - bottom, Page: a.Page}
}

func linesBBox(lines []*entities.TextLine) entities.BoundingBox {
	if len(lines) == 0 {
		return entities.BoundingBox{}
	}
	box := lines[0].BBox
	for _, line := range lines[1:] {
		box = unionBox(box, line.BBox)
	}
	return box
}

func objectText(obj entities.IObject) string {
	switch v := obj.(type) {
	case *entities.TextLine:
		return strings.TrimSpace(v.GetText())
	case *entities.SemanticParagraph:
		return strings.TrimSpace(joinLines(v.Lines))
	case *entities.SemanticHeading:
		return strings.TrimSpace(joinLines(v.Lines))
	case *entities.SemanticCaption:
		return strings.TrimSpace(v.Text)
	default:
		return ""
	}
}

func joinLines(lines []*entities.TextLine) string {
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		if text := strings.TrimSpace(line.GetText()); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
