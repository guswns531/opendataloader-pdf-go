package paragraph

import (
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/heuristics/text"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// IDAllocator provides document-local node IDs.
type IDAllocator interface {
	NewNodeID() model.NodeID
}

const (
	defaultMergeVerticalGap       = 12.0
	defaultMergeHorizontalOverlap = 0.25
)

// Assemble converts grouped text paragraphs into semantic paragraph nodes.
//
// The helper performs a conservative adjacent merge pass first so obvious
// paragraph continuations become a single model node before IDs are assigned.
func Assemble(allocator IDAllocator, paragraphs []text.Paragraph) []*model.Paragraph {
	merged := MergeAdjacent(paragraphs)
	out := make([]*model.Paragraph, 0, len(merged))
	for i, paragraph := range merged {
		out = append(out, ToNode(nodeID(allocator), i, paragraph))
	}
	return out
}

// MergeAdjacent merges consecutive grouped paragraphs when they look like the
// same paragraph split by layout noise.
func MergeAdjacent(paragraphs []text.Paragraph) []text.Paragraph {
	if len(paragraphs) == 0 {
		return []text.Paragraph{}
	}

	out := make([]text.Paragraph, 0, len(paragraphs))
	current := paragraphs[0]
	for _, next := range paragraphs[1:] {
		if CanMerge(current, next) {
			current = mergeParagraphs(current, next)
			continue
		}
		out = append(out, current)
		current = next
	}
	out = append(out, current)
	return out
}

// CanMerge reports whether two grouped paragraphs should be combined.
func CanMerge(prev, next text.Paragraph) bool {
	if prev.PageIndex != next.PageIndex || prev.PageNumber != next.PageNumber {
		return false
	}
	if prev.Bounds.IsZero() || next.Bounds.IsZero() {
		return false
	}

	prevBox := prev.Bounds.Normalize()
	nextBox := next.Bounds.Normalize()
	verticalGap := prevBox.Bottom - nextBox.Top
	if verticalGap < 0 {
		verticalGap = 0
	}
	if verticalGap > defaultMergeVerticalGap {
		return false
	}

	overlap := horizontalOverlapRatio(prevBox, nextBox)
	return overlap >= defaultMergeHorizontalOverlap
}

// ToNode converts a grouped paragraph into a model paragraph node.
func ToNode(id model.NodeID, index int, paragraph text.Paragraph) *model.Paragraph {
	bounds := paragraph.Bounds.Normalize()
	if paragraph.Bounds.IsZero() {
		bounds = model.Box{}
	}

	return &model.Paragraph{
		TextNode: model.TextNode{
			BaseNode: model.BaseNode{
				ID:         id,
				Type:       model.ElementTypeParagraph,
				Index:      index,
				PageIndex:  paragraph.PageIndex,
				PageNumber: paragraph.PageNumber,
				Bounds:     bounds,
			},
			TextProperties: model.TextProperties{
				Content: strings.TrimSpace(paragraph.Text),
			},
		},
	}
}

func nodeID(allocator IDAllocator) model.NodeID {
	if allocator == nil {
		return 0
	}
	return allocator.NewNodeID()
}

func mergeParagraphs(prev, next text.Paragraph) text.Paragraph {
	merged := prev
	merged.Lines = append(append([]text.Line(nil), prev.Lines...), next.Lines...)
	merged.Text = mergeText(prev.Text, next.Text)
	merged.Bounds = prev.Bounds.Union(next.Bounds)
	return merged
}

func mergeText(prev, next string) string {
	prev = strings.TrimSpace(prev)
	next = strings.TrimSpace(next)
	if prev == "" {
		return next
	}
	if next == "" {
		return prev
	}

	if strings.HasSuffix(prev, "-") && len(next) > 0 {
		return strings.TrimSuffix(prev, "-") + next
	}
	return prev + " " + next
}

func horizontalOverlapRatio(a, b model.Box) float64 {
	a = a.Normalize()
	b = b.Normalize()

	left := maxFloat(a.Left, b.Left)
	right := minFloat(a.Right, b.Right)
	overlap := right - left
	if overlap <= 0 {
		return 0
	}

	width := minFloat(a.Width(), b.Width())
	if width <= 0 {
		return 0
	}
	return overlap / width
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
