// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const (
	labelSectionHeader = "section_header"
	labelPageHeader    = "page_header"
	labelPageFooter    = "page_footer"
	labelFormula       = "formula"
	coordOriginTopLeft = "TOPLEFT"
)

type doclingRoot struct {
	Document *struct {
		JSONContent json.RawMessage `json:"json_content"`
	} `json:"document"`
	Status      string          `json:"status"`
	Errors      json.RawMessage `json:"errors"`
	FailedPages []int           `json:"failed_pages"`
}

type doclingDocument struct {
	Pages    map[string]doclingPageMeta `json:"pages"`
	Texts    []doclingText              `json:"texts"`
	Tables   []doclingTable             `json:"tables"`
	Pictures []doclingPicture           `json:"pictures"`
}

type doclingPageMeta struct {
	Size *struct {
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	} `json:"size"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type doclingText struct {
	Label string         `json:"label"`
	Text  string         `json:"text"`
	Orig  string         `json:"orig"`
	Meta  map[string]any `json:"meta"`
	Prov  []doclingProv  `json:"prov"`
}

type doclingPicture struct {
	Prov        []doclingProv       `json:"prov"`
	Annotations []doclingAnnotation `json:"annotations"`
}

type doclingAnnotation struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
}

type doclingTable struct {
	Prov []doclingProv     `json:"prov"`
	Data *doclingTableData `json:"data"`
}

type doclingTableData struct {
	Grid       [][]any            `json:"grid"`
	TableCells []doclingTableCell `json:"table_cells"`
}

type doclingTableCell struct {
	StartRow int    `json:"start_row_offset_idx"`
	StartCol int    `json:"start_col_offset_idx"`
	RowSpan  int    `json:"row_span"`
	ColSpan  int    `json:"col_span"`
	Text     string `json:"text"`
}

type doclingProv struct {
	PageNo int          `json:"page_no"`
	BBox   *doclingBBox `json:"bbox"`
}

type doclingBBox struct {
	L           float64 `json:"l"`
	T           float64 `json:"t"`
	R           float64 `json:"r"`
	B           float64 `json:"b"`
	CoordOrigin string  `json:"coord_origin"`
}

func TransformDoclingResponse(data []byte) ([]*entities.Page, error) {
	parsed, err := parseDoclingResponse(data)
	if err != nil {
		return nil, err
	}
	return parsed.Pages, nil
}

func transformDoclingDocument(data []byte) ([]*entities.Page, error) {
	var doc doclingDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	pageHeights := map[int]float64{}
	pageSizes := map[int][2]float64{}
	maxPage := 0
	for key, meta := range doc.Pages {
		var pageNum int
		if _, err := fmt.Sscanf(key, "%d", &pageNum); err != nil {
			continue
		}
		width := meta.Width
		height := meta.Height
		if meta.Size != nil {
			if meta.Size.Width > 0 {
				width = meta.Size.Width
			}
			if meta.Size.Height > 0 {
				height = meta.Size.Height
			}
		}
		pageHeights[pageNum] = height
		pageSizes[pageNum] = [2]float64{width, height}
		if pageNum > maxPage {
			maxPage = pageNum
		}
	}

	updateMaxPage := func(pageNo int) {
		if pageNo > maxPage {
			maxPage = pageNo
		}
	}
	for _, t := range doc.Texts {
		if len(t.Prov) > 0 {
			updateMaxPage(t.Prov[0].PageNo)
		}
	}
	for _, t := range doc.Tables {
		if len(t.Prov) > 0 {
			updateMaxPage(t.Prov[0].PageNo)
		}
	}
	for _, p := range doc.Pictures {
		if len(p.Prov) > 0 {
			updateMaxPage(p.Prov[0].PageNo)
		}
	}

	pages := make([]*entities.Page, maxPage)
	for i := 0; i < maxPage; i++ {
		size := pageSizes[i+1]
		pages[i] = &entities.Page{
			PageMetadata: entities.PageMetadata{
				Number: i,
				Width:  size[0],
				Height: size[1],
			},
			Elements: []entities.IObject{},
		}
	}

	idCounter := 0
	nextID := func(prefix string) string {
		idCounter++
		return fmt.Sprintf("%s_%d", prefix, idCounter)
	}

	for _, node := range doc.Texts {
		if node.Label == labelPageHeader || node.Label == labelPageFooter || len(node.Prov) == 0 {
			continue
		}
		pageNo := node.Prov[0].PageNo
		if pageNo <= 0 || pageNo > len(pages) {
			continue
		}
		bbox := extractDoclingBBox(node.Prov[0].BBox, pageNo-1, pageHeights[pageNo])
		text := node.Text
		if text == "" {
			text = node.Orig
		}
		var object entities.IObject
		switch node.Label {
		case labelSectionHeader:
			level := 1
			if raw, ok := node.Meta["level"].(float64); ok && raw >= 1 {
				level = int(raw)
			}
			object = &entities.SemanticHeading{
				BaseObject: entities.BaseObject{ID: nextID("heading"), BBox: bbox},
				Lines:      []*entities.TextLine{newTextLine(nextID("line"), nextID("chunk"), bbox, text)},
				Level:      level,
				FontSize:   12.0,
			}
		case labelFormula:
			object = &entities.SemanticFormula{
				BaseObject: entities.BaseObject{ID: nextID("formula"), BBox: bbox},
				LaTeX:      text,
			}
		default:
			object = &entities.SemanticParagraph{
				BaseObject: entities.BaseObject{ID: nextID("paragraph"), BBox: bbox},
				Lines:      []*entities.TextLine{newTextLine(nextID("line"), nextID("chunk"), bbox, text)},
			}
		}
		pages[pageNo-1].Elements = append(pages[pageNo-1].Elements, object)
	}

	for _, picture := range doc.Pictures {
		if len(picture.Prov) == 0 {
			continue
		}
		pageNo := picture.Prov[0].PageNo
		if pageNo <= 0 || pageNo > len(pages) {
			continue
		}
		bbox := extractDoclingBBox(picture.Prov[0].BBox, pageNo-1, pageHeights[pageNo])
		alt := ""
		for _, annotation := range picture.Annotations {
			if annotation.Kind == "description" {
				alt = annotation.Text
				break
			}
		}
		pages[pageNo-1].Elements = append(pages[pageNo-1].Elements, &entities.SemanticImage{
			BaseObject: entities.BaseObject{ID: nextID("image"), BBox: bbox},
			Alt:        alt,
			Width:      bbox.Width,
			Height:     bbox.Height,
		})
	}

	for _, table := range doc.Tables {
		if len(table.Prov) == 0 || table.Data == nil || len(table.Data.Grid) == 0 || len(table.Data.Grid[0]) == 0 {
			continue
		}
		pageNo := table.Prov[0].PageNo
		if pageNo <= 0 || pageNo > len(pages) {
			continue
		}
		tableBBox := extractDoclingBBox(table.Prov[0].BBox, pageNo-1, pageHeights[pageNo])
		rowCount := len(table.Data.Grid)
		colCount := len(table.Data.Grid[0])
		rowHeight := tableBBox.Height / float64(rowCount)
		colWidth := tableBBox.Width / float64(colCount)
		cellMap := map[[2]int]doclingTableCell{}
		for _, cell := range table.Data.TableCells {
			cellMap[[2]int{cell.StartRow, cell.StartCol}] = cell
		}

		rows := make([]*entities.TableRow, 0, rowCount)
		for row := 0; row < rowCount; row++ {
			rowTop := bboxTop(tableBBox) - float64(row)*rowHeight
			rowBottom := rowTop - rowHeight
			tableRow := &entities.TableRow{
				Cells: make([]*entities.TableCell, 0, colCount),
				BBox: entities.BoundingBox{
					X:      bboxLeft(tableBBox),
					Y:      rowBottom,
					Width:  tableBBox.Width,
					Height: rowHeight,
					Page:   pageNo - 1,
				},
			}
			for col := 0; col < colCount; col++ {
				cellInfo, ok := cellMap[[2]int{row, col}]
				rowSpan, colSpan, text := 1, 1, ""
				if ok {
					if cellInfo.RowSpan > 0 {
						rowSpan = cellInfo.RowSpan
					}
					if cellInfo.ColSpan > 0 {
						colSpan = cellInfo.ColSpan
					}
					text = cellInfo.Text
				}
				cellLeft := bboxLeft(tableBBox) + float64(col)*colWidth
				cellTop := bboxTop(tableBBox) - float64(row)*rowHeight
				cellBox := entities.BoundingBox{
					X:      cellLeft,
					Y:      cellTop - float64(rowSpan)*rowHeight,
					Width:  float64(colSpan) * colWidth,
					Height: float64(rowSpan) * rowHeight,
					Page:   pageNo - 1,
				}
				cell := &entities.TableCell{
					Rowspan: rowSpan,
					Colspan: colSpan,
					BBox:    cellBox,
				}
				if text != "" {
					cell.Content = []entities.IObject{
						&entities.SemanticParagraph{
							BaseObject: entities.BaseObject{ID: nextID("paragraph"), BBox: cellBox},
							Lines:      []*entities.TextLine{newTextLine(nextID("line"), nextID("chunk"), cellBox, text)},
						},
					}
				}
				tableRow.Cells = append(tableRow.Cells, cell)
			}
			rows = append(rows, tableRow)
		}
		pages[pageNo-1].Elements = append(pages[pageNo-1].Elements, &entities.SemanticTable{
			BaseObject: entities.BaseObject{ID: nextID("table"), BBox: tableBBox},
			Rows:       rows,
		})
	}

	for _, page := range pages {
		sort.Slice(page.Elements, func(i, j int) bool {
			left := page.Elements[i].GetBBox()
			right := page.Elements[j].GetBBox()
			topDiff := bboxTop(right) - bboxTop(left)
			if mathAbs(topDiff) > 5.0 {
				return topDiff < 0
			}
			return bboxLeft(left) < bboxLeft(right)
		})
	}

	return pages, nil
}

func newTextLine(lineID, chunkID string, bbox entities.BoundingBox, text string) *entities.TextLine {
	chunk := &entities.TextChunk{
		BaseObject: entities.BaseObject{ID: chunkID, BBox: bbox},
		Text:       text,
		Baseline:   bbox.Y,
	}
	return &entities.TextLine{
		BaseObject: entities.BaseObject{ID: lineID, BBox: bbox},
		Chunks:     []*entities.TextChunk{chunk},
		Baseline:   chunk.Baseline,
	}
}

func extractDoclingBBox(b *doclingBBox, pageIndex int, pageHeight float64) entities.BoundingBox {
	if b == nil {
		return entities.BoundingBox{Page: pageIndex}
	}
	left, bottom, right, top := b.L, b.B, b.R, b.T
	if b.CoordOrigin == coordOriginTopLeft && pageHeight > 0 {
		top = pageHeight - b.T
		bottom = pageHeight - b.B
	}
	return entities.BoundingBox{
		X:      left,
		Y:      bottom,
		Width:  right - left,
		Height: top - bottom,
		Page:   pageIndex,
	}
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
