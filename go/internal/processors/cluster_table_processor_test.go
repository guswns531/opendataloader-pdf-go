package processors

import (
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

func TestClusterColumnStarts(t *testing.T) {
	input := []float64{10, 12, 17, 40, 44, 70}

	got := clusterColumnStarts(input, 5)

	want := []float64{10, 40, 70}
	if len(got) != len(want) {
		t.Fatalf("clusterColumnStarts() len = %d, want %d (%v)", len(got), len(want), got)
	}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("clusterColumnStarts()[%d] = %v, want %v (full=%v)", idx, got[idx], want[idx], got)
		}
	}
}

func TestHasPossibleTable(t *testing.T) {
	paragraphLike := []entities.IObject{
		testChunk("a", 10, 100, 10, 10, 100),
		testChunk("b", 23, 100, 10, 10, 100),
	}
	if hasPossibleTable(paragraphLike) {
		t.Fatal("hasPossibleTable() = true for paragraph-like spacing")
	}

	tableLike := []entities.IObject{
		testChunk("a", 10, 100, 10, 10, 100),
		testChunk("b", 55, 100, 10, 10, 100),
	}
	if !hasPossibleTable(tableLike) {
		t.Fatal("hasPossibleTable() = false for table-like spacing")
	}
}

func testChunk(id string, x, y, width, height, baseline float64) *entities.TextChunk {
	return &entities.TextChunk{
		BaseObject: entities.BaseObject{
			ID: id,
			BBox: entities.BoundingBox{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
				Page:   0,
			},
		},
		Text:     id,
		Baseline: baseline,
	}
}
