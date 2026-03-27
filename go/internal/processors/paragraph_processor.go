// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors

import (
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const (
	paragraphMergeThreshold      = 0.75
	paragraphCloseGapMultiplier  = 1.15
	paragraphNormalGapMultiplier = 1.45
)

var paragraphLabelPattern = regexp.MustCompile(`^(?:[\p{N}\p{L}]+[.)\]])\s+|^(?:[-*+•◦▪‣])\s*`)

type ParagraphProcessor struct{}

type paragraphBlock struct {
	lines        []*entities.TextLine
	alignment    string
	hasStartLine bool
	hasEndLine   bool
}

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
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].BBox.Page != filtered[j].BBox.Page {
			return filtered[i].BBox.Page < filtered[j].BBox.Page
		}
		if !closeEnough(filtered[i].Baseline, filtered[j].Baseline, textLineBaselineTolerance) {
			return filtered[i].Baseline > filtered[j].Baseline
		}
		return filtered[i].BBox.X < filtered[j].BBox.X
	})

	blocks := make([]*paragraphBlock, 0, len(filtered))
	for _, line := range filtered {
		blocks = append(blocks, newParagraphBlock(line))
	}

	blocks = mergeParagraphBlocks(blocks, tryMergeJustifiedBlocks)
	blocks = mergeParagraphBlocks(blocks, tryMergeLeftAlignedBlocks)
	blocks = mergeParagraphBlocks(blocks, tryMergeFirstLeftLines)
	blocks = mergeParagraphBlocks(blocks, tryMergeCenteredBlocks)

	return blocksToParagraphs(blocks, ctx)
}

func mergeParagraphBlocks(blocks []*paragraphBlock, fn func(*paragraphBlock, *paragraphBlock) bool) []*paragraphBlock {
	if len(blocks) < 2 {
		return blocks
	}
	merged := make([]*paragraphBlock, 0, len(blocks))
	merged = append(merged, blocks[0])
	for i := 1; i < len(blocks); i++ {
		current := blocks[i]
		previous := merged[len(merged)-1]
		if fn(previous, current) {
			previous.lines = append(previous.lines, current.lines...)
			previous.alignment = detectParagraphAlignment(previous.lines)
			previous.hasStartLine = previous.hasStartLine || current.hasStartLine
			previous.hasEndLine = previous.hasEndLine || current.hasEndLine
			continue
		}
		merged = append(merged, current)
	}
	return merged
}

func tryMergeJustifiedBlocks(previous, next *paragraphBlock) bool {
	if blockPairAlignment(previous, next) != entities.AlignJustify {
		return false
	}
	if !canMergeByLeading(previous, next, false) || !sameTextSize(previous, next) {
		return false
	}
	previous.alignment = entities.AlignJustify
	return true
}

func tryMergeCenteredBlocks(previous, next *paragraphBlock) bool {
	if blockPairAlignment(previous, next) != entities.AlignCenter {
		return false
	}
	if !canMergeByLeading(previous, next, false) || !sameTextSize(previous, next) {
		return false
	}
	if previous.linesCount() > 1 && previous.effectiveAlignment() != entities.AlignCenter {
		return false
	}
	if next.linesCount() > 1 && next.effectiveAlignment() != entities.AlignCenter {
		return false
	}
	previous.alignment = entities.AlignCenter
	return true
}

func tryMergeLeftAlignedBlocks(previous, next *paragraphBlock) bool {
	if blockPairAlignment(previous, next) != entities.AlignLeft {
		return false
	}
	if isLabeledLine(next.firstLine()) || !sameTextSize(previous, next) {
		return false
	}
	if !supportsLeftAlignment(previous) || !supportsLeftAlignment(next) {
		return false
	}
	shouldBeClose := previous.effectiveAlignment() == entities.AlignJustify || next.effectiveAlignment() == entities.AlignJustify
	if !canMergeByLeading(previous, next, shouldBeClose) {
		return false
	}
	if !hasConsistentLeftIndent(previous, next) {
		return false
	}
	previous.alignment = entities.AlignLeft
	previous.hasEndLine = false
	return true
}

func tryMergeFirstLeftLines(previous, next *paragraphBlock) bool {
	if previous.linesCount() != 1 || next.linesCount() == 0 {
		return false
	}
	if isLabeledLine(next.firstLine()) || !sameTextSize(previous, next) {
		return false
	}
	if next.hasStartLine || next.effectiveAlignment() != entities.AlignLeft {
		return false
	}
	if !canMergeByLeading(previous, next, false) {
		return false
	}
	if horizontalOverlapRatio(previous.lastLine().BBox, next.firstLine().BBox) <= 0.50 {
		return false
	}
	previous.alignment = entities.AlignLeft
	previous.hasStartLine = true
	return true
}

func canMergeByLeading(previous, next *paragraphBlock, shouldBeClose bool) bool {
	last := previous.lastLine()
	first := next.firstLine()
	if last == nil || first == nil || last.BBox.Page != first.BBox.Page {
		return false
	}
	if !sameTextSize(previous, next) {
		return false
	}
	if horizontalOverlapRatio(last.BBox, first.BBox) <= 0.50 {
		return false
	}
	leading := expectedLeading(previous, next)
	if leading <= 0 {
		return false
	}
	gap := lineVerticalGap(last.BBox, first.BBox)
	multiplier := paragraphNormalGapMultiplier
	if shouldBeClose {
		multiplier = paragraphCloseGapMultiplier
	}
	return gap <= leading*multiplier
}

func expectedLeading(previous, next *paragraphBlock) float64 {
	samples := append(internalLeadingSamples(previous.lines), internalLeadingSamples(next.lines)...)
	if len(samples) == 0 {
		lineHeight := math.Max(previous.lastLine().BBox.Height, next.firstLine().BBox.Height)
		if lineHeight <= 0 {
			lineHeight = math.Max(averageFontSize(previous.lastLine().Chunks), averageFontSize(next.firstLine().Chunks))
		}
		if lineHeight <= 0 {
			return 1
		}
		return lineHeight
	}
	sort.Float64s(samples)
	return samples[len(samples)/2]
}

func internalLeadingSamples(lines []*entities.TextLine) []float64 {
	if len(lines) < 2 {
		return nil
	}
	samples := make([]float64, 0, len(lines)-1)
	for i := 1; i < len(lines); i++ {
		gap := lineVerticalGap(lines[i-1].BBox, lines[i].BBox)
		if gap > 0 {
			samples = append(samples, gap)
		}
	}
	return samples
}

func hasConsistentLeftIndent(previous, next *paragraphBlock) bool {
	expectedLeft := dominantLeft(previous.lines)
	nextLeft := next.firstLine().BBox.X
	tolerance := alignmentTolerance(previous.lastLine(), next.firstLine())
	if nextLeft > expectedLeft+tolerance {
		return false
	}
	if nextLeft < expectedLeft-tolerance {
		return false
	}
	return true
}

func dominantLeft(lines []*entities.TextLine) float64 {
	if len(lines) == 0 {
		return 0
	}
	values := make([]float64, 0, len(lines))
	start := 0
	if len(lines) > 1 {
		start = 1
	}
	for i := start; i < len(lines); i++ {
		values = append(values, lines[i].BBox.X)
	}
	if len(values) == 0 {
		return lines[0].BBox.X
	}
	sort.Float64s(values)
	return values[len(values)/2]
}

func supportsLeftAlignment(block *paragraphBlock) bool {
	if block == nil {
		return false
	}
	if block.linesCount() == 1 {
		return true
	}
	switch block.effectiveAlignment() {
	case entities.AlignLeft, entities.AlignJustify:
		return true
	default:
		return false
	}
}

func sameTextSize(previous, next *paragraphBlock) bool {
	for _, size1 := range textSizes(previous.lines) {
		for _, size2 := range textSizes(next.lines) {
			if similarFontSize(size1, size2) {
				return true
			}
		}
	}
	return false
}

func textSizes(lines []*entities.TextLine) []float64 {
	out := make([]float64, 0, len(lines))
	for _, line := range lines {
		size := averageFontSize(line.Chunks)
		if size > 0 {
			out = append(out, size)
		}
	}
	return out
}

func blockPairAlignment(previous, next *paragraphBlock) string {
	return detectPairAlignment(previous.lastLine(), next.firstLine())
}

func newParagraphBlock(line *entities.TextLine) *paragraphBlock {
	return &paragraphBlock{
		lines:     []*entities.TextLine{line},
		alignment: entities.AlignLeft,
	}
}

func (b *paragraphBlock) firstLine() *entities.TextLine {
	if b == nil || len(b.lines) == 0 {
		return nil
	}
	return b.lines[0]
}

func (b *paragraphBlock) lastLine() *entities.TextLine {
	if b == nil || len(b.lines) == 0 {
		return nil
	}
	return b.lines[len(b.lines)-1]
}

func (b *paragraphBlock) linesCount() int {
	if b == nil {
		return 0
	}
	return len(b.lines)
}

func (b *paragraphBlock) effectiveAlignment() string {
	if b == nil {
		return entities.AlignLeft
	}
	if b.alignment != "" {
		return b.alignment
	}
	return detectParagraphAlignment(b.lines)
}

func blocksToParagraphs(blocks []*paragraphBlock, ctx *containers.ProcessorContext) []*entities.SemanticParagraph {
	paragraphs := make([]*entities.SemanticParagraph, 0, len(blocks))
	for _, block := range blocks {
		if block == nil || len(block.lines) == 0 {
			continue
		}
		paragraph := &entities.SemanticParagraph{
			BaseObject: entities.BaseObject{
				ID:   nextObjectID(ctx),
				BBox: unionLines(block.lines),
			},
			Lines:     block.lines,
			Alignment: block.effectiveAlignment(),
		}
		paragraph.Lines[0].IsFirstLine = true
		paragraph.Lines[len(paragraph.Lines)-1].IsLastLine = true
		paragraphs = append(paragraphs, paragraph)
	}
	return paragraphs
}

func unionLines(lines []*entities.TextLine) entities.BoundingBox {
	box := lines[0].BBox
	for _, line := range lines[1:] {
		box = unionBoundingBox(box, line.BBox)
	}
	return box
}

func detectPairAlignment(line1, line2 *entities.TextLine) string {
	if line1 == nil || line2 == nil {
		return ""
	}
	tolerance := alignmentTolerance(line1, line2)
	leftAligned := closeEnough(line1.BBox.X, line2.BBox.X, tolerance)
	rightAligned := closeEnough(line1.BBox.X+line1.BBox.Width, line2.BBox.X+line2.BBox.Width, tolerance)
	center1 := line1.BBox.X + line1.BBox.Width/2
	center2 := line2.BBox.X + line2.BBox.Width/2
	centerAligned := closeEnough(center1, center2, tolerance)
	maxWidth := math.Max(line1.BBox.Width, line2.BBox.Width)
	justifyAligned := maxWidth > 0 &&
		len(line1.Chunks) > 0 &&
		len(line2.Chunks) > 0 &&
		line1.BBox.Width >= maxWidth*0.9 &&
		line2.BBox.Width >= maxWidth*0.9 &&
		leftAligned &&
		rightAligned

	switch {
	case justifyAligned:
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
	if len(lines) <= 1 {
		return entities.AlignLeft
	}

	leftAligned := true
	rightAligned := true
	centerAligned := true
	justifyAligned := len(lines) >= 2
	base := lines[0]
	tolerance := math.Max(1.5, base.BBox.Height*0.5)
	baseCenter := base.BBox.X + base.BBox.Width/2
	baseRight := base.BBox.X + base.BBox.Width
	maxWidth := base.BBox.Width

	for _, line := range lines[1:] {
		if line.BBox.Width > maxWidth {
			maxWidth = line.BBox.Width
		}
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
	for _, line := range lines {
		if maxWidth <= 0 || line.BBox.Width < maxWidth*0.9 {
			justifyAligned = false
			break
		}
	}

	switch {
	case justifyAligned && leftAligned && rightAligned:
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

func alignmentTolerance(line1, line2 *entities.TextLine) float64 {
	height1 := 0.0
	height2 := 0.0
	if line1 != nil {
		height1 = line1.BBox.Height
	}
	if line2 != nil {
		height2 = line2.BBox.Height
	}
	return math.Max(1.5, math.Min(height1, height2)*0.5)
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

func horizontalOverlapRatio(a, b entities.BoundingBox) float64 {
	overlap := horizontalOverlap(a, b)
	if overlap <= 0 {
		return 0
	}
	minWidth := math.Min(a.Width, b.Width)
	if minWidth <= 0 {
		return 0
	}
	return overlap / minWidth
}

func lineVerticalGap(a, b entities.BoundingBox) float64 {
	if a.Y >= b.Y {
		return math.Abs(a.Y - (b.Y + b.Height))
	}
	return math.Abs(b.Y - (a.Y + a.Height))
}

func similarFontSize(a, b float64) bool {
	if a <= 0 || b <= 0 {
		return false
	}
	diff := math.Abs(a - b)
	maxSize := math.Max(a, b)
	return diff/maxSize <= 0.10
}

func closeEnough(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
