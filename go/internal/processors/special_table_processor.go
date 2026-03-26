// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors

import (
	"regexp"
	"sort"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

type SpecialTableProcessor struct {
	AbstractTableProcessor
}

var koreanSpecialTablePattern = regexp.MustCompile(`^\(?(수신|경유|제목)\)?.*`)

func (p *SpecialTableProcessor) Process(elements []entities.IObject, ctx *containers.ProcessorContext) []entities.IObject {
	koreanDetected := detectSpecialKoreanTables(elements, ctx)
	if len(koreanDetected) > 0 {
		elements = koreanDetected
	}

	lines := collectTextLines(elements)
	if len(lines) < 2 {
		return append([]entities.IObject(nil), elements...)
	}

	tables, consumed := detectAlignmentTables(lines, ctx)
	if len(tables) == 0 {
		return append([]entities.IObject(nil), elements...)
	}

	used := make(map[string]struct{})
	result := make([]entities.IObject, 0, len(elements))
	for _, table := range tables {
		result = append(result, table)
	}
	for _, line := range consumed {
		if line != nil {
			used[line.GetID()] = struct{}{}
		}
	}
	for _, element := range elements {
		if _, ok := used[element.GetID()]; !ok {
			result = append(result, element)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		left := result[i].GetBBox()
		right := result[j].GetBBox()
		if !areClose(left.Y, right.Y, tableAlignmentTolerance) {
			return left.Y > right.Y
		}
		return left.X < right.X
	})
	return result
}

func detectSpecialKoreanTables(elements []entities.IObject, ctx *containers.ProcessorContext) []entities.IObject {
	result := append([]entities.IObject(nil), elements...)
	lines := make([]*entities.TextLine, 0)
	startIndex := -1

	flush := func() {
		if len(lines) == 0 || startIndex < 0 {
			lines = nil
			startIndex = -1
			return
		}
		result[startIndex] = buildSpecialKoreanTable(lines, ctx)
		for idx := startIndex + 1; idx < startIndex+len(lines); idx++ {
			result[idx] = nil
		}
		lines = nil
		startIndex = -1
	}

	for idx, element := range result {
		line, ok := element.(*entities.TextLine)
		if !ok {
			flush()
			continue
		}
		if koreanSpecialTablePattern.MatchString(strings.TrimSpace(line.GetText())) {
			if startIndex < 0 {
				startIndex = idx
			}
			lines = append(lines, line)
			continue
		}
		flush()
	}
	flush()

	filtered := make([]entities.IObject, 0, len(result))
	for _, element := range result {
		if element != nil {
			filtered = append(filtered, element)
		}
	}
	return filtered
}

func collectTextLines(elements []entities.IObject) []*entities.TextLine {
	lines := make([]*entities.TextLine, 0)
	for _, element := range elements {
		if line, ok := element.(*entities.TextLine); ok && strings.TrimSpace(line.GetText()) != "" {
			lines = append(lines, line)
		}
	}
	sort.SliceStable(lines, func(i, j int) bool {
		if !areClose(lines[i].GetBBox().Y, lines[j].GetBBox().Y, tableAlignmentTolerance) {
			return lines[i].GetBBox().Y > lines[j].GetBBox().Y
		}
		return lines[i].GetBBox().X < lines[j].GetBBox().X
	})
	return lines
}

func detectAlignmentTables(lines []*entities.TextLine, ctx *containers.ProcessorContext) ([]*entities.SemanticTable, []*entities.TextLine) {
	tables := make([]*entities.SemanticTable, 0)
	consumed := make([]*entities.TextLine, 0)
	var group []*entities.TextLine
	var signature []float64

	flush := func() {
		if len(group) < 2 || len(signature) < 2 {
			group = nil
			signature = nil
			return
		}
		if table := buildAlignedTextTable(group, signature, ctx); table != nil && len(table.Rows) >= 2 {
			tables = append(tables, table)
			consumed = append(consumed, group...)
		}
		group = nil
		signature = nil
	}

	for _, line := range lines {
		positions := lineColumnPositions(line)
		if len(positions) < 2 {
			flush()
			continue
		}
		if len(group) == 0 {
			group = append(group, line)
			signature = positions
			continue
		}
		if columnPositionsMatch(signature, positions) {
			group = append(group, line)
			signature = mergePositions(signature, positions)
			continue
		}
		flush()
		group = append(group, line)
		signature = positions
	}
	flush()

	return tables, consumed
}

func lineColumnPositions(line *entities.TextLine) []float64 {
	if len(line.Chunks) < 2 {
		return nil
	}
	values := make([]float64, 0, len(line.Chunks))
	for _, chunk := range line.Chunks {
		if chunk == nil || isWhitespaceChunk(chunk) {
			continue
		}
		values = append(values, chunk.GetBBox().X)
	}
	return uniqueSorted(values, tableAlignmentTolerance, false)
}

func columnPositionsMatch(expected, actual []float64) bool {
	if len(expected) != len(actual) {
		return false
	}
	for idx := range expected {
		if !areClose(expected[idx], actual[idx], tableAlignmentTolerance) {
			return false
		}
	}
	return true
}

func mergePositions(a, b []float64) []float64 {
	merged := make([]float64, len(a))
	for idx := range a {
		merged[idx] = (a[idx] + b[idx]) / 2
	}
	return merged
}

func buildAlignedTextTable(lines []*entities.TextLine, positions []float64, ctx *containers.ProcessorContext) *entities.SemanticTable {
	rowBounds := make([]float64, 0, len(lines)+1)
	colBounds := append([]float64(nil), positions...)

	for rowIdx, line := range lines {
		top := bboxTop(line.GetBBox())
		bottom := line.GetBBox().Y
		if rowIdx == 0 {
			rowBounds = append(rowBounds, top)
		}
		rowBounds = append(rowBounds, bottom)
	}

	lastRight := bboxRight(lines[0].GetBBox())
	for _, line := range lines[1:] {
		lastRight = max(lastRight, bboxRight(line.GetBBox()))
	}
	colBounds = append(colBounds, lastRight)
	colBounds = uniqueSorted(colBounds, tableAlignmentTolerance, false)
	if len(colBounds) < 3 {
		return nil
	}

	cells := make(map[[2]int][]entities.IObject)
	for rowIdx, line := range lines {
		for _, chunk := range line.Chunks {
			if chunk == nil || isWhitespaceChunk(chunk) {
				continue
			}
			col := nearestColumn(colBounds, chunk.GetBBox().X)
			if col < 0 {
				continue
			}
			cells[[2]int{rowIdx, col}] = append(cells[[2]int{rowIdx, col}], chunk)
		}
	}
	return buildTableFromGrid(rowBounds, colBounds, cells, lines[0].GetBBox().Page, ctx)
}

func buildSpecialKoreanTable(lines []*entities.TextLine, ctx *containers.ProcessorContext) *entities.SemanticTable {
	rows := make([]*entities.TableRow, 0, len(lines))
	var tableBox entities.BoundingBox
	for rowIdx, line := range lines {
		text := []rune(line.GetText())
		colon := -1
		for idx, r := range text {
			if r == ':' {
				colon = idx
				break
			}
		}
		row := &entities.TableRow{
			Cells: make([]*entities.TableCell, 0, 2),
			BBox:  line.GetBBox(),
		}
		if colon < 0 {
			cell := entities.NewTableCell(rowIdx, 0, 1, 2, line.GetBBox(), []entities.IObject{line})
			row.Cells = append(row.Cells, cell)
		} else {
			left := textChunkSliceFromLine(line, 0, colon)
			right := textChunkSliceFromLine(line, colon+1, len(text))
			row.Cells = append(row.Cells,
				entities.NewTableCell(rowIdx, 0, 1, 1, textBBoxOrLine(left, line.GetBBox()), objectsForTextChunk(left)),
				entities.NewTableCell(rowIdx, 1, 1, 1, textBBoxOrLine(right, line.GetBBox()), objectsForTextChunk(right)),
			)
		}
		rows = append(rows, row)
		if rowIdx == 0 {
			tableBox = line.GetBBox()
		} else {
			tableBox = unionBBox(tableBox, line.GetBBox())
		}
	}
	return normalizeTable(&entities.SemanticTable{
		BaseObject: entities.BaseObject{
			ID:   nextObjectID(ctx),
			BBox: tableBox,
		},
		Rows: rows,
	}, ctx)
}

func textChunkSliceFromLine(line *entities.TextLine, start, end int) *entities.TextChunk {
	text := []rune(line.GetText())
	if start < 0 {
		start = 0
	}
	if end > len(text) {
		end = len(text)
	}
	if start >= end {
		return nil
	}
	box := line.GetBBox()
	charWidth := box.Width / float64(len(text))
	if charWidth <= 0 {
		charWidth = box.Width
	}
	value := strings.TrimSpace(string(text[start:end]))
	if value == "" {
		return nil
	}
	leftOffset := float64(start) * charWidth
	width := float64(end-start) * charWidth
	return &entities.TextChunk{
		BaseObject: entities.BaseObject{
			BBox: entities.BoundingBox{
				X:      box.X + leftOffset,
				Y:      box.Y,
				Width:  width,
				Height: box.Height,
				Page:   box.Page,
			},
		},
		Text:     value,
		Baseline: line.Baseline,
	}
}

func objectsForTextChunk(chunk *entities.TextChunk) []entities.IObject {
	if chunk == nil {
		return nil
	}
	return []entities.IObject{chunk}
}

func textBBoxOrLine(chunk *entities.TextChunk, fallback entities.BoundingBox) entities.BoundingBox {
	if chunk == nil {
		return fallback
	}
	return chunk.GetBBox()
}

func nearestColumn(bounds []float64, x float64) int {
	if len(bounds) < 2 {
		return -1
	}

	bestIdx := -1
	bestDistance := 0.0
	for idx := 0; idx < len(bounds)-1; idx++ {
		distance := x - bounds[idx]
		if distance < -tableAlignmentTolerance {
			continue
		}
		if bestIdx == -1 || distance < bestDistance {
			bestIdx = idx
			bestDistance = distance
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}
	return -1
}
