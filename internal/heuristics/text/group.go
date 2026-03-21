package text

import (
	"math"
	"sort"
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// Line groups adjacent raw text artifacts that appear on the same visual line.
type Line struct {
	Artifacts  []*model.RawArtifact
	Text       string
	Bounds     model.Box
	PageIndex  model.PageIndex
	PageNumber model.PageNumber
}

// Paragraph groups adjacent lines that form a single text paragraph.
type Paragraph struct {
	Lines      []Line
	Text       string
	Bounds     model.Box
	PageIndex  model.PageIndex
	PageNumber model.PageNumber
}

type artifactItem struct {
	artifact *model.RawArtifact
	text     string
	bounds   model.Box
}

type lineItem struct {
	line   Line
	bounds model.Box
}

// GroupArtifactsToLines groups raw text artifacts into reading-order lines.
func GroupArtifactsToLines(artifacts []*model.RawArtifact) []Line {
	items := make([]artifactItem, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifact == nil {
			continue
		}
		text := artifactText(artifact)
		if strings.TrimSpace(text) == "" {
			continue
		}
		items = append(items, artifactItem{
			artifact: artifact,
			text:     text,
			bounds:   artifactBounds(artifact),
		})
	}
	if len(items) == 0 {
		return []Line{}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return artifactBefore(items[i], items[j])
	})

	lines := make([]Line, 0, len(items))
	var current *Line
	var last artifactItem

	flush := func() {
		if current == nil {
			return
		}
		lines = append(lines, *current)
		current = nil
	}

	for _, item := range items {
		if current == nil {
			line := newLineFromArtifact(item.artifact, item.text, item.bounds)
			current = &line
			last = item
			continue
		}
		if sameLine(last, item, current.Bounds) {
			current.Artifacts = append(current.Artifacts, item.artifact)
			current.Text = joinText(current.Text, item.text)
			current.Bounds = current.Bounds.Union(item.bounds)
			last = item
			continue
		}
		flush()
		line := newLineFromArtifact(item.artifact, item.text, item.bounds)
		current = &line
		last = item
	}
	flush()
	return lines
}

// GroupLinesToParagraphs groups lines into simple paragraphs.
func GroupLinesToParagraphs(lines []Line) []Paragraph {
	items := make([]lineItem, 0, len(lines))
	for _, line := range lines {
		if len(line.Artifacts) == 0 && strings.TrimSpace(line.Text) == "" {
			continue
		}
		items = append(items, lineItem{
			line:   line,
			bounds: line.Bounds,
		})
	}
	if len(items) == 0 {
		return []Paragraph{}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return lineBefore(items[i], items[j])
	})

	paragraphs := make([]Paragraph, 0, len(items))
	var current *Paragraph
	var last lineItem

	flush := func() {
		if current == nil {
			return
		}
		paragraphs = append(paragraphs, *current)
		current = nil
	}

	for _, item := range items {
		if isBlankLine(item.line) {
			flush()
			continue
		}
		if current == nil {
			paragraph := newParagraphFromLine(item.line, item.bounds)
			current = &paragraph
			last = item
			continue
		}
		if sameParagraph(last, item) {
			current.Lines = append(current.Lines, item.line)
			current.Text = joinText(current.Text, item.line.Text)
			current.Bounds = current.Bounds.Union(item.bounds)
			last = item
			continue
		}
		flush()
		paragraph := newParagraphFromLine(item.line, item.bounds)
		current = &paragraph
		last = item
	}
	flush()
	return paragraphs
}

// GroupArtifactsToParagraphs is a convenience wrapper over the two basic grouping steps.
func GroupArtifactsToParagraphs(artifacts []*model.RawArtifact) []Paragraph {
	return GroupLinesToParagraphs(GroupArtifactsToLines(artifacts))
}

func newLineFromArtifact(artifact *model.RawArtifact, text string, bounds model.Box) Line {
	line := Line{
		Artifacts:  []*model.RawArtifact{artifact},
		Text:       text,
		Bounds:     bounds,
		PageIndex:  artifact.PageIndex,
		PageNumber: artifact.PageNumber,
	}
	return line
}

func newParagraphFromLine(line Line, bounds model.Box) Paragraph {
	return Paragraph{
		Lines:      []Line{line},
		Text:       line.Text,
		Bounds:     bounds,
		PageIndex:  line.PageIndex,
		PageNumber: line.PageNumber,
	}
}

func artifactText(artifact *model.RawArtifact) string {
	if artifact == nil {
		return ""
	}
	if text := strings.TrimSpace(artifact.Text); text != "" {
		return text
	}
	return strings.TrimSpace(artifact.Style.Content)
}

func artifactBounds(artifact *model.RawArtifact) model.Box {
	if artifact == nil {
		return model.Box{}
	}
	if !artifact.Bounds.IsZero() {
		return artifact.Bounds.Normalize()
	}
	if bounds, ok := artifact.Boxes.Bounds(); ok {
		return bounds.Normalize()
	}
	return model.Box{}
}

func artifactBefore(a, b artifactItem) bool {
	if a.artifact.PageIndex != b.artifact.PageIndex {
		return a.artifact.PageIndex < b.artifact.PageIndex
	}
	if a.artifact.PageNumber != b.artifact.PageNumber {
		return a.artifact.PageNumber < b.artifact.PageNumber
	}
	if a.bounds.IsZero() != b.bounds.IsZero() {
		return !a.bounds.IsZero()
	}
	if !a.bounds.IsZero() && !b.bounds.IsZero() {
		if !floatEqual(a.bounds.Top, b.bounds.Top) {
			return a.bounds.Top > b.bounds.Top
		}
		if !floatEqual(a.bounds.Left, b.bounds.Left) {
			return a.bounds.Left < b.bounds.Left
		}
		if !floatEqual(a.bounds.Bottom, b.bounds.Bottom) {
			return a.bounds.Bottom > b.bounds.Bottom
		}
		if !floatEqual(a.bounds.Right, b.bounds.Right) {
			return a.bounds.Right < b.bounds.Right
		}
	}
	if a.artifact.Sequence != b.artifact.Sequence {
		return a.artifact.Sequence < b.artifact.Sequence
	}
	if a.artifact.ID != b.artifact.ID {
		return a.artifact.ID < b.artifact.ID
	}
	return a.text < b.text
}

func lineBefore(a, b lineItem) bool {
	if a.line.PageIndex != b.line.PageIndex {
		return a.line.PageIndex < b.line.PageIndex
	}
	if a.line.PageNumber != b.line.PageNumber {
		return a.line.PageNumber < b.line.PageNumber
	}
	if a.bounds.IsZero() != b.bounds.IsZero() {
		return !a.bounds.IsZero()
	}
	if !a.bounds.IsZero() && !b.bounds.IsZero() {
		if !floatEqual(a.bounds.Top, b.bounds.Top) {
			return a.bounds.Top > b.bounds.Top
		}
		if !floatEqual(a.bounds.Left, b.bounds.Left) {
			return a.bounds.Left < b.bounds.Left
		}
	}
	if len(a.line.Artifacts) != len(b.line.Artifacts) {
		return len(a.line.Artifacts) > len(b.line.Artifacts)
	}
	return a.line.Text < b.line.Text
}

func sameLine(prev, curr artifactItem, lineBounds model.Box) bool {
	if prev.artifact.PageIndex != curr.artifact.PageIndex || prev.artifact.PageNumber != curr.artifact.PageNumber {
		return false
	}
	if prev.bounds.IsZero() || curr.bounds.IsZero() {
		return curr.artifact.Sequence == prev.artifact.Sequence+1
	}

	verticalOverlap := boxVerticalOverlap(prev.bounds, curr.bounds)
	if verticalOverlap < 0.5 {
		return false
	}

	horizontalGap := boxHorizontalGap(lineBounds, curr.bounds)
	return horizontalGap <= lineGapTolerance(prev, curr)*6
}

func sameParagraph(prev, curr lineItem) bool {
	if prev.line.PageIndex != curr.line.PageIndex || prev.line.PageNumber != curr.line.PageNumber {
		return false
	}
	if prev.bounds.IsZero() || curr.bounds.IsZero() {
		return true
	}
	if strings.TrimSpace(curr.line.Text) == "" {
		return false
	}
	verticalGap := boxVerticalGap(prev.bounds, curr.bounds)
	if verticalGap > paragraphGapTolerance(prev, curr) {
		return false
	}
	return horizontalOverlapRatio(prev.bounds, curr.bounds) >= paragraphOverlapTolerance
}

const paragraphOverlapTolerance = 0.25

func lineGapTolerance(a, b artifactItem) float64 {
	tolerance := 1.5
	if font := math.Max(a.artifact.Style.FontSize, b.artifact.Style.FontSize); font > 0 {
		tolerance = math.Max(tolerance, font*0.4)
	}
	if height := math.Min(nonZero(a.bounds.Height()), nonZero(b.bounds.Height())); height > 0 {
		tolerance = math.Max(tolerance, height*0.5)
	}
	return tolerance
}

func paragraphGapTolerance(a, b lineItem) float64 {
	tolerance := 3.0
	if height := math.Min(nonZero(a.bounds.Height()), nonZero(b.bounds.Height())); height > 0 {
		tolerance = math.Max(tolerance, height*1.5)
	}
	return tolerance
}

func boxVerticalGap(a, b model.Box) float64 {
	a = a.Normalize()
	b = b.Normalize()
	if a.Top < b.Bottom {
		return b.Bottom - a.Top
	}
	if b.Top < a.Bottom {
		return a.Bottom - b.Top
	}
	return 0
}

func boxVerticalOverlap(a, b model.Box) float64 {
	a = a.Normalize()
	b = b.Normalize()
	top := math.Min(a.Top, b.Top)
	bottom := math.Max(a.Bottom, b.Bottom)
	return top - bottom
}

func boxHorizontalGap(a, b model.Box) float64 {
	a = a.Normalize()
	b = b.Normalize()
	if a.Right < b.Left {
		return b.Left - a.Right
	}
	if b.Right < a.Left {
		return a.Left - b.Right
	}
	return 0
}

func horizontalOverlapRatio(a, b model.Box) float64 {
	a = a.Normalize()
	b = b.Normalize()
	left := math.Max(a.Left, b.Left)
	right := math.Min(a.Right, b.Right)
	overlap := right - left
	if overlap <= 0 {
		return 0
	}

	width := math.Min(nonZero(a.Width()), nonZero(b.Width()))
	if width <= 0 {
		return 0
	}
	return overlap / width
}

func joinText(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if text := strings.TrimSpace(part); text != "" {
			filtered = append(filtered, text)
		}
	}
	return strings.Join(filtered, " ")
}

func isBlankLine(line Line) bool {
	return strings.TrimSpace(line.Text) == ""
}

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) <= 0.01
}

func nonZero(value float64) float64 {
	if value <= 0 {
		return 0
	}
	return value
}
