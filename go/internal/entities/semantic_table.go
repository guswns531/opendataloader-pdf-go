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

package entities

type TableCell struct {
	Content      []IObject
	Rowspan      int
	Colspan      int
	IsOriginCell bool
	OriginRow    int
	OriginCol    int
	BBox         BoundingBox
}

type TableRow struct {
	Cells []*TableCell
	BBox  BoundingBox
}

type SemanticTable struct {
	BaseObject
	Rows        []*TableRow
	PrevTableID string
	NextTableID string
}

func (t *SemanticTable) GetObjectType() ObjectType { return ObjectTypeTable }

func NewTableCell(row, col, rowspan, colspan int, bbox BoundingBox, content []IObject) *TableCell {
	if rowspan <= 0 {
		rowspan = 1
	}
	if colspan <= 0 {
		colspan = 1
	}
	return &TableCell{
		Content:      content,
		Rowspan:      rowspan,
		Colspan:      colspan,
		IsOriginCell: true,
		OriginRow:    row,
		OriginCol:    col,
		BBox:         bbox,
	}
}

func NewCoveredTableCell(originRow, originCol int, bbox BoundingBox) *TableCell {
	return &TableCell{
		Rowspan:      1,
		Colspan:      1,
		IsOriginCell: false,
		OriginRow:    originRow,
		OriginCol:    originCol,
		BBox:         bbox,
	}
}

func (c *TableCell) EffectiveRowSpan() int {
	if c == nil || c.Rowspan <= 0 {
		return 1
	}
	return c.Rowspan
}

func (c *TableCell) EffectiveColSpan() int {
	if c == nil || c.Colspan <= 0 {
		return 1
	}
	return c.Colspan
}

func (c *TableCell) OriginPosition(defaultRow, defaultCol int) (int, int) {
	if c == nil {
		return defaultRow, defaultCol
	}
	if c.IsOriginCell || c.OriginRow != 0 || c.OriginCol != 0 {
		return c.OriginRow, c.OriginCol
	}
	return defaultRow, defaultCol
}

func (c *TableCell) IsOrigin(defaultRow, defaultCol int) bool {
	if c == nil {
		return false
	}
	if c.IsOriginCell {
		return true
	}
	if c.OriginRow != 0 || c.OriginCol != 0 {
		return c.OriginRow == defaultRow && c.OriginCol == defaultCol
	}
	return true
}
