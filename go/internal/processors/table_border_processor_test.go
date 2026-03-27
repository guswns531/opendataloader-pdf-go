package processors

import (
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

func TestDetectFallbackEnclosingTableRejectsBorderOnlyBox(t *testing.T) {
	lineArts := []*entities.LineArtChunk{
		lineArt(10, 100, 90, 1, true, false),
		lineArt(10, 10, 90, 1, true, false),
		lineArt(10, 10, 1, 90, false, true),
		lineArt(100, 10, 1, 90, false, true),
	}

	_, ok := detectFallbackEnclosingTable(lineArts)
	if ok {
		t.Fatalf("expected border-only enclosing box to be rejected")
	}
}

func TestDetectFallbackEnclosingTableRequiresInternalStructure(t *testing.T) {
	lineArts := []*entities.LineArtChunk{
		lineArt(10, 100, 90, 1, true, false),
		lineArt(10, 10, 90, 1, true, false),
		lineArt(10, 55, 90, 1, true, false),
		lineArt(10, 10, 1, 90, false, true),
		lineArt(100, 10, 1, 90, false, true),
		lineArt(55, 10, 1, 90, false, true),
	}

	table, ok := detectFallbackEnclosingTable(lineArts)
	if !ok {
		t.Fatalf("expected enclosing box with internal dividers to be accepted")
	}
	if len(table.rowBounds) != 3 {
		t.Fatalf("expected 3 row bounds, got %d", len(table.rowBounds))
	}
	if len(table.colBounds) != 3 {
		t.Fatalf("expected 3 col bounds, got %d", len(table.colBounds))
	}
}

func TestTopLevelTablesDiscardsNestedCandidates(t *testing.T) {
	outer := detectedTable{
		bbox:      entities.BoundingBox{X: 10, Y: 10, Width: 90, Height: 90, Page: 1},
		rowBounds: []float64{100, 55, 10},
		colBounds: []float64{10, 55, 100},
	}
	inner := detectedTable{
		bbox:      entities.BoundingBox{X: 10, Y: 55, Width: 45, Height: 45, Page: 1},
		rowBounds: []float64{100, 55},
		colBounds: []float64{10, 55},
	}

	result := topLevelTables([]detectedTable{inner, outer})
	if len(result) != 1 {
		t.Fatalf("expected only outer table to remain, got %d", len(result))
	}
	if !sameBBox(result[0].bbox, outer.bbox) {
		t.Fatalf("expected outer table to remain")
	}
}

func TestIsMeaningfulBorderTableRequiresAtLeastTwoRowsAndColumns(t *testing.T) {
	if isMeaningfulBorderTable(detectedTable{
		rowBounds: []float64{100, 50},
		colBounds: []float64{10, 60, 100},
	}) {
		t.Fatalf("expected single-row candidate to be rejected")
	}
	if isMeaningfulBorderTable(detectedTable{
		rowBounds: []float64{100, 50, 10},
		colBounds: []float64{10, 100},
	}) {
		t.Fatalf("expected single-column candidate to be rejected")
	}
}

func lineArt(x, y, width, height float64, horizontal, vertical bool) *entities.LineArtChunk {
	return &entities.LineArtChunk{
		BaseObject: entities.BaseObject{
			BBox: entities.BoundingBox{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
				Page:   1,
			},
		},
		IsHorizontal: horizontal,
		IsVertical:   vertical,
	}
}
