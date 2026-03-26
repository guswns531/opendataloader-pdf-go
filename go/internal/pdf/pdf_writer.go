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

package pdf

import (
	"fmt"
	"os"
	"strings"

	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcpu_model "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/annotation"
	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/ocg"
)

type PDFWriter struct {
	optionalContents map[PDFLayer]*ocg.OptionalContentGroup
}

func NewPDFWriter() *PDFWriter {
	return &PDFWriter{
		optionalContents: make(map[PDFLayer]*ocg.OptionalContentGroup),
	}
}

// UpdatePDF adds semantic annotations to an existing PDF and saves to outputPath.
func (w *PDFWriter) UpdatePDF(
	inputPath string,
	password string,
	outputPath string,
	doc *entities.Document,
) error {
	if doc == nil {
		return fmt.Errorf("pdf: nil document")
	}

	pdfapi.DisableConfigDir()

	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	conf := pdfcpu_model.NewDefaultConfiguration()
	conf.UserPW = password
	conf.OwnerPW = password

	ctx, err := pdfapi.ReadContext(f, conf)
	if err != nil {
		return err
	}

	if err := ocg.SetupOCProperties(ctx); err != nil {
		return err
	}

	for pageIdx, page := range doc.Pages {
		if page == nil {
			continue
		}
		for _, element := range page.Elements {
			if err := w.drawContent(ctx, pageIdx, element, PDFLayerContent); err != nil {
				return err
			}
		}
	}

	return pdfapi.WriteContextFile(ctx, outputPath)
}

func (w *PDFWriter) drawContent(ctx *pdfcpu_model.Context, pageIdx int, obj entities.IObject, layer PDFLayer) error {
	if obj == nil {
		return nil
	}
	if obj.GetObjectType() == entities.ObjectTypeTextLine {
		return nil
	}

	entry, err := w.getOptionalContent(ctx, layer)
	if err != nil {
		return err
	}

	bbox := obj.GetBBox()
	if err := annotation.AddSquareAnnotation(ctx, pageIdx, &annotation.SquareAnnotation{
		X:        bbox.X,
		Y:        bbox.Y,
		Width:    bbox.Width,
		Height:   bbox.Height,
		Color:    getColor(obj.GetObjectType()),
		Opacity:  0.4,
		Contents: annotationContents(obj),
		OCGRef:   entry.Ref,
	}); err != nil {
		return err
	}

	switch v := obj.(type) {
	case *entities.SemanticTable:
		return w.drawTableCells(ctx, pageIdx, v)
	case *entities.PDFList:
		return w.drawListItems(ctx, pageIdx, v)
	case *entities.SemanticHeaderFooter:
		for _, line := range v.Lines {
			if err := w.drawContent(ctx, pageIdx, line, PDFLayerHeaderFooterContent); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *PDFWriter) drawTableCells(ctx *pdfcpu_model.Context, pageIdx int, table *entities.SemanticTable) error {
	if table == nil {
		return nil
	}

	for rowIdx, row := range table.Rows {
		if row == nil {
			continue
		}
		for colIdx, cell := range row.Cells {
			if cell == nil {
				continue
			}

			var parts []string
			for _, item := range cell.Content {
				parts = append(parts, extractText(item))
			}

			entry, err := w.getOptionalContent(ctx, PDFLayerTableCells)
			if err != nil {
				return err
			}

			if err := annotation.AddSquareAnnotation(ctx, pageIdx, &annotation.SquareAnnotation{
				X:       cell.BBox.X,
				Y:       cell.BBox.Y,
				Width:   cell.BBox.Width,
				Height:  cell.BBox.Height,
				Color:   getColor(entities.ObjectTypeTable),
				Opacity: 0.4,
				Contents: fmt.Sprintf(
					"Table cell: row number %d, column number %d, row span %d, column span %d, text content %q",
					rowIdx+1,
					colIdx+1,
					cell.Rowspan,
					cell.Colspan,
					strings.TrimSpace(strings.Join(parts, " ")),
				),
				OCGRef: entry.Ref,
			}); err != nil {
				return err
			}

			for _, item := range cell.Content {
				if err := w.drawContent(ctx, pageIdx, item, PDFLayerTableContent); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (w *PDFWriter) drawListItems(ctx *pdfcpu_model.Context, pageIdx int, list *entities.PDFList) error {
	if list == nil {
		return nil
	}

	for _, item := range list.Items {
		if item == nil {
			continue
		}

		entry, err := w.getOptionalContent(ctx, PDFLayerListItems)
		if err != nil {
			return err
		}

		if err := annotation.AddSquareAnnotation(ctx, pageIdx, &annotation.SquareAnnotation{
			X:        item.BBox.X,
			Y:        item.BBox.Y,
			Width:    item.BBox.Width,
			Height:   item.BBox.Height,
			Color:    getColor(entities.ObjectTypeList),
			Opacity:  0.4,
			Contents: fmt.Sprintf("List item: text content %q", strings.TrimSpace(extractText(item))),
			OCGRef:   entry.Ref,
		}); err != nil {
			return err
		}

		for _, content := range item.Content {
			if err := w.drawContent(ctx, pageIdx, content, PDFLayerListContent); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *PDFWriter) getOptionalContent(ctx *pdfcpu_model.Context, layer PDFLayer) (*ocg.OptionalContentGroup, error) {
	if entry, ok := w.optionalContents[layer]; ok {
		return entry, nil
	}

	entry, err := ocg.AddOCG(ctx, string(layer))
	if err != nil {
		return nil, err
	}
	w.optionalContents[layer] = entry

	return entry, nil
}

func getContents(obj entities.IObject) string {
	switch v := obj.(type) {
	case *entities.SemanticTable:
		return fmt.Sprintf("Table: %d rows, %d columns, previous table id %s, next table id %s",
			len(v.Rows), maxColumns(v.Rows), v.PrevTableID, v.NextTableID)
	case *entities.PDFList:
		return fmt.Sprintf("List: number of items %d", len(v.Items))
	case *entities.SemanticHeaderFooter:
		if v.IsHeader {
			return "header"
		}
		return "footer"
	case *entities.SemanticCaption:
		if v.RefType != "" {
			return fmt.Sprintf("%s, connected with object type = %s", v.Text, v.RefType)
		}
		return v.Text
	case *entities.SemanticHeading:
		return fmt.Sprintf("%s, heading level %d", textLinesString(v.Lines), v.Level)
	case *entities.SemanticParagraph:
		return textLinesString(v.Lines)
	case *entities.SemanticImage:
		return fmt.Sprintf("Image: height %.2f, width %.2f", v.Height, v.Width)
	case *entities.SemanticFormula:
		return v.LaTeX
	case *entities.TextLine:
		return v.GetText()
	case *entities.TextChunk:
		return v.Text
	default:
		return ""
	}
}

func annotationContents(obj entities.IObject) string {
	parts := make([]string, 0, 3)
	if obj != nil && obj.GetID() != "" {
		parts = append(parts, "id = "+obj.GetID())
	}
	if level, ok := objectLevel(obj); ok {
		parts = append(parts, fmt.Sprintf("level = %s", level))
	}

	contents := getContents(obj)
	if contents != "" {
		parts = append(parts, contents)
	}

	return strings.Join(parts, ", ")
}

func getColor(objType entities.ObjectType) [3]float64 {
	switch objType {
	case entities.ObjectTypeHeading, entities.ObjectTypeHeaderFooter:
		return [3]float64{0, 0, 1}
	case entities.ObjectTypeList:
		return [3]float64{0, 1, 0}
	case entities.ObjectTypeParagraph:
		return [3]float64{0, 1, 1}
	case entities.ObjectTypeImage:
		return [3]float64{1, 0, 0}
	case entities.ObjectTypeTable:
		return [3]float64{1, 0, 1}
	case entities.ObjectTypeCaption:
		return [3]float64{1, 1, 0}
	default:
		return [3]float64{0.9, 0.9, 0.9}
	}
}

func objectLevel(obj entities.IObject) (string, bool) {
	switch v := obj.(type) {
	case *entities.SemanticHeading:
		if v.Level > 0 {
			return fmt.Sprintf("%d", v.Level), true
		}
	case *entities.ListItem:
		if v.Level > 0 {
			return fmt.Sprintf("%d", v.Level), true
		}
	}

	return "", false
}

func extractText(obj entities.IObject) string {
	switch v := obj.(type) {
	case *entities.TextChunk:
		return v.Text
	case *entities.TextLine:
		return v.GetText()
	case *entities.SemanticParagraph:
		return textLinesString(v.Lines)
	case *entities.SemanticHeading:
		return textLinesString(v.Lines)
	case *entities.SemanticCaption:
		return v.Text
	case *entities.SemanticHeaderFooter:
		return textLinesString(v.Lines)
	case *entities.ListItem:
		var parts []string
		for _, item := range v.Content {
			parts = append(parts, extractText(item))
		}
		return strings.TrimSpace(strings.Join(parts, " "))
	case *entities.SemanticFormula:
		return v.LaTeX
	default:
		return getContents(obj)
	}
}

func textLinesString(lines []*entities.TextLine) string {
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == nil {
			continue
		}
		text := strings.TrimSpace(line.GetText())
		if text != "" {
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, " ")
}

func maxColumns(rows []*entities.TableRow) int {
	max := 0
	for _, row := range rows {
		if row != nil && len(row.Cells) > max {
			max = len(row.Cells)
		}
	}
	return max
}
