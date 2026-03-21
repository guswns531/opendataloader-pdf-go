package readingorder

import (
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

type testElement struct {
	model.BaseNode
	label string
}

func newTestElement(label string, box model.Box) model.ContentElement {
	return &testElement{
		BaseNode: model.BaseNode{
			Type:   model.ElementTypeParagraph,
			Bounds: box,
		},
		label: label,
	}
}

func labels(elements []model.ContentElement) []string {
	out := make([]string, len(elements))
	for i, element := range elements {
		out[i] = element.(*testElement).label
	}
	return out
}

func TestSortPrefersColumnSplit(t *testing.T) {
	input := []model.ContentElement{
		newTestElement("left-a", model.NewBox(20, 120, 180, 160)),
		newTestElement("right-a", model.NewBox(260, 130, 420, 170)),
		newTestElement("left-b", model.NewBox(20, 60, 180, 100)),
		newTestElement("right-b", model.NewBox(260, 50, 420, 90)),
	}

	got := Sort(input)
	want := []string{"left-a", "left-b", "right-a", "right-b"}

	if gotLabels := labels(got); !equalStrings(gotLabels, want) {
		t.Fatalf("Sort() = %v, want %v", gotLabels, want)
	}
}

func TestSortPreservesStableOrderForTies(t *testing.T) {
	box := model.NewBox(40, 80, 140, 140)
	input := []model.ContentElement{
		newTestElement("first", box),
		newTestElement("second", box),
		newTestElement("third", box),
	}

	got := Sort(input)
	want := []string{"first", "second", "third"}

	if gotLabels := labels(got); !equalStrings(gotLabels, want) {
		t.Fatalf("Sort() = %v, want %v", gotLabels, want)
	}
}

func TestSortPrefersBalancedVerticalSplitForInterleavedColumns(t *testing.T) {
	input := []model.ContentElement{
		newTestElement("right-a", model.NewBox(330, 740, 520, 760)),
		newTestElement("left-a", model.NewBox(40, 700, 230, 720)),
		newTestElement("right-b", model.NewBox(330, 580, 520, 600)),
		newTestElement("left-b", model.NewBox(40, 540, 230, 560)),
	}

	got := Sort(input)
	want := []string{"left-a", "left-b", "right-a", "right-b"}

	if gotLabels := labels(got); !equalStrings(gotLabels, want) {
		t.Fatalf("Sort() = %v, want %v", gotLabels, want)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
