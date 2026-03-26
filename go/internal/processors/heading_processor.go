// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package processors

import (
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
	bodyFontSize := 0.0
	fontCounts := map[float64]int{}

	for _, line := range lines {
		if line == nil {
			continue
		}
		size := lineFontSize(line)
		weight := lineFontWeight(line)
		stats.Add(size, weight)
		if size > 0 {
			fontCounts[size]++
			if fontCounts[size] > fontCounts[bodyFontSize] {
				bodyFontSize = size
			}
		}
	}

	headings := make([]*entities.SemanticHeading, 0)
	for _, line := range lines {
		if line == nil {
			continue
		}
		score := p.headingScore(line, bodyFontSize, stats)
		if score <= headingProbabilityThreshold {
			continue
		}
		heading := &entities.SemanticHeading{
			BaseObject: entities.BaseObject{ID: nextID(ctx), BBox: line.BBox},
			Lines:      []*entities.TextLine{line},
			FontSize:   lineFontSize(line),
			IsBold:     lineIsBold(line),
		}
		headings = append(headings, heading)
		if ctx != nil {
			ctx.Headings = append(ctx.Headings, heading)
		}
	}
	return headings
}

func (p *HeadingProcessor) headingScore(line *entities.TextLine, bodyFontSize float64, stats *utils.TextNodeStatistics) float64 {
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

func nextID(ctx *containers.ProcessorContext) string {
	if ctx == nil {
		return ""
	}
	return ctx.NextID()
}
