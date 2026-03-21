package htmlout

import (
	"bytes"
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestEmitterNameAndFormat(t *testing.T) {
	emitter := New()

	if got := emitter.Name(); got != "html" {
		t.Fatalf("Name() = %q, want html", got)
	}
	if got := emitter.Format(); got != core.OutputFormatHTML {
		t.Fatalf("Format() = %q, want %q", got, core.OutputFormatHTML)
	}
}

func TestEmitRendersSemanticHTML(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{
		FileName: "sample.pdf",
		Title:    strPtr("Fixture <Title>"),
	})
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
					Content: `First paragraph with <tags> & "quotes".`,
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
						cellWithParagraph("Beta & Gamma"),
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := New().Emit(nil, doc, &buf); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	want := strings.TrimSpace(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Fixture &lt;Title&gt;</title>
  </head>
  <body>
    <main>
      <h6>Overview</h6>
      <p>First paragraph with &lt;tags&gt; &amp; &quot;quotes&quot;.</p>
      <ol>
        <li>
          <p>First item</p>
        </li>
        <li>
          <p>Second item</p>
          <p>Nested paragraph</p>
          <ul>
            <li>
              <p>Nested bullet</p>
            </li>
          </ul>
        </li>
      </ol>
      <table>
        <tbody>
          <tr>
            <td>
              <p>Title</p>
            </td>
            <td>
              <p>Value</p>
            </td>
          </tr>
          <tr>
            <td>
              <p>Alpha</p>
            </td>
            <td>
              <p>Beta &amp; Gamma</p>
            </td>
          </tr>
        </tbody>
      </table>
    </main>
  </body>
</html>`)

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

	want := strings.TrimSpace(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>sample.pdf</title>
  </head>
  <body>
    <main>
      <p>Page scoped paragraph</p>
    </main>
  </body>
</html>`)

	if got := strings.TrimSpace(buf.String()); got != want {
		t.Fatalf("Emit() output mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestEmitRejectsNilInputs(t *testing.T) {
	if err := New().Emit(nil, nil, &bytes.Buffer{}); err == nil {
		t.Fatal("Emit(nil document) = nil error, want error")
	}

	if err := New().Emit(nil, model.NewDocument(model.DocumentMetadata{}), nil); err == nil {
		t.Fatal("Emit(nil writer) = nil error, want error")
	}

	var emitter *Emitter
	if err := emitter.Emit(nil, model.NewDocument(model.DocumentMetadata{}), &bytes.Buffer{}); err == nil {
		t.Fatal("nil emitter = nil error, want error")
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

func strPtr(s string) *string {
	return &s
}
