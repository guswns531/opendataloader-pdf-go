// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors

import (
	"math"
	"regexp"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const paragraphMergeThreshold = 0.75

var paragraphLabelPattern = regexp.MustCompile(`^(?:[\p{N}\p{L}]+[.)\]])\s+|^(?:[-*+•◦▪‣])\s*`)

type ParagraphProcessor struct{}

func (p *ParagraphProcessor) Process(lines []*entities.TextLine, ctx *containers.ProcessorContext) []*entities.SemanticParagraph {
	filtered := make([]*entities.TextLine, 0, len(lines))
	for _, line := range lines {
		if line != nil && len(line.Chunks) > 0 {
			filtered = append(filtered, line)
		}
	}
	if len(filtered) == 0 {
		return nil
	}

	paragraphs := make([]*entities.SemanticParagraph, 0)
	current := newParagraph(filtered[0], ctx)
	for i := 1; i < len(filtered); i++ {
		line := filtered[i]
		if canMerge(current.Lines[len(current.Lines)-1], line) {
			current.Lines = append(current.Lines, line)
			current.BBox = unionBoundingBox(current.BBox, line.BBox)
		} else {
			finalizeParagraph(current)
			paragraphs = append(paragraphs, current)
			current = newParagraph(line, ctx)
		}
	}
	finalizeParagraph(current)
	paragraphs = append(paragraphs, current)
	return paragraphs
}

func canMerge(line1, line2 *entities.TextLine) bool {
	if line1 == nil || line2 == nil {
		return false
	}
	if isLabeledLine(line2) {
		return false
	}

	probability := mergeProbability(line1, line2)
	return probability >= paragraphMergeThreshold
}

func mergeProbability(line1, line2 *entities.TextLine) float64 {
	font1 := averageFontSize(line1.Chunks)
	font2 := averageFontSize(line2.Chunks)
	avgFont := (font1 + font2) / 2
	if avgFont <= 0 {
		avgFont = 1
	}

	var score float64
	if closeEnough(font1, font2, 0.5) {
		score += 0.2
	}
	if shareComparableWeight(line1, line2) {
		score += 0.1
	}

	gap := line1.BBox.Y - (line2.BBox.Y + line2.BBox.Height)
	if gap < 0 {
		gap = line2.BBox.Y - (line1.BBox.Y + line1.BBox.Height)
	}
	if gap <= avgFont*0.75 {
		score += 0.35
	} else if gap <= avgFont*1.5 {
		score += 0.2
	}

	if horizontalOverlap(line1.BBox, line2.BBox) > 0 {
		score += 0.15
	}

	switch detectPairAlignment(line1, line2) {
	case entities.AlignJustify:
		score += 0.2
	case entities.AlignLeft, entities.AlignRight, entities.AlignCenter:
		score += 0.25
	}

	return score
}

func detectPairAlignment(line1, line2 *entities.TextLine) string {
	tolerance := math.Max(1.5, math.Min(line1.BBox.Height, line2.BBox.Height)*0.5)
	leftAligned := closeEnough(line1.BBox.X, line2.BBox.X, tolerance)
	rightAligned := closeEnough(line1.BBox.X+line1.BBox.Width, line2.BBox.X+line2.BBox.Width, tolerance)
	center1 := line1.BBox.X + line1.BBox.Width/2
	center2 := line2.BBox.X + line2.BBox.Width/2
	centerAligned := closeEnough(center1, center2, tolerance)

	switch {
	case leftAligned && rightAligned:
		return entities.AlignJustify
	case leftAligned:
		return entities.AlignLeft
	case rightAligned:
		return entities.AlignRight
	case centerAligned:
		return entities.AlignCenter
	default:
		return ""
	}
}

func detectParagraphAlignment(lines []*entities.TextLine) string {
	if len(lines) == 0 {
		return entities.AlignLeft
	}
	if len(lines) == 1 {
		return entities.AlignLeft
	}

	leftAligned := true
	rightAligned := true
	centerAligned := true
	base := lines[0]
	tolerance := math.Max(1.5, base.BBox.Height*0.5)
	baseCenter := base.BBox.X + base.BBox.Width/2
	baseRight := base.BBox.X + base.BBox.Width

	for _, line := range lines[1:] {
		if !closeEnough(base.BBox.X, line.BBox.X, tolerance) {
			leftAligned = false
		}
		if !closeEnough(baseRight, line.BBox.X+line.BBox.Width, tolerance) {
			rightAligned = false
		}
		if !closeEnough(baseCenter, line.BBox.X+line.BBox.Width/2, tolerance) {
			centerAligned = false
		}
	}

	switch {
	case leftAligned && rightAligned:
		return entities.AlignJustify
	case leftAligned:
		return entities.AlignLeft
	case rightAligned:
		return entities.AlignRight
	case centerAligned:
		return entities.AlignCenter
	default:
		return entities.AlignLeft
	}
}

func newParagraph(line *entities.TextLine, ctx *containers.ProcessorContext) *entities.SemanticParagraph {
	return &entities.SemanticParagraph{
		BaseObject: entities.BaseObject{
			ID:   nextObjectID(ctx),
			BBox: line.BBox,
		},
		Lines: []*entities.TextLine{line},
	}
}

func finalizeParagraph(paragraph *entities.SemanticParagraph) {
	if len(paragraph.Lines) == 0 {
		return
	}
	paragraph.Alignment = detectParagraphAlignment(paragraph.Lines)
	paragraph.Lines[0].IsFirstLine = true
	paragraph.Lines[len(paragraph.Lines)-1].IsLastLine = true
}

func isLabeledLine(line *entities.TextLine) bool {
	if line == nil {
		return false
	}
	if line.LineArtBullet != nil {
		return true
	}
	return paragraphLabelPattern.MatchString(strings.TrimSpace(line.GetText()))
}

func shareComparableWeight(line1, line2 *entities.TextLine) bool {
	chunk1 := firstVisibleChunk(line1)
	chunk2 := firstVisibleChunk(line2)
	if chunk1 == nil || chunk2 == nil {
		return false
	}

	weight1 := chunk1.FontStyle.FontWeight
	weight2 := chunk2.FontStyle.FontWeight
	if weight1 == 0 && weight2 == 0 {
		return chunk1.FontStyle.Bold == chunk2.FontStyle.Bold
	}
	return closeEnough(weight1, weight2, 0.1)
}

func firstVisibleChunk(line *entities.TextLine) *entities.TextChunk {
	for _, chunk := range line.Chunks {
		if chunk != nil && strings.TrimSpace(chunk.Text) != "" {
			return chunk
		}
	}
	return nil
}

func horizontalOverlap(a, b entities.BoundingBox) float64 {
	left := math.Max(a.X, b.X)
	right := math.Min(a.X+a.Width, b.X+b.Width)
	return right - left
}

func closeEnough(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
