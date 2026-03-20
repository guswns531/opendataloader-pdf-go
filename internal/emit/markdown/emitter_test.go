package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestEmitterNameAndFormat(t *testing.T) {
	emitter := New()

	if got := emitter.Name(); got != "markdown" {
		t.Fatalf("Name() = %q, want markdown", got)
	}
	if got := emitter.Format(); got != core.OutputFormatMarkdown {
		t.Fatalf("Format() = %q, want %q", got, core.OutputFormatMarkdown)
	}
}

func TestEmitRendersDeterministicMarkdown(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{
		FileName:  "sample.pdf",
		Author:    strPtr("Hancom"),
		Title:     strPtr("Fixture Sample"),
		PageCount: 1,
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

	if page := doc.AddPage(&model.Page{
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
				Text:       "page artifact",
				Data:       []byte{1, 2, 3},
				Style: model.TextProperties{
					Content:  "artifact style",
					Font:     "Helvetica",
					FontSize: 10,
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
			&model.Table{
				BaseNode: model.BaseNode{
					ID:         77,
					Type:       model.ElementTypeTable,
					PageIndex:  0,
					PageNumber: 1,
				},
				NumberOfRows:    1,
				NumberOfColumns: 1,
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
								},
								RowNumber:    1,
								ColumnNumber: 1,
								RowSpan:      1,
								ColumnSpan:   1,
								Kids: []model.ContentElement{
									&model.Caption{
										TextNode: model.TextNode{
											BaseNode: model.BaseNode{
												ID:         13,
												Type:       model.ElementTypeCaption,
												PageIndex:  0,
												PageNumber: 1,
											},
											TextProperties: model.TextProperties{
												Content: "Inside cell",
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
	}); page == nil {
		t.Fatal("AddPage returned nil")
	}

	var buf bytes.Buffer
	if err := New().Emit(nil, doc, &buf); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	want := strings.TrimSpace(`# Document
- File: sample.pdf
- Page count: 1
- Author: Hancom
- Title: Fixture Sample
## Document Content
- Heading #7
  - ID: 7
  - Type: heading
  - Index: 0
  - Page index: 0
  - Page number: 1
  - Content: Document heading
  - Heading level: 1
## Pages
### Page 1
  - Index: 0
  - Number: 1
  - Label: 1
  - Size: 612.00 x 792.00
  - Bounds: 0.00, 0.00, 612.00, 792.00
  - Rotation: 90
  #### Artifacts
  - Artifact 91 (text)
    - ID: 91
    - Page index: 0
    - Page number: 1
    - Bounds: 10.00, 20.00, 30.00, 40.00
    - Text: page artifact
    - Data bytes: 3
    - Content: artifact style
    - Font: Helvetica
    - Font size: 10
    - Links
      - Linked content: 77
    - Sequence: 2
  #### Content
  - Paragraph #11
    - ID: 11
    - Type: paragraph
    - Index: 0
    - Page index: 0
    - Page number: 1
    - Bounds: 1.00, 2.00, 3.00, 4.00
    - Links
      - Next: 77
    - Content: Hello fixture
    - Font: Times-Roman
    - Font size: 12
  - Table #77
    - ID: 77
    - Type: table
    - Index: 0
    - Page index: 0
    - Page number: 1
    - Rows: 1
    - Columns: 1
    - Row 1
      - Type: table row
      - Table cell #12
        - ID: 12
        - Type: table cell
        - Index: 0
        - Page index: 0
        - Page number: 1
        - Row number: 1
        - Column number: 1
        - Row span: 1
        - Column span: 1
        - Caption #13
          - ID: 13
          - Type: caption
          - Index: 0
          - Page index: 0
          - Page number: 1
          - Content: Inside cell`)

	if got := strings.TrimSpace(buf.String()); got != want {
		t.Fatalf("Emit() output mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestEmitRejectsNilDocument(t *testing.T) {
	if err := New().Emit(nil, nil, &bytes.Buffer{}); err == nil {
		t.Fatal("Emit() with nil document = nil error, want error")
	}
}

func strPtr(s string) *string {
	return &s
}

func nodeIDPtr(id model.NodeID) *model.NodeID {
	return &id
}
