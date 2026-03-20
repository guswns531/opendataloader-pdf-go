package paragraph

import (
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/heuristics/text"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestAssembleConvertsGroupedParagraphsIntoModelNodes(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{})
	input := []text.Paragraph{
		{
			Text:       "First paragraph",
			Bounds:     model.NewBox(10, 80, 100, 100),
			PageIndex:  0,
			PageNumber: 1,
		},
		{
			Text:       "Second paragraph",
			Bounds:     model.NewBox(12, 50, 102, 70),
			PageIndex:  1,
			PageNumber: 2,
		},
	}

	got := Assemble(doc, input)
	if got == nil {
		t.Fatalf("Assemble() returned nil")
	}
	if gotLen, want := len(got), 2; gotLen != want {
		t.Fatalf("len(Assemble()) = %d, want %d", gotLen, want)
	}

	first := got[0]
	if first.ID != 1 {
		t.Fatalf("first ID = %d, want 1", first.ID)
	}
	if first.Type != model.ElementTypeParagraph {
		t.Fatalf("first type = %q, want %q", first.Type, model.ElementTypeParagraph)
	}
	if first.Index != 0 {
		t.Fatalf("first index = %d, want 0", first.Index)
	}
	if first.PageIndex != 0 || first.PageNumber != 1 {
		t.Fatalf("first page metadata = (%d, %d), want (0, 1)", first.PageIndex, first.PageNumber)
	}
	if first.Bounds != model.NewBox(10, 80, 100, 100) {
		t.Fatalf("first bounds = %#v, want %#v", first.Bounds, model.NewBox(10, 80, 100, 100))
	}
	if first.Content != "First paragraph" {
		t.Fatalf("first content = %q, want %q", first.Content, "First paragraph")
	}

	second := got[1]
	if second.ID != 2 {
		t.Fatalf("second ID = %d, want 2", second.ID)
	}
	if second.Index != 1 {
		t.Fatalf("second index = %d, want 1", second.Index)
	}
	if second.PageIndex != 1 || second.PageNumber != 2 {
		t.Fatalf("second page metadata = (%d, %d), want (1, 2)", second.PageIndex, second.PageNumber)
	}
	if second.Content != "Second paragraph" {
		t.Fatalf("second content = %q, want %q", second.Content, "Second paragraph")
	}
}

func TestCanMergeUsesSimpleGeometryAndPageChecks(t *testing.T) {
	prev := text.Paragraph{
		Text:       "First line",
		Bounds:     model.NewBox(10, 80, 100, 100),
		PageIndex:  0,
		PageNumber: 1,
	}
	closeNext := text.Paragraph{
		Text:       "continuation",
		Bounds:     model.NewBox(12, 66, 98, 76),
		PageIndex:  0,
		PageNumber: 1,
	}
	farNext := text.Paragraph{
		Text:       "new paragraph",
		Bounds:     model.NewBox(12, 40, 98, 50),
		PageIndex:  0,
		PageNumber: 1,
	}
	otherPage := text.Paragraph{
		Text:       "other page",
		Bounds:     model.NewBox(12, 66, 98, 76),
		PageIndex:  1,
		PageNumber: 2,
	}

	if !CanMerge(prev, closeNext) {
		t.Fatalf("CanMerge() = false, want true for close paragraphs on same page")
	}
	if CanMerge(prev, farNext) {
		t.Fatalf("CanMerge() = true, want false for far paragraphs")
	}
	if CanMerge(prev, otherPage) {
		t.Fatalf("CanMerge() = true, want false for different pages")
	}
}

func TestMergeAdjacentAndAssembleCollapseContinuations(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{})
	input := []text.Paragraph{
		{
			Text:       "For ease of reference",
			Bounds:     model.NewBox(0, 90, 100, 100),
			PageIndex:  0,
			PageNumber: 1,
		},
		{
			Text:       "and exposition",
			Bounds:     model.NewBox(0, 78, 100, 88),
			PageIndex:  0,
			PageNumber: 1,
		},
		{
			Text:       "next page text",
			Bounds:     model.NewBox(0, 90, 100, 100),
			PageIndex:  1,
			PageNumber: 2,
		},
	}

	merged := MergeAdjacent(input)
	if got, want := len(merged), 2; got != want {
		t.Fatalf("len(MergeAdjacent()) = %d, want %d", got, want)
	}
	if got, want := merged[0].Text, "For ease of reference and exposition"; got != want {
		t.Fatalf("merged text = %q, want %q", got, want)
	}
	if got, want := merged[0].Bounds, model.NewBox(0, 78, 100, 100); got != want {
		t.Fatalf("merged bounds = %#v, want %#v", got, want)
	}

	nodes := Assemble(doc, input)
	if got, want := len(nodes), 2; got != want {
		t.Fatalf("len(Assemble()) = %d, want %d", got, want)
	}
	if got, want := nodes[0].Content, "For ease of reference and exposition"; got != want {
		t.Fatalf("assembled content = %q, want %q", got, want)
	}
	if got, want := nodes[0].Bounds, model.NewBox(0, 78, 100, 100); got != want {
		t.Fatalf("assembled bounds = %#v, want %#v", got, want)
	}
	if got, want := nodes[1].PageNumber, model.PageNumber(2); got != want {
		t.Fatalf("second page number = %d, want %d", got, want)
	}
}
