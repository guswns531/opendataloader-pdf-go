// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors

import (
	"sort"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

type SpecialTableProcessor struct {
	AbstractTableProcessor
}

func (p *SpecialTableProcessor) Process(elements []entities.IObject, ctx *containers.ProcessorContext) []entities.IObject {
	lines := collectTextLines(elements)
	if len(lines) < 2 {
		return append([]entities.IObject(nil), elements...)
	}

	tables := detectAlignmentTables(lines, ctx)
	if len(tables) == 0 {
		return append([]entities.IObject(nil), elements...)
	}

	used := make(map[string]struct{})
	result := make([]entities.IObject, 0, len(elements))
	for _, table := range tables {
		result = append(result, table)
		for _, row := range table.Rows {
			for _, cell := range row.Cells {
				for _, content := range cell.Content {
					used[content.GetID()] = struct{}{}
				}
			}
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

func detectAlignmentTables(lines []*entities.TextLine, ctx *containers.ProcessorContext) []*entities.SemanticTable {
	tables := make([]*entities.SemanticTable, 0)
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

	return tables
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
	cells := make(map[[2]int][]entities.IObject)

	for rowIdx, line := range lines {
		top := bboxTop(line.GetBBox())
		bottom := line.GetBBox().Y
		if rowIdx == 0 {
			rowBounds = append(rowBounds, top)
		}
		rowBounds = append(rowBounds, bottom)
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

	lastRight := bboxRight(lines[0].GetBBox())
	for _, line := range lines[1:] {
		lastRight = max(lastRight, bboxRight(line.GetBBox()))
	}
	colBounds = append(colBounds, lastRight)
	colBounds = uniqueSorted(colBounds, tableAlignmentTolerance, false)
	if len(colBounds) < 3 {
		return nil
	}
	return buildTableFromGrid(rowBounds, colBounds, cells, lines[0].GetBBox().Page, ctx)
}

func nearestColumn(bounds []float64, x float64) int {
	for idx := 0; idx < len(bounds)-1; idx++ {
		if x >= bounds[idx]-tableAlignmentTolerance && x <= bounds[idx+1]+tableAlignmentTolerance {
			return idx
		}
	}
	if len(bounds) == 0 {
		return -1
	}
	if x >= bounds[len(bounds)-1]-tableAlignmentTolerance {
		return len(bounds) - 2
	}
	return -1
}
