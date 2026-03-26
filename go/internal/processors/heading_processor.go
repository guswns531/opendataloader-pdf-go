// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package processors

import (
	"math"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/utils"
)

const (
	headingProbabilityThreshold        = 0.75
	bulletedHeadingProbabilityIncrease = 0.10
)

type HeadingProcessor struct{}

func (p *HeadingProcessor) Process(lines []*entities.TextLine, ctx *containers.ProcessorContext) []*entities.SemanticHeading {
	stats := utils.NewTextNodeStatistics()
	for _, line := range lines {
		if line == nil {
			continue
		}
		size := lineFontSize(line)
		weight := lineFontWeight(line)
		stats.Add(size, weight)
	}
	bodyFontSize := stats.FontSizeMode()

	headings := make([]*entities.SemanticHeading, 0)
	for idx, line := range lines {
		if line == nil {
			continue
		}
		var prevLine, nextLine *entities.TextLine
		if idx > 0 {
			prevLine = lines[idx-1]
		}
		if idx+1 < len(lines) {
			nextLine = lines[idx+1]
		}
		score := p.headingScore(line, prevLine, nextLine, bodyFontSize, stats)
		if score <= headingProbabilityThreshold {
			continue
		}
		heading := &entities.SemanticHeading{
			BaseObject: entities.BaseObject{ID: nextID(ctx), BBox: line.BBox},
			Lines:      []*entities.TextLine{line},
			FontSize:   lineFontSize(line),
			IsBold:     lineIsBold(line),
			IsItalic:   lineIsItalic(line),
			FontFamily: lineFontFamily(line),
		}
		headings = append(headings, heading)
		if ctx != nil {
			ctx.Headings = append(ctx.Headings, heading)
		}
	}
	return headings
}

func (p *HeadingProcessor) headingScore(line, prevLine, nextLine *entities.TextLine, bodyFontSize float64, stats *utils.TextNodeStatistics) float64 {
	text := strings.TrimSpace(line.GetText())
	if text == "" {
		return 0
	}
	if len([]rune(text)) >= 120 {
		return 0
	}

	size := lineFontSize(line)
	weight := lineFontWeight(line)
	if bodyFontSize > 0 && size <= bodyFontSize*1.1 && !lineIsBold(line) {
		return 0
	}
	score := 0.0

	if bodyFontSize > 0 && size > bodyFontSize*1.1 {
		score += min(0.45, (size-bodyFontSize)/max(bodyFontSize, 1.0)+0.2)
	}
	score += stats.FontSizeRarityBoost(size)
	score += stats.FontWeightRarityBoost(weight)

	if lineIsBold(line) {
		score += 0.2
	}
	if isIsolatedHeadingLine(line, prevLine, nextLine) {
		score += 0.22
	}
	if line.BBox.Page == 0 && line.BBox.Y+line.BBox.Height >= 700 {
		score += 0.08
	}
	length := len([]rune(text))
	switch {
	case length <= 8:
		score += 0.25
	case length <= 30:
		score += 0.18
	case length <= 80:
		score += 0.05
	default:
		score -= 0.15
	}

	if utils.IsOrderedBullet(text) || utils.IsUnorderedBullet(text) {
		score += bulletedHeadingProbabilityIncrease
	}
	if strings.HasSuffix(text, ".") || strings.HasSuffix(text, ",") {
		score -= 0.1
	}
	if strings.Count(text, " ") > 12 {
		score -= 0.1
	}
	return score
}

func isIsolatedHeadingLine(line, prevLine, nextLine *entities.TextLine) bool {
	if line == nil {
		return false
	}
	threshold := math.Max(line.BBox.Height*0.6, 8)
	prevGapOK := prevLine == nil || prevLine.BBox.Y-line.BBox.Y >= threshold || !sameColumn(prevLine, line)
	nextGapOK := nextLine == nil || line.BBox.Y-nextLine.BBox.Y >= threshold || !sameColumn(nextLine, line)
	return prevGapOK && nextGapOK
}

func sameColumn(a, b *entities.TextLine) bool {
	if a == nil || b == nil {
		return false
	}
	centerA := a.BBox.X + a.BBox.Width/2
	centerB := b.BBox.X + b.BBox.Width/2
	halfWidth := math.Max(math.Min(a.BBox.Width, b.BBox.Width)/2, 10)
	return math.Abs(centerA-centerB) <= halfWidth
}

func nextID(ctx *containers.ProcessorContext) string {
	if ctx == nil {
		return ""
	}
	return ctx.NextID()
}
