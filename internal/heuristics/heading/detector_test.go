package heading

import (
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestDetectFindsHeadingsAndClustersLevels(t *testing.T) {
	elements := []model.ContentElement{
		paragraph(1, "Document Title", 22, true, model.NewBox(48, 760, 360, 792)),
		paragraph(2, "1. Introduction", 16, true, model.NewBox(48, 710, 220, 732)),
		paragraph(3, "2. Methods", 16, true, model.NewBox(48, 676, 180, 698)),
		paragraph(4, "This is a body paragraph that should stay a paragraph because it is longer, plain, and sentence-like.", 10, false, model.NewBox(48, 620, 520, 640)),
	}

	detections := Detect(elements)
	if got, want := len(detections), 3; got != want {
		t.Fatalf("detections = %d, want %d", got, want)
	}

	byText := map[string]Detection{}
	for _, detection := range detections {
		byText[detection.Heading.Content] = detection
		if detection.Heading.Content == "This is a body paragraph that should stay a paragraph because it is longer, plain, and sentence-like." {
			t.Fatal("body paragraph was incorrectly classified as a heading")
		}
	}

	assertLevel(t, byText, "Document Title", 1)
	assertLevel(t, byText, "1. Introduction", 2)
	assertLevel(t, byText, "2. Methods", 2)

	if got, want := byText["Document Title"].Score, 0.0; got <= want {
		t.Fatalf("title score = %v, want > %v", got, want)
	}
}

func TestDetectRejectsLongPlainParagraph(t *testing.T) {
	elements := []model.ContentElement{
		paragraph(1, "This is a long body paragraph with no heading cues, even though it uses a short page-wide box and looks like normal prose.", 11, false, model.NewBox(48, 640, 520, 664)),
	}

	detections := Detect(elements)
	if len(detections) != 0 {
		t.Fatalf("detections = %d, want 0", len(detections))
	}
}

func assertLevel(t *testing.T, byText map[string]Detection, text string, want int) {
	t.Helper()
	detection, ok := byText[text]
	if !ok {
		t.Fatalf("missing detection for %q", text)
	}
	if got := detection.Heading.HeadingLevel; got != want {
		t.Fatalf("heading level for %q = %d, want %d", text, got, want)
	}
}

func paragraph(id int64, text string, size float64, bold bool, bounds model.Box) *model.Paragraph {
	return &model.Paragraph{
		TextNode: model.TextNode{
			BaseNode: model.BaseNode{
				ID:         model.NodeID(id),
				Type:       model.ElementTypeParagraph,
				PageIndex:  0,
				PageNumber: 1,
				Bounds:     bounds,
			},
			TextProperties: model.TextProperties{
				Content:  text,
				FontSize: size,
				Bold:     bold,
			},
		},
	}
}
