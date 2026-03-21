package textclean

import (
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestCleanerString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		replacement string
		want        string
	}{
		{
			name:        "replaces invalid bytes",
			input:       string([]byte{'a', 0xff, 'b', 0xfe, 'c'}),
			replacement: "?",
			want:        "a?b?c",
		},
		{
			name:        "replaces replacement and invisible characters",
			input:       "x\uFFFBy\u200Bz\t\n\rw",
			replacement: "<bad>",
			want:        "x<bad>y<bad>z\t\n\rw",
		},
		{
			name:        "can remove characters entirely",
			input:       "a\x00b\x1fc",
			replacement: "",
			want:        "abc",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := New(tt.replacement).String(tt.input)
			if got != tt.want {
				t.Fatalf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCleanerDocument(t *testing.T) {
	t.Parallel()

	doc := &model.Document{
		Pages: []*model.Page{
			nil,
			{
				Metadata: model.PageMetadata{
					Index:  1,
					Number: 2,
					Label:  "keep-me",
				},
				Artifacts: []*model.RawArtifact{
					nil,
					{
						Kind: model.ArtifactKindText,
						Text: string([]byte{'p', 0xff, 'g'}),
						Style: model.TextProperties{
							Content: "style\uFFFD\u200B",
						},
					},
				},
				Kids: []model.ContentElement{
					&model.Paragraph{
						TextNode: model.TextNode{
							TextProperties: model.TextProperties{
								Content: string([]byte{'p', 0xff, '1'}),
							},
						},
					},
					&model.Heading{
						TextNode: model.TextNode{
							TextProperties: model.TextProperties{
								Content: "head\uFFFDto",
							},
						},
					},
					&model.Caption{
						TextNode: model.TextNode{
							TextProperties: model.TextProperties{
								Content: "cap\u200Btion",
							},
						},
					},
					&model.List{
						ListItems: []*model.ListItem{
							nil,
							{
								TextNode: model.TextNode{
									TextProperties: model.TextProperties{
										Content: "item\x00one",
									},
								},
								Kids: []model.ContentElement{
									&model.Paragraph{
										TextNode: model.TextNode{
											TextProperties: model.TextProperties{
												Content: "child\uFFFDbad",
											},
										},
									},
								},
							},
						},
					},
					&model.TextBlock{
						Kids: []model.ContentElement{
							&model.HeaderFooter{
								Kids: []model.ContentElement{
									&model.Paragraph{
										TextNode: model.TextNode{
											TextProperties: model.TextProperties{
												Content: "header\x1ffooter",
											},
										},
									},
								},
							},
						},
					},
					&model.Table{
						Rows: []model.TableRow{
							{
								Type: model.ElementTypeTableRow,
								Cells: []*model.TableCell{
									{
										Kids: []model.ContentElement{
											&model.Paragraph{
												TextNode: model.TextNode{
													TextProperties: model.TextProperties{
														Content: "cell\x00text",
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
			},
		},
		Kids: []model.ContentElement{
			&model.Paragraph{
				TextNode: model.TextNode{
					TextProperties: model.TextProperties{
						Content: "doc\x00root",
					},
				},
			},
		},
		Artifacts: []*model.RawArtifact{
			{
				Kind: model.ArtifactKindText,
				Text: "doc\uFFFDbad",
			},
		},
	}

	cleaner := New("*")
	if !cleaner.Document(doc) {
		t.Fatalf("Document() = false, want true")
	}

	page := doc.Pages[1]
	if page == nil {
		t.Fatalf("expected non-nil page after cleaning")
	}
	if got, want := page.Metadata.Label, "keep-me"; got != want {
		t.Fatalf("page metadata changed: got %q, want %q", got, want)
	}

	if got, want := page.Artifacts[1].Text, "p*g"; got != want {
		t.Fatalf("page artifact text = %q, want %q", got, want)
	}
	if got, want := page.Artifacts[1].Style.Content, "style**"; got != want {
		t.Fatalf("page artifact style content = %q, want %q", got, want)
	}

	if got, want := doc.Kids[0].(*model.Paragraph).TextProperties.Content, "doc*root"; got != want {
		t.Fatalf("document root paragraph content = %q, want %q", got, want)
	}
	if got, want := doc.Artifacts[0].Text, "doc*bad"; got != want {
		t.Fatalf("document artifact text = %q, want %q", got, want)
	}

	paragraph := page.Kids[0].(*model.Paragraph)
	if got, want := paragraph.TextProperties.Content, "p*1"; got != want {
		t.Fatalf("paragraph content = %q, want %q", got, want)
	}

	heading := page.Kids[1].(*model.Heading)
	if got, want := heading.TextProperties.Content, "head*to"; got != want {
		t.Fatalf("heading content = %q, want %q", got, want)
	}

	caption := page.Kids[2].(*model.Caption)
	if got, want := caption.TextProperties.Content, "cap*tion"; got != want {
		t.Fatalf("caption content = %q, want %q", got, want)
	}

	list := page.Kids[3].(*model.List)
	if got, want := list.ListItems[1].TextProperties.Content, "item*one"; got != want {
		t.Fatalf("list item content = %q, want %q", got, want)
	}
	if got, want := list.ListItems[1].Kids[0].(*model.Paragraph).TextProperties.Content, "child*bad"; got != want {
		t.Fatalf("nested list item content = %q, want %q", got, want)
	}

	headerFooter := page.Kids[4].(*model.TextBlock).Kids[0].(*model.HeaderFooter)
	if got, want := headerFooter.Kids[0].(*model.Paragraph).TextProperties.Content, "header*footer"; got != want {
		t.Fatalf("header/footer content = %q, want %q", got, want)
	}

	table := page.Kids[5].(*model.Table)
	if got, want := table.Rows[0].Cells[0].Kids[0].(*model.Paragraph).TextProperties.Content, "cell*text"; got != want {
		t.Fatalf("table cell content = %q, want %q", got, want)
	}
}
