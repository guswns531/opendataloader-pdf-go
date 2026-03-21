package layout

import (
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestApplyDropsOffPageAndTinyContent(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{FileName: "sample.pdf"})
	doc.Pages = []*model.Page{
		{
			Metadata: model.PageMetadata{
				Index:  0,
				Number: 1,
				Bounds: model.NewBox(0, 0, 100, 100),
			},
		},
	}

	doc.Kids = []model.ContentElement{
		newParagraph(1, 0, "keep", model.NewBox(10, 10, 20, 20), 12),
		newParagraph(1, 0, "drop-off", model.NewBox(120, 10, 130, 20), 12),
		newParagraph(1, 0, "drop-tiny", model.NewBox(30, 30, 31, 31), 0.5),
	}
	doc.Artifacts = []*model.RawArtifact{
		{
			PageNumber: 1,
			Kind:       model.ArtifactKindText,
			Bounds:     model.NewBox(110, 10, 120, 20),
			Text:       "drop-off",
			Style:      model.TextProperties{FontSize: 0.5},
		},
		{
			PageNumber: 1,
			Kind:       model.ArtifactKindText,
			Bounds:     model.NewBox(40, 40, 41, 41),
			Text:       "drop-tiny",
			Style:      model.TextProperties{FontSize: 0.5},
		},
		{
			PageNumber: 1,
			Kind:       model.ArtifactKindImage,
			Bounds:     model.NewBox(5, 5, 15, 15),
			Format:     model.ImageFormatPNG,
		},
	}

	doc.Pages[0].Kids = []model.ContentElement{
		&model.TextBlock{
			BaseNode: model.BaseNode{
				PageNumber: 1,
				PageIndex:  0,
				Bounds:     model.NewBox(0, 0, 100, 100),
			},
			Kids: []model.ContentElement{
				newParagraph(1, 0, "block-keep", model.NewBox(15, 15, 25, 25), 12),
				newParagraph(1, 0, "block-drop-off", model.NewBox(-20, 0, -10, 10), 12),
				newParagraph(1, 0, "block-drop-tiny", model.NewBox(20, 20, 21, 21), 0.5),
			},
		},
		&model.TextBlock{
			BaseNode: model.BaseNode{
				PageNumber: 1,
				PageIndex:  0,
				Bounds:     model.NewBox(120, 0, 140, 20),
			},
			Kids: []model.ContentElement{
				newParagraph(1, 0, "never-kept", model.NewBox(121, 1, 130, 10), 12),
			},
		},
	}
	doc.Pages[0].Artifacts = []*model.RawArtifact{
		{
			PageNumber: 1,
			Kind:       model.ArtifactKindImage,
			Bounds:     model.NewBox(5, 5, 15, 15),
			Format:     model.ImageFormatPNG,
		},
		{
			PageNumber: 1,
			Kind:       model.ArtifactKindText,
			Bounds:     model.NewBox(50, 50, 51, 51),
			Text:       "tiny",
			Style:      model.TextProperties{FontSize: 0.5},
		},
	}

	if err := Apply(doc); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if got, want := len(doc.Kids), 1; got != want {
		t.Fatalf("len(doc.Kids) = %d, want %d", got, want)
	}
	if got := paragraphText(doc.Kids[0]); got != "keep" {
		t.Fatalf("doc.Kids[0] text = %q, want %q", got, "keep")
	}

	if got, want := len(doc.Artifacts), 1; got != want {
		t.Fatalf("len(doc.Artifacts) = %d, want %d", got, want)
	}
	if doc.Artifacts[0].Kind != model.ArtifactKindImage {
		t.Fatalf("doc.Artifacts[0].Kind = %q, want %q", doc.Artifacts[0].Kind, model.ArtifactKindImage)
	}

	if got, want := len(doc.Pages[0].Kids), 1; got != want {
		t.Fatalf("len(page.Kids) = %d, want %d", got, want)
	}
	block, ok := doc.Pages[0].Kids[0].(*model.TextBlock)
	if !ok {
		t.Fatalf("page.Kids[0] type = %T, want *model.TextBlock", doc.Pages[0].Kids[0])
	}
	if got, want := len(block.Kids), 1; got != want {
		t.Fatalf("len(block.Kids) = %d, want %d", got, want)
	}
	if got := paragraphText(block.Kids[0]); got != "block-keep" {
		t.Fatalf("block.Kids[0] text = %q, want %q", got, "block-keep")
	}

	if got, want := len(doc.Pages[0].Artifacts), 1; got != want {
		t.Fatalf("len(page.Artifacts) = %d, want %d", got, want)
	}
	if doc.Pages[0].Artifacts[0].Kind != model.ArtifactKindImage {
		t.Fatalf("page.Artifacts[0].Kind = %q, want %q", doc.Pages[0].Artifacts[0].Kind, model.ArtifactKindImage)
	}
}

func TestApplyReturnsErrorOnNilDocument(t *testing.T) {
	if err := Apply(nil); err == nil {
		t.Fatal("Apply(nil) should fail")
	}
}

func TestApplyWithOptionsUsesFallbackThresholds(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{FileName: "sample.pdf"})
	doc.Pages = []*model.Page{
		{
			Metadata: model.PageMetadata{
				Index:  0,
				Number: 1,
				Bounds: model.NewBox(0, 0, 100, 100),
			},
		},
	}
	doc.Kids = []model.ContentElement{
		newParagraph(1, 0, "tiny", model.NewBox(10, 10, 11, 11), 0.75),
	}

	if err := ApplyWithOptions(doc, Options{}); err != nil {
		t.Fatalf("ApplyWithOptions() error = %v", err)
	}
	if got := len(doc.Kids); got != 0 {
		t.Fatalf("len(doc.Kids) = %d, want 0", got)
	}
}

func newParagraph(pageNumber model.PageNumber, pageIndex model.PageIndex, text string, bounds model.Box, fontSize float64) *model.Paragraph {
	return &model.Paragraph{
		TextNode: model.TextNode{
			BaseNode: model.BaseNode{
				PageNumber: pageNumber,
				PageIndex:  pageIndex,
				Bounds:     bounds,
			},
			TextProperties: model.TextProperties{
				Content:  text,
				FontSize: fontSize,
			},
		},
	}
}

func paragraphText(element model.ContentElement) string {
	if element == nil {
		return ""
	}
	switch node := element.(type) {
	case *model.Paragraph:
		return node.Content
	case *model.ListItem:
		return node.Content
	case *model.Heading:
		return node.Content
	case *model.Caption:
		return node.Content
	default:
		return ""
	}
}
