package markdown

import (
	"strings"
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

func TestGenerateSimpleTableUsesPipeSyntaxWithSeparatorRow(t *testing.T) {
	doc := &entities.Document{
		Pages: []*entities.Page{
			{
				Elements: []entities.IObject{
					&entities.SemanticTable{
						Rows: []*entities.TableRow{
							{
								Cells: []*entities.TableCell{
									newTextCell(0, 0, "Col A"),
									newTextCell(0, 1, "Col B"),
								},
							},
							{
								Cells: []*entities.TableCell{
									newTextCell(1, 0, "A1"),
									newTextCell(1, 1, "B1"),
								},
							},
						},
					},
				},
			},
		},
	}

	output, err := NewMarkdownGenerator(nil, false, false).Generate(doc)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if strings.Contains(output, "<table>") {
		t.Fatalf("expected pipe table output, got HTML: %q", output)
	}
	if !strings.Contains(output, "| Col A | Col B |") {
		t.Fatalf("expected header row in pipe syntax, got %q", output)
	}
	if !strings.Contains(output, "| --- | --- |") {
		t.Fatalf("expected markdown separator row, got %q", output)
	}
	if !strings.Contains(output, "| A1 | B1 |") {
		t.Fatalf("expected body row in pipe syntax, got %q", output)
	}
}

func TestGenerateSpanTableUsesHTMLWithTableSections(t *testing.T) {
	doc := &entities.Document{
		Pages: []*entities.Page{
			{
				Elements: []entities.IObject{
					&entities.SemanticTable{
						Rows: []*entities.TableRow{
							{
								Cells: []*entities.TableCell{
									newSpanTextCell(0, 0, 2, 1, "No."),
									newSpanTextCell(0, 1, 1, 2, "Group"),
								},
							},
							{
								Cells: []*entities.TableCell{
									entities.NewCoveredTableCell(0, 0, entities.BoundingBox{}),
									newTextCell(1, 1, "Left"),
									newTextCell(1, 2, "Right"),
								},
							},
							{
								Cells: []*entities.TableCell{
									newTextCell(2, 0, "1"),
									newTextCell(2, 1, "A"),
									newTextCell(2, 2, "B"),
								},
							},
						},
					},
				},
			},
		},
	}

	output, err := NewMarkdownGenerator(nil, false, false).Generate(doc)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for _, want := range []string{
		"<table>",
		"<thead>",
		"<th rowspan=\"2\">No.</th>",
		"<th colspan=\"2\">Group</th>",
		"</thead>",
		"<tbody>",
		"<td>1</td>",
		"</tbody>",
		"</table>",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got %q", want, output)
		}
	}

	if strings.Contains(output, "| --- |") {
		t.Fatalf("expected HTML output for span table, got %q", output)
	}
}

func newTextCell(row, col int, text string) *entities.TableCell {
	return newSpanTextCell(row, col, 1, 1, text)
}

func newSpanTextCell(row, col, rowspan, colspan int, text string) *entities.TableCell {
	return entities.NewTableCell(row, col, rowspan, colspan, entities.BoundingBox{}, []entities.IObject{
		&entities.TextChunk{Text: text},
	})
}
