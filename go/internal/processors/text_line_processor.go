// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors

import (
	"math"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const (
	textLineBaselineTolerance = 0.5
	textLineSpaceRatio        = 0.3
	textLineTabRatio          = 2.0
	textLineWordBoundaryRatio = 0.05
	textLineDenseProseRatio   = 0.03
	textLineDisplayTextRatio  = 0.02
	textLineWordOverlapRatio  = 0.08
	textLineDisplayFontMin    = 14.0
	textLineGutterGapRatio    = 3.0
	textLineNarrowWidthRatio  = 0.35
	listLabelHeightEpsilon    = 1.5
	lineArtBulletGapMax       = 20.0
	lineArtBaselineTolerance  = 5.0
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
		line.Chunks = addSyntheticSpacing(line.Chunks)
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
		if chunk == nil || strings.TrimSpace(chunk.Text) == "" {
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
			if math.Abs(line.Baseline-chunk.Baseline) <= textLineBaselineTolerance &&
				canMergeChunkIntoLine(line, chunk) {
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

func canMergeChunkIntoLine(line *entities.TextLine, chunk *entities.TextChunk) bool {
	if line == nil || chunk == nil || len(line.Chunks) == 0 {
		return true
	}

	last := line.Chunks[len(line.Chunks)-1]
	if last == nil {
		return true
	}

	gap := chunk.BBox.X - (last.BBox.X + last.BBox.Width)
	if gap <= 0 {
		return true
	}

	height := math.Max(chunk.BBox.Height, line.BBox.Height)
	if height <= 0 || gap <= height*textLineGutterGapRatio {
		return true
	}

	lineWidth := line.BBox.Width
	chunkWidth := chunk.BBox.Width
	if lineWidth <= 0 || chunkWidth <= 0 {
		return true
	}

	narrower := math.Min(lineWidth, chunkWidth)
	wider := math.Max(lineWidth, chunkWidth)
	return narrower/wider > textLineNarrowWidthRatio
}

func addSyntheticSpacing(chunks []*entities.TextChunk) []*entities.TextChunk {
	if len(chunks) <= 1 {
		return chunks
	}

	avgCharWidth := averageCharWidth(chunks)
	if avgCharWidth <= 0 {
		avgCharWidth = textLineBaselineTolerance
	}
	avgFontSize := averageFontSize(chunks)
	spaceThreshold := avgCharWidth * textLineSpaceRatio
	tabThreshold := avgCharWidth * textLineTabRatio

	result := make([]*entities.TextChunk, 0, len(chunks)*2)
	result = append(result, chunks[0])
	previousEnd := chunks[0].BBox.X + chunks[0].BBox.Width
	for i := 1; i < len(chunks); i++ {
		previous := chunks[i-1]
		current := chunks[i]
		currentStart := current.BBox.X
		gap := currentStart - previousEnd
		if shouldInsertSyntheticSpaceBeforeSingletonContinuation(chunks, i, previousEnd, avgCharWidth, spaceThreshold) ||
			shouldInsertSyntheticSpace(previous, current, gap, avgCharWidth, avgFontSize, spaceThreshold) {
			spacingText := " "
			if gap > tabThreshold {
				spacingText = "\t"
			}
			spacingBox := entities.BoundingBox{
				X:      previousEnd,
				Y:      current.BBox.Y,
				Width:  gap,
				Height: current.BBox.Height,
				Page:   current.BBox.Page,
			}
			result = append(result, &entities.TextChunk{
				BaseObject: entities.BaseObject{BBox: spacingBox},
				Text:       spacingText,
				FontStyle:  entities.FontStyle{FontSize: current.FontStyle.FontSize},
				Baseline:   current.Baseline,
			})
		}
		result = append(result, current)
		previousEnd = current.BBox.X + current.BBox.Width
	}

	return result
}

func shouldInsertSyntheticSpaceBeforeSingletonContinuation(chunks []*entities.TextChunk, idx int, previousEnd, avgCharWidth, spaceThreshold float64) bool {
	if idx <= 0 || idx+1 >= len(chunks) {
		return false
	}
	previous := chunks[idx-1]
	current := chunks[idx]
	next := chunks[idx+1]
	if previous == nil || current == nil || next == nil {
		return false
	}
	currentRune, ok := singleNonSpaceRune(current)
	if !ok || !unicode.IsLower(currentRune) {
		return false
	}
	nextRune, ok := firstNonSpaceRune(next)
	if !ok || !unicode.IsLower(nextRune) || trimmedRuneCount(next) <= 1 {
		return false
	}
	prevRune, ok := lastNonSpaceRune(previous)
	if !ok || !isWordBoundaryRune(prevRune) || prevRune == '-' {
		return false
	}
	prevGap := current.BBox.X - previousEnd
	if prevGap <= 0 || prevGap > avgCharWidth*textLineWordBoundaryRatio {
		return false
	}
	nextGap := next.BBox.X - (current.BBox.X + current.BBox.Width)
	if nextGap <= 0 {
		return false
	}
	return nextGap > spaceThreshold || looksLikeInlineWordBoundary(current, next)
}

func shouldInsertSyntheticSpace(previous, current *entities.TextChunk, gap, avgCharWidth, avgFontSize, spaceThreshold float64) bool {
	if looksLikeHyphenatedContinuation(previous, current) {
		return false
	}
	if looksLikeLowercaseContinuationFragment(previous, current) {
		return false
	}
	if gap <= 0 {
		if !looksLikeInlineWordBoundary(previous, current) || !isRecoverableWordOverlap(gap, avgCharWidth) {
			return false
		}
		gap = 0
	}
	if gap > spaceThreshold {
		return true
	}
	if avgCharWidth <= 0 {
		return false
	}
	boundaryRatio := textLineWordBoundaryRatio
	if looksLikeDenseProseWordBoundary(previous, current) {
		boundaryRatio = textLineDenseProseRatio
	}
	if looksLikeDisplayTextWordBoundary(previous, current, avgFontSize) {
		boundaryRatio = math.Min(boundaryRatio, textLineDisplayTextRatio)
	}
	if gap <= avgCharWidth*boundaryRatio {
		return false
	}
	return looksLikeInlineWordBoundary(previous, current)
}

func isRecoverableWordOverlap(gap, avgCharWidth float64) bool {
	if gap >= 0 {
		return true
	}
	if avgCharWidth <= 0 {
		return false
	}
	overlapTolerance := math.Min(avgCharWidth*textLineWordOverlapRatio, 0.5)
	return math.Abs(gap) <= overlapTolerance
}

func looksLikeInlineWordBoundary(previous, current *entities.TextChunk) bool {
	prevRune, ok := lastNonSpaceRune(previous)
	if !ok {
		return false
	}
	nextRune, ok := firstNonSpaceRune(current)
	if !ok {
		return false
	}
	return isWordBoundaryRune(prevRune) && isWordBoundaryRune(nextRune)
}

func looksLikeDenseProseWordBoundary(previous, current *entities.TextChunk) bool {
	if !looksLikeInlineWordBoundary(previous, current) {
		return false
	}
	return trimmedRuneCount(previous) > 1 && trimmedRuneCount(current) > 1
}

func looksLikeDisplayTextWordBoundary(previous, current *entities.TextChunk, avgFontSize float64) bool {
	if avgFontSize < textLineDisplayFontMin || !looksLikeInlineWordBoundary(previous, current) {
		return false
	}
	return isAlphabeticDisplayWord(previous) && isAlphabeticDisplayWord(current)
}

func looksLikeLowercaseContinuationFragment(previous, current *entities.TextChunk) bool {
	prevRune, ok := singleNonSpaceRune(previous)
	if !ok || !unicode.IsLower(prevRune) {
		return false
	}
	nextRune, ok := firstNonSpaceRune(current)
	if !ok || !unicode.IsLower(nextRune) {
		return false
	}
	return trimmedRuneCount(current) > 1
}

func looksLikeHyphenatedContinuation(previous, current *entities.TextChunk) bool {
	prevRune, ok := lastNonSpaceRune(previous)
	if !ok || prevRune != '-' {
		return false
	}
	currentRune, ok := singleNonSpaceRune(current)
	return ok && unicode.IsLower(currentRune)
}

func isAlphabeticDisplayWord(chunk *entities.TextChunk) bool {
	text := strings.TrimSpace(chunk.Text)
	if trimmedRuneCountFromString(text) < 3 {
		return false
	}
	for _, r := range text {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func lastNonSpaceRune(chunk *entities.TextChunk) (rune, bool) {
	if chunk == nil {
		return 0, false
	}
	text := strings.TrimRightFunc(chunk.Text, unicode.IsSpace)
	if text == "" {
		return 0, false
	}
	r, _ := utf8.DecodeLastRuneInString(text)
	if r == utf8.RuneError {
		return 0, false
	}
	return r, true
}

func firstNonSpaceRune(chunk *entities.TextChunk) (rune, bool) {
	if chunk == nil {
		return 0, false
	}
	text := strings.TrimLeftFunc(chunk.Text, unicode.IsSpace)
	if text == "" {
		return 0, false
	}
	r, _ := utf8.DecodeRuneInString(text)
	if r == utf8.RuneError {
		return 0, false
	}
	return r, true
}

func singleNonSpaceRune(chunk *entities.TextChunk) (rune, bool) {
	if chunk == nil {
		return 0, false
	}
	text := strings.TrimSpace(chunk.Text)
	if trimmedRuneCountFromString(text) != 1 {
		return 0, false
	}
	r, _ := utf8.DecodeRuneInString(text)
	if r == utf8.RuneError {
		return 0, false
	}
	return r, true
}

func trimmedRuneCount(chunk *entities.TextChunk) int {
	if chunk == nil {
		return 0
	}
	return trimmedRuneCountFromString(strings.TrimSpace(chunk.Text))
}

func trimmedRuneCountFromString(text string) int {
	return utf8.RuneCountInString(text)
}

func isWordBoundaryRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
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

	used := make([]bool, len(lineArts))
	for _, line := range lines {
		bestIdx := -1
		bestGap := math.MaxFloat64
		for idx, art := range lineArts {
			if used[idx] || !isLineConnectedWithLineArt(line, art) {
				continue
			}
			gap := line.BBox.X - (art.BBox.X + art.BBox.Width)
			if gap < bestGap {
				bestGap = gap
				bestIdx = idx
			}
		}
		if bestIdx >= 0 {
			line.LineArtBullet = lineArts[bestIdx]
			used[bestIdx] = true
		}
	}
}

func isLineConnectedWithLineArt(line *entities.TextLine, lineArt *entities.LineArtChunk) bool {
	lineHeight := line.BBox.Height
	if lineHeight <= 0 {
		return false
	}

	artRight := lineArt.BBox.X + lineArt.BBox.Width
	if artRight > line.BBox.X {
		return false
	}
	if line.BBox.X-artRight > lineArtBulletGapMax {
		return false
	}

	lineArtY := lineArt.BBox.Y + lineArt.BBox.Height/2
	return math.Abs(line.Baseline-lineArtY) <= lineArtBaselineTolerance &&
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

func averageCharWidth(chunks []*entities.TextChunk) float64 {
	totalWidth := 0.0
	totalChars := 0
	for _, chunk := range chunks {
		if chunk == nil {
			continue
		}
		runes := []rune(chunk.Text)
		if len(runes) == 0 {
			continue
		}
		totalWidth += chunk.BBox.Width
		totalChars += len(runes)
	}
	if totalChars == 0 {
		return averageFontSize(chunks) * 0.5
	}
	return totalWidth / float64(totalChars)
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
