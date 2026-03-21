package pagefilter

import (
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestApplyReordersPagesAndDocumentContent(t *testing.T) {
	author := "A. Author"
	title := "Sample Report"
	doc := model.NewDocument(model.DocumentMetadata{
		FileName:  "sample.pdf",
		Author:    &author,
		Title:     &title,
		Keywords:  []string{"alpha", "beta"},
		PageCount: 3,
	})

	page1 := page(1, 0, "page-1", "artifact-1", "kid-1")
	page2 := page(2, 1, "page-2", "artifact-2", "kid-2")
	page3 := page(3, 2, "page-3", "artifact-3", "kid-3")
	doc.Pages = []*model.Page{page1, page2, page3}
	doc.Artifacts = []*model.RawArtifact{
		{PageNumber: 1, Text: "artifact-1"},
		{PageNumber: 2, Text: "artifact-2"},
		{PageNumber: 3, Text: "artifact-3"},
	}
	doc.Kids = []model.ContentElement{
		paragraph(1, 0, "kid-1"),
		paragraph(2, 1, "kid-2"),
		paragraph(3, 2, "kid-3"),
	}

	if err := Apply(doc, []model.PageNumber{3, 1, 3}); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if doc.Metadata.FileName != "sample.pdf" {
		t.Fatalf("file name = %q, want %q", doc.Metadata.FileName, "sample.pdf")
	}
	if doc.Metadata.Author == nil || *doc.Metadata.Author != author {
		t.Fatalf("author = %v, want %q", doc.Metadata.Author, author)
	}
	if doc.Metadata.Title == nil || *doc.Metadata.Title != title {
		t.Fatalf("title = %v, want %q", doc.Metadata.Title, title)
	}
	if got, want := doc.Metadata.Keywords, []string{"alpha", "beta"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("keywords = %v, want %v", got, want)
	}
	if doc.Metadata.PageCount != 3 {
		t.Fatalf("page count = %d, want 3", doc.Metadata.PageCount)
	}

	if got, want := len(doc.Pages), 3; got != want {
		t.Fatalf("len(doc.Pages) = %d, want %d", got, want)
	}
	if got, want := doc.Pages[0].Metadata.Number, model.PageNumber(3); got != want {
		t.Fatalf("doc.Pages[0] number = %d, want %d", got, want)
	}
	if got, want := doc.Pages[1].Metadata.Number, model.PageNumber(1); got != want {
		t.Fatalf("doc.Pages[1] number = %d, want %d", got, want)
	}
	if got, want := doc.Pages[2].Metadata.Number, model.PageNumber(3); got != want {
		t.Fatalf("doc.Pages[2] number = %d, want %d", got, want)
	}

	if got, want := len(doc.Kids), 3; got != want {
		t.Fatalf("len(doc.Kids) = %d, want %d", got, want)
	}
	if got, want := contentText(doc.Kids[0]), "kid-3"; got != want {
		t.Fatalf("doc.Kids[0] = %q, want %q", got, want)
	}
	if got, want := contentText(doc.Kids[1]), "kid-1"; got != want {
		t.Fatalf("doc.Kids[1] = %q, want %q", got, want)
	}
	if got, want := contentText(doc.Kids[2]), "kid-3"; got != want {
		t.Fatalf("doc.Kids[2] = %q, want %q", got, want)
	}

	if got, want := len(doc.Artifacts), 3; got != want {
		t.Fatalf("len(doc.Artifacts) = %d, want %d", got, want)
	}
	if got, want := doc.Artifacts[0].Text, "artifact-3"; got != want {
		t.Fatalf("doc.Artifacts[0] = %q, want %q", got, want)
	}
	if got, want := doc.Artifacts[1].Text, "artifact-1"; got != want {
		t.Fatalf("doc.Artifacts[1] = %q, want %q", got, want)
	}
	if got, want := doc.Artifacts[2].Text, "artifact-3"; got != want {
		t.Fatalf("doc.Artifacts[2] = %q, want %q", got, want)
	}
}

func TestApplyRejectsMissingOrInvalidPages(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{FileName: "sample.pdf"})
	doc.Pages = []*model.Page{
		page(1, 0, "page-1", "artifact-1", "kid-1"),
	}
	doc.Artifacts = []*model.RawArtifact{{PageNumber: 1, Text: "artifact-1"}}
	doc.Kids = []model.ContentElement{paragraph(1, 0, "kid-1")}

	if err := Apply(doc, []model.PageNumber{0}); err == nil {
		t.Fatal("Apply() with page 0 should fail")
	}
	if err := Apply(doc, []model.PageNumber{2}); err == nil {
		t.Fatal("Apply() with missing page should fail")
	}
	if got, want := len(doc.Pages), 1; got != want {
		t.Fatalf("document should remain unchanged on error, len(doc.Pages) = %d, want %d", got, want)
	}
}

func page(number model.PageNumber, index model.PageIndex, pageText, artifactText, kidText string) *model.Page {
	return &model.Page{
		Metadata: model.PageMetadata{
			Index:  index,
			Number: number,
			Label:  pageText,
		},
		Artifacts: []*model.RawArtifact{{PageNumber: number, Text: artifactText}},
		Kids:      []model.ContentElement{paragraph(number, index, kidText)},
	}
}

func paragraph(number model.PageNumber, index model.PageIndex, text string) model.ContentElement {
	return &model.Paragraph{
		TextNode: model.TextNode{
			BaseNode: model.BaseNode{
				PageNumber: number,
				PageIndex:  index,
			},
			TextProperties: model.TextProperties{Content: text},
		},
	}
}

func contentText(element model.ContentElement) string {
	if element == nil {
		return ""
	}
	switch node := element.(type) {
	case *model.Paragraph:
		return node.Content
	default:
		return ""
	}
}
