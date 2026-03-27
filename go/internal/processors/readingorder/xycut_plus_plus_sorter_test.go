package readingorder

import (
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

func TestXYCutSortsTwoColumnsLeftThenRight(t *testing.T) {
	objects := []entities.IObject{
		testTextChunk("left-top", 20, 160, 40, 10),
		testTextChunk("right-top", 120, 160, 40, 10),
		testTextChunk("left-bottom", 20, 120, 40, 10),
		testTextChunk("right-bottom", 120, 120, 40, 10),
	}

	sorted := XYCutPlusPlusSorter{}.Sort(objects, 200, 240)

	assertObjectTexts(t, sorted, []string{
		"left-top",
		"left-bottom",
		"right-top",
		"right-bottom",
	})
}

func TestXYCutRecursivelySortsMultipleColumnsLeftToRight(t *testing.T) {
	objects := []entities.IObject{
		testTextChunk("col1-top", 20, 160, 30, 10),
		testTextChunk("col2-top", 120, 160, 30, 10),
		testTextChunk("col3-top", 220, 160, 30, 10),
		testTextChunk("col1-bottom", 20, 120, 30, 10),
		testTextChunk("col2-bottom", 120, 120, 30, 10),
		testTextChunk("col3-bottom", 220, 120, 30, 10),
	}

	sorted := XYCutPlusPlusSorter{}.Sort(objects, 320, 240)

	assertObjectTexts(t, sorted, []string{
		"col1-top",
		"col1-bottom",
		"col2-top",
		"col2-bottom",
		"col3-top",
		"col3-bottom",
	})
}

func TestXYCutDetectsTwoColumnsWhenGutterLandsOnHistogramBoundary(t *testing.T) {
	objects := []entities.IObject{
		testTextChunk("left-top", 150, 160, 46, 10),
		testTextChunk("right-top", 211, 160, 46, 10),
		testTextChunk("left-bottom", 150, 120, 46, 10),
		testTextChunk("right-bottom", 211, 120, 46, 10),
	}

	sorted := XYCutPlusPlusSorter{}.Sort(objects, 500, 240)

	assertObjectTexts(t, sorted, []string{
		"left-top",
		"left-bottom",
		"right-top",
		"right-bottom",
	})
}

func TestXYCutKeepsBasicTwoColumnsGroupedWithModestGutter(t *testing.T) {
	objects := []entities.IObject{
		testTextChunk("left-top", 50, 160, 250, 20),
		testTextChunk("right-top", 310, 160, 250, 20),
		testTextChunk("left-bottom", 50, 120, 250, 20),
		testTextChunk("right-bottom", 310, 120, 250, 20),
	}

	sorted := XYCutPlusPlusSorter{}.Sort(objects, 600, 240)

	assertObjectTexts(t, sorted, []string{
		"left-top",
		"left-bottom",
		"right-top",
		"right-bottom",
	})
}

func TestXYCutBalancesImbalancedColumnStartAndEnd(t *testing.T) {
	objects := []entities.IObject{
		testTextChunk("right-lead", 120, 200, 40, 10),
		testTextChunk("left-top", 20, 160, 40, 10),
		testTextChunk("right-top", 120, 160, 40, 10),
		testTextChunk("left-bottom", 20, 120, 40, 10),
		testTextChunk("right-bottom", 120, 120, 40, 10),
		testTextChunk("left-trail", 20, 80, 40, 10),
	}

	sorted := XYCutPlusPlusSorter{}.Sort(objects, 200, 260)

	assertObjectTexts(t, sorted, []string{
		"right-lead",
		"left-top",
		"left-bottom",
		"right-top",
		"right-bottom",
		"left-trail",
	})
}

func TestXYCutTreatsGutterSpanningBlocksAsNeutralDuringColumnMerge(t *testing.T) {
	objects := []entities.IObject{
		testTextChunk("center-heading", 70, 180, 60, 10),
		testTextChunk("left-top", 20, 160, 40, 10),
		testTextChunk("right-top", 120, 160, 40, 10),
		testTextChunk("left-bottom", 20, 120, 40, 10),
		testTextChunk("right-bottom", 120, 120, 40, 10),
	}

	sorted := XYCutPlusPlusSorter{}.Sort(objects, 200, 240)

	assertObjectTexts(t, sorted, []string{
		"center-heading",
		"left-top",
		"left-bottom",
		"right-top",
		"right-bottom",
	})
}

func TestXYCutDefersMarginalSidebarUntilAfterMainBodyBand(t *testing.T) {
	objects := []entities.IObject{
		testTextChunk("sidebar", 10, 95, 18, 85),
		testTextChunk("left-top", 70, 170, 35, 10),
		testTextChunk("right-top", 125, 170, 35, 10),
		testTextChunk("left-bottom", 70, 130, 35, 10),
		testTextChunk("right-bottom", 125, 130, 35, 10),
		testTextChunk("footer", 70, 60, 90, 10),
	}

	sorted := XYCutPlusPlusSorter{}.Sort(objects, 200, 240)

	assertObjectTexts(t, sorted, []string{
		"left-top",
		"left-bottom",
		"right-top",
		"right-bottom",
		"sidebar",
		"footer",
	})
}

func testTextChunk(id string, x, y, width, height float64) *entities.TextChunk {
	return &entities.TextChunk{
		BaseObject: entities.BaseObject{
			ID: id,
			BBox: entities.BoundingBox{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
			},
		},
		Text: id,
	}
}

func assertObjectTexts(t *testing.T, objects []entities.IObject, want []string) {
	t.Helper()

	if len(objects) != len(want) {
		t.Fatalf("expected %d objects, got %d", len(want), len(objects))
	}

	for idx, object := range objects {
		chunk, ok := object.(*entities.TextChunk)
		if !ok {
			t.Fatalf("object %d has unexpected type %T", idx, object)
		}
		if chunk.Text != want[idx] {
			t.Fatalf("expected order %v, got %v", want, objectTexts(objects))
		}
	}
}

func objectTexts(objects []entities.IObject) []string {
	out := make([]string, 0, len(objects))
	for _, object := range objects {
		if chunk, ok := object.(*entities.TextChunk); ok {
			out = append(out, chunk.Text)
		}
	}
	return out
}
