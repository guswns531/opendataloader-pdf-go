package text

import (
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestGroupArtifactsToLines(t *testing.T) {
	artifacts := []*model.RawArtifact{
		{
			ID:         1,
			Kind:       model.ArtifactKindText,
			PageIndex:  0,
			PageNumber: 1,
			Bounds:     model.NewBox(0, 90, 10, 100),
			Text:       "Hello",
			Sequence:   1,
		},
		{
			ID:         2,
			Kind:       model.ArtifactKindText,
			PageIndex:  0,
			PageNumber: 1,
			Bounds:     model.NewBox(12, 90, 22, 100),
			Style:      model.TextProperties{Content: "world"},
			Sequence:   2,
		},
		{
			ID:         3,
			Kind:       model.ArtifactKindText,
			PageIndex:  0,
			PageNumber: 1,
			Bounds:     model.NewBox(0, 78, 10, 88),
			Text:       "Second",
			Sequence:   3,
		},
		{
			ID:         4,
			Kind:       model.ArtifactKindText,
			PageIndex:  0,
			PageNumber: 1,
			Bounds:     model.NewBox(12, 78, 22, 88),
			Text:       "line",
			Sequence:   4,
		},
		{
			ID:         5,
			Kind:       model.ArtifactKindText,
			PageIndex:  1,
			PageNumber: 2,
			Bounds:     model.NewBox(0, 90, 10, 100),
			Text:       "Page",
			Sequence:   1,
		},
		{
			ID:         6,
			Kind:       model.ArtifactKindText,
			PageIndex:  1,
			PageNumber: 2,
			Bounds:     model.NewBox(12, 90, 18, 100),
			Text:       "two",
			Sequence:   2,
		},
	}

	lines := GroupArtifactsToLines(artifacts)
	if got, want := len(lines), 3; got != want {
		t.Fatalf("lines = %d, want %d", got, want)
	}
	if got, want := lines[0].Text, "Hello world"; got != want {
		t.Fatalf("first line text = %q, want %q", got, want)
	}
	if got, want := lines[1].Text, "Second line"; got != want {
		t.Fatalf("second line text = %q, want %q", got, want)
	}
	if got, want := lines[2].Text, "Page two"; got != want {
		t.Fatalf("third line text = %q, want %q", got, want)
	}

	paragraphs := GroupLinesToParagraphs(lines)
	if got, want := len(paragraphs), 2; got != want {
		t.Fatalf("paragraphs = %d, want %d", got, want)
	}
	if got, want := len(paragraphs[0].Lines), 2; got != want {
		t.Fatalf("first paragraph lines = %d, want %d", got, want)
	}
	if got, want := paragraphs[0].Text, "Hello world Second line"; got != want {
		t.Fatalf("first paragraph text = %q, want %q", got, want)
	}
	if got, want := paragraphs[1].Text, "Page two"; got != want {
		t.Fatalf("second paragraph text = %q, want %q", got, want)
	}
}

func TestGroupLinesToParagraphsRespectsBlankSeparators(t *testing.T) {
	lines := []Line{
		{
			Artifacts: []*model.RawArtifact{{ID: 1, Kind: model.ArtifactKindText}},
			Text:      "First line",
			Bounds:    model.NewBox(0, 90, 20, 100),
		},
		{
			Text:   "   ",
			Bounds: model.NewBox(0, 80, 0, 80),
		},
		{
			Artifacts: []*model.RawArtifact{{ID: 2, Kind: model.ArtifactKindText}},
			Text:      "Second line",
			Bounds:    model.NewBox(0, 60, 20, 70),
		},
	}

	paragraphs := GroupLinesToParagraphs(lines)
	if got, want := len(paragraphs), 2; got != want {
		t.Fatalf("paragraphs = %d, want %d", got, want)
	}
}
