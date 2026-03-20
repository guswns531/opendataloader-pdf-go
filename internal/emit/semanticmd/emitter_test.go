package semanticmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestEmitterNameAndFormat(t *testing.T) {
	emitter := New()

	if got := emitter.Name(); got != "semanticmd" {
		t.Fatalf("Name() = %q, want semanticmd", got)
	}
	if got := emitter.Format(); got != core.OutputFormatMarkdown {
		t.Fatalf("Format() = %q, want %q", got, core.OutputFormatMarkdown)
	}
}

func TestEmitRendersSemanticMarkdown(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{FileName: "sample.pdf"})
	doc.Kids = []model.ContentElement{
		&model.Heading{
			TextNode: model.TextNode{
				TextProperties: model.TextProperties{
					Content: "Overview",
				},
			},
			HeadingLevel: 7,
		},
		&model.Paragraph{
			TextNode: model.TextNode{
				TextProperties: model.TextProperties{
					Content: "First paragraph with markdown *chars* and a pipe | marker.",
				},
			},
		},
		&model.List{
			BaseNode: model.BaseNode{
				Type: model.ElementTypeList,
			},
			NumberingStyle: "arabic",
			ListItems: []*model.ListItem{
				{
					TextNode: model.TextNode{
						TextProperties: model.TextProperties{
							Content: "First item",
						},
					},
				},
				{
					TextNode: model.TextNode{
						TextProperties: model.TextProperties{
							Content: "Second item",
						},
					},
					Kids: []model.ContentElement{
						&model.Paragraph{
							TextNode: model.TextNode{
								TextProperties: model.TextProperties{
									Content: "Nested paragraph",
								},
							},
						},
						&model.List{
							BaseNode: model.BaseNode{
								Type: model.ElementTypeList,
							},
							NumberingStyle: "unordered",
							ListItems: []*model.ListItem{
								{
									TextNode: model.TextNode{
										TextProperties: model.TextProperties{
											Content: "Nested bullet",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		&model.Table{
			BaseNode: model.BaseNode{
				Type: model.ElementTypeTable,
			},
			Rows: []model.TableRow{
				{
					Type:      model.ElementTypeTableRow,
					RowNumber: 1,
					Cells: []*model.TableCell{
						cellWithParagraph("Title"),
						cellWithParagraph("Value"),
					},
				},
				{
					Type:      model.ElementTypeTableRow,
					RowNumber: 2,
					Cells: []*model.TableCell{
						cellWithParagraph("Alpha"),
						cellWithParagraph("Beta | Gamma"),
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := New().Emit(nil, doc, &buf); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	want := strings.TrimSpace(`###### Overview

First paragraph with markdown \*chars\* and a pipe \| marker.

1. First item

1. Second item

    Nested paragraph

    - Nested bullet

| Title | Value |
| --- | --- |
| Alpha | Beta \| Gamma |`)

	if got := strings.TrimSpace(buf.String()); got != want {
		t.Fatalf("Emit() output mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
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

	if got, want := strings.TrimSpace(buf.String()), "Page scoped paragraph"; got != want {
		t.Fatalf("Emit() output = %q, want %q", got, want)
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

func cellWithParagraph(text string) *model.TableCell {
	return &model.TableCell{
		Kids: []model.ContentElement{
			&model.Paragraph{
				TextNode: model.TextNode{
					TextProperties: model.TextProperties{
						Content: text,
					},
				},
			},
		},
	}
}
