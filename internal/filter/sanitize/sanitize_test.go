package sanitize

import (
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestApplySanitizesSensitiveDataAcrossDocumentGraph(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{FileName: "sample.pdf"})
	page := &model.Page{
		Metadata: model.PageMetadata{
			Index:  0,
			Number: 1,
		},
		Artifacts: []*model.RawArtifact{
			{
				Text: "Reach me at alice@example.org or mailto:alice@example.org",
				Style: model.TextProperties{
					Content: "See https://internal.example.test/report",
				},
			},
		},
		Kids: []model.ContentElement{
			paragraph("Email jane.doe@example.com"),
			heading("Server 192.168.1.10"),
			caption("Call +1 (202) 555-0198"),
			&model.TextBlock{
				Kids: []model.ContentElement{
					&model.ListItem{
						TextNode: textNode("Card 4111 1111 1111 1111"),
						Kids: []model.ContentElement{
							&model.TableCell{
								Kids: []model.ContentElement{
									paragraph("IPv6 fe80::1"),
								},
							},
						},
					},
				},
			},
			&model.HeaderFooter{
				Kids: []model.ContentElement{
					paragraph("Visit www.example.org"),
				},
			},
			&model.Table{
				Rows: []model.TableRow{
					{
						Cells: []*model.TableCell{
							{
								Kids: []model.ContentElement{
									paragraph("Backup card 4242 4242 4242 4242"),
								},
							},
						},
					},
				},
			},
		},
	}
	doc.Pages = []*model.Page{page}
	doc.Artifacts = []*model.RawArtifact{
		{
			Text: "Backup server 10.0.0.5",
			Style: model.TextProperties{
				Content: "Open https://backup.example.net",
			},
		},
	}

	if err := Apply(doc); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if got, want := doc.Artifacts[0].Text, "Backup server [IP]"; got != want {
		t.Fatalf("doc artifact text = %q, want %q", got, want)
	}
	if got, want := doc.Artifacts[0].Style.Content, "Open [URL]"; got != want {
		t.Fatalf("doc artifact style content = %q, want %q", got, want)
	}

	if got, want := page.Artifacts[0].Text, "Reach me at [EMAIL] or [URL]"; got != want {
		t.Fatalf("page artifact text = %q, want %q", got, want)
	}
	if got, want := page.Artifacts[0].Style.Content, "See [URL]"; got != want {
		t.Fatalf("page artifact style content = %q, want %q", got, want)
	}

	if got, want := paragraphContent(page.Kids[0]), "Email [EMAIL]"; got != want {
		t.Fatalf("paragraph content = %q, want %q", got, want)
	}
	if got, want := headingContent(page.Kids[1]), "Server [IP]"; got != want {
		t.Fatalf("heading content = %q, want %q", got, want)
	}
	if got, want := captionContent(page.Kids[2]), "Call [PHONE]"; got != want {
		t.Fatalf("caption content = %q, want %q", got, want)
	}

	block := page.Kids[3].(*model.TextBlock)
	item := block.Kids[0].(*model.ListItem)
	if got, want := item.Content, "Card [CARD]"; got != want {
		t.Fatalf("list item content = %q, want %q", got, want)
	}
	cell := item.Kids[0].(*model.TableCell)
	if got, want := paragraphContent(cell.Kids[0]), "IPv6 [IP]"; got != want {
		t.Fatalf("nested paragraph content = %q, want %q", got, want)
	}

	header := page.Kids[4].(*model.HeaderFooter)
	if got, want := paragraphContent(header.Kids[0]), "Visit [URL]"; got != want {
		t.Fatalf("header/footer content = %q, want %q", got, want)
	}

	table := page.Kids[5].(*model.Table)
	if got, want := paragraphContent(table.Rows[0].Cells[0].Kids[0]), "Backup card [CARD]"; got != want {
		t.Fatalf("table cell content = %q, want %q", got, want)
	}
}

func TestApplyRejectsNilDocument(t *testing.T) {
	if err := Apply(nil); err == nil {
		t.Fatal("Apply(nil) should fail")
	}
}

func paragraphContent(element model.ContentElement) string {
	node, ok := element.(*model.Paragraph)
	if !ok {
		return ""
	}
	return node.Content
}

func headingContent(element model.ContentElement) string {
	node, ok := element.(*model.Heading)
	if !ok {
		return ""
	}
	return node.Content
}

func captionContent(element model.ContentElement) string {
	node, ok := element.(*model.Caption)
	if !ok {
		return ""
	}
	return node.Content
}

func textNode(content string) model.TextNode {
	return model.TextNode{
		TextProperties: model.TextProperties{
			Content: content,
		},
	}
}

func paragraph(content string) model.ContentElement {
	return &model.Paragraph{TextNode: textNode(content)}
}

func heading(content string) model.ContentElement {
	return &model.Heading{TextNode: textNode(content)}
}

func caption(content string) model.ContentElement {
	return &model.Caption{TextNode: textNode(content)}
}
