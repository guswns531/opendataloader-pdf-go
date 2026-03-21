package table

import (
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestDetectConvertsAlignedGridIntoTable(t *testing.T) {
	input := []model.ContentElement{
		paragraph("This report summarizes the sample set.", 20, model.NewBox(40, 744, 260, 760)),
		paragraph("North", 40, model.NewBox(40, 700, 120, 716)),
		paragraph("18", 180, model.NewBox(180, 700, 220, 716)),
		paragraph("South", 40, model.NewBox(40, 670, 120, 686)),
		paragraph("24", 180, model.NewBox(180, 670, 220, 686)),
		paragraph("Additional notes follow below.", 20, model.NewBox(40, 620, 260, 636)),
	}

	got := Detect(input)
	if gotLen, want := len(got), 3; gotLen != want {
		t.Fatalf("Detect() len = %d, want %d", gotLen, want)
	}

	if _, ok := got[0].(*model.Paragraph); !ok {
		t.Fatalf("first element type = %T, want *model.Paragraph", got[0])
	}

	table, ok := got[1].(*model.Table)
	if !ok {
		t.Fatalf("second element type = %T, want *model.Table", got[1])
	}
	if table.Type != model.ElementTypeTable {
		t.Fatalf("table type = %q, want %q", table.Type, model.ElementTypeTable)
	}
	if table.NumberOfRows != 2 {
		t.Fatalf("table row count = %d, want 2", table.NumberOfRows)
	}
	if table.NumberOfColumns != 2 {
		t.Fatalf("table column count = %d, want 2", table.NumberOfColumns)
	}
	if got, want := table.Bounds, model.NewBox(40, 670, 220, 716); got != want {
		t.Fatalf("table bounds = %#v, want %#v", got, want)
	}
	if len(table.Rows) != 2 {
		t.Fatalf("table rows len = %d, want 2", len(table.Rows))
	}

	row0 := table.Rows[0]
	if row0.Type != model.ElementTypeTableRow {
		t.Fatalf("row0 type = %q, want %q", row0.Type, model.ElementTypeTableRow)
	}
	if row0.RowNumber != 1 {
		t.Fatalf("row0 number = %d, want 1", row0.RowNumber)
	}
	if len(row0.Cells) != 2 {
		t.Fatalf("row0 cells len = %d, want 2", len(row0.Cells))
	}

	cell00 := row0.Cells[0]
	if cell00.Type != model.ElementTypeTableCell {
		t.Fatalf("cell00 type = %q, want %q", cell00.Type, model.ElementTypeTableCell)
	}
	if cell00.RowNumber != 1 || cell00.ColumnNumber != 1 {
		t.Fatalf("cell00 position = (%d, %d), want (1, 1)", cell00.RowNumber, cell00.ColumnNumber)
	}
	if len(cell00.Kids) != 1 {
		t.Fatalf("cell00 kids len = %d, want 1", len(cell00.Kids))
	}
	if para, ok := cell00.Kids[0].(*model.Paragraph); !ok || para.Content != "North" {
		t.Fatalf("cell00 kid = %T %q, want *model.Paragraph %q", cell00.Kids[0], paragraphContent(cell00.Kids[0]), "North")
	}

	cell11 := table.Rows[1].Cells[1]
	if cell11.RowNumber != 2 || cell11.ColumnNumber != 2 {
		t.Fatalf("cell11 position = (%d, %d), want (2, 2)", cell11.RowNumber, cell11.ColumnNumber)
	}
	if para, ok := cell11.Kids[0].(*model.Paragraph); !ok || para.Content != "24" {
		t.Fatalf("cell11 kid = %T %q, want *model.Paragraph %q", cell11.Kids[0], paragraphContent(cell11.Kids[0]), "24")
	}

	if _, ok := got[2].(*model.Paragraph); !ok {
		t.Fatalf("third element type = %T, want *model.Paragraph", got[2])
	}
}

func TestDetectRejectsSentenceLikeAlignedRows(t *testing.T) {
	input := []model.ContentElement{
		paragraph("The system processes incoming requests.", 40, model.NewBox(40, 700, 260, 716)),
		paragraph("It then stores the normalized output.", 180, model.NewBox(180, 700, 380, 716)),
		paragraph("The job completes without incident.", 40, model.NewBox(40, 670, 260, 686)),
		paragraph("It returns the final response to the caller.", 180, model.NewBox(180, 670, 380, 686)),
	}

	got := Detect(input)
	if gotLen, want := len(got), len(input); gotLen != want {
		t.Fatalf("Detect() len = %d, want %d", gotLen, want)
	}
	for i, element := range got {
		if _, ok := element.(*model.Paragraph); !ok {
			t.Fatalf("element %d type = %T, want *model.Paragraph", i, element)
		}
	}
}

func TestDetectRejectsMisalignedRuns(t *testing.T) {
	input := []model.ContentElement{
		paragraph("A", 40, model.NewBox(40, 700, 120, 716)),
		paragraph("B", 180, model.NewBox(180, 700, 220, 716)),
		paragraph("C", 70, model.NewBox(70, 670, 150, 686)),
		paragraph("D", 180, model.NewBox(180, 670, 220, 686)),
	}

	got := Detect(input)
	if gotLen, want := len(got), len(input); gotLen != want {
		t.Fatalf("Detect() len = %d, want %d", gotLen, want)
	}
	if _, ok := got[0].(*model.Paragraph); !ok {
		t.Fatalf("first element type = %T, want *model.Paragraph", got[0])
	}
	if _, ok := got[1].(*model.Paragraph); !ok {
		t.Fatalf("second element type = %T, want *model.Paragraph", got[1])
	}
}

func TestDetectLeavesSingleRowAsParagraphs(t *testing.T) {
	input := []model.ContentElement{
		paragraph("Left", 40, model.NewBox(40, 700, 120, 716)),
		paragraph("Right", 180, model.NewBox(180, 700, 220, 716)),
	}

	got := Detect(input)
	if gotLen, want := len(got), len(input); gotLen != want {
		t.Fatalf("Detect() len = %d, want %d", gotLen, want)
	}
	for i, element := range got {
		if _, ok := element.(*model.Paragraph); !ok {
			t.Fatalf("element %d type = %T, want *model.Paragraph", i, element)
		}
	}
}

func paragraph(text string, left float64, bounds model.Box) *model.Paragraph {
	return &model.Paragraph{
		TextNode: model.TextNode{
			BaseNode: model.BaseNode{
				Type:       model.ElementTypeParagraph,
				PageIndex:  0,
				PageNumber: 1,
				Bounds:     bounds,
			},
			TextProperties: model.TextProperties{
				Content: text,
			},
		},
	}
}

func paragraphContent(element model.ContentElement) string {
	para, ok := element.(*model.Paragraph)
	if !ok {
		return ""
	}
	return para.Content
}
