// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package processors

import (
	"math"
	"regexp"
	"strings"
	"unicode"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/utils"
)

const (
	headingProbabilityThreshold        = 0.75
	bulletedHeadingProbabilityIncrease = 0.10
	sectionHeadingProbabilityIncrease  = 0.22
	bodyStylePenalty                   = 0.35
	distinctFontFamilyHeadingBoost     = 0.55
)

var sectionHeadingPrefixPattern = regexp.MustCompile(`^(?:(?:(?:\d+(?:\.\d+)*)\.?|[A-Z]\.|[IVXLCDM]+\.)\s+\p{L}|[IVXLCDM]{2,}\s+\p{L})`)

type HeadingProcessor struct{}

func (p *HeadingProcessor) Process(lines []*entities.TextLine, ctx *containers.ProcessorContext) []*entities.SemanticHeading {
	stats := utils.NewTextNodeStatistics()
	for _, line := range lines {
		if line == nil {
			continue
		}
		if shouldSkipHeadingLine(line) {
			continue
		}
		size := lineFontSize(line)
		weight := lineFontWeight(line)
		stats.Add(size, weight)
	}
	bodyFontSize := stats.FontSizeMode()
	bodyFontFamily := dominantLineFontFamily(lines)

	headings := make([]*entities.SemanticHeading, 0)
	for idx, line := range lines {
		if line == nil {
			continue
		}
		if shouldSkipHeadingLine(line) {
			continue
		}
		var prevLine, nextLine *entities.TextLine
		if idx > 0 {
			prevLine = lines[idx-1]
		}
		if idx+1 < len(lines) {
			nextLine = lines[idx+1]
		}
		score := p.headingScore(line, prevLine, nextLine, bodyFontSize, bodyFontFamily, stats)
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

func (p *HeadingProcessor) headingScore(line, prevLine, nextLine *entities.TextLine, bodyFontSize float64, bodyFontFamily string, stats *utils.TextNodeStatistics) float64 {
	text := strings.TrimSpace(line.GetText())
	if text == "" {
		return 0
	}
	if len([]rune(text)) >= 120 {
		return 0
	}
	if !hasMinimumHeadingContent(text) {
		return 0
	}
	if isFormulaLikeHeadingFragment(text) {
		return 0
	}

	size := lineFontSize(line)
	weight := lineFontWeight(line)
	family := lineFontFamily(line)
	score := 0.0

	if bodyFontSize > 0 && size > bodyFontSize*1.1 {
		score += min(0.45, (size-bodyFontSize)/max(bodyFontSize, 1.0)+0.2)
	}
	score += stats.FontSizeRarityBoost(size)
	score += stats.FontWeightRarityBoost(weight)

	if lineIsBold(line) {
		score += 0.2
	}
	if family != "" && bodyFontFamily != "" && family != bodyFontFamily {
		score += distinctFontFamilyHeadingBoost
	}
	if bodyFontSize > 0 && size <= bodyFontSize*1.1 && !lineIsBold(line) && !lineIsItalic(line) &&
		(bodyFontFamily == "" || family == "" || family == bodyFontFamily) {
		score -= bodyStylePenalty
	}
	if isIsolatedHeadingLine(line, prevLine, nextLine) {
		score += 0.22
	}
	if line.BBox.Page == 0 && line.BBox.Y+line.BBox.Height >= 700 {
		score += 0.08
	}
	length := len([]rune(text))
	switch {
	case length <= 5:
		score -= 0.2
	case length <= 8:
		score -= 0.02
	case length <= 30:
		score += 0.1
	case length <= 80:
		score += 0.03
	default:
		score -= 0.15
	}
	if headingWordCount(text) == 1 && !looksLikeSectionHeadingText(text) {
		score -= 0.28
	}

	if utils.IsOrderedBullet(text) || utils.IsUnorderedBullet(text) {
		score += bulletedHeadingProbabilityIncrease
	}
	if looksLikeSectionHeadingText(text) {
		score += sectionHeadingProbabilityIncrease
	}
	if isColonSuffixedHeadingText(line, text) {
		score += 0.15
	}
	if strings.HasSuffix(text, ".") || strings.HasSuffix(text, ",") {
		score -= 0.1
	}
	if strings.Count(text, " ") > 12 {
		score -= 0.1
	}
	return score
}

func (p *HeadingProcessor) PromoteListHeadings(elements []entities.IObject, ctx *containers.ProcessorContext) ([]entities.IObject, []*entities.SemanticHeading) {
	if len(elements) == 0 {
		return elements, nil
	}

	stats, bodyFontSize, bodyFontFamily := buildHeadingStatsFromObjects(elements)
	out := make([]entities.IObject, 0, len(elements))
	headings := make([]*entities.SemanticHeading, 0)

	for idx, element := range elements {
		list, ok := element.(*entities.PDFList)
		if !ok || list == nil || len(list.Items) == 0 {
			out = append(out, element)
			continue
		}

		prevLine := nearestTextLineBefore(elements, idx)
		nextLine := nearestTextLineAfter(elements, idx)
		replacement, promoted := p.promoteList(list, prevLine, nextLine, bodyFontSize, bodyFontFamily, stats, ctx)
		if len(promoted) == 0 {
			out = append(out, element)
			continue
		}
		out = append(out, replacement...)
		headings = append(headings, promoted...)
	}

	return out, headings
}

func (p *HeadingProcessor) promoteList(list *entities.PDFList, prevLine, nextLine *entities.TextLine, bodyFontSize float64, bodyFontFamily string, stats *utils.TextNodeStatistics, ctx *containers.ProcessorContext) ([]entities.IObject, []*entities.SemanticHeading) {
	if list == nil {
		return nil, nil
	}

	replacement := make([]entities.IObject, 0, len(list.Items))
	promotedHeadings := make([]*entities.SemanticHeading, 0)
	remainingItems := make([]*entities.ListItem, 0, len(list.Items))

	for itemIndex, item := range list.Items {
		candidate, body, ok := listItemHeadingCandidate(item)
		if !ok {
			remainingItems = append(remainingItems, item)
			continue
		}

		itemPrev := prevLine
		itemNext := nextLine
		if itemIndex > 0 {
			itemPrev = lastTextLineInListItem(list.Items[itemIndex-1])
		}
		if itemIndex+1 < len(list.Items) {
			itemNext = firstTextLineInListItem(list.Items[itemIndex+1])
		}

		score := p.headingScore(candidate, itemPrev, itemNext, bodyFontSize, bodyFontFamily, stats)
		if item.IsOrdered && looksLikeSectionHeadingText(strings.TrimSpace(candidate.GetText())) {
			score += 0.75
		}
		if score <= headingProbabilityThreshold || !looksLikeSectionHeadingText(strings.TrimSpace(candidate.GetText())) || !hasListHeadingStyle(item, candidate, itemPrev, itemNext, bodyFontSize) {
			remainingItems = append(remainingItems, item)
			continue
		}

		heading := &entities.SemanticHeading{
			BaseObject: entities.BaseObject{ID: nextID(ctx), BBox: candidate.BBox},
			Lines:      []*entities.TextLine{candidate},
			FontSize:   lineFontSize(candidate),
			IsBold:     lineIsBold(candidate),
			IsItalic:   lineIsItalic(candidate),
			FontFamily: lineFontFamily(candidate),
		}
		replacement = append(replacement, heading)
		promotedHeadings = append(promotedHeadings, heading)
		if ctx != nil {
			ctx.Headings = append(ctx.Headings, heading)
		}

		if len(body) > 0 {
			bodyObjects, _ := transformTextRuns(body, ctx)
			replacement = append(replacement, bodyObjects...)
		}
	}

	if len(remainingItems) > 0 {
		cloned := *list
		cloned.Items = remainingItems
		cloned.BBox = listBBox(remainingItems)
		replacement = append(replacement, &cloned)
	}

	return replacement, promotedHeadings
}

func dominantLineFontFamily(lines []*entities.TextLine) string {
	counts := make(map[string]int)
	bestFamily := ""
	bestCount := 0
	for _, line := range lines {
		family := lineFontFamily(line)
		if family == "" {
			continue
		}
		counts[family]++
		if counts[family] > bestCount {
			bestCount = counts[family]
			bestFamily = family
		}
	}
	return bestFamily
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

func shouldSkipHeadingLine(line *entities.TextLine) bool {
	return line != nil && (line.InListItem || line.InTableCell)
}

func buildHeadingStatsFromObjects(elements []entities.IObject) (*utils.TextNodeStatistics, float64, string) {
	stats := utils.NewTextNodeStatistics()
	lines := make([]*entities.TextLine, 0)
	for _, element := range elements {
		for _, line := range textLinesFromObject(element) {
			if line == nil || line.InTableCell {
				continue
			}
			stats.Add(lineFontSize(line), lineFontWeight(line))
			lines = append(lines, line)
		}
	}
	return stats, stats.FontSizeMode(), dominantLineFontFamily(lines)
}

func textLinesFromObject(obj entities.IObject) []*entities.TextLine {
	switch v := obj.(type) {
	case *entities.TextLine:
		return []*entities.TextLine{v}
	case *entities.SemanticParagraph:
		return v.Lines
	case *entities.SemanticHeading:
		return v.Lines
	case *entities.SemanticHeaderFooter:
		return v.Lines
	case *entities.PDFList:
		lines := make([]*entities.TextLine, 0)
		for _, item := range v.Items {
			lines = append(lines, textLinesFromObjects(item.Content)...)
		}
		return lines
	default:
		return nil
	}
}

func textLinesFromObjects(objects []entities.IObject) []*entities.TextLine {
	lines := make([]*entities.TextLine, 0)
	for _, obj := range objects {
		lines = append(lines, textLinesFromObject(obj)...)
	}
	return lines
}

func nearestTextLineBefore(elements []entities.IObject, index int) *entities.TextLine {
	for i := index - 1; i >= 0; i-- {
		lines := textLinesFromObject(elements[i])
		if len(lines) > 0 {
			return lines[len(lines)-1]
		}
	}
	return nil
}

func nearestTextLineAfter(elements []entities.IObject, index int) *entities.TextLine {
	for i := index + 1; i < len(elements); i++ {
		lines := textLinesFromObject(elements[i])
		if len(lines) > 0 {
			return lines[0]
		}
	}
	return nil
}

func listItemHeadingCandidate(item *entities.ListItem) (*entities.TextLine, []entities.IObject, bool) {
	if item == nil || !item.IsOrdered && item.BulletText == "" {
		return nil, nil, false
	}
	var firstLine *entities.TextLine
	firstLineIndex := -1
	for idx, obj := range item.Content {
		line, ok := obj.(*entities.TextLine)
		if !ok || line == nil || strings.TrimSpace(line.GetText()) == "" {
			continue
		}
		firstLine = line
		firstLineIndex = idx
		break
	}
	if firstLine == nil {
		return nil, nil, false
	}

	fullText := strings.TrimSpace(strings.TrimSpace(item.BulletText) + " " + strings.TrimSpace(firstLine.GetText()))
	if !looksLikeSectionHeadingText(fullText) {
		return nil, nil, false
	}

	candidate := cloneLineWithText(firstLine, fullText)
	candidate.InListItem = false

	body := make([]entities.IObject, 0, len(item.Content))
	if firstLineIndex >= 0 {
		body = append(body, item.Content[:firstLineIndex]...)
		body = append(body, item.Content[firstLineIndex+1:]...)
	}
	return candidate, body, true
}

func firstTextLineInListItem(item *entities.ListItem) *entities.TextLine {
	if item == nil {
		return nil
	}
	for _, obj := range item.Content {
		if line, ok := obj.(*entities.TextLine); ok && line != nil {
			return line
		}
	}
	return nil
}

func lastTextLineInListItem(item *entities.ListItem) *entities.TextLine {
	if item == nil {
		return nil
	}
	for idx := len(item.Content) - 1; idx >= 0; idx-- {
		if line, ok := item.Content[idx].(*entities.TextLine); ok && line != nil {
			return line
		}
	}
	return nil
}

func listBBox(items []*entities.ListItem) entities.BoundingBox {
	if len(items) == 0 {
		return entities.BoundingBox{}
	}
	box := items[0].BBox
	for _, item := range items[1:] {
		if item != nil {
			box = unionBox(box, item.BBox)
		}
	}
	return box
}

func hasListHeadingStyle(item *entities.ListItem, line, prevLine, nextLine *entities.TextLine, bodyFontSize float64) bool {
	if item == nil || line == nil {
		return false
	}
	size := lineFontSize(line)
	if lineIsBold(line) || (bodyFontSize > 0 && size > bodyFontSize*1.08) || isIsolatedHeadingLine(line, prevLine, nextLine) {
		return true
	}
	if !item.IsOrdered || !looksLikeSectionHeadingText(strings.TrimSpace(line.GetText())) {
		return false
	}
	return bodyFontSize == 0 || size >= bodyFontSize*0.98
}

func looksLikeSectionHeadingText(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	return sectionHeadingPrefixPattern.MatchString(text)
}

func hasMinimumHeadingContent(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	if headingWordCount(text) >= 2 {
		return true
	}
	alnum := 0
	letterCount := 0
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			alnum++
		}
		if unicode.IsLetter(r) {
			letterCount++
		}
	}
	if alnum < 3 {
		return false
	}
	if looksLikeSectionHeadingText(text) {
		return letterCount >= 3
	}
	return headingWordCount(text) == 1 && letterCount >= 4
}

func isFormulaLikeHeadingFragment(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	if headingWordCount(text) >= 2 {
		return false
	}
	letters := 0
	digits := 0
	operators := 0
	for _, r := range text {
		switch {
		case unicode.IsLetter(r):
			letters++
		case unicode.IsDigit(r):
			digits++
		case strings.ContainsRune("=+-−*/^_()[]{}∂∫√≤≥≈∞′″", r):
			operators++
		}
	}
	return letters+digits <= 3 || operators >= 2
}

func headingWordCount(text string) int {
	return len(strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune(".,;:!?()[]{}<>/\\|\"'“”‘’•*-_=+`~", r)
	}))
}

func isColonSuffixedHeadingText(line *entities.TextLine, text string) bool {
	if line == nil || line.InListItem || line.InTableCell {
		return false
	}
	text = strings.TrimSpace(text)
	if !strings.HasSuffix(text, ":") {
		return false
	}
	wordCount := headingWordCount(text)
	if wordCount < 2 || wordCount > 8 {
		return false
	}
	if utils.IsOrderedBullet(text) || utils.IsUnorderedBullet(text) {
		return false
	}
	return !startsWithListLikePrefix(text)
}

func startsWithListLikePrefix(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	for _, prefix := range []string{"(", "[", "{", "\"", "'", "“", "‘"} {
		text = strings.TrimLeft(text, prefix)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return false
	}
	first := fields[0]
	if first == "" {
		return false
	}
	last := rune(first[len(first)-1])
	if last != '.' && last != ')' && last != ':' {
		return false
	}
	token := strings.TrimRight(first, ".):")
	if token == "" {
		return false
	}
	hasDigit := false
	for _, r := range token {
		if unicode.IsDigit(r) {
			hasDigit = true
			continue
		}
		if r != '.' {
			return false
		}
	}
	return hasDigit
}

func nextID(ctx *containers.ProcessorContext) string {
	if ctx == nil {
		return ""
	}
	return ctx.NextID()
}
