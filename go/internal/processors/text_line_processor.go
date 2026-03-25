// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors

import (
	"math"
	"sort"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const (
	textLineBaselineTolerance = 0.5
	textLineSpaceRatio        = 0.33
	listLabelHeightEpsilon    = 1.5
)

type TextLineProcessor struct{}

func (p *TextLineProcessor) Process(
	chunks []*entities.TextChunk,
	lineArts []*entities.LineArtChunk,
	ctx *containers.ProcessorContext,
) []*entities.TextLine {
	lines := groupTextChunksIntoLines(chunks, ctx)
	for _, line := range lines {
		sort.Slice(line.Chunks, func(i, j int) bool {
			return line.Chunks[i].BBox.X < line.Chunks[j].BBox.X
		})
		rebuildLineGeometry(line)
	}

	sort.Slice(lines, func(i, j int) bool {
		if math.Abs(lines[i].Baseline-lines[j].Baseline) > textLineBaselineTolerance {
			return lines[i].Baseline > lines[j].Baseline
		}
		return lines[i].BBox.X < lines[j].BBox.X
	})
	linkTextLinesWithConnectedLineArt(lines, lineArts)
	return lines
}

func groupTextChunksIntoLines(chunks []*entities.TextChunk, ctx *containers.ProcessorContext) []*entities.TextLine {
	sorted := make([]*entities.TextChunk, 0, len(chunks))
	for _, chunk := range chunks {
		if chunk == nil || chunk.Text == "" {
			continue
		}
		sorted = append(sorted, chunk)
	}

	sort.Slice(sorted, func(i, j int) bool {
		if math.Abs(sorted[i].Baseline-sorted[j].Baseline) > textLineBaselineTolerance {
			return sorted[i].Baseline > sorted[j].Baseline
		}
		return sorted[i].BBox.X < sorted[j].BBox.X
	})

	lines := make([]*entities.TextLine, 0)
	for _, chunk := range sorted {
		var target *entities.TextLine
		for _, line := range lines {
			if math.Abs(line.Baseline-chunk.Baseline) <= textLineBaselineTolerance {
				target = line
				break
			}
		}
		if target == nil {
			target = &entities.TextLine{
				BaseObject: entities.BaseObject{ID: nextObjectID(ctx), BBox: chunk.BBox},
				Chunks:     []*entities.TextChunk{},
				Baseline:   chunk.Baseline,
			}
			lines = append(lines, target)
		}
		target.Chunks = append(target.Chunks, chunk)
		rebuildLineGeometry(target)
	}

	return lines
}

func addSyntheticSpaces(chunks []*entities.TextChunk, fontSize float64) []*entities.TextChunk {
	if len(chunks) <= 1 {
		return chunks
	}

	threshold := fontSize * textLineSpaceRatio
	if threshold <= 0 {
		threshold = textLineBaselineTolerance
	}

	result := make([]*entities.TextChunk, 0, len(chunks)*2)
	result = append(result, chunks[0])
	previousEnd := chunks[0].BBox.X + chunks[0].BBox.Width
	for i := 1; i < len(chunks); i++ {
		current := chunks[i]
		currentStart := current.BBox.X
		if currentStart-previousEnd > threshold {
			spaceBox := entities.BoundingBox{
				X:      previousEnd,
				Y:      current.BBox.Y,
				Width:  currentStart - previousEnd,
				Height: current.BBox.Height,
				Page:   current.BBox.Page,
			}
			result = append(result, &entities.TextChunk{
				BaseObject: entities.BaseObject{BBox: spaceBox},
				Text:       " ",
				FontStyle:  entities.FontStyle{FontSize: fontSize},
				Baseline:   current.Baseline,
			})
		}
		result = append(result, current)
		previousEnd = current.BBox.X + current.BBox.Width
	}

	return result
}

func linkTextLinesWithConnectedLineArt(lines []*entities.TextLine, lineArts []*entities.LineArtChunk) {
	if len(lines) == 0 || len(lineArts) == 0 {
		return
	}

	sort.Slice(lineArts, func(i, j int) bool {
		lineI := lineArts[i].BBox.Y + lineArts[i].BBox.Height
		lineJ := lineArts[j].BBox.Y + lineArts[j].BBox.Height
		if math.Abs(lineI-lineJ) > textLineBaselineTolerance {
			return lineI > lineJ
		}
		return lineArts[i].BBox.X < lineArts[j].BBox.X
	})

	for _, line := range lines {
		for _, art := range lineArts {
			if isLineConnectedWithLineArt(line, art) {
				line.LineArtBullet = art
				break
			}
		}
	}
}

func isLineConnectedWithLineArt(line *entities.TextLine, lineArt *entities.LineArtChunk) bool {
	lineHeight := line.BBox.Height
	if lineHeight <= 0 {
		return false
	}

	artRight := lineArt.BBox.X + lineArt.BBox.Width
	lineMidY := line.BBox.Y + line.BBox.Height/2
	artMidY := lineArt.BBox.Y + lineArt.BBox.Height/2

	return artRight <= line.BBox.X &&
		math.Abs(lineMidY-artMidY) <= lineHeight &&
		lineArt.BBox.Height < listLabelHeightEpsilon*lineHeight
}

func rebuildLineGeometry(line *entities.TextLine) {
	if len(line.Chunks) == 0 {
		return
	}

	box := line.Chunks[0].BBox
	hasStrikethrough := false
	for _, chunk := range line.Chunks {
		box = unionBoundingBox(box, chunk.BBox)
		if chunk.IsStrikethrough {
			hasStrikethrough = true
		}
	}
	line.BBox = box
	line.HasStrikethrough = hasStrikethrough
}

func averageFontSize(chunks []*entities.TextChunk) float64 {
	if len(chunks) == 0 {
		return 0
	}

	var sum float64
	for _, chunk := range chunks {
		sum += chunk.FontStyle.FontSize
	}
	return sum / float64(len(chunks))
}

func unionBoundingBox(a, b entities.BoundingBox) entities.BoundingBox {
	left := math.Min(a.X, b.X)
	bottom := math.Min(a.Y, b.Y)
	right := math.Max(a.X+a.Width, b.X+b.Width)
	top := math.Max(a.Y+a.Height, b.Y+b.Height)
	return entities.BoundingBox{
		X:      left,
		Y:      bottom,
		Width:  right - left,
		Height: top - bottom,
		Page:   a.Page,
	}
}

func nextObjectID(ctx *containers.ProcessorContext) string {
	if ctx == nil {
		return ""
	}
	return ctx.NextID()
}
