package processors

import (
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/utils"
)

func TestPromoteListHeadingsReclassifiesOrderedSectionTitle(t *testing.T) {
	processor := &HeadingProcessor{}
	ctx := containers.NewProcessorContext()

	intro := testHeadingLine("Intro paragraph", 10, 720, 180, 12, 12, false)
	sectionLine := testHeadingLine("Diesel and biodiesel use", 30, 680, 180, 12, 12, false)
	bodyLine := testHeadingLine("The consumption of diesel fuel in Indonesia...", 30, 664, 260, 12, 12, false)
	outro := testHeadingLine("Body continues after the section.", 10, 620, 220, 12, 12, false)

	list := &entities.PDFList{
		BaseObject: entities.BaseObject{ID: "list-1", BBox: mergeBoxes(sectionLine.BBox, bodyLine.BBox)},
		IsOrdered:  true,
		Items: []*entities.ListItem{
			{
				BaseObject: entities.BaseObject{ID: "item-1", BBox: mergeBoxes(sectionLine.BBox, bodyLine.BBox)},
				BulletText: "2.1.",
				IsOrdered:  true,
				Level:      1,
				Content:    []entities.IObject{sectionLine, bodyLine},
			},
		},
	}

	elements := []entities.IObject{intro, list, outro}
	promoted, headings := processor.PromoteListHeadings(elements, ctx)

	if got := len(headings); got != 1 {
		t.Fatalf("expected 1 promoted heading, got %d", got)
	}
	if got := headings[0].Lines[0].GetText(); got != "2.1. Diesel and biodiesel use" {
		t.Fatalf("unexpected promoted heading text: %q", got)
	}
	if _, ok := promoted[1].(*entities.SemanticHeading); !ok {
		t.Fatalf("expected promoted heading at index 1, got %T", promoted[1])
	}
	if _, ok := promoted[2].(*entities.SemanticParagraph); !ok {
		t.Fatalf("expected list body to remain as paragraph, got %T", promoted[2])
	}
}

func TestHeadingScoreRejectsSingleTokenMathFragments(t *testing.T) {
	processor := &HeadingProcessor{}
	stats := utils.NewTextNodeStatistics()
	body := testHeadingLine("Regular body text", 10, 700, 180, 12, 12, false)
	stats.Add(lineFontSize(body), lineFontWeight(body))

	candidate := testHeadingLine("p", 10, 680, 12, 12, 12, false)
	score := processor.headingScore(candidate, body, nil, 12, lineFontFamily(body), stats)
	if score != 0 {
		t.Fatalf("expected math fragment to be rejected with zero score, got %.2f", score)
	}
}

func TestLooksLikeSectionHeadingTextMatchesPrefixesWithoutTrailingDot(t *testing.T) {
	cases := []string{
		"8 Choosing between Observer Models and Rejecting Participants",
		"12 Conclusion",
		"1.2 Background",
		"IV Methods",
		"IV. Methods",
	}
	for _, text := range cases {
		if !looksLikeSectionHeadingText(text) {
			t.Fatalf("expected section heading match for %q", text)
		}
	}
}

func TestLooksLikeSectionHeadingTextRejectsFootnoteLikeAndProsePrefixes(t *testing.T) {
	cases := []string{
		"18 ., <SimultaneityNoisyCriteriaMultistart 225-386>",
		"I have presented two variants of a latency-based observer model",
		"12 .",
	}
	for _, text := range cases {
		if looksLikeSectionHeadingText(text) {
			t.Fatalf("expected non-heading prefix rejection for %q", text)
		}
	}
}

func TestHeadingScoreBoostsColonSuffixedTitleLines(t *testing.T) {
	processor := &HeadingProcessor{}
	stats := utils.NewTextNodeStatistics()
	body := testHeadingLine("Regular body text for statistics", 10, 700, 220, 12, 12, false)
	stats.Add(lineFontSize(body), lineFontWeight(body))

	colonTitle := testHeadingLine("Steps for Using the Microscope:", 10, 680, 190, 12, 12, false)
	plainLine := testHeadingLine("Steps for Using the Microscope", 10, 680, 190, 12, 12, false)

	colonScore := processor.headingScore(colonTitle, body, nil, 12, lineFontFamily(body), stats)
	plainScore := processor.headingScore(plainLine, body, nil, 12, lineFontFamily(body), stats)

	if colonScore <= plainScore {
		t.Fatalf("expected colon-suffixed title to score higher: colon=%.2f plain=%.2f", colonScore, plainScore)
	}
}

func TestHeadingScoreDoesNotBoostColonSuffixedNumberedListItems(t *testing.T) {
	processor := &HeadingProcessor{}
	stats := utils.NewTextNodeStatistics()
	body := testHeadingLine("Regular body text for statistics", 10, 700, 220, 12, 12, false)
	stats.Add(lineFontSize(body), lineFontWeight(body))

	numbered := testHeadingLine("1. Steps for Using the Microscope:", 10, 680, 210, 12, 12, false)
	plain := testHeadingLine("1. Steps for Using the Microscope", 10, 680, 210, 12, 12, false)

	numberedScore := processor.headingScore(numbered, body, nil, 12, lineFontFamily(body), stats)
	plainScore := processor.headingScore(plain, body, nil, 12, lineFontFamily(body), stats)

	if numberedScore != plainScore {
		t.Fatalf("expected numbered colon line to avoid boost: colon=%.2f plain=%.2f", numberedScore, plainScore)
	}
}

func testHeadingLine(text string, x, y, width, height, fontSize float64, bold bool) *entities.TextLine {
	chunk := &entities.TextChunk{
		BaseObject: entities.BaseObject{
			ID: "chunk-" + text,
			BBox: entities.BoundingBox{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
				Page:   1,
			},
		},
		Text: text,
		FontStyle: entities.FontStyle{
			FontName: "Times-Roman",
			FontSize: fontSize,
			Bold:     bold,
		},
		Baseline: y + height,
	}
	return &entities.TextLine{
		BaseObject: entities.BaseObject{
			ID:   "line-" + text,
			BBox: chunk.BBox,
		},
		Chunks:   []*entities.TextChunk{chunk},
		Baseline: chunk.Baseline,
	}
}
