// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors

import (
	"sort"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const (
	suspiciousYDifferenceEpsilon = 0.1
	suspiciousXDifferenceEpsilon = 3.0
	tableAlignmentTolerance      = 3.0
)

type TableProcessor interface {
	Process(elements []entities.IObject, ctx *containers.ProcessorContext) []entities.IObject
}

type AbstractTableProcessor struct{}

func getPagesWithPossibleTables(contents [][]entities.IObject) []int {
	pageNumbers := make([]int, 0)
	for pageNumber, pageContents := range contents {
		if hasPossibleTable(pageContents) {
			pageNumbers = append(pageNumbers, pageNumber)
		}
	}
	return pageNumbers
}

func hasPossibleTable(contents []entities.IObject) bool {
	var previous *entities.TextChunk
	for _, content := range contents {
		current, ok := content.(*entities.TextChunk)
		if !ok || isWhitespaceChunk(current) {
			continue
		}
		if previous != nil && areSuspiciousTextChunks(previous, current) {
			return true
		}
		previous = current
	}
	return false
}

func areSuspiciousTextChunks(previous, current *entities.TextChunk) bool {
	if bboxTop(previous.GetBBox()) < bboxBottom(current.GetBBox()) {
		return true
	}
	if areClose(previous.Baseline, current.Baseline, current.GetBBox().Height*suspiciousYDifferenceEpsilon) {
		if bboxLeft(current.GetBBox())-bboxRight(previous.GetBBox()) > current.GetBBox().Height*suspiciousXDifferenceEpsilon {
			return true
		}
	}
	return false
}

func normalizeTable(table *entities.SemanticTable, ctx *containers.ProcessorContext) *entities.SemanticTable {
	if table == nil {
		return nil
	}

	table.Rows = compactRows(table.Rows)
	for _, row := range table.Rows {
		row.Cells = compactCells(row.Cells)
		row.BBox = unionCells(row.Cells)
		for _, cell := range row.Cells {
			cell.Content = normalizeCellContent(cell.Content)
		}
	}
	table.BBox = unionRows(table.Rows)
	if table.ID == "" && ctx != nil {
		table.ID = ctx.NextID()
	}
	return table
}

func compactRows(rows []*entities.TableRow) []*entities.TableRow {
	compacted := make([]*entities.TableRow, 0, len(rows))
	for _, row := range rows {
		if row == nil || len(row.Cells) == 0 {
			continue
		}
		if isEmptyRow(row) && len(compacted) > 0 {
			grown := map[*entities.TableCell]struct{}{}
			for col, cell := range row.Cells {
				if cell == nil {
					continue
				}
				origin := resolveOriginCell(compacted, len(compacted)-1, col)
				if origin == nil {
					continue
				}
				if _, ok := grown[origin]; !ok {
					origin.Rowspan += cell.EffectiveRowSpan()
					grown[origin] = struct{}{}
				}
				origin.BBox = unionBBox(origin.BBox, cell.BBox)
				markCoveredCell(cell, origin)
			}
		}
		row.BBox = unionCells(row.Cells)
		compacted = append(compacted, row)
	}
	return compacted
}

func compactCells(cells []*entities.TableCell) []*entities.TableCell {
	for col, cell := range cells {
		if cell == nil {
			continue
		}
		if cell.Rowspan <= 0 {
			cell.Rowspan = 1
		}
		if cell.Colspan <= 0 {
			cell.Colspan = 1
		}
		if !cell.IsOriginCell {
			continue
		}
		if col > 0 && isEmptyCell(cell) {
			prev := resolveOriginCellInRow(cells, col-1)
			if prev != nil {
				prev.Colspan += cell.EffectiveColSpan()
				prev.BBox = unionBBox(prev.BBox, cell.BBox)
				markCoveredCell(cell, prev)
			}
		}
	}
	return cells
}

func normalizeCellContent(contents []entities.IObject) []entities.IObject {
	if len(contents) == 0 {
		return nil
	}
	sorted := append([]entities.IObject(nil), contents...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left := sorted[i].GetBBox()
		right := sorted[j].GetBBox()
		if !areClose(left.Y, right.Y, tableAlignmentTolerance) {
			return left.Y > right.Y
		}
		return left.X < right.X
	})

	result := make([]entities.IObject, 0, len(sorted))
	for _, content := range sorted {
		switch typed := content.(type) {
		case *entities.TextChunk:
			if typed.Text == "" || isWhitespaceChunk(typed) {
				continue
			}
		case *entities.TextLine:
			if strings.TrimSpace(typed.GetText()) == "" {
				continue
			}
		}
		result = append(result, content)
	}
	return result
}

func buildTableFromGrid(rowBounds, colBounds []float64, cells map[[2]int][]entities.IObject, page int, ctx *containers.ProcessorContext) *entities.SemanticTable {
	if len(rowBounds) < 2 || len(colBounds) < 2 {
		return nil
	}

	rows := make([]*entities.TableRow, 0, len(rowBounds)-1)
	for row := 0; row < len(rowBounds)-1; row++ {
		tableRow := &entities.TableRow{
			Cells: make([]*entities.TableCell, 0, len(colBounds)-1),
			BBox: entities.BoundingBox{
				X:      colBounds[0],
				Y:      rowBounds[row+1],
				Width:  colBounds[len(colBounds)-1] - colBounds[0],
				Height: rowBounds[row] - rowBounds[row+1],
				Page:   page,
			},
		}
		for col := 0; col < len(colBounds)-1; col++ {
			cellBox := entities.BoundingBox{
				X:      colBounds[col],
				Y:      rowBounds[row+1],
				Width:  colBounds[col+1] - colBounds[col],
				Height: rowBounds[row] - rowBounds[row+1],
				Page:   page,
			}
			tableRow.Cells = append(tableRow.Cells, entities.NewTableCell(
				row,
				col,
				1,
				1,
				cellBox,
				normalizeCellContent(cells[[2]int{row, col}]),
			))
		}
		rows = append(rows, tableRow)
	}

	table := &entities.SemanticTable{
		BaseObject: entities.BaseObject{
			BBox: entities.BoundingBox{
				X:      colBounds[0],
				Y:      rowBounds[len(rowBounds)-1],
				Width:  colBounds[len(colBounds)-1] - colBounds[0],
				Height: rowBounds[0] - rowBounds[len(rowBounds)-1],
				Page:   page,
			},
		},
		Rows: rows,
	}
	return normalizeTable(table, ctx)
}

func bboxLeft(box entities.BoundingBox) float64 { return bboxMin(box.X, box.X+box.Width) }

func bboxRight(box entities.BoundingBox) float64 { return bboxMax(box.X, box.X+box.Width) }

func bboxBottom(box entities.BoundingBox) float64 { return bboxMin(box.Y, box.Y+box.Height) }

func bboxTop(box entities.BoundingBox) float64 { return bboxMax(box.Y, box.Y+box.Height) }

func bboxCenterX(box entities.BoundingBox) float64 { return (bboxLeft(box) + bboxRight(box)) / 2 }

func bboxCenterY(box entities.BoundingBox) float64 { return (bboxBottom(box) + bboxTop(box)) / 2 }

func areClose(a, b, epsilon float64) bool {
	if a > b {
		return a-b <= epsilon
	}
	return b-a <= epsilon
}

func isWhitespaceChunk(chunk *entities.TextChunk) bool {
	return strings.TrimSpace(chunk.Text) == ""
}

func bboxContains(outer, inner entities.BoundingBox) bool {
	return inner.Page == outer.Page &&
		bboxLeft(inner) >= bboxLeft(outer)-tableAlignmentTolerance &&
		bboxRight(inner) <= bboxRight(outer)+tableAlignmentTolerance &&
		bboxBottom(inner) >= bboxBottom(outer)-tableAlignmentTolerance &&
		bboxTop(inner) <= bboxTop(outer)+tableAlignmentTolerance
}

func bboxIntersects(a, b entities.BoundingBox) bool {
	if a.Page != b.Page {
		return false
	}
	return bboxLeft(a) <= bboxRight(b) &&
		bboxRight(a) >= bboxLeft(b) &&
		bboxBottom(a) <= bboxTop(b) &&
		bboxTop(a) >= bboxBottom(b)
}

func unionBBox(a, b entities.BoundingBox) entities.BoundingBox {
	if a.Width == 0 && a.Height == 0 {
		return b
	}
	if b.Width == 0 && b.Height == 0 {
		return a
	}
	minX := bboxMin(bboxLeft(a), bboxLeft(b))
	minY := bboxMin(bboxBottom(a), bboxBottom(b))
	maxX := bboxMax(bboxRight(a), bboxRight(b))
	maxY := bboxMax(bboxTop(a), bboxTop(b))
	return entities.BoundingBox{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY, Page: a.Page}
}

func unionRows(rows []*entities.TableRow) entities.BoundingBox {
	var box entities.BoundingBox
	for _, row := range rows {
		if row == nil {
			continue
		}
		box = unionBBox(box, row.BBox)
	}
	return box
}

func unionCells(cells []*entities.TableCell) entities.BoundingBox {
	var box entities.BoundingBox
	for _, cell := range cells {
		if cell == nil {
			continue
		}
		box = unionBBox(box, cell.BBox)
	}
	return box
}

func isEmptyRow(row *entities.TableRow) bool {
	for _, cell := range row.Cells {
		if !isEmptyCell(cell) {
			return false
		}
	}
	return true
}

func isEmptyCell(cell *entities.TableCell) bool {
	return cell == nil || len(normalizeCellContent(cell.Content)) == 0
}

func resolveOriginCell(rows []*entities.TableRow, row, col int) *entities.TableCell {
	if row < 0 || row >= len(rows) || rows[row] == nil || col < 0 || col >= len(rows[row].Cells) {
		return nil
	}
	cell := rows[row].Cells[col]
	for steps := 0; cell != nil && !cell.IsOriginCell && steps < len(rows)*maxRowWidth(rows); steps++ {
		if cell.OriginRow < 0 || cell.OriginRow >= len(rows) {
			return nil
		}
		originRow := rows[cell.OriginRow]
		if originRow == nil || cell.OriginCol < 0 || cell.OriginCol >= len(originRow.Cells) {
			return nil
		}
		next := originRow.Cells[cell.OriginCol]
		if next == cell {
			break
		}
		cell = next
	}
	return cell
}

func resolveOriginCellInRow(cells []*entities.TableCell, col int) *entities.TableCell {
	if col < 0 || col >= len(cells) {
		return nil
	}
	cell := cells[col]
	for steps := 0; cell != nil && !cell.IsOriginCell && steps < len(cells); steps++ {
		if cell.OriginCol < 0 || cell.OriginCol >= len(cells) {
			return nil
		}
		next := cells[cell.OriginCol]
		if next == cell {
			break
		}
		cell = next
	}
	return cell
}

func markCoveredCell(cell, origin *entities.TableCell) {
	if cell == nil || origin == nil {
		return
	}
	cell.IsOriginCell = false
	cell.OriginRow = origin.OriginRow
	cell.OriginCol = origin.OriginCol
	cell.Rowspan = 1
	cell.Colspan = 1
	cell.Content = nil
}

func maxRowWidth(rows []*entities.TableRow) int {
	maxWidth := 0
	for _, row := range rows {
		if row != nil && len(row.Cells) > maxWidth {
			maxWidth = len(row.Cells)
		}
	}
	if maxWidth == 0 {
		return 1
	}
	return maxWidth
}

func uniqueSorted(values []float64, tolerance float64, descending bool) []float64 {
	if len(values) == 0 {
		return nil
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	uniq := make([]float64, 0, len(sorted))
	for _, value := range sorted {
		if len(uniq) == 0 || !areClose(uniq[len(uniq)-1], value, tolerance) {
			uniq = append(uniq, value)
		}
	}
	if descending {
		for i, j := 0, len(uniq)-1; i < j; i, j = i+1, j-1 {
			uniq[i], uniq[j] = uniq[j], uniq[i]
		}
	}
	return uniq
}

func bboxMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func bboxMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
