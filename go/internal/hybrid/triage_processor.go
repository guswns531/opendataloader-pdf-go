// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid

import (
	"math"
	"sort"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const (
	LineRatioThreshold        = 0.3
	AlignedLineGroupsMin      = 5
	GridGapMultiplier         = 3.0
	MinLineCountForTable      = 8
	MinGridLines              = 3
	MinRowSeparatorPatterns   = 5
	MinAlignedShortLines      = 2
	MinConsecutivePatterns    = 2
	LargeImageRatio           = 0.11
	ImageAspectRatioThreshold = 1.75
)

const (
	baselineEpsilon        = 0.1
	lineLengthTolerance    = 0.05
	highPatternCount       = 30
	minTablePatterns       = 3
	minPatternDensity      = 0.10
	minPatternsForDensity  = 2
	multiColumnXShiftRatio = 2.0
	xDifferenceEpsilon     = 1.5
)

type TriageSignals struct {
	LineChunkCount         int
	TextChunkCount         int
	LineToTextRatio        float64
	AlignedLineGroups      int
	HasTableBorder         bool
	HasSuspiciousPattern   bool
	HorizontalLineCount    int
	VerticalLineCount      int
	LineArtCount           int
	HasGridLines           bool
	HasTableBorderLines    bool
	HasRowSeparatorPattern bool
	HasAlignedShortLines   bool
	TablePatternCount      int
	MaxConsecutiveStreak   int
	PatternDensity         float64
	HasConsecutivePatterns bool
	LargeImageRatio        float64
	LargeImageAspectRatio  float64
}

func (s TriageSignals) hasVectorTableSignal() bool {
	return s.HasGridLines || s.HasTableBorderLines || s.LineArtCount >= MinLineCountForTable ||
		s.HasRowSeparatorPattern || s.HasAlignedShortLines
}

func (s TriageSignals) hasTextTablePattern() bool {
	hasHighPatternCount := s.TablePatternCount >= highPatternCount
	meetsPatternThreshold := s.TablePatternCount >= minTablePatterns ||
		(s.PatternDensity >= minPatternDensity && s.TablePatternCount >= minPatternsForDensity)
	return (s.HasConsecutivePatterns || hasHighPatternCount) && meetsPatternThreshold
}

func (s TriageSignals) hasLargeImage() bool {
	return s.LargeImageRatio >= LargeImageRatio && s.LargeImageAspectRatio >= ImageAspectRatioThreshold
}

type TriageResult struct {
	Decision   string
	Confidence float64
	Signals    TriageSignals
}

type TriageProcessor struct{}

func (p *TriageProcessor) Triage(page *entities.Page) *TriageResult {
	signals := extractSignals(page)

	switch {
	case signals.HasTableBorder:
		return &TriageResult{Decision: TriageDecisionBackend, Confidence: 1.0, Signals: signals}
	case signals.hasVectorTableSignal():
		return &TriageResult{Decision: TriageDecisionBackend, Confidence: 0.95, Signals: signals}
	case signals.hasTextTablePattern():
		return &TriageResult{Decision: TriageDecisionBackend, Confidence: 0.9, Signals: signals}
	case signals.hasLargeImage():
		return &TriageResult{Decision: TriageDecisionBackend, Confidence: 0.85, Signals: signals}
	case signals.LineToTextRatio > LineRatioThreshold:
		return &TriageResult{Decision: TriageDecisionBackend, Confidence: 0.8, Signals: signals}
	default:
		return &TriageResult{Decision: TriageDecisionJava, Confidence: 0.9, Signals: signals}
	}
}

func extractSignals(page *entities.Page) TriageSignals {
	if page == nil {
		return TriageSignals{}
	}

	acc := signalAccumulator{}
	for _, chunk := range page.Chunks {
		acc.processTextChunk(chunk)
	}
	for _, line := range page.LineArts {
		acc.processLineArtChunk(line)
	}
	for _, element := range page.Elements {
		acc.processElement(element, page)
	}

	totalCount := len(page.Chunks) + len(page.LineArts) + len(page.Elements)
	lineToTextRatio := 0.0
	if totalCount > 0 {
		lineToTextRatio = float64(acc.lineChunkCount) / float64(totalCount)
	}

	patternDensity := 0.0
	if acc.nonWhitespaceTextCount > 0 {
		patternDensity = float64(acc.tablePatternCount) / float64(acc.nonWhitespaceTextCount)
	}

	signals := TriageSignals{
		LineChunkCount:         acc.lineChunkCount,
		TextChunkCount:         acc.textChunkCount,
		LineToTextRatio:        lineToTextRatio,
		AlignedLineGroups:      countAlignedLineGroups(acc.textChunks, GridGapMultiplier),
		HasTableBorder:         acc.hasTableBorder,
		HasSuspiciousPattern:   checkSuspiciousPatterns(acc.textChunks),
		HorizontalLineCount:    acc.horizontalLineCount,
		VerticalLineCount:      acc.verticalLineCount,
		LineArtCount:           acc.lineArtCount,
		HasGridLines:           acc.horizontalLineCount >= MinGridLines && acc.verticalLineCount >= MinGridLines,
		HasTableBorderLines:    acc.horizontalLineCount+acc.verticalLineCount >= MinLineCountForTable,
		HasRowSeparatorPattern: acc.rowSeparatorPatternCount >= MinRowSeparatorPatterns,
		HasAlignedShortLines:   acc.hasAlignedShortHorizontalLines(),
		TablePatternCount:      acc.tablePatternCount,
		MaxConsecutiveStreak:   acc.maxConsecutiveStreak,
		PatternDensity:         patternDensity,
		HasConsecutivePatterns: acc.maxConsecutiveStreak >= MinConsecutivePatterns,
		LargeImageAspectRatio:  acc.maxImageAspectRatio,
	}

	pageArea := page.Width * page.Height
	if pageArea > 0 && acc.maxImageArea > 0 {
		signals.LargeImageRatio = acc.maxImageArea / pageArea
	}

	return signals
}

type signalAccumulator struct {
	lineChunkCount           int
	textChunkCount           int
	nonWhitespaceTextCount   int
	horizontalLineCount      int
	verticalLineCount        int
	lineArtCount             int
	tablePatternCount        int
	currentConsecutiveStreak int
	maxConsecutiveStreak     int
	rowSeparatorPatternCount int
	lastWasHorizontalLine    bool
	previousTextChunk        *entities.TextChunk
	textChunks               []*entities.TextChunk
	shortHorizontalLines     [][2]float64
	maxImageArea             float64
	maxImageAspectRatio      float64
	hasTableBorder           bool
}

func (a *signalAccumulator) processTextChunk(chunk *entities.TextChunk) {
	if chunk == nil {
		return
	}
	a.textChunkCount++
	a.textChunks = append(a.textChunks, chunk)
	if strings.TrimSpace(chunk.Text) == "" {
		return
	}

	a.nonWhitespaceTextCount++
	a.lastWasHorizontalLine = false

	if a.previousTextChunk != nil {
		if areSuspiciousTextChunks(a.previousTextChunk, chunk) {
			a.tablePatternCount++
			a.currentConsecutiveStreak++
			if a.currentConsecutiveStreak > a.maxConsecutiveStreak {
				a.maxConsecutiveStreak = a.currentConsecutiveStreak
			}
		} else {
			a.currentConsecutiveStreak = 0
		}
	}
	a.previousTextChunk = chunk
}

func (a *signalAccumulator) processLineArtChunk(line *entities.LineArtChunk) {
	if line == nil {
		return
	}
	a.lineChunkCount++
	a.lineArtCount++
	left, bottom, right, top := bboxEdges(line.BBox)
	width := right - left
	height := top - bottom
	if line.IsHorizontal || width > height*3 {
		a.horizontalLineCount++
		if !a.lastWasHorizontalLine {
			a.rowSeparatorPatternCount++
		}
		a.shortHorizontalLines = append(a.shortHorizontalLines, [2]float64{left, width})
		a.lastWasHorizontalLine = true
		return
	}
	if line.IsVertical || height > width*3 {
		a.verticalLineCount++
	}
}

func (a *signalAccumulator) processElement(element entities.IObject, page *entities.Page) {
	switch v := element.(type) {
	case *entities.SemanticTable:
		a.hasTableBorder = true
	case *entities.SemanticImage:
		a.processImageBBox(v.BBox, page)
	}
}

func (a *signalAccumulator) processImageBBox(box entities.BoundingBox, page *entities.Page) {
	left, bottom, right, top := bboxEdges(box)
	width := right - left
	height := top - bottom
	area := width * height
	if area > a.maxImageArea {
		a.maxImageArea = area
		if height > 0 {
			a.maxImageAspectRatio = width / height
		} else {
			a.maxImageAspectRatio = 0
		}
	}
	_ = page
}

func (a *signalAccumulator) hasAlignedShortHorizontalLines() bool {
	if len(a.shortHorizontalLines) < MinAlignedShortLines {
		return false
	}
	for i := range a.shortHorizontalLines {
		ref := a.shortHorizontalLines[i]
		matchCount := 1
		for j := i + 1; j < len(a.shortHorizontalLines); j++ {
			cur := a.shortHorizontalLines[j]
			maxLen := math.Max(ref[1], cur[1])
			if maxLen == 0 {
				continue
			}
			xMatches := math.Abs(ref[0]-cur[0])/maxLen <= lineLengthTolerance
			lenMatches := math.Abs(ref[1]-cur[1])/maxLen <= lineLengthTolerance
			if xMatches && lenMatches {
				matchCount++
				if matchCount >= MinAlignedShortLines {
					return true
				}
			}
		}
	}
	return false
}

func checkSuspiciousPatterns(chunks []*entities.TextChunk) bool {
	if len(chunks) < 2 {
		return false
	}
	var previous *entities.TextChunk
	for _, current := range chunks {
		if current == nil || strings.TrimSpace(current.Text) == "" {
			continue
		}
		if previous != nil && areOnSameBaseline(previous, current) {
			gap := bboxLeft(current.BBox) - bboxRight(previous.BBox)
			avgHeight := (current.BBox.Height + previous.BBox.Height) / 2
			if gap > avgHeight*3.0 {
				return true
			}
		}
		previous = current
	}
	return false
}

func areSuspiciousTextChunks(previous, current *entities.TextChunk) bool {
	if bboxTop(previous.BBox) < bboxBottom(current.BBox) {
		xShift := bboxLeft(previous.BBox) - bboxLeft(current.BBox)
		textWidth := bboxRight(previous.BBox) - bboxLeft(previous.BBox)
		if textWidth > 0 && xShift > textWidth*multiColumnXShiftRatio {
			return false
		}
		return true
	}
	if areOnSameBaseline(previous, current) {
		return bboxLeft(current.BBox)-bboxRight(previous.BBox) > current.BBox.Height*xDifferenceEpsilon
	}
	return false
}

func areOnSameBaseline(a, b *entities.TextChunk) bool {
	avgHeight := (a.BBox.Height + b.BBox.Height) / 2
	return math.Abs(a.Baseline-b.Baseline) < avgHeight*baselineEpsilon
}

func countAlignedLineGroups(chunks []*entities.TextChunk, gapMultiplier float64) int {
	if len(chunks) < 2 {
		return 0
	}

	groups := map[float64][]*entities.TextChunk{}
	for _, chunk := range chunks {
		if chunk == nil || strings.TrimSpace(chunk.Text) == "" {
			continue
		}
		rounded := math.Round(chunk.Baseline*10.0) / 10.0
		var matched *float64
		for key := range groups {
			if math.Abs(key-rounded) < chunk.BBox.Height*baselineEpsilon {
				v := key
				matched = &v
				break
			}
		}
		if matched != nil {
			groups[*matched] = append(groups[*matched], chunk)
		} else {
			groups[rounded] = append(groups[rounded], chunk)
		}
	}

	aligned := 0
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		sort.Slice(group, func(i, j int) bool {
			return bboxLeft(group[i].BBox) < bboxLeft(group[j].BBox)
		})
		hasLargeGap := false
		for i := 1; i < len(group); i++ {
			prev := group[i-1]
			curr := group[i]
			gap := bboxLeft(curr.BBox) - bboxRight(prev.BBox)
			avgHeight := (prev.BBox.Height + curr.BBox.Height) / 2
			if gap > avgHeight*gapMultiplier {
				hasLargeGap = true
				break
			}
		}
		if hasLargeGap {
			aligned++
		}
	}
	return aligned
}

func bboxLeft(box entities.BoundingBox) float64   { return math.Min(box.X, box.X+box.Width) }
func bboxRight(box entities.BoundingBox) float64  { return math.Max(box.X, box.X+box.Width) }
func bboxBottom(box entities.BoundingBox) float64 { return math.Min(box.Y, box.Y+box.Height) }
func bboxTop(box entities.BoundingBox) float64    { return math.Max(box.Y, box.Y+box.Height) }

func bboxEdges(box entities.BoundingBox) (float64, float64, float64, float64) {
	return bboxLeft(box), bboxBottom(box), bboxRight(box), bboxTop(box)
}
