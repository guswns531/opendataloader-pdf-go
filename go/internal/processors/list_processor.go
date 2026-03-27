// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package processors

import (
	"regexp"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/utils"
)

var (
	orderedBulletPattern   = regexp.MustCompile(`^(\d+(?:\.\d+)*\.|\([a-zA-Z0-9]+\)|[a-zA-Z]\.)`)
	unorderedBulletPattern = regexp.MustCompile(`^[•\-※◦▪▸►\*]`)
	koreanAttachPattern    = regexp.MustCompile(`^붙\s*임`)
)

const baselineDiffThreshold = 1.2

type ListProcessor struct{}

func (p *ListProcessor) Process(elements []entities.IObject, ctx *containers.ProcessorContext) []entities.IObject {
	if len(elements) == 0 {
		return elements
	}

	out := make([]entities.IObject, 0, len(elements))
	for i := 0; i < len(elements); {
		line, ok := elements[i].(*entities.TextLine)
		if !ok || !isBulletLine(line) {
			out = append(out, elements[i])
			i++
			continue
		}

		list := &entities.PDFList{
			BaseObject: entities.BaseObject{ID: nextID(ctx), BBox: line.BBox},
			IsOrdered:  utils.IsOrderedBullet(strings.TrimSpace(line.GetText())),
		}

		for i < len(elements) {
			current, ok := elements[i].(*entities.TextLine)
			if !ok || !isBulletLine(current) {
				break
			}
			item, next := p.consumeListItem(elements, i, list.IsOrdered, ctx)
			list.Items = append(list.Items, item)
			list.BBox = unionBox(list.BBox, item.BBox)
			i = next
		}

		out = append(out, list)
	}
	return out
}

func MergeListsAcrossPages(pages []*entities.Page) {
	if len(pages) < 2 {
		return
	}

	for idx := 1; idx < len(pages); idx++ {
		prev := pages[idx-1]
		curr := pages[idx]
		if prev == nil || curr == nil {
			continue
		}

		prevList, prevPos := trailingList(prev.Elements)
		currList, currPos := leadingList(curr.Elements)
		if prevList == nil || currList == nil {
			continue
		}
		if !sameListPattern(prevList, currList) {
			continue
		}

		prevList.Items = append(prevList.Items, currList.Items...)
		curr.Elements = append(append([]entities.IObject{}, curr.Elements[:currPos]...), curr.Elements[currPos+1:]...)
		if prevPos >= 0 {
			prev.Elements[prevPos] = prevList
		}
	}
}

func (p *ListProcessor) consumeListItem(elements []entities.IObject, start int, isOrdered bool, ctx *containers.ProcessorContext) (*entities.ListItem, int) {
	line := elements[start].(*entities.TextLine)
	bulletText, bodyText := splitBulletText(strings.TrimSpace(line.GetText()))
	bodyLine := cloneLineWithText(line, bodyText)
	bodyLine.InListItem = true
	content := []entities.IObject{bodyLine}
	box := bodyLine.BBox
	next := start + 1
	previousLine := bodyLine

	for next < len(elements) {
		nextLine, ok := elements[next].(*entities.TextLine)
		if ok {
			if isBulletLine(nextLine) || !isListContinuation(line, previousLine, nextLine) {
				break
			}
			nextLine.InListItem = true
			content = append(content, nextLine)
			box = unionBox(box, nextLine.BBox)
			previousLine = nextLine
			next++
			continue
		}
		content = append(content, elements[next])
		box = unionBox(box, elements[next].GetBBox())
		next++
	}

	return &entities.ListItem{
		BaseObject: entities.BaseObject{ID: nextID(ctx), BBox: box},
		Content:    content,
		BulletText: bulletText,
		IsOrdered:  isOrdered,
		Level:      1,
	}, next
}

func isListContinuation(first, previous, current *entities.TextLine) bool {
	if current == nil || first == nil || previous == nil {
		return false
	}
	maxXGap := max(lineFontSize(first)*0.3, 4)
	if current.BBox.X < first.BBox.X-maxXGap {
		return false
	}
	if utils.IsLabeledLine(current) {
		return false
	}
	if len(current.Chunks) > 0 && current.Chunks[0] != nil && current.Chunks[0].IsHidden {
		return false
	}
	if previous != first {
		if current.BBox.X+maxXGap < previous.BBox.X {
			return false
		}
		return true
	}
	lineGap := abs(previous.Baseline - current.Baseline)
	nextGap := max(current.BBox.Height, previous.BBox.Height)
	if nextGap <= 0 {
		return true
	}
	return lineGap <= baselineDiffThreshold*nextGap || current.BBox.X > first.BBox.X+first.BBox.Height
}

func isBulletLine(line *entities.TextLine) bool {
	if line == nil {
		return false
	}
	text := strings.TrimSpace(line.GetText())
	if text == "" {
		return false
	}
	if unorderedBulletPattern.MatchString(text) || orderedBulletPattern.MatchString(text) {
		return true
	}
	return utils.IsOrderedBullet(text) || utils.IsUnorderedBullet(text) || line.LineArtBullet != nil || koreanAttachPattern.MatchString(text)
}

func trailingList(elements []entities.IObject) (*entities.PDFList, int) {
	for i := len(elements) - 1; i >= 0; i-- {
		if list, ok := elements[i].(*entities.PDFList); ok {
			return list, i
		}
	}
	return nil, -1
}

func leadingList(elements []entities.IObject) (*entities.PDFList, int) {
	for i, element := range elements {
		if list, ok := element.(*entities.PDFList); ok {
			return list, i
		}
	}
	return nil, -1
}

func sameListPattern(left, right *entities.PDFList) bool {
	if left == nil || right == nil || left.IsOrdered != right.IsOrdered {
		return false
	}
	if len(left.Items) == 0 || len(right.Items) == 0 {
		return false
	}
	return bulletPatternKey(left.Items[0].BulletText, left.IsOrdered) == bulletPatternKey(right.Items[0].BulletText, right.IsOrdered)
}

func bulletPatternKey(text string, ordered bool) string {
	value := strings.TrimSpace(text)
	switch {
	case ordered:
		switch {
		case regexp.MustCompile(`^\d+\.$`).MatchString(value):
			return "ordered-numeric"
		case regexp.MustCompile(`^\([A-Za-z0-9]+\)$`).MatchString(value):
			return "ordered-paren"
		case regexp.MustCompile(`^[A-Za-z]\.$`).MatchString(value):
			return "ordered-alpha"
		default:
			return "ordered-other"
		}
	default:
		if value == "" {
			return "unordered-empty"
		}
		r := []rune(value)
		return "unordered-" + string(r[0])
	}
}

func splitBulletText(text string) (string, string) {
	text = strings.TrimSpace(text)
	if match := orderedBulletPattern.FindString(text); match != "" {
		return strings.TrimSpace(match), strings.TrimSpace(strings.TrimPrefix(text, match))
	}
	if match := unorderedBulletPattern.FindString(text); match != "" {
		return strings.TrimSpace(match), strings.TrimSpace(strings.TrimPrefix(text, match))
	}
	if koreanAttachPattern.MatchString(text) {
		match := koreanAttachPattern.FindString(text)
		return strings.TrimSpace(match), strings.TrimSpace(strings.TrimPrefix(text, match))
	}
	fields := strings.Fields(text)
	if len(fields) > 1 {
		return fields[0], strings.Join(fields[1:], " ")
	}
	return text, ""
}

func cloneLineWithText(line *entities.TextLine, text string) *entities.TextLine {
	if line == nil {
		return nil
	}
	cloned := *line
	if len(line.Chunks) == 0 {
		return &cloned
	}
	chunks := make([]*entities.TextChunk, len(line.Chunks))
	for i, chunk := range line.Chunks {
		if chunk == nil {
			continue
		}
		c := *chunk
		if i == 0 {
			c.Text = text
		} else {
			c.Text = ""
		}
		chunks[i] = &c
	}
	cloned.Chunks = chunks
	return &cloned
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
