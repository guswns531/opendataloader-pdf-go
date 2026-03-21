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

func TestDetectRejectsSentenceLikeParagraphWithHeadingStyling(t *testing.T) {
	elements := []model.ContentElement{
		paragraph(1, "Document Title", 22, true, model.NewBox(48, 760, 360, 792)),
		paragraph(2, "Important note for readers.", 16, true, model.NewBox(48, 714, 320, 736)),
		paragraph(3, "This is a supporting body paragraph that should remain untouched by the heading detector.", 10, false, model.NewBox(48, 650, 520, 670)),
	}

	detections := Detect(elements)
	if got, want := len(detections), 1; got != want {
		t.Fatalf("detections = %d, want %d", got, want)
	}
	if got := detections[0].Heading.Content; got != "Document Title" {
		t.Fatalf("detected heading = %q, want %q", got, "Document Title")
	}
}

func TestDetectAssignsSameLevelToSameSizeHeadings(t *testing.T) {
	elements := []model.ContentElement{
		paragraph(1, "Document Title", 24, true, model.NewBox(48, 760, 380, 792)),
		paragraph(2, "1. Overview", 17, false, model.NewBox(48, 716, 210, 736)),
		paragraph(3, "Details", 17, true, model.NewBox(48, 680, 220, 700)),
		paragraph(4, "This is a longer body paragraph with no heading cues and should not be detected.", 10, false, model.NewBox(48, 620, 520, 640)),
		paragraph(5, "Another ordinary paragraph that should keep the body font dominant on the page.", 10, false, model.NewBox(48, 580, 520, 600)),
	}

	detections := Detect(elements)
	if got, want := len(detections), 3; got != want {
		t.Fatalf("detections = %d, want %d", got, want)
	}

	byText := map[string]Detection{}
	for _, detection := range detections {
		byText[detection.Heading.Content] = detection
	}

	assertLevel(t, byText, "Document Title", 1)
	assertLevel(t, byText, "1. Overview", 2)
	assertLevel(t, byText, "Details", 2)
}

func TestDetectUsesRareSizeBucketAsAdditionalHeadingCue(t *testing.T) {
	elements := []model.ContentElement{
		paragraph(1, "Body paragraph one with normal dominant style.", 10, false, model.NewBox(48, 640, 520, 660)),
		paragraph(2, "Body paragraph two keeps the same body font size.", 10, false, model.NewBox(48, 610, 520, 630)),
		paragraph(3, "Body paragraph three still dominates the page style.", 10, false, model.NewBox(48, 580, 520, 600)),
		paragraph(4, "Methods", 12, false, model.NewBox(48, 710, 220, 728)),
	}

	detections := Detect(elements)
	if got, want := len(detections), 1; got != want {
		t.Fatalf("detections = %d, want %d", got, want)
	}
	if got := detections[0].Heading.Content; got != "Methods" {
		t.Fatalf("detected heading = %q, want %q", got, "Methods")
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
