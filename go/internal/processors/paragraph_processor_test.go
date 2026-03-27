package processors

import (
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

func TestParagraphProcessorBreaksOnLargeLeading(t *testing.T) {
	processor := &ParagraphProcessor{}
	ctx := containers.NewProcessorContext()
	lines := []*entities.TextLine{
		testParagraphLine("First line", 10, 700, 100, 12, 12),
		testParagraphLine("Second line", 10, 684, 100, 12, 12),
		testParagraphLine("New paragraph", 10, 650, 100, 12, 12),
	}

	paragraphs := processor.Process(lines, ctx)

	if got := len(paragraphs); got != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", got)
	}
	if got := len(paragraphs[0].Lines); got != 2 {
		t.Fatalf("expected first paragraph to have 2 lines, got %d", got)
	}
	if paragraphs[1].Lines[0].GetText() != "New paragraph" {
		t.Fatalf("expected third line to start a new paragraph, got %q", paragraphs[1].Lines[0].GetText())
	}
}

func TestParagraphProcessorBreaksOnIndentChange(t *testing.T) {
	processor := &ParagraphProcessor{}
	ctx := containers.NewProcessorContext()
	lines := []*entities.TextLine{
		testParagraphLine("Body line one", 10, 700, 100, 12, 12),
		testParagraphLine("Body line two", 10, 684, 100, 12, 12),
		testParagraphLine("Indented start", 24, 668, 100, 12, 12),
	}

	paragraphs := processor.Process(lines, ctx)

	if got := len(paragraphs); got != 2 {
		t.Fatalf("expected indent change to split paragraphs, got %d", got)
	}
	if paragraphs[1].Lines[0].BBox.X != 24 {
		t.Fatalf("expected indented line to start second paragraph")
	}
}

func TestParagraphProcessorMergesCenteredLinesSeparatelyFromLeftAlignedBody(t *testing.T) {
	processor := &ParagraphProcessor{}
	ctx := containers.NewProcessorContext()
	lines := []*entities.TextLine{
		testParagraphLine("Centered title", 30, 700, 100, 12, 12),
		testParagraphLine("Centered subtitle", 50, 684, 60, 12, 12),
		testParagraphLine("Body starts", 10, 668, 100, 12, 12),
		testParagraphLine("Body continues", 10, 652, 92, 12, 12),
	}

	paragraphs := processor.Process(lines, ctx)

	if got := len(paragraphs); got != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", got)
	}
	if got := paragraphs[0].Alignment; got != entities.AlignCenter {
		t.Fatalf("expected centered paragraph alignment, got %q", got)
	}
	if got := paragraphs[1].Alignment; got != entities.AlignLeft {
		t.Fatalf("expected left paragraph alignment, got %q", got)
	}
}

func testParagraphLine(text string, x, y, width, height, fontSize float64) *entities.TextLine {
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
			FontSize: fontSize,
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
