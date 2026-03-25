// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package processors

import "github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"

const (
	verticalCenterTolerance    = 0.2
	minHorizontalOverlapRatio  = 0.8
	maxLineToTextWidthRatio    = 1.5
	maxStrokeToTextHeightRatio = 1.3
)

type StrikethroughProcessor struct{}

func (p *StrikethroughProcessor) Process(lines []*entities.TextLine, lineArts []*entities.LineArtChunk) []*entities.TextLine {
	if len(lines) == 0 || len(lineArts) == 0 {
		return lines
	}
	for _, lineArt := range lineArts {
		if lineArt == nil || !lineArt.IsHorizontal {
			continue
		}
		for _, line := range lines {
			if line == nil || !isStrikethroughLine(lineArt, line) {
				continue
			}
			line.HasStrikethrough = true
			for _, chunk := range line.Chunks {
				if chunk != nil {
					chunk.IsStrikethrough = true
				}
			}
		}
	}
	return lines
}

func isStrikethroughLine(lineArt *entities.LineArtChunk, line *entities.TextLine) bool {
	textBox := line.BBox
	textHeight := textBox.Height
	if textHeight <= 0 {
		return false
	}
	if lineArt.LineWidth/textHeight > maxStrokeToTextHeightRatio {
		return false
	}

	lineY := lineArt.BBox.Y + lineArt.BBox.Height/2
	textCenterY := textBox.Y + textBox.Height/2
	if abs(lineY-textCenterY) > textHeight*verticalCenterTolerance {
		return false
	}

	lineLeft := lineArt.BBox.X
	lineRight := lineArt.BBox.X + lineArt.BBox.Width
	textLeft := textBox.X
	textRight := textBox.X + textBox.Width
	overlap := min(lineRight, textRight) - max(lineLeft, textLeft)
	if overlap <= 0 || textBox.Width <= 0 {
		return false
	}
	if overlap/textBox.Width < minHorizontalOverlapRatio {
		return false
	}
	if lineArt.BBox.Width/textBox.Width > maxLineToTextWidthRatio {
		return false
	}
	return true
}
