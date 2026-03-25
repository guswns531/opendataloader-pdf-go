// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors

import (
	"sort"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const maxTableDepth = 10

type TableBorderProcessor struct{}

type detectedTable struct {
	bbox      entities.BoundingBox
	rowBounds []float64
	colBounds []float64
}

func (p *TableBorderProcessor) Process(elements []entities.IObject, lineArts []*entities.LineArtChunk, ctx *containers.ProcessorContext) []entities.IObject {
	return p.processNode(elements, lineArts, ctx, 0)
}

func (p *TableBorderProcessor) processNode(elements []entities.IObject, lineArts []*entities.LineArtChunk, ctx *containers.ProcessorContext, depth int) []entities.IObject {
	if depth >= maxTableDepth {
		return append([]entities.IObject(nil), elements...)
	}

	tables := detectBorderTables(lineArts)
	if len(tables) == 0 {
		if fallback, ok := detectFallbackEnclosingTable(lineArts); ok {
			tables = []detectedTable{fallback}
		} else {
			return append([]entities.IObject(nil), elements...)
		}
	}

	used := make([]bool, len(elements))
	result := make([]entities.IObject, 0, len(elements))

	for _, table := range tables {
		cellContents := make(map[[2]int][]entities.IObject)
		for idx, element := range elements {
			if !bboxContains(table.bbox, element.GetBBox()) {
				continue
			}
			used[idx] = true
			row, col, ok := assignToCell(table, element.GetBBox())
			if !ok {
				continue
			}
			cellContents[[2]int{row, col}] = append(cellContents[[2]int{row, col}], element)
		}

		innerLineArts := filterNestedLineArts(lineArts, table.bbox)
		semanticTable := buildTableFromGrid(table.rowBounds, table.colBounds, cellContents, table.bbox.Page, ctx)
		if semanticTable == nil {
			continue
		}

		for rowIdx, row := range semanticTable.Rows {
			for colIdx, cell := range row.Cells {
				cellLineArts := filterLineArtsForCell(innerLineArts, cell.BBox)
				if len(cellLineArts) == 0 {
					continue
				}
				row.Cells[colIdx].Content = p.processNode(cell.Content, cellLineArts, ctx, depth+1)
				if nested, ok := firstSemanticTable(row.Cells[colIdx].Content); ok {
					row.Cells[colIdx].BBox = unionBBox(row.Cells[colIdx].BBox, nested.GetBBox())
				}
			}
			semanticTable.Rows[rowIdx].BBox = unionCells(row.Cells)
		}
		semanticTable.BBox = unionRows(semanticTable.Rows)
		result = append(result, semanticTable)
	}

	for idx, element := range elements {
		if !used[idx] {
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

func detectBorderTables(lineArts []*entities.LineArtChunk) []detectedTable {
	horizontal := make([]*entities.LineArtChunk, 0)
	vertical := make([]*entities.LineArtChunk, 0)
	for _, line := range lineArts {
		box := line.GetBBox()
		width := bboxRight(box) - bboxLeft(box)
		height := bboxTop(box) - bboxBottom(box)
		switch {
		case line.IsHorizontal || width >= height:
			horizontal = append(horizontal, line)
		case line.IsVertical || height > width:
			vertical = append(vertical, line)
		}
	}

	var candidates []detectedTable
	for _, left := range vertical {
		for _, right := range vertical {
			if left == right || left.GetBBox().Page != right.GetBBox().Page {
				continue
			}
			leftX := bboxLeft(left.GetBBox())
			rightX := bboxLeft(right.GetBBox())
			if rightX <= leftX+tableAlignmentTolerance {
				continue
			}
			for _, topLine := range horizontal {
				if topLine.GetBBox().Page != left.GetBBox().Page {
					continue
				}
				topY := bboxBottom(topLine.GetBBox())
				if !horizontalSpans(topLine, leftX, rightX) || !verticalTouches(left, topY) || !verticalTouches(right, topY) {
					continue
				}
				for _, bottomLine := range horizontal {
					if bottomLine == topLine || bottomLine.GetBBox().Page != left.GetBBox().Page {
						continue
					}
					bottomY := bboxBottom(bottomLine.GetBBox())
					if topY <= bottomY+tableAlignmentTolerance {
						continue
					}
					if !horizontalSpans(bottomLine, leftX, rightX) || !verticalTouches(left, bottomY) || !verticalTouches(right, bottomY) {
						continue
					}

					rowBounds := collectHorizontalBounds(horizontal, leftX, rightX, bottomY, topY)
					colBounds := collectVerticalBounds(vertical, leftX, rightX, bottomY, topY)
					if len(rowBounds) < 2 || len(colBounds) < 2 {
						continue
					}

					candidate := detectedTable{
						bbox: entities.BoundingBox{
							X:      leftX,
							Y:      bottomY,
							Width:  rightX - leftX,
							Height: topY - bottomY,
							Page:   left.GetBBox().Page,
						},
						rowBounds: rowBounds,
						colBounds: colBounds,
					}
					if !containsDetectedTable(candidates, candidate) {
						candidates = append(candidates, candidate)
					}
				}
			}
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if !areClose(candidates[i].bbox.Y, candidates[j].bbox.Y, tableAlignmentTolerance) {
			return candidates[i].bbox.Y > candidates[j].bbox.Y
		}
		return candidates[i].bbox.X < candidates[j].bbox.X
	})
	return topLevelTables(candidates)
}

func detectFallbackEnclosingTable(lineArts []*entities.LineArtChunk) (detectedTable, bool) {
	horizontal := make([]*entities.LineArtChunk, 0)
	vertical := make([]*entities.LineArtChunk, 0)
	for _, line := range lineArts {
		box := line.GetBBox()
		width := bboxRight(box) - bboxLeft(box)
		height := bboxTop(box) - bboxBottom(box)
		switch {
		case line.IsHorizontal || width >= height:
			horizontal = append(horizontal, line)
		case line.IsVertical || height > width:
			vertical = append(vertical, line)
		}
	}
	if len(horizontal) < 2 || len(vertical) < 2 {
		return detectedTable{}, false
	}

	minX, maxX := bboxLeft(vertical[0].GetBBox()), bboxLeft(vertical[0].GetBBox())
	minY, maxY := bboxBottom(horizontal[0].GetBBox()), bboxBottom(horizontal[0].GetBBox())
	page := horizontal[0].GetBBox().Page
	for _, line := range vertical[1:] {
		x := bboxLeft(line.GetBBox())
		minX = min(minX, x)
		maxX = max(maxX, x)
	}
	for _, line := range horizontal[1:] {
		y := bboxBottom(line.GetBBox())
		minY = min(minY, y)
		maxY = max(maxY, y)
	}
	if maxX-minX <= tableAlignmentTolerance || maxY-minY <= tableAlignmentTolerance {
		return detectedTable{}, false
	}

	rowBounds := uniqueSorted([]float64{maxY, minY}, tableAlignmentTolerance, true)
	colBounds := uniqueSorted([]float64{minX, maxX}, tableAlignmentTolerance, false)
	if len(rowBounds) < 2 || len(colBounds) < 2 {
		return detectedTable{}, false
	}

	return detectedTable{
		bbox: entities.BoundingBox{
			X:      minX,
			Y:      minY,
			Width:  maxX - minX,
			Height: maxY - minY,
			Page:   page,
		},
		rowBounds: rowBounds,
		colBounds: colBounds,
	}, true
}

func containsDetectedTable(existing []detectedTable, candidate detectedTable) bool {
	for _, item := range existing {
		if areClose(item.bbox.X, candidate.bbox.X, tableAlignmentTolerance) &&
			areClose(item.bbox.Y, candidate.bbox.Y, tableAlignmentTolerance) &&
			areClose(item.bbox.Width, candidate.bbox.Width, tableAlignmentTolerance) &&
			areClose(item.bbox.Height, candidate.bbox.Height, tableAlignmentTolerance) {
			return true
		}
	}
	return false
}

func topLevelTables(candidates []detectedTable) []detectedTable {
	result := make([]detectedTable, 0, len(candidates))
	for idx, candidate := range candidates {
		contained := false
		for jdx, other := range candidates {
			if idx == jdx {
				continue
			}
			if sameBBox(candidate.bbox, other.bbox) {
				continue
			}
			if bboxContains(other.bbox, candidate.bbox) {
				contained = true
				break
			}
		}
		if !contained {
			result = append(result, candidate)
		}
	}
	return result
}

func findHorizontalBoundary(horizontal []*entities.LineArtChunk, leftX, rightX, minY, maxY float64) (*entities.LineArtChunk, *entities.LineArtChunk) {
	var topLine, bottomLine *entities.LineArtChunk
	for _, line := range horizontal {
		if !horizontalSpans(line, leftX, rightX) {
			continue
		}
		y := bboxBottom(line.GetBBox())
		if y < minY-tableAlignmentTolerance || y > maxY+tableAlignmentTolerance {
			continue
		}
		if topLine == nil || y > bboxBottom(topLine.GetBBox()) {
			topLine = line
		}
		if bottomLine == nil || y < bboxBottom(bottomLine.GetBBox()) {
			bottomLine = line
		}
	}
	return topLine, bottomLine
}

func collectHorizontalBounds(horizontal []*entities.LineArtChunk, leftX, rightX, minY, maxY float64) []float64 {
	values := make([]float64, 0)
	for _, line := range horizontal {
		if !horizontalSpans(line, leftX, rightX) {
			continue
		}
		y := bboxBottom(line.GetBBox())
		if y < minY-tableAlignmentTolerance || y > maxY+tableAlignmentTolerance {
			continue
		}
		values = append(values, y)
	}
	return uniqueSorted(values, tableAlignmentTolerance, true)
}

func collectVerticalBounds(vertical []*entities.LineArtChunk, minX, maxX, bottomY, topY float64) []float64 {
	values := make([]float64, 0)
	for _, line := range vertical {
		box := line.GetBBox()
		x := bboxLeft(box)
		if x < minX-tableAlignmentTolerance || x > maxX+tableAlignmentTolerance {
			continue
		}
		if bboxBottom(box) > bottomY+tableAlignmentTolerance || bboxTop(box) < topY-tableAlignmentTolerance {
			continue
		}
		values = append(values, x)
	}
	return uniqueSorted(values, tableAlignmentTolerance, false)
}

func horizontalSpans(line *entities.LineArtChunk, leftX, rightX float64) bool {
	box := line.GetBBox()
	return bboxLeft(box) <= leftX+tableAlignmentTolerance && bboxRight(box) >= rightX-tableAlignmentTolerance
}

func verticalTouches(line *entities.LineArtChunk, y float64) bool {
	box := line.GetBBox()
	return bboxBottom(box) <= y+tableAlignmentTolerance && bboxTop(box) >= y-tableAlignmentTolerance
}

func assignToCell(table detectedTable, box entities.BoundingBox) (int, int, bool) {
	centerX := bboxCenterX(box)
	centerY := bboxCenterY(box)
	for row := 0; row < len(table.rowBounds)-1; row++ {
		if centerY <= table.rowBounds[row]+tableAlignmentTolerance && centerY >= table.rowBounds[row+1]-tableAlignmentTolerance {
			for col := 0; col < len(table.colBounds)-1; col++ {
				if centerX >= table.colBounds[col]-tableAlignmentTolerance && centerX <= table.colBounds[col+1]+tableAlignmentTolerance {
					return row, col, true
				}
			}
		}
	}
	return 0, 0, false
}

func filterNestedLineArts(lineArts []*entities.LineArtChunk, outer entities.BoundingBox) []*entities.LineArtChunk {
	result := make([]*entities.LineArtChunk, 0)
	for _, line := range lineArts {
		if bboxContains(outer, line.GetBBox()) && !sameBBox(outer, line.GetBBox()) {
			result = append(result, line)
		}
	}
	return result
}

func filterLineArtsForCell(lineArts []*entities.LineArtChunk, cell entities.BoundingBox) []*entities.LineArtChunk {
	result := make([]*entities.LineArtChunk, 0)
	for _, line := range lineArts {
		if bboxContains(cell, line.GetBBox()) && !sameBBox(cell, line.GetBBox()) {
			result = append(result, line)
		}
	}
	return result
}

func sameBBox(a, b entities.BoundingBox) bool {
	return areClose(a.X, b.X, tableAlignmentTolerance) &&
		areClose(a.Y, b.Y, tableAlignmentTolerance) &&
		areClose(a.Width, b.Width, tableAlignmentTolerance) &&
		areClose(a.Height, b.Height, tableAlignmentTolerance)
}

func firstSemanticTable(contents []entities.IObject) (*entities.SemanticTable, bool) {
	for _, content := range contents {
		table, ok := content.(*entities.SemanticTable)
		if ok {
			return table, true
		}
	}
	return nil, false
}
