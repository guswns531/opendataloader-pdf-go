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
	if t.PrevTableID != "" {
		out[jsonPreviousTableID] = t.PrevTableID
	}
	if t.NextTableID != "" {
		out[jsonNextTableID] = t.NextTableID
	}

	rows := make([]interface{}, 0, len(t.Rows))
	for rowIndex, row := range t.Rows {
		if row == nil {
			continue
		}
		rows = append(rows, serializeTableRow(row, rowIndex))
	}
	out[jsonRows] = rows
	return out
}

func serializeTableRow(row *entities.TableRow, rowIndex int) map[string]interface{} {
	out := map[string]interface{}{
		jsonType:      "table row",
		jsonRowNumber: rowIndex + 1,
	}
	cells := make([]interface{}, 0, len(row.Cells))
	for colIndex, cell := range row.Cells {
		if cell == nil {
			continue
		}
		cells = append(cells, serializeTableCell(cell, rowIndex, colIndex))
	}
	out[jsonCells] = cells
	return out
}

func serializeTableCell(cell *entities.TableCell, rowIndex, colIndex int) map[string]interface{} {
	out := map[string]interface{}{
		jsonType:         "table cell",
		jsonRowNumber:    rowIndex + 1,
		jsonColumnNumber: colIndex + 1,
		jsonRowSpan:      defaultSpan(cell.Rowspan),
		jsonColumnSpan:   defaultSpan(cell.Colspan),
		jsonBoundingBox: map[string]interface{}{
			"left":   serializeDouble(cell.BBox.X),
			"bottom": serializeDouble(cell.BBox.Y),
			"right":  serializeDouble(cell.BBox.X + cell.BBox.Width),
			"top":    serializeDouble(cell.BBox.Y + cell.BBox.Height),
		},
	}
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
