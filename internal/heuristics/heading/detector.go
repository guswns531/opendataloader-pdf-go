package heading

import (
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/guswns531/opendataloader-pdf-go/internal/heuristics/readingorder"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// Detector applies lightweight heading heuristics to the current model graph.
//
// The implementation is intentionally pragmatic: it scores text-bearing layout
// nodes with a small set of style, position, and length signals, then groups the
// surviving candidates into heading levels by font/style similarity.
type Detector struct {
	MinScore        float64
	MaxWords        int
	MaxChars        int
	SizeBucketStep  float64
	TopShareBoost   float64
	WidthShareBoost float64
}

// Detection describes a heading candidate that passed the heuristic filter.
type Detection struct {
	Source  model.ContentElement
	Heading *model.Heading
	Score   float64
}

// NewDetector returns a detector with conservative defaults.
func NewDetector() Detector {
	return Detector{
		MinScore:        0.50,
		MaxWords:        24,
		MaxChars:        180,
		SizeBucketStep:  0.5,
		TopShareBoost:   0.10,
		WidthShareBoost: 0.05,
	}
}

// Detect scans the supplied content elements and returns ordered heading matches.
func Detect(elements []model.ContentElement) []Detection {
	return NewDetector().Detect(elements)
}

// Detect scans the supplied content elements and returns ordered heading matches.
func (d Detector) Detect(elements []model.ContentElement) []Detection {
	if d.MinScore <= 0 {
		d.MinScore = 0.50
	}
	if d.MaxWords <= 0 {
		d.MaxWords = 24
	}
	if d.MaxChars <= 0 {
		d.MaxChars = 180
	}
	if d.SizeBucketStep <= 0 {
		d.SizeBucketStep = 0.5
	}

	ordered := readingorder.Sort(elements)
	fontBaseline := dominantFontSize(ordered, d.SizeBucketStep)
	pageStats := collectPageStats(ordered)
	stats := collectStyleStats(ordered, d.SizeBucketStep)

	candidates := make([]candidate, 0, len(ordered))
	for i, element := range ordered {
		cand, ok := buildCandidate(element, i)
		if !ok {
			continue
		}
		cand.baselineFontSize = fontBaseline
		cand.pageStats = pageStats[cand.pageIndex]
		cand.styleStats = stats
		cand.score = scoreCandidate(cand, d)
		if cand.score < d.MinScore {
			continue
		}
		candidates = append(candidates, cand)
	}

	if len(candidates) == 0 {
		return nil
	}

	assignHeadingLevels(candidates, d.SizeBucketStep)

	out := make([]Detection, 0, len(candidates))
	for _, cand := range candidates {
		out = append(out, Detection{
			Source:  cand.source,
			Heading: cand.heading,
			Score:   cand.score,
		})
	}
	return out
}

type candidate struct {
	source           model.ContentElement
	heading          *model.Heading
	text             string
	bounds           model.Box
	pageIndex        model.PageIndex
	fontSize         float64
	bold             bool
	wordCount        int
	charCount        int
	numbered         bool
	allCaps          bool
	endsWithSentence bool
	baselineFontSize float64
	pageStats        *pageStats
	styleStats       styleStats
	score            float64
	existingLevel    int
	originalIndex    int
}

type headingStyleKey struct {
	sizeBucket float64
}

type pageStats struct {
	hasBounds bool
	minLeft   float64
	maxRight  float64
	minBottom float64
	maxTop    float64
}

type styleStats struct {
	total       int
	sizeBuckets map[float64]int
	boldCount   int
}

var numberedHeadingPattern = regexp.MustCompile(`^(?:\d+(?:\.\d+)*|[IVXLCDM]+|[A-Z])(?:[.)])?\s+`)

func buildCandidate(element model.ContentElement, index int) (candidate, bool) {
	if element == nil {
		return candidate{}, false
	}

	switch node := element.(type) {
	case *model.Paragraph:
		return candidateFromTextNode(node.ContentType(), &node.TextNode, node, index)
	case *model.Heading:
		cand, ok := candidateFromTextNode(node.ContentType(), &node.TextNode, node, index)
		if ok {
			cand.existingLevel = node.HeadingLevel
		}
		return cand, ok
	default:
		return candidate{}, false
	}
}

func candidateFromTextNode(_ model.ElementType, textNode *model.TextNode, source model.ContentElement, index int) (candidate, bool) {
	if textNode == nil {
		return candidate{}, false
	}

	text := strings.TrimSpace(textNode.TextProperties.Content)
	if text == "" {
		return candidate{}, false
	}

	wordCount := len(strings.Fields(text))
	if wordCount == 0 {
		return candidate{}, false
	}

	base := source.NodeBase()
	if base == nil {
		return candidate{}, false
	}

	fontSize := textNode.TextProperties.FontSize

	cand := candidate{
		source:           source,
		text:             text,
		bounds:           normalizeBox(base.Bounds),
		pageIndex:        base.PageIndex,
		fontSize:         fontSize,
		bold:             textNode.TextProperties.Bold,
		wordCount:        wordCount,
		charCount:        len(text),
		numbered:         numberedHeadingPattern.MatchString(text),
		allCaps:          looksLikeAllCaps(text),
		endsWithSentence: hasSentenceTerminal(text),
		originalIndex:    index,
	}

	heading := &model.Heading{
		TextNode: model.TextNode{
			BaseNode: copyBaseNode(base),
			TextProperties: model.TextProperties{
				Font:       textNode.TextProperties.Font,
				FontSize:   textNode.TextProperties.FontSize,
				TextColor:  textNode.TextProperties.TextColor,
				Content:    textNode.TextProperties.Content,
				HiddenText: textNode.TextProperties.HiddenText,
				Bold:       textNode.TextProperties.Bold,
				Italic:     textNode.TextProperties.Italic,
				Underline:  textNode.TextProperties.Underline,
			},
		},
		HeadingLevel: 0,
	}
	heading.Type = model.ElementTypeHeading
	cand.heading = heading

	return cand, true
}

func scoreCandidate(cand candidate, d Detector) float64 {
	if cand.charCount == 0 || cand.wordCount == 0 {
		return 0
	}
	if cand.charCount > d.MaxChars || cand.wordCount > d.MaxWords {
		return 0
	}
	if !cand.numbered && !cand.allCaps && cand.fontSize <= 0 && !cand.bold {
		return 0
	}
	if !cand.numbered && !cand.allCaps && cand.endsWithSentence && cand.wordCount > 3 {
		return 0
	}

	score := 0.0

	if cand.fontSize > 0 && cand.baselineFontSize > 0 {
		switch {
		case cand.fontSize >= cand.baselineFontSize+8:
			score += 0.40
		case cand.fontSize >= cand.baselineFontSize+4:
			score += 0.30
		case cand.fontSize >= cand.baselineFontSize+2:
			score += 0.22
		case cand.fontSize >= cand.baselineFontSize+1:
			score += 0.12
		case cand.fontSize < cand.baselineFontSize:
			score -= 0.05
		}
	}

	if cand.bold {
		score += 0.15
	}
	if cand.numbered {
		score += 0.18
	}
	if cand.allCaps {
		score += 0.10
	}

	if topShare := cand.topShare(); topShare >= 0.60 {
		score += d.TopShareBoost
	}
	if leftShare := cand.leftShare(); leftShare <= 0.15 {
		score += 0.05
	}
	if widthShare := cand.widthShare(); widthShare >= 0.35 {
		score += d.WidthShareBoost
	}
	if cand.rareSizeBucket() {
		score += 0.08
	}
	if cand.bold && cand.rareBoldStyle() {
		score += 0.04
	}

	switch {
	case cand.wordCount <= 4:
		score += 0.13
	case cand.wordCount <= 8:
		score += 0.07
	case cand.wordCount <= 12:
		score += 0.00
	default:
		score -= 0.14
	}

	if cand.endsWithSentence {
		score -= 0.12
	}

	return score
}

func collectStyleStats(elements []model.ContentElement, bucketStep float64) styleStats {
	stats := styleStats{
		sizeBuckets: make(map[float64]int),
	}
	for _, element := range elements {
		cand, ok := buildCandidate(element, 0)
		if !ok {
			continue
		}
		stats.total++
		stats.sizeBuckets[bucketFontSize(cand.fontSize, bucketStep)]++
		if cand.bold {
			stats.boldCount++
		}
	}
	return stats
}

func (cand candidate) rareSizeBucket() bool {
	if cand.styleStats.total == 0 {
		return false
	}
	bucket := bucketFontSize(cand.fontSize, 0.5)
	count := cand.styleStats.sizeBuckets[bucket]
	if count == 0 {
		return false
	}
	return float64(count)/float64(cand.styleStats.total) <= 0.25
}

func (cand candidate) rareBoldStyle() bool {
	if cand.styleStats.total == 0 || cand.styleStats.boldCount == 0 {
		return false
	}
	return float64(cand.styleStats.boldCount)/float64(cand.styleStats.total) <= 0.25
}

func assignHeadingLevels(candidates []candidate, sizeBucketStep float64) {
	keys := make([]headingStyleKey, 0, len(candidates))
	seen := make(map[headingStyleKey]struct{}, len(candidates))
	for _, cand := range candidates {
		if cand.existingLevel > 0 {
			continue
		}
		key := headingStyleKey{
			sizeBucket: bucketFontSize(cand.heading.FontSize, sizeBucketStep),
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		if keys[i].sizeBucket != keys[j].sizeBucket {
			return keys[i].sizeBucket > keys[j].sizeBucket
		}
		return false
	})

	levels := make(map[headingStyleKey]int, len(keys))
	for i, key := range keys {
		levels[key] = i + 1
	}

	for i := range candidates {
		if candidates[i].existingLevel > 0 {
			candidates[i].heading.HeadingLevel = candidates[i].existingLevel
			continue
		}
		key := headingStyleKey{
			sizeBucket: bucketFontSize(candidates[i].heading.FontSize, sizeBucketStep),
		}
		candidates[i].heading.HeadingLevel = levels[key]
	}
}

func dominantFontSize(elements []model.ContentElement, step float64) float64 {
	if step <= 0 {
		step = 0.5
	}

	counts := make(map[float64]int)
	bestSize := 0.0
	bestCount := 0

	for _, element := range elements {
		if _, ok := element.(*model.Heading); ok {
			continue
		}
		textNode := textNodeFor(element)
		if textNode == nil {
			continue
		}
		size := textNode.FontSize
		if size <= 0 {
			continue
		}
		bucket := bucketFontSize(size, step)
		counts[bucket]++
		if counts[bucket] > bestCount || (counts[bucket] == bestCount && (bestSize == 0 || bucket < bestSize)) {
			bestSize = bucket
			bestCount = counts[bucket]
		}
	}

	return bestSize
}

func collectPageStats(elements []model.ContentElement) map[model.PageIndex]*pageStats {
	statsByPage := make(map[model.PageIndex]*pageStats)
	for _, element := range elements {
		if textNodeFor(element) == nil {
			continue
		}
		base := elementBase(element)
		if base == nil {
			continue
		}
		box := normalizeBox(base.Bounds)
		if box.IsZero() {
			continue
		}
		stats, ok := statsByPage[base.PageIndex]
		if !ok {
			stats = &pageStats{
				hasBounds: true,
				minLeft:   box.Left,
				maxRight:  box.Right,
				minBottom: box.Bottom,
				maxTop:    box.Top,
			}
			statsByPage[base.PageIndex] = stats
			continue
		}
		if box.Left < stats.minLeft {
			stats.minLeft = box.Left
		}
		if box.Right > stats.maxRight {
			stats.maxRight = box.Right
		}
		if box.Bottom < stats.minBottom {
			stats.minBottom = box.Bottom
		}
		if box.Top > stats.maxTop {
			stats.maxTop = box.Top
		}
	}
	return statsByPage
}

func (c candidate) topShare() float64 {
	if c.pageStats == nil || !c.pageStats.hasBounds {
		return 0.5
	}
	span := c.pageStats.maxTop - c.pageStats.minBottom
	if span <= 0 {
		return 0.5
	}
	return (c.bounds.Top - c.pageStats.minBottom) / span
}

func (c candidate) widthShare() float64 {
	if c.pageStats == nil || !c.pageStats.hasBounds {
		return 0
	}
	span := c.pageStats.maxRight - c.pageStats.minLeft
	if span <= 0 {
		return 0
	}
	return c.bounds.Width() / span
}

func (c candidate) leftShare() float64 {
	if c.pageStats == nil || !c.pageStats.hasBounds {
		return 0.5
	}
	span := c.pageStats.maxRight - c.pageStats.minLeft
	if span <= 0 {
		return 0.5
	}
	return (c.bounds.Left - c.pageStats.minLeft) / span
}

func textNodeFor(element model.ContentElement) *model.TextProperties {
	switch node := element.(type) {
	case *model.Paragraph:
		return &node.TextProperties
	case *model.Heading:
		return &node.TextProperties
	default:
		return nil
	}
}

func elementBase(element model.ContentElement) *model.BaseNode {
	if element == nil {
		return nil
	}
	return element.NodeBase()
}

func copyBaseNode(base *model.BaseNode) model.BaseNode {
	if base == nil {
		return model.BaseNode{}
	}
	copy := *base
	if len(base.Boxes) > 0 {
		copy.Boxes = append(model.MultiBox(nil), base.Boxes...)
	}
	return copy
}

func bucketFontSize(size, step float64) float64 {
	if size <= 0 {
		return 0
	}
	if step <= 0 {
		step = 0.5
	}
	return float64(int((size/step)+0.5)) * step
}

func normalizeBox(box model.Box) model.Box {
	box = box.Normalize()
	if box.Left != box.Left || box.Right != box.Right || box.Top != box.Top || box.Bottom != box.Bottom {
		return model.Box{}
	}
	return box
}

func hasSentenceTerminal(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	last := text[len(text)-1]
	return last == '.' || last == '!' || last == '?'
}

func looksLikeAllCaps(text string) bool {
	hasLetter := false
	for _, r := range text {
		if unicode.IsLetter(r) {
			hasLetter = true
			if unicode.IsLower(r) {
				return false
			}
		}
	}
	return hasLetter
}
