package schemajson

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestEmitterNameAndFormat(t *testing.T) {
	emitter := New()

	if got := emitter.Name(); got != "schemajson" {
		t.Fatalf("Name() = %q, want schemajson", got)
	}
	if got := emitter.Format(); got != core.OutputFormatJSON {
		t.Fatalf("Format() = %q, want %q", got, core.OutputFormatJSON)
	}
}

func TestEmitWritesSchemaAlignedDocument(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{
		FileName:         "sample.pdf",
		Author:           strPtr("Author"),
		Title:            strPtr("Title"),
		CreationDate:     strPtr("2025-03-20T00:00:00Z"),
		ModificationDate: strPtr("2025-03-21T00:00:00Z"),
		PageCount:        2,
	})

	listPrevID := model.NodeID(11)
	listNextID := model.NodeID(12)
	linkedContentID := model.NodeID(21)
	tablePrevID := model.NodeID(31)
	tableNextID := model.NodeID(32)

	doc.Kids = []model.ContentElement{
		&model.Paragraph{
			TextNode: model.TextNode{
				BaseNode: model.BaseNode{
					ID:         1,
					Type:       model.ElementTypeParagraph,
					PageNumber: 1,
					Bounds:     model.NewBox(10, 20, 30, 40),
				},
				TextProperties: model.TextProperties{
					Font:      "Arial",
					FontSize:  9.5,
					TextColor: "#000000",
					Content:   "Paragraph",
				},
			},
		},
		&model.Heading{
			TextNode: model.TextNode{
				BaseNode: model.BaseNode{
					ID:         2,
					Type:       model.ElementTypeHeading,
					Level:      "1",
					PageNumber: 1,
					Bounds:     model.NewBox(40, 50, 60, 70),
				},
				TextProperties: model.TextProperties{
					Font:      "Arial-Bold",
					FontSize:  12,
					TextColor: "#111111",
					Content:   "Heading",
				},
			},
			HeadingLevel: 2,
		},
		&model.Caption{
			TextNode: model.TextNode{
				BaseNode: model.BaseNode{
					ID:         3,
					Type:       model.ElementTypeCaption,
					PageNumber: 1,
					Bounds:     model.NewBox(70, 80, 90, 100),
					Links: model.LinkField{
						LinkedContentID: &linkedContentID,
					},
				},
				TextProperties: model.TextProperties{
					Font:      "Arial",
					FontSize:  8,
					TextColor: "#222222",
					Content:   "Caption",
				},
			},
		},
		&model.List{
			BaseNode: model.BaseNode{
				ID:         4,
				Type:       model.ElementTypeList,
				PageNumber: 1,
				Bounds:     model.NewBox(100, 110, 120, 130),
				Links: model.LinkField{
					PreviousID: &listPrevID,
					NextID:     &listNextID,
				},
			},
			NumberingStyle: "bullet",
			NumberOfItems:  2,
			ListItems: []*model.ListItem{
				{
					TextNode: model.TextNode{
						BaseNode: model.BaseNode{
							ID:         5,
							Type:       model.ElementTypeListItem,
							PageNumber: 1,
							Bounds:     model.NewBox(101, 111, 102, 112),
						},
						TextProperties: model.TextProperties{
							Font:      "Arial",
							FontSize:  8,
							TextColor: "#333333",
							Content:   "Item 1",
						},
					},
					Kids: []model.ContentElement{
						&model.Paragraph{
							TextNode: model.TextNode{
								BaseNode: model.BaseNode{
									ID:         6,
									Type:       model.ElementTypeParagraph,
									PageNumber: 1,
									Bounds:     model.NewBox(103, 113, 104, 114),
								},
								TextProperties: model.TextProperties{
									Font:      "Arial",
									FontSize:  7,
									TextColor: "#444444",
									Content:   "Nested paragraph",
								},
							},
						},
					},
				},
				{
					TextNode: model.TextNode{
						BaseNode: model.BaseNode{
							ID:         7,
							Type:       model.ElementTypeListItem,
							PageNumber: 1,
							Bounds:     model.NewBox(105, 115, 106, 116),
						},
						TextProperties: model.TextProperties{
							Font:      "Arial",
							FontSize:  8,
							TextColor: "#555555",
							Content:   "Item 2",
						},
					},
				},
			},
		},
		&model.TextBlock{
			BaseNode: model.BaseNode{
				ID:         8,
				Type:       model.ElementTypeTextBlock,
				PageNumber: 1,
				Bounds:     model.NewBox(130, 140, 150, 160),
			},
			Kids: []model.ContentElement{
				&model.Paragraph{
					TextNode: model.TextNode{
						BaseNode: model.BaseNode{
							ID:         9,
							Type:       model.ElementTypeParagraph,
							PageNumber: 1,
							Bounds:     model.NewBox(131, 141, 132, 142),
						},
						TextProperties: model.TextProperties{
							Font:      "Arial",
							FontSize:  7,
							TextColor: "#666666",
							Content:   "Text block child",
						},
					},
				},
			},
		},
		&model.Image{
			BaseNode: model.BaseNode{
				ID:         10,
				Type:       model.ElementTypeImage,
				PageNumber: 1,
				Bounds:     model.NewBox(160, 170, 180, 190),
			},
			Source: "images/figure.png",
			Data:   "data-uri",
			Format: model.ImageFormatPNG,
		},
		&model.HeaderFooter{
			BaseNode: model.BaseNode{
				ID:         13,
				Type:       model.ElementTypeFooter,
				PageNumber: 1,
				Bounds:     model.NewBox(190, 200, 210, 220),
			},
			Kids: []model.ContentElement{
				&model.Paragraph{
					TextNode: model.TextNode{
						BaseNode: model.BaseNode{
							ID:         14,
							Type:       model.ElementTypeParagraph,
							PageNumber: 1,
							Bounds:     model.NewBox(191, 201, 192, 202),
						},
						TextProperties: model.TextProperties{
							Font:      "Arial",
							FontSize:  6,
							TextColor: "#777777",
							Content:   "Footer child",
						},
					},
				},
			},
		},
		&model.Table{
			BaseNode: model.BaseNode{
				ID:         15,
				Type:       model.ElementTypeTable,
				PageNumber: 1,
				Bounds:     model.NewBox(230, 240, 250, 260),
			},
			NumberOfRows:    1,
			NumberOfColumns: 1,
			PreviousTableID: &tablePrevID,
			NextTableID:     &tableNextID,
			Rows: []model.TableRow{
				{
					Type:      model.ElementTypeTableRow,
					RowNumber: 1,
					Cells: []*model.TableCell{
						{
							BaseNode: model.BaseNode{
								ID:         16,
								Type:       model.ElementTypeTableCell,
								PageNumber: 1,
								Bounds:     model.NewBox(231, 241, 232, 242),
							},
							RowNumber:    1,
							ColumnNumber: 1,
							RowSpan:      2,
							ColumnSpan:   3,
							Kids: []model.ContentElement{
								&model.Paragraph{
									TextNode: model.TextNode{
										BaseNode: model.BaseNode{
											ID:         17,
											Type:       model.ElementTypeParagraph,
											PageNumber: 1,
											Bounds:     model.NewBox(233, 243, 234, 244),
										},
										TextProperties: model.TextProperties{
											Font:      "Arial",
											FontSize:  6,
											TextColor: "#888888",
											Content:   "Cell child",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := New().Emit(core.NewProcessingContext(doc, core.ProcessingOptions{}), doc, &buf); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	root := mustMap(t, decodeJSON(t, buf.Bytes()))
	wantKeys := []string{"file name", "number of pages", "author", "title", "creation date", "modification date", "kids"}
	assertExactKeys(t, root, wantKeys)
	if root["file name"] != "sample.pdf" {
		t.Fatalf("file name = %v, want sample.pdf", root["file name"])
	}
	if got := int(root["number of pages"].(float64)); got != 2 {
		t.Fatalf("number of pages = %d, want 2", got)
	}

	kids := mustSlice(t, root["kids"])
	if len(kids) != 8 {
		t.Fatalf("kids length = %d, want 8", len(kids))
	}

	paragraph := mustMap(t, kids[0])
	assertExactKeys(t, paragraph, []string{"type", "id", "page number", "bounding box", "font", "font size", "text color", "content"})
	if paragraph["type"] != "paragraph" {
		t.Fatalf("paragraph type = %v, want paragraph", paragraph["type"])
	}
	assertBox(t, paragraph["bounding box"], 10, 20, 30, 40)
	if _, ok := paragraph["hidden text"]; ok {
		t.Fatal("paragraph unexpectedly emitted hidden text")
	}

	heading := mustMap(t, kids[1])
	if heading["heading level"] != float64(2) {
		t.Fatalf("heading level = %v, want 2", heading["heading level"])
	}

	caption := mustMap(t, kids[2])
	if got := int(caption["linked content id"].(float64)); got != 21 {
		t.Fatalf("linked content id = %d, want 21", got)
	}

	list := mustMap(t, kids[3])
	assertExactKeys(t, list, []string{"type", "id", "page number", "bounding box", "numbering style", "number of list items", "previous list id", "next list id", "list items"})
	if got := int(list["number of list items"].(float64)); got != 2 {
		t.Fatalf("number of list items = %d, want 2", got)
	}
	listItems := mustSlice(t, list["list items"])
	if len(listItems) != 2 {
		t.Fatalf("list items length = %d, want 2", len(listItems))
	}
	nestedListItem := mustMap(t, listItems[0])
	if nestedListItem["content"] != "Item 1" {
		t.Fatalf("list item content = %v, want Item 1", nestedListItem["content"])
	}
	nestedKids := mustSlice(t, nestedListItem["kids"])
	if len(nestedKids) != 1 {
		t.Fatalf("nested list item kids length = %d, want 1", len(nestedKids))
	}

	textBlock := mustMap(t, kids[4])
	if len(mustSlice(t, textBlock["kids"])) != 1 {
		t.Fatal("text block should contain one child")
	}

	image := mustMap(t, kids[5])
	if image["format"] != "png" {
		t.Fatalf("image format = %v, want png", image["format"])
	}
	if image["source"] != "images/figure.png" {
		t.Fatalf("image source = %v, want images/figure.png", image["source"])
	}

	footer := mustMap(t, kids[6])
	if footer["type"] != "footer" {
		t.Fatalf("header/footer type = %v, want footer", footer["type"])
	}
	if len(mustSlice(t, footer["kids"])) != 1 {
		t.Fatal("footer should contain one child")
	}

	table := mustMap(t, kids[7])
	if got := int(table["number of rows"].(float64)); got != 1 {
		t.Fatalf("number of rows = %d, want 1", got)
	}
	if got := int(table["number of columns"].(float64)); got != 1 {
		t.Fatalf("number of columns = %d, want 1", got)
	}
	rows := mustSlice(t, table["rows"])
	if len(rows) != 1 {
		t.Fatalf("rows length = %d, want 1", len(rows))
	}
	row := mustMap(t, rows[0])
	if row["type"] != "table row" {
		t.Fatalf("row type = %v, want table row", row["type"])
	}
	cells := mustSlice(t, row["cells"])
	if len(cells) != 1 {
		t.Fatalf("cells length = %d, want 1", len(cells))
	}
	cell := mustMap(t, cells[0])
	if got := int(cell["row span"].(float64)); got != 2 {
		t.Fatalf("row span = %d, want 2", got)
	}
	if got := int(cell["column span"].(float64)); got != 3 {
		t.Fatalf("column span = %d, want 3", got)
	}
	if len(mustSlice(t, cell["kids"])) != 1 {
		t.Fatal("table cell should contain one child")
	}
}

func TestEmitFallsBackToPageKids(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{
		FileName:  "page-only.pdf",
		PageCount: 0,
	})

	doc.AddPage(&model.Page{
		Metadata: model.PageMetadata{
			Index:  0,
			Number: 1,
		},
		Kids: []model.ContentElement{
			&model.Paragraph{
				TextNode: model.TextNode{
					BaseNode: model.BaseNode{
						ID:         1,
						Type:       model.ElementTypeParagraph,
						PageNumber: 1,
						Bounds:     model.NewBox(1, 2, 3, 4),
					},
					TextProperties: model.TextProperties{
						Font:      "Arial",
						FontSize:  7,
						TextColor: "#000000",
						Content:   "Page child",
					},
				},
			},
		},
	})

	var buf bytes.Buffer
	if err := New().Emit(nil, doc, &buf); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	root := mustMap(t, decodeJSON(t, buf.Bytes()))
	if got := int(root["number of pages"].(float64)); got != 1 {
		t.Fatalf("number of pages = %d, want 1", got)
	}
	kids := mustSlice(t, root["kids"])
	if len(kids) != 1 {
		t.Fatalf("kids length = %d, want 1", len(kids))
	}
	paragraph := mustMap(t, kids[0])
	if paragraph["content"] != "Page child" {
		t.Fatalf("content = %v, want Page child", paragraph["content"])
	}
}

func TestEmitInfersPageCountFromDocumentKids(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{
		FileName: "kid-only.pdf",
	})
	doc.Kids = []model.ContentElement{
		&model.Paragraph{
			TextNode: model.TextNode{
				BaseNode: model.BaseNode{
					ID:         1,
					Type:       model.ElementTypeParagraph,
					PageNumber: 4,
					Bounds:     model.NewBox(1, 2, 3, 4),
				},
				TextProperties: model.TextProperties{
					Font:      "Arial",
					FontSize:  7,
					TextColor: "#000000",
					Content:   "Kid child",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := New().Emit(nil, doc, &buf); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	root := mustMap(t, decodeJSON(t, buf.Bytes()))
	if got := int(root["number of pages"].(float64)); got != 4 {
		t.Fatalf("number of pages = %d, want 4", got)
	}
}

func decodeJSON(t *testing.T, data []byte) any {
	t.Helper()

	var out any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("output is not valid json: %v", err)
	}
	return out
}

func mustMap(t *testing.T, value any) map[string]any {
	t.Helper()

	m, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value is %T, want map[string]any", value)
	}
	return m
}

func mustSlice(t *testing.T, value any) []any {
	t.Helper()

	s, ok := value.([]any)
	if !ok {
		t.Fatalf("value is %T, want []any", value)
	}
	return s
}

func assertExactKeys(t *testing.T, value map[string]any, want []string) {
	t.Helper()

	if len(value) != len(want) {
		t.Fatalf("key count = %d, want %d (%v)", len(value), len(want), value)
	}
	for _, key := range want {
		if _, ok := value[key]; !ok {
			t.Fatalf("missing key %q in %v", key, value)
		}
	}
}

func assertBox(t *testing.T, value any, left, bottom, right, top float64) {
	t.Helper()

	coords := mustSlice(t, value)
	if len(coords) != 4 {
		t.Fatalf("bounding box length = %d, want 4", len(coords))
	}
	if coords[0].(float64) != left || coords[1].(float64) != bottom || coords[2].(float64) != right || coords[3].(float64) != top {
		t.Fatalf("bounding box = %v, want [%v %v %v %v]", coords, left, bottom, right, top)
	}
}

func strPtr(value string) *string {
	return &value
}
