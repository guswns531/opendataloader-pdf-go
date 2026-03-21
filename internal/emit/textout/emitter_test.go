package textout

import (
	"bytes"
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestEmitterNameAndFormat(t *testing.T) {
	emitter := New()

	if got := emitter.Name(); got != "textout" {
		t.Fatalf("Name() = %q, want textout", got)
	}
	if got := emitter.Format(); got != core.OutputFormatText {
		t.Fatalf("Format() = %q, want %q", got, core.OutputFormatText)
	}
}

func TestEmitRendersDeterministicPlainText(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{
		FileName:         "sample.pdf",
		Author:           strPtr("Hancom"),
		Title:            strPtr("Fixture Sample"),
		CreationDate:     strPtr("2026-01-01"),
		ModificationDate: strPtr("2026-01-02"),
		Producer:         strPtr("PDF Producer"),
		Creator:          strPtr("PDF Creator"),
		Subject:          strPtr("Example Subject"),
		Language:         strPtr("en"),
		Keywords:         []string{"alpha", "beta"},
		PageCount:        1,
	})

	doc.Kids = []model.ContentElement{
		&model.Heading{
			TextNode: model.TextNode{
				BaseNode: model.BaseNode{
					ID:         7,
					Type:       model.ElementTypeHeading,
					PageIndex:  0,
					PageNumber: 1,
				},
				TextProperties: model.TextProperties{
					Content: "Document heading",
				},
			},
			HeadingLevel: 1,
		},
	}

	prevTableID := model.NodeID(21)
	nextTableID := model.NodeID(22)
	page := doc.AddPage(&model.Page{
		Metadata: model.PageMetadata{
			Index:    0,
			Number:   1,
			Label:    "1",
			Size:     model.PageSize{Width: 612, Height: 792},
			Bounds:   model.NewBox(0, 0, 612, 792),
			Rotation: 90,
		},
		Artifacts: []*model.RawArtifact{
			{
				ID:         91,
				Kind:       model.ArtifactKindText,
				PageIndex:  0,
				PageNumber: 1,
				Bounds:     model.NewBox(10, 20, 30, 40),
				Boxes: model.MultiBox{
					model.NewBox(10, 20, 15, 25),
					model.NewBox(16, 26, 30, 40),
				},
				Text:   "page artifact",
				Format: model.ImageFormatPNG,
				Data:   []byte{1, 2, 3},
				Style: model.TextProperties{
					Content:    "artifact style",
					Font:       "Helvetica",
					FontSize:   10,
					TextColor:  "#000000",
					HiddenText: true,
					Bold:       true,
					Italic:     true,
					Underline:  true,
				},
				Links: model.LinkField{
					LinkedContentID: nodeIDPtr(77),
				},
				Sequence: 2,
			},
		},
		Kids: []model.ContentElement{
			&model.Paragraph{
				TextNode: model.TextNode{
					BaseNode: model.BaseNode{
						ID:         11,
						Type:       model.ElementTypeParagraph,
						PageIndex:  0,
						PageNumber: 1,
						Bounds:     model.NewBox(1, 2, 3, 4),
						Links: model.LinkField{
							NextID: nodeIDPtr(77),
						},
					},
					TextProperties: model.TextProperties{
						Content:  "Hello fixture",
						Font:     "Times-Roman",
						FontSize: 12,
					},
				},
			},
			&model.List{
				BaseNode: model.BaseNode{
					ID:         55,
					Type:       model.ElementTypeList,
					PageIndex:  0,
					PageNumber: 1,
				},
				NumberingStyle: "arabic",
				NumberOfItems:  2,
				ListItems: []*model.ListItem{
					{
						TextNode: model.TextNode{
							BaseNode: model.BaseNode{
								ID:         56,
								Type:       model.ElementTypeListItem,
								PageIndex:  0,
								PageNumber: 1,
							},
							TextProperties: model.TextProperties{
								Content: "First item",
							},
						},
						Kids: []model.ContentElement{
							&model.Caption{
								TextNode: model.TextNode{
									BaseNode: model.BaseNode{
										ID:         57,
										Type:       model.ElementTypeCaption,
										PageIndex:  0,
										PageNumber: 1,
									},
									TextProperties: model.TextProperties{
										Content: "Nested caption",
									},
								},
							},
						},
					},
				},
			},
			&model.Table{
				BaseNode: model.BaseNode{
					ID:         77,
					Type:       model.ElementTypeTable,
					PageIndex:  0,
					PageNumber: 1,
				},
				NumberOfRows:    1,
				NumberOfColumns: 1,
				PreviousTableID: &prevTableID,
				NextTableID:     &nextTableID,
				Rows: []model.TableRow{
					{
						Type:      model.ElementTypeTableRow,
						RowNumber: 1,
						Cells: []*model.TableCell{
							{
								BaseNode: model.BaseNode{
									ID:         12,
									Type:       model.ElementTypeTableCell,
									PageIndex:  0,
									PageNumber: 1,
									Bounds:     model.NewBox(55, 65, 150, 120),
									Links: model.LinkField{
										PreviousID: nodeIDPtr(99),
									},
								},
								RowNumber:    1,
								ColumnNumber: 1,
								RowSpan:      2,
								ColumnSpan:   3,
								Kids: []model.ContentElement{
									&model.Paragraph{
										TextNode: model.TextNode{
											TextProperties: model.TextProperties{
												Content: "Cell text",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if page == nil {
		t.Fatal("AddPage returned nil")
	}

	doc.Artifacts = append(doc.Artifacts, &model.RawArtifact{
		ID:         101,
		Kind:       model.ArtifactKindImage,
		PageIndex:  0,
		PageNumber: 1,
		Text:       "document artifact",
		Sequence:   9,
	})

	var buf bytes.Buffer
	if err := New().Emit(core.NewProcessingContext(doc, core.ProcessingOptions{}), doc, &buf); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	want := []string{
		"Document",
		"File: sample.pdf",
		"Page count: 1",
		"Author: Hancom",
		"Title: Fixture Sample",
		"Creation date: 2026-01-01",
		"Modification date: 2026-01-02",
		"Producer: PDF Producer",
		"Creator: PDF Creator",
		"Subject: Example Subject",
		"Language: en",
		"Keywords: alpha, beta",
		"Document content",
		"Heading #7",
		"ID: 7",
		"Type: heading",
		"Index: 0",
		"Page index: 0",
		"Page number: 1",
		"Content: Document heading",
		"Heading level: 1",
		"Document artifacts",
		"Artifact 101 (image)",
		"ID: 101",
		"Page index: 0",
		"Page number: 1",
		"Text: document artifact",
		"Sequence: 9",
		"Pages",
		"Page 1",
		"Index: 0",
		"Number: 1",
		"Label: 1",
		"Size: 612.00 x 792.00",
		"Bounds: 0.00, 0.00, 612.00, 792.00",
		"Rotation: 90",
		"Artifacts",
		"Artifact 91 (text)",
		"ID: 91",
		"Page index: 0",
		"Page number: 1",
		"Bounds: 10.00, 20.00, 30.00, 40.00",
		"Boxes: 10.00, 20.00, 15.00, 25.00; 16.00, 26.00, 30.00, 40.00",
		"Text: page artifact",
		"Format: png",
		"Data bytes: 3",
		"Content: artifact style",
		"Font: Helvetica",
		"Font size: 10",
		"Text color: #000000",
		"Bold: true",
		"Italic: true",
		"Underline: true",
		"Hidden text: true",
		"Links",
		"Linked content: 77",
		"Sequence: 2",
		"Content",
		"Paragraph #11",
		"ID: 11",
		"Type: paragraph",
		"Index: 0",
		"Page index: 0",
		"Page number: 1",
		"Bounds: 1.00, 2.00, 3.00, 4.00",
		"Links",
		"Next: 77",
		"Content: Hello fixture",
		"Font: Times-Roman",
		"Font size: 12",
		"List #55",
		"ID: 55",
		"Type: list",
		"Index: 0",
		"Page index: 0",
		"Page number: 1",
		"Numbering style: arabic",
		"Number of items: 2",
		"List item 1 #56",
		"Content: First item",
		"Caption #57",
		"ID: 57",
		"Type: caption",
		"Index: 0",
		"Page index: 0",
		"Page number: 1",
		"Content: Nested caption",
		"Table #77",
		"ID: 77",
		"Type: table",
		"Index: 0",
		"Page index: 0",
		"Page number: 1",
		"Rows: 1",
		"Columns: 1",
		"Previous table ID: 21",
		"Next table ID: 22",
		"Row 1",
		"Type: table row",
		"Table cell #12",
		"ID: 12",
		"Type: table cell",
		"Index: 0",
		"Page index: 0",
		"Page number: 1",
		"Bounds: 55.00, 65.00, 150.00, 120.00",
		"Links",
		"Previous: 99",
		"Row number: 1",
		"Column number: 1",
		"Row span: 2",
		"Column span: 3",
		"Paragraph",
		"ID: 0",
		"Index: 0",
		"Page index: 0",
		"Page number: 0",
		"Content: Cell text",
	}

	if got := normalizedLines(buf.String()); !slicesEqual(got, want) {
		t.Fatalf("Emit() normalized lines mismatch\nwant:\n%v\n\ngot:\n%v", want, got)
	}
}

func TestEmitFallsBackToPageContent(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{FileName: "sample.pdf"})
	page := doc.AddPage(&model.Page{
		Metadata: model.PageMetadata{
			Index:  0,
			Number: 1,
		},
	})
	page.Kids = []model.ContentElement{
		&model.Paragraph{
			TextNode: model.TextNode{
				TextProperties: model.TextProperties{
					Content: "Page scoped paragraph",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := New().Emit(nil, doc, &buf); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	want := []string{
		"Document",
		"File: sample.pdf",
		"Page count: 0",
		"Document content",
		"Paragraph",
		"ID: 0",
		"Index: 0",
		"Page index: 0",
		"Page number: 0",
		"Content: Page scoped paragraph",
		"Pages",
		"Page 1",
		"Index: 0",
		"Number: 1",
		"Rotation: 0",
		"Content",
		"Paragraph",
		"ID: 0",
		"Index: 0",
		"Page index: 0",
		"Page number: 0",
		"Content: Page scoped paragraph",
	}

	if got := normalizedLines(buf.String()); !slicesEqual(got, want) {
		t.Fatalf("Emit() normalized lines mismatch\nwant:\n%v\n\ngot:\n%v", want, got)
	}
}

func TestEmitRejectsNilInputs(t *testing.T) {
	if err := New().Emit(nil, nil, &bytes.Buffer{}); err == nil {
		t.Fatal("Emit(nil document) = nil error, want error")
	}

	if err := New().Emit(nil, model.NewDocument(model.DocumentMetadata{}), nil); err == nil {
		t.Fatal("Emit(nil writer) = nil error, want error")
	}
}

func strPtr(s string) *string {
	return &s
}

func nodeIDPtr(id model.NodeID) *model.NodeID {
	return &id
}

func normalizedLines(s string) []string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return lines
}

func slicesEqual[T comparable](left, right []T) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
