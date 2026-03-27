package processors

import (
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

func TestSpecialTableProcessorAcceptsTwoRowAlignedTable(t *testing.T) {
	processor := &SpecialTableProcessor{}
	ctx := containers.NewProcessorContext()

	elements := []entities.IObject{
		testLine("row1", specialTestChunk("DAVAO", 50, 120, 40, 10, 120), specialTestChunk("750", 120, 120, 20, 10, 120), specialTestChunk("17,807", 170, 120, 30, 10, 120)),
		testLine("row2", specialTestChunk("ILOILO", 50, 100, 45, 10, 100), specialTestChunk("212", 120, 100, 20, 10, 100), specialTestChunk("24,381", 170, 100, 30, 10, 100)),
	}

	result := processor.Process(elements, ctx)
	if len(result) != 1 {
		t.Fatalf("expected a table, got %d elements", len(result))
	}
	table, ok := result[0].(*entities.SemanticTable)
	if !ok {
		t.Fatalf("expected semantic table, got %T", result[0])
	}
	if len(table.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(table.Rows))
	}
}

func TestSpecialTableProcessorAcceptsCaptionedTwoColumnTableWithoutNumbers(t *testing.T) {
	processor := &SpecialTableProcessor{}
	ctx := containers.NewProcessorContext()

	elements := []entities.IObject{
		testLine("caption", specialTestChunk("Table 2: Example", 40, 170, 90, 10, 170)),
		testLine("row1", specialTestChunk("Label A", 50, 140, 35, 10, 140), specialTestChunk("Value one", 130, 140, 45, 10, 140)),
		testLine("row2", specialTestChunk("Label B", 50, 120, 35, 10, 120), specialTestChunk("Value two", 130, 120, 45, 10, 120)),
		testLine("row3", specialTestChunk("Label C", 50, 100, 35, 10, 100), specialTestChunk("Value three", 130, 100, 55, 10, 100)),
	}

	result := processor.Process(elements, ctx)
	if len(result) != 2 {
		t.Fatalf("expected table plus caption, got %d elements", len(result))
	}
	foundTable := false
	for _, element := range result {
		if _, ok := element.(*entities.SemanticTable); ok {
			foundTable = true
			break
		}
	}
	if !foundTable {
		t.Fatalf("expected one semantic table in result, got %T and %T", result[0], result[1])
	}
}

func specialTestChunk(text string, x, y, width, height, baseline float64) *entities.TextChunk {
	return &entities.TextChunk{
		BaseObject: entities.BaseObject{
			ID: text,
			BBox: entities.BoundingBox{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
				Page:   0,
			},
		},
		Text:     text,
		Baseline: baseline,
		FontStyle: entities.FontStyle{
			FontSize: 10,
		},
	}
}

func testLine(id string, chunks ...*entities.TextChunk) *entities.TextLine {
	box := chunks[0].BBox
	for _, c := range chunks[1:] {
		if c.BBox.X < box.X {
			box.X = c.BBox.X
		}
		right := c.BBox.X + c.BBox.Width
		boxRight := box.X + box.Width
		if right > boxRight {
			box.Width = right - box.X
		}
		if c.BBox.Y < box.Y {
			box.Y = c.BBox.Y
		}
		top := c.BBox.Y + c.BBox.Height
		boxTop := box.Y + box.Height
		if top > boxTop {
			box.Height = top - box.Y
		}
	}
	return &entities.TextLine{
		BaseObject: entities.BaseObject{
			ID:   id,
			BBox: box,
		},
		Chunks:   chunks,
		Baseline: chunks[0].Baseline,
	}
}
