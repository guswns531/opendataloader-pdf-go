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
	orderedBulletPattern   = regexp.MustCompile(`^\s*(?:\(?\d+\)|\d+[\.\)]|[가-힣A-Za-z][\.\)]|[Ⅰ-Ⅻⅰ-ⅻ①-⑳⑴-⒇⒈-⒛❶-❿➀-➉➊-➓])\s*`)
	unorderedBulletPattern = regexp.MustCompile(`^\s*[∘*+\-•‣∙▪■□▢▣▤▥▦▧▨▩◆◇○●◦◯☐☑☒✓✔❍❏❐❑⬛⬜⭐]\s*`)
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

		if len(list.Items) == 1 {
			out = append(out, list.Items[0].Content...)
			continue
		}
		out = append(out, list)
	}
	return out
}

func (p *ListProcessor) consumeListItem(elements []entities.IObject, start int, isOrdered bool, ctx *containers.ProcessorContext) (*entities.ListItem, int) {
	line := elements[start].(*entities.TextLine)
	bulletText, bodyText := splitBulletText(strings.TrimSpace(line.GetText()))
	bodyLine := cloneLineWithText(line, bodyText)
	content := []entities.IObject{bodyLine}
	box := bodyLine.BBox
	next := start + 1

	for next < len(elements) {
		nextLine, ok := elements[next].(*entities.TextLine)
		if !ok || isBulletLine(nextLine) {
			break
		}
		if !isListContinuation(line, nextLine) {
			break
		}
		content = append(content, nextLine)
		box = unionBox(box, nextLine.BBox)
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

func isListContinuation(first, current *entities.TextLine) bool {
	if current == nil || first == nil {
		return false
	}
	if current.BBox.X+0.01 < first.BBox.X {
		return false
	}
	lineGap := abs(current.Baseline - first.Baseline)
	height := max(first.BBox.Height, current.BBox.Height)
	if height == 0 {
		return true
	}
	return lineGap <= baselineDiffThreshold*height || current.BBox.X > first.BBox.X+first.BBox.Height
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
