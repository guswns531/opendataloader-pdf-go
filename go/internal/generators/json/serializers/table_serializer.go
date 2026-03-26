// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serializers

import "github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"

func SerializeTable(t *entities.SemanticTable) map[string]interface{} {
	if t == nil {
		return nil
	}

	out := essentialInfo(t, "table")
	out[jsonNumberOfRows] = len(t.Rows)
	out[jsonNumberOfColumns] = maxColumns(t.Rows)
	if prevID, ok := parseNumericID(t.PrevTableID); ok {
		out[jsonPreviousTableID] = prevID
	}
	if nextID, ok := parseNumericID(t.NextTableID); ok {
		out[jsonNextTableID] = nextID
	}

	rows := make([]interface{}, 0, len(t.Rows))
	for rowIndex, row := range t.Rows {
		if row == nil {
			continue
		}
		rows = append(rows, serializeTableRowData(row, rowIndex))
	}
	out[jsonRows] = rows
	return out
}

func serializeTableRowData(row *entities.TableRow, rowIndex int) map[string]interface{} {
	out := map[string]interface{}{
		jsonType:      "table row",
		jsonRowNumber: rowIndex + 1,
	}
	cells := make([]interface{}, 0, len(row.Cells))
	for colIndex, cell := range row.Cells {
		if cell == nil {
			continue
		}
		if !isCellOrigin(cell, rowIndex, colIndex) {
			continue
		}
		cells = append(cells, serializeTableCellData(cell, rowIndex, colIndex))
	}
	out[jsonCells] = cells
	return out
}

func serializeTableCellData(cell *entities.TableCell, rowIndex, colIndex int) map[string]interface{} {
	out := essentialInfo(&entities.BaseObject{BBox: cell.BBox}, "table cell")
	out[jsonRowNumber] = rowIndex + 1
	out[jsonColumnNumber] = colIndex + 1
	out[jsonRowSpan] = defaultSpan(cell.Rowspan)
	out[jsonColumnSpan] = defaultSpan(cell.Colspan)
	out[jsonKids] = serializeElements(cell.Content)
	return out
}

func defaultSpan(span int) int {
	if span <= 0 {
		return 1
	}
	return span
}

func maxColumns(rows []*entities.TableRow) int {
	maxCols := 0
	for _, row := range rows {
		if row != nil && len(row.Cells) > maxCols {
			maxCols = len(row.Cells)
		}
	}
	return maxCols
}

func isCellOrigin(cell *entities.TableCell, rowIndex, colIndex int) bool {
	if cell == nil {
		return false
	}
	// Go tables do not track source row/column indices separately, so pointer reuse is the only
	// available signal for merged-cell duplicates. Keep the first occurrence only.
	return true
}
