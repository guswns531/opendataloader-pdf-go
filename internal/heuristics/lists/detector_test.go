package lists

import (
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestDetectBuildsNestedListsFromNumberedAndIndentedItems(t *testing.T) {
	input := []model.ContentElement{
		paragraph("Intro", 12),
		paragraph("1. Parent", 24),
		paragraph("2. Branch", 24),
		paragraph("a) Child one", 58),
		paragraph("b) Child two", 58),
		paragraph("3. End", 24),
		paragraph("Conclusion", 12),
	}

	got := Detect(input)
	if gotLen := len(got); gotLen != 3 {
		t.Fatalf("Detect() length = %d, want 3", gotLen)
	}

	if _, ok := got[0].(*model.Paragraph); !ok {
		t.Fatalf("first element type = %T, want *model.Paragraph", got[0])
	}
	list, ok := got[1].(*model.List)
	if !ok {
		t.Fatalf("second element type = %T, want *model.List", got[1])
	}
	if list.NumberingStyle != "arabic" {
		t.Fatalf("list numbering style = %q, want %q", list.NumberingStyle, "arabic")
	}
	if list.NumberOfItems != 3 {
		t.Fatalf("list number of items = %d, want 3", list.NumberOfItems)
	}
	if got := list.ListItems[0].Content; got != "Parent" {
		t.Fatalf("first item content = %q, want %q", got, "Parent")
	}
	if got := list.ListItems[1].Content; got != "Branch" {
		t.Fatalf("second item content = %q, want %q", got, "Branch")
	}

	if len(list.ListItems[1].Kids) != 1 {
		t.Fatalf("second item kids length = %d, want 1", len(list.ListItems[1].Kids))
	}
	nested, ok := list.ListItems[1].Kids[0].(*model.List)
	if !ok {
		t.Fatalf("nested kid type = %T, want *model.List", list.ListItems[1].Kids[0])
	}
	if nested.NumberingStyle != "lower-alpha" {
		t.Fatalf("nested numbering style = %q, want %q", nested.NumberingStyle, "lower-alpha")
	}
	if nested.NumberOfItems != 2 {
		t.Fatalf("nested item count = %d, want 2", nested.NumberOfItems)
	}
	if got := nested.ListItems[0].Content; got != "Child one" {
		t.Fatalf("nested first item content = %q, want %q", got, "Child one")
	}
	if got := nested.ListItems[1].Content; got != "Child two" {
		t.Fatalf("nested second item content = %q, want %q", got, "Child two")
	}

	if _, ok := got[2].(*model.Paragraph); !ok {
		t.Fatalf("third element type = %T, want *model.Paragraph", got[2])
	}
}

func TestDetectSeparatesDifferentListStyles(t *testing.T) {
	input := []model.ContentElement{
		paragraph("• One", 20),
		paragraph("• Two", 20),
		paragraph("1. First", 20),
		paragraph("2. Second", 20),
	}

	got := Detect(input)
	if gotLen := len(got); gotLen != 2 {
		t.Fatalf("Detect() length = %d, want 2", gotLen)
	}

	first, ok := got[0].(*model.List)
	if !ok {
		t.Fatalf("first element type = %T, want *model.List", got[0])
	}
	if first.NumberingStyle != "unordered" {
		t.Fatalf("first list numbering style = %q, want %q", first.NumberingStyle, "unordered")
	}
	if first.NumberOfItems != 2 {
		t.Fatalf("first list item count = %d, want 2", first.NumberOfItems)
	}

	second, ok := got[1].(*model.List)
	if !ok {
		t.Fatalf("second element type = %T, want *model.List", got[1])
	}
	if second.NumberingStyle != "arabic" {
		t.Fatalf("second list numbering style = %q, want %q", second.NumberingStyle, "arabic")
	}
	if second.NumberOfItems != 2 {
		t.Fatalf("second list item count = %d, want 2", second.NumberOfItems)
	}
}

func TestDetectLeavesSentenceLikePrefixesAsParagraphs(t *testing.T) {
	input := []model.ContentElement{
		paragraph("1.5 million people", 24),
		paragraph("Still prose", 24),
	}

	got := Detect(input)
	if gotLen := len(got); gotLen != 2 {
		t.Fatalf("Detect() length = %d, want 2", gotLen)
	}
	first, ok := got[0].(*model.Paragraph)
	if !ok {
		t.Fatalf("first element type = %T, want *model.Paragraph", got[0])
	}
	if first.Content != "1.5 million people" {
		t.Fatalf("first paragraph content = %q, want %q", first.Content, "1.5 million people")
	}
}

func TestDetectAttachesIndentedContinuationToPreviousListItem(t *testing.T) {
	input := []model.ContentElement{
		paragraph("1. Parent item", 24),
		paragraph("Continuation paragraph", 42),
		paragraph("2. Next item", 24),
	}

	got := Detect(input)
	if gotLen := len(got); gotLen != 1 {
		t.Fatalf("Detect() length = %d, want 1", gotLen)
	}
	list, ok := got[0].(*model.List)
	if !ok {
		t.Fatalf("first element type = %T, want *model.List", got[0])
	}
	if list.NumberOfItems != 2 {
		t.Fatalf("list number of items = %d, want 2", list.NumberOfItems)
	}
	if len(list.ListItems[0].Kids) != 1 {
		t.Fatalf("first list item kids = %d, want 1", len(list.ListItems[0].Kids))
	}
	child, ok := list.ListItems[0].Kids[0].(*model.Paragraph)
	if !ok {
		t.Fatalf("continuation child type = %T, want *model.Paragraph", list.ListItems[0].Kids[0])
	}
	if child.Content != "Continuation paragraph" {
		t.Fatalf("continuation child content = %q, want %q", child.Content, "Continuation paragraph")
	}
}

func paragraph(text string, left float64) *model.Paragraph {
	return &model.Paragraph{
		TextNode: model.TextNode{
			BaseNode: model.BaseNode{
				Type:       model.ElementTypeParagraph,
				PageIndex:  0,
				PageNumber: 1,
				Bounds:     model.NewBox(left, 0, left+120, 20),
			},
			TextProperties: model.TextProperties{
				Content: text,
			},
		},
	}
}
