package readingorder

import (
	"math"
	"sort"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// XYCut sorts content elements into a practical reading order using bounding boxes.
//
// The implementation is intentionally small: it prefers recursive whitespace cuts
// when a clear horizontal or vertical separation exists, and otherwise falls back
// to a stable top-to-bottom, left-to-right order.
type XYCut struct{}

// Sort returns a new slice ordered in reading order.
func Sort(elements []model.ContentElement) []model.ContentElement {
	return XYCut{}.Sort(elements)
}

// Sort returns a new slice ordered in reading order.
func (XYCut) Sort(elements []model.ContentElement) []model.ContentElement {
	items := make([]sortableElement, 0, len(elements))
	for i, element := range elements {
		box, ok := elementBox(element)
		items = append(items, sortableElement{
			element: element,
			box:     box,
			valid:   ok,
			index:   i,
		})
	}

	ordered := sortItems(items)
	out := make([]model.ContentElement, len(ordered))
	for i, item := range ordered {
		out[i] = item.element
	}
	return out
}

type sortableElement struct {
	element model.ContentElement
	box     model.Box
	valid   bool
	index   int
}

type splitAxis int

const (
	axisNone splitAxis = iota
	axisHorizontal
	axisVertical
)

type splitCandidate struct {
	axis     splitAxis
	cut      float64
	gap      float64
	span     float64
	validity float64
	balance  float64
}

func sortItems(items []sortableElement) []sortableElement {
	if len(items) <= 1 {
		return append([]sortableElement(nil), items...)
	}

	if candidate, ok := bestSplit(items); ok {
		left, right, splitOK := partition(items, candidate)
		if splitOK && len(left) > 0 && len(right) > 0 {
			left = sortItems(left)
			right = sortItems(right)
			return append(left, right...)
		}
	}

	out := append([]sortableElement(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		return readingOrderLess(out[i], out[j])
	})
	return out
}

func bestSplit(items []sortableElement) (splitCandidate, bool) {
	if len(items) < 2 {
		return splitCandidate{}, false
	}

	xGap, xCut, xSpan, xOK := largestGap(items, true)
	yGap, yCut, ySpan, yOK := largestGap(items, false)

	if !xOK && !yOK {
		return splitCandidate{}, false
	}

	candidates := make([]splitCandidate, 0, 2)
	var vertical *splitCandidate
	if xOK {
		score := xGap / xSpan
		if score >= splitRatio {
			if candidate, ok := scoredCandidate(items, splitCandidate{
				axis: axisVertical, cut: xCut, gap: xGap, span: xSpan, validity: score,
			}); ok {
				candidates = append(candidates, candidate)
				vertical = &candidates[len(candidates)-1]
			}
		}
	}
	if yOK {
		score := yGap / ySpan
		if score >= splitRatio {
			if candidate, ok := scoredCandidate(items, splitCandidate{
				axis: axisHorizontal, cut: yCut, gap: yGap, span: ySpan, validity: score,
			}); ok {
				candidates = append(candidates, candidate)
			}
		}
	}

	if len(candidates) == 0 {
		return splitCandidate{}, false
	}
	if vertical != nil && vertical.balance >= 0.25 {
		return *vertical, true
	}

	best := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidateScore(candidate) > candidateScore(best) {
			best = candidate
		}
	}
	return best, true
}

const splitRatio = 0.12
const verticalBalanceBonus = 0.08

func scoredCandidate(items []sortableElement, candidate splitCandidate) (splitCandidate, bool) {
	left, right, ok := partition(items, candidate)
	if !ok || len(left) == 0 || len(right) == 0 {
		return splitCandidate{}, false
	}
	candidate.balance = partitionBalance(len(left), len(right))
	return candidate, true
}

func candidateScore(candidate splitCandidate) float64 {
	score := candidate.validity + candidate.balance*0.05
	if candidate.axis == axisVertical && candidate.balance >= 0.25 {
		score += verticalBalanceBonus
	}
	return score
}

func partitionBalance(left, right int) float64 {
	total := left + right
	if total == 0 {
		return 0
	}
	smaller := math.Min(float64(left), float64(right))
	return smaller / float64(total)
}

func largestGap(items []sortableElement, horizontal bool) (gap, cut, span float64, ok bool) {
	intervals := make([]interval, 0, len(items))
	for _, item := range items {
		if !item.valid {
			return 0, 0, 0, false
		}
		box := item.box.Normalize()
		if horizontal {
			intervals = append(intervals, interval{start: box.Left, end: box.Right})
		} else {
			intervals = append(intervals, interval{start: box.Bottom, end: box.Top})
		}
	}

	if len(intervals) < 2 {
		return 0, 0, 0, false
	}

	sort.Slice(intervals, func(i, j int) bool {
		if intervals[i].start == intervals[j].start {
			return intervals[i].end < intervals[j].end
		}
		return intervals[i].start < intervals[j].start
	})

	minStart := intervals[0].start
	maxEnd := intervals[0].end
	bestGap := 0.0
	bestCut := 0.0

	for _, iv := range intervals[1:] {
		if iv.start > maxEnd {
			currentGap := iv.start - maxEnd
			if currentGap > bestGap {
				bestGap = currentGap
				bestCut = (maxEnd + iv.start) / 2
			}
		}
		if iv.end > maxEnd {
			maxEnd = iv.end
		}
		if iv.start < minStart {
			minStart = iv.start
		}
	}

	span = maxEnd - minStart
	if span <= 0 || bestGap <= 0 {
		return 0, 0, 0, false
	}
	return bestGap, bestCut, span, true
}

func partition(items []sortableElement, candidate splitCandidate) (left, right []sortableElement, ok bool) {
	left = make([]sortableElement, 0, len(items))
	right = make([]sortableElement, 0, len(items))

	for _, item := range items {
		box := item.box.Normalize()
		switch candidate.axis {
		case axisVertical:
			if box.Right <= candidate.cut {
				left = append(left, item)
				continue
			}
			if box.Left >= candidate.cut {
				right = append(right, item)
				continue
			}
			return nil, nil, false
		case axisHorizontal:
			if box.Top <= candidate.cut {
				right = append(right, item)
				continue
			}
			if box.Bottom >= candidate.cut {
				left = append(left, item)
				continue
			}
			return nil, nil, false
		default:
			return nil, nil, false
		}
	}

	return left, right, true
}

func readingOrderLess(a, b sortableElement) bool {
	if a.valid != b.valid {
		return a.valid
	}
	if a.valid && b.valid {
		aBox := a.box.Normalize()
		bBox := b.box.Normalize()

		if aBox.Top != bBox.Top {
			return aBox.Top > bBox.Top
		}
		if aBox.Left != bBox.Left {
			return aBox.Left < bBox.Left
		}
		if aBox.Bottom != bBox.Bottom {
			return aBox.Bottom > bBox.Bottom
		}
		if aBox.Right != bBox.Right {
			return aBox.Right < bBox.Right
		}
	}
	return a.index < b.index
}

type interval struct {
	start float64
	end   float64
}

func elementBox(element model.ContentElement) (model.Box, bool) {
	if element == nil {
		return model.Box{}, false
	}
	base := element.NodeBase()
	if base == nil {
		return model.Box{}, false
	}

	if box := normalizeBox(base.Bounds); !box.IsZero() {
		return box, true
	}

	if box, ok := base.Boxes.Bounds(); ok {
		box = normalizeBox(box)
		if !box.IsZero() {
			return box, true
		}
	}

	return model.Box{}, false
}

func normalizeBox(box model.Box) model.Box {
	box = box.Normalize()
	if math.IsNaN(box.Left) || math.IsNaN(box.Right) || math.IsNaN(box.Top) || math.IsNaN(box.Bottom) {
		return model.Box{}
	}
	return box
}
