// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors

import (
	"math"
	"sort"
	"strings"
	"unicode/utf8"

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
	remainders := make([]entities.IObject, 0)

	for _, table := range tables {
		if !isMeaningfulBorderTable(table) {
			continue
		}

		cellContents := make(map[[2]int][]entities.IObject)
		for idx, element := range elements {
			if used[idx] || !bboxIntersects(table.bbox, element.GetBBox()) {
				continue
			}

			switch typed := element.(type) {
			case *entities.TextChunk:
				assigned, outside := splitTextChunkAcrossTable(table, typed)
				if len(assigned) == 0 {
					continue
				}
				used[idx] = true
				for _, piece := range outside {
					if piece != nil {
						remainders = append(remainders, piece)
					}
				}
				for cellKey, contents := range assigned {
					cellContents[cellKey] = append(cellContents[cellKey], contents...)
				}
			default:
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
	result = append(result, remainders...)

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

	rowBounds := collectHorizontalBounds(horizontal, minX, maxX, minY, maxY)
	colBounds := collectVerticalBounds(vertical, minX, maxX, minY, maxY)
	if !hasInternalDivider(rowBounds, maxY, minY) && !hasInternalDivider(colBounds, minX, maxX) {
		return detectedTable{}, false
	}
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
	if len(candidates) <= 1 {
		return candidates
	}

	sorted := append([]detectedTable(nil), candidates...)
	sort.SliceStable(sorted, func(i, j int) bool {
		leftArea := bboxArea(sorted[i].bbox)
		rightArea := bboxArea(sorted[j].bbox)
		if !areClose(leftArea, rightArea, tableAlignmentTolerance) {
			return leftArea > rightArea
		}
		leftGrid := tableGridSize(sorted[i])
		rightGrid := tableGridSize(sorted[j])
		if leftGrid != rightGrid {
			return leftGrid > rightGrid
		}
		if !areClose(sorted[i].bbox.Y, sorted[j].bbox.Y, tableAlignmentTolerance) {
			return sorted[i].bbox.Y > sorted[j].bbox.Y
		}
		return sorted[i].bbox.X < sorted[j].bbox.X
	})

	result := make([]detectedTable, 0, len(sorted))
	for _, candidate := range sorted {
		contained := false
		for _, kept := range result {
			if sameBBox(candidate.bbox, kept.bbox) || bboxSubsetOf(candidate.bbox, kept.bbox) {
				contained = true
				break
			}
		}
		if !contained {
			result = append(result, candidate)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		if !areClose(result[i].bbox.Y, result[j].bbox.Y, tableAlignmentTolerance) {
			return result[i].bbox.Y > result[j].bbox.Y
		}
		return result[i].bbox.X < result[j].bbox.X
	})
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
		if bboxStrictlyContains(outer, line.GetBBox()) {
			result = append(result, line)
		}
	}
	return result
}

func filterLineArtsForCell(lineArts []*entities.LineArtChunk, cell entities.BoundingBox) []*entities.LineArtChunk {
	result := make([]*entities.LineArtChunk, 0)
	for _, line := range lineArts {
		if bboxStrictlyContains(cell, line.GetBBox()) {
			result = append(result, line)
		}
	}
	return result
}

func bboxStrictlyContains(outer, inner entities.BoundingBox) bool {
	if outer.Page != inner.Page {
		return false
	}
	return bboxLeft(inner) > bboxLeft(outer)+tableAlignmentTolerance &&
		bboxRight(inner) < bboxRight(outer)-tableAlignmentTolerance &&
		bboxBottom(inner) > bboxBottom(outer)+tableAlignmentTolerance &&
		bboxTop(inner) < bboxTop(outer)-tableAlignmentTolerance
}

func sameBBox(a, b entities.BoundingBox) bool {
	return areClose(a.X, b.X, tableAlignmentTolerance) &&
		areClose(a.Y, b.Y, tableAlignmentTolerance) &&
		areClose(a.Width, b.Width, tableAlignmentTolerance) &&
		areClose(a.Height, b.Height, tableAlignmentTolerance)
}

func bboxSubsetOf(inner, outer entities.BoundingBox) bool {
	if inner.Page != outer.Page || sameBBox(inner, outer) {
		return false
	}
	return bboxLeft(inner) >= bboxLeft(outer)-tableAlignmentTolerance &&
		bboxRight(inner) <= bboxRight(outer)+tableAlignmentTolerance &&
		bboxBottom(inner) >= bboxBottom(outer)-tableAlignmentTolerance &&
		bboxTop(inner) <= bboxTop(outer)+tableAlignmentTolerance
}

func hasInternalDivider(bounds []float64, outerA, outerB float64) bool {
	for _, value := range bounds {
		if !areClose(value, outerA, tableAlignmentTolerance) && !areClose(value, outerB, tableAlignmentTolerance) {
			return true
		}
	}
	return false
}

func isMeaningfulBorderTable(table detectedTable) bool {
	return len(table.rowBounds) >= 3 && len(table.colBounds) >= 3
}

func bboxArea(box entities.BoundingBox) float64 {
	return math.Abs(box.Width * box.Height)
}

func tableGridSize(table detectedTable) int {
	return (len(table.rowBounds) - 1) * (len(table.colBounds) - 1)
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

func splitTextChunkAcrossTable(table detectedTable, chunk *entities.TextChunk) (map[[2]int][]entities.IObject, []*entities.TextChunk) {
	assigned := make(map[[2]int][]entities.IObject)
	pieces := make([]*entities.TextChunk, 0)
	box := chunk.GetBBox()
	if !bboxIntersects(table.bbox, box) {
		return assigned, pieces
	}

	if before := sliceTextChunkByX(chunk, bboxLeft(box), bboxLeft(table.bbox)); before != nil {
		pieces = append(pieces, before)
	}
	if after := sliceTextChunkByX(chunk, bboxRight(table.bbox), bboxRight(box)); after != nil {
		pieces = append(pieces, after)
	}

	for row := 0; row < len(table.rowBounds)-1; row++ {
		cellBottom := table.rowBounds[row+1]
		cellTop := table.rowBounds[row]
		if bboxTop(box) < cellBottom-tableAlignmentTolerance || bboxBottom(box) > cellTop+tableAlignmentTolerance {
			continue
		}
		for col := 0; col < len(table.colBounds)-1; col++ {
			cellLeft := table.colBounds[col]
			cellRight := table.colBounds[col+1]
			piece := sliceTextChunkByX(chunk, cellLeft, cellRight)
			if piece == nil {
				continue
			}
			assigned[[2]int{row, col}] = append(assigned[[2]int{row, col}], piece)
		}
	}

	if len(assigned) == 0 && bboxContains(table.bbox, box) {
		if row, col, ok := assignToCell(table, box); ok {
			assigned[[2]int{row, col}] = append(assigned[[2]int{row, col}], chunk)
		}
	}
	return assigned, pieces
}

func sliceTextChunkByX(chunk *entities.TextChunk, startX, endX float64) *entities.TextChunk {
	box := chunk.GetBBox()
	segmentLeft := math.Max(bboxLeft(box), startX)
	segmentRight := math.Min(bboxRight(box), endX)
	if segmentRight-segmentLeft <= tableAlignmentTolerance/10 {
		return nil
	}

	runes := []rune(chunk.Text)
	if len(runes) == 0 {
		return nil
	}

	charWidth := box.Width / float64(len(runes))
	if charWidth <= 0 {
		return nil
	}

	startIdx := int(math.Floor((segmentLeft - bboxLeft(box)) / charWidth))
	endIdx := int(math.Ceil((segmentRight - bboxLeft(box)) / charWidth))
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(runes) {
		endIdx = len(runes)
	}
	if startIdx >= endIdx {
		return nil
	}

	text := string(runes[startIdx:endIdx])
	leftTrim := len(text) - len(strings.TrimLeft(text, " \t\r\n"))
	rightTrim := len(text) - len(strings.TrimRight(text, " \t\r\n"))
	startIdx += utf8.RuneCountInString(text[:leftTrim])
	endIdx -= utf8.RuneCountInString(text[len(text)-rightTrim:])
	if startIdx >= endIdx {
		return nil
	}
	text = string(runes[startIdx:endIdx])
	if strings.TrimSpace(text) == "" {
		return nil
	}

	left := bboxLeft(box) + float64(startIdx)*charWidth
	right := bboxLeft(box) + float64(endIdx)*charWidth
	if right <= left {
		return nil
	}

	return &entities.TextChunk{
		BaseObject: entities.BaseObject{
			ID: chunk.GetID(),
			BBox: entities.BoundingBox{
				X:      left,
				Y:      box.Y,
				Width:  right - left,
				Height: box.Height,
				Page:   box.Page,
			},
		},
		Text:            text,
		FontStyle:       chunk.FontStyle,
		Baseline:        chunk.Baseline,
		CharSpacing:     chunk.CharSpacing,
		IsHidden:        chunk.IsHidden,
		IsHiddenOCG:     chunk.IsHiddenOCG,
		IsOffPage:       chunk.IsOffPage,
		IsTiny:          chunk.IsTiny,
		IsStrikethrough: chunk.IsStrikethrough,
	}
}
