package readingorder

import (
	"fmt"
	"reflect"
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

func TestXYCutFallsBackAfterDeepDegenerateRecursion(t *testing.T) {
	const objectCount = maxSegmentDepth + 32

	objects := make([]entities.IObject, 0, objectCount)
	for i := 0; i < objectCount; i++ {
		x := 20.0 + float64(i)*15.0
		objects = append(objects, testTextChunk(fmt.Sprintf("item-%03d", i), x, 100, 8, 10))
	}

	sorter := XYCutPlusPlusSorter{}
	want := objectTexts(sorter.Sort(objects, 4000, 200))
	if len(want) != objectCount {
		t.Fatalf("expected %d objects, got %d", objectCount, len(want))
	}

	for run := 0; run < 5; run++ {
		got := objectTexts(sorter.Sort(objects, 4000, 200))
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("run %d produced unstable order: want %v, got %v", run, want, got)
		}
	}
}

func TestXYCutKeepsRightColumnImageGridGroupedAfterLeftTrailBlock(t *testing.T) {
	objects := []entities.IObject{
		testTextChunk("667", 14.04, 232.0, 22.32, 345.52),
		testTextChunk("646", 130.151, 652.242, 334.926, 36.597),
		testTextChunk("647", 82.271, 567.65, 434.445, 62.673),
		testTextChunk("648", 145.995, 528.628, 44.485, 15.554),
		testTextChunk("649", 50.112, 173.942, 236.25, 338.206),
		testTextChunk("650", 50.112, 129.636, 236.247, 27.13),
		testTextChunk("653", 315.944, 496.162, 70.864, 42.52),
		testTextChunk("654", 315.945, 452.648, 70.862, 42.519),
		testTextChunk("655", 315.945, 409.132, 70.863, 42.52),
		testTextChunk("657", 392.918, 496.162, 70.865, 42.52),
		testTextChunk("658", 392.918, 452.646, 70.865, 42.52),
		testTextChunk("659", 392.918, 409.13, 70.865, 42.52),
		testTextChunk("660", 469.89, 496.163, 70.87, 42.52),
		testTextChunk("661", 469.889, 452.647, 70.871, 42.52),
		testTextChunk("662", 469.89, 409.131, 70.869, 42.521),
		testTextChunk("656", 308.862, 360.946, 236.253, 49.36),
		testTextChunk("663", 308.862, 328.315, 76.836, 15.554),
		testTextChunk("664", 308.862, 200.233, 236.247, 121.538),
		testTextChunk("665", 308.862, 105.211, 236.247, 94.44),
		testTextChunk("666", 308.862, 77.935, 236.247, 26.694),
	}

	sorted := XYCutPlusPlusSorter{}.Sort(objects, 595, 842)

	positions := map[string]int{}
	for idx, text := range objectTexts(sorted) {
		positions[text] = idx
	}

	imageIDs := []string{"653", "654", "655", "657", "658", "659", "660", "661", "662"}
	minImagePos, maxImagePos := len(sorted), -1
	for _, id := range imageIDs {
		pos, ok := positions[id]
		if !ok {
			t.Fatalf("missing image %s in sort order %v", id, objectTexts(sorted))
		}
		if pos < minImagePos {
			minImagePos = pos
		}
		if pos > maxImagePos {
			maxImagePos = pos
		}
	}

	if maxImagePos-minImagePos != len(imageIDs)-1 {
		t.Fatalf("expected image grid to stay consecutive, got %v", objectTexts(sorted))
	}
	if positions["650"] >= minImagePos {
		t.Fatalf("expected 650 before image grid, got %v", objectTexts(sorted))
	}
	if maxImagePos >= positions["656"] {
		t.Fatalf("expected image grid before 656, got %v", objectTexts(sorted))
	}
	if !(positions["667"] < positions["650"] || positions["667"] > positions["666"]) {
		t.Fatalf("expected 667 outside the 650..666 body/grid band, got %v", objectTexts(sorted))
	}
	if !(positions["663"] < positions["664"] && positions["664"] < positions["665"] && positions["665"] < positions["666"]) {
		t.Fatalf("expected right-column continuation order after 656, got %v", objectTexts(sorted))
	}
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
