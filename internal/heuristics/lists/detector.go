package lists

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// Detector groups paragraph-like content into list and list item nodes.
//
// The detector is intentionally small. It relies on explicit bullet/number
// prefixes, plus indentation changes when those hints are available, and it
// only groups consecutive items.
type Detector struct {
	IndentTolerance             float64
	NestedIndentThreshold       float64
	ContinuationIndentThreshold float64
}

const (
	defaultIndentTolerance             = 4.0
	defaultNestedIndentThreshold       = 18.0
	defaultContinuationIndentThreshold = 10.0
)

// Detect converts a slice of content elements into a new slice where detected
// list runs become model.List containers with model.ListItem children.
func Detect(elements []model.ContentElement) []model.ContentElement {
	return Detector{}.Detect(elements)
}

// Detect converts a slice of content elements into a new slice where detected
// list runs become model.List containers with model.ListItem children.
func (d Detector) Detect(elements []model.ContentElement) []model.ContentElement {
	d = d.withDefaults()

	out := make([]model.ContentElement, 0, len(elements))
	var frames []*listFrame

	for _, element := range elements {
		if element == nil {
			continue
		}

		switch node := element.(type) {
		case *model.List:
			frames = nil
			out = append(out, node)
			continue
		case *model.ListItem:
			frames = nil
			out = append(out, node)
			continue
		}

		candidate, ok := d.candidateFromElement(element)
		if ok {
			out, frames = d.placeCandidate(out, frames, candidate)
			continue
		}

		if d.attachContinuation(frames, element) {
			continue
		}

		frames = nil
		out = append(out, element)
	}

	return out
}

type markerKind int

const (
	markerUnknown markerKind = iota
	markerUnordered
	markerOrdered
)

type listCandidate struct {
	element        model.ContentElement
	base           model.BaseNode
	props          model.TextProperties
	text           string
	indent         float64
	numberingStyle string
	kind           markerKind
}

type listFrame struct {
	list       *model.List
	indent     float64
	numbering  string
	kind       markerKind
	parentItem *model.ListItem
	lastItem   *model.ListItem
	pageIndex  model.PageIndex
	pageNumber model.PageNumber
}

func (d Detector) withDefaults() Detector {
	if d.IndentTolerance <= 0 {
		d.IndentTolerance = defaultIndentTolerance
	}
	if d.NestedIndentThreshold <= 0 {
		d.NestedIndentThreshold = defaultNestedIndentThreshold
	}
	if d.ContinuationIndentThreshold <= 0 {
		d.ContinuationIndentThreshold = defaultContinuationIndentThreshold
	}
	return d
}

func (d Detector) placeCandidate(out []model.ContentElement, frames []*listFrame, candidate listCandidate) ([]model.ContentElement, []*listFrame) {
	for {
		if len(frames) == 0 {
			frames = append(frames, d.openList(&out, nil, candidate))
			return out, frames
		}

		top := frames[len(frames)-1]
		if !samePage(top, candidate) {
			frames = frames[:0]
			frames = append(frames, d.openList(&out, nil, candidate))
			return out, frames
		}

		if candidate.indent > top.indent+d.NestedIndentThreshold && top.lastItem != nil {
			frames = append(frames, d.openList(&out, top.lastItem, candidate))
			return out, frames
		}

		if candidate.indent < top.indent-d.IndentTolerance {
			frames = frames[:len(frames)-1]
			continue
		}

		if top.matches(candidate, d.IndentTolerance) {
			top.append(candidate)
			return out, frames
		}

		if len(frames) > 1 {
			frames = frames[:len(frames)-1]
			continue
		}

		frames = frames[:0]
		frames = append(frames, d.openList(&out, nil, candidate))
		return out, frames
	}
}

func (d Detector) openList(out *[]model.ContentElement, parentItem *model.ListItem, candidate listCandidate) *listFrame {
	frame := &listFrame{
		list: &model.List{
			BaseNode: model.BaseNode{
				Type:       model.ElementTypeList,
				PageIndex:  candidate.base.PageIndex,
				PageNumber: candidate.base.PageNumber,
				Bounds:     candidate.base.Bounds,
			},
			NumberingStyle: candidate.numberingStyle,
			NumberOfItems:  0,
			ListItems:      make([]*model.ListItem, 0, 4),
		},
		indent:     candidate.indent,
		numbering:  candidate.numberingStyle,
		kind:       candidate.kind,
		parentItem: parentItem,
		pageIndex:  candidate.base.PageIndex,
		pageNumber: candidate.base.PageNumber,
	}

	if parentItem == nil {
		*out = append(*out, frame.list)
	} else {
		parentItem.Kids = append(parentItem.Kids, frame.list)
	}

	frame.append(candidate)
	return frame
}

func (f *listFrame) matches(candidate listCandidate, tolerance float64) bool {
	if f == nil {
		return false
	}
	if f.kind != candidate.kind {
		return false
	}
	if f.numbering != candidate.numberingStyle {
		return false
	}
	if candidate.base.PageIndex != f.pageIndex || candidate.base.PageNumber != f.pageNumber {
		return false
	}
	return floatAbs(f.indent-candidate.indent) <= tolerance
}

func (f *listFrame) append(candidate listCandidate) {
	if f == nil || f.list == nil {
		return
	}

	item := &model.ListItem{
		TextNode: model.TextNode{
			BaseNode: model.BaseNode{
				Type:       model.ElementTypeListItem,
				PageIndex:  candidate.base.PageIndex,
				PageNumber: candidate.base.PageNumber,
				Bounds:     candidate.base.Bounds,
			},
			TextProperties: candidate.textProperties(),
		},
	}

	f.list.ListItems = append(f.list.ListItems, item)
	f.list.NumberOfItems++
	if f.list.BaseNode.Bounds.IsZero() {
		f.list.BaseNode.Bounds = candidate.base.Bounds
	} else {
		f.list.BaseNode.Bounds = f.list.BaseNode.Bounds.Union(candidate.base.Bounds)
	}
	f.lastItem = item
}

func (c listCandidate) textProperties() model.TextProperties {
	props := model.TextProperties{
		Font:       c.props.Font,
		FontSize:   c.props.FontSize,
		TextColor:  c.props.TextColor,
		Content:    c.text,
		HiddenText: c.props.HiddenText,
		Bold:       c.props.Bold,
		Italic:     c.props.Italic,
		Underline:  c.props.Underline,
	}
	return props
}

func (d Detector) attachContinuation(frames []*listFrame, element model.ContentElement) bool {
	if len(frames) == 0 {
		return false
	}

	top := frames[len(frames)-1]
	if top == nil || top.lastItem == nil {
		return false
	}

	base := element.NodeBase()
	if base == nil {
		return false
	}
	if base.PageIndex != top.pageIndex || base.PageNumber != top.pageNumber {
		return false
	}

	indent := elementIndent(element)
	if indent < top.indent+d.ContinuationIndentThreshold {
		return false
	}

	top.lastItem.Kids = append(top.lastItem.Kids, element)
	return true
}

func (d Detector) candidateFromElement(element model.ContentElement) (listCandidate, bool) {
	base := element.NodeBase()
	if base == nil {
		return listCandidate{}, false
	}

	text, props, ok := elementText(element)
	if !ok {
		return listCandidate{}, false
	}

	indent := elementIndent(element)
	bounds := elementBounds(element)
	kind, numberingStyle, stripped, ok := parseListMarker(text)
	if !ok || strings.TrimSpace(stripped) == "" {
		return listCandidate{}, false
	}

	baseCopy := *base
	if !bounds.IsZero() {
		baseCopy.Bounds = bounds
	}
	return listCandidate{
		element:        element,
		base:           baseCopy,
		props:          props,
		text:           stripped,
		indent:         indent,
		numberingStyle: numberingStyle,
		kind:           kind,
	}, true
}

func elementText(element model.ContentElement) (string, model.TextProperties, bool) {
	switch node := element.(type) {
	case *model.Paragraph:
		return textFromNode(node.TextNode), node.TextProperties, true
	case *model.Caption:
		return textFromNode(node.TextNode), node.TextProperties, true
	case *model.ListItem:
		return textFromNode(node.TextNode), node.TextProperties, true
	case *model.TextBlock:
		return textFromNode(model.TextNode{TextProperties: model.TextProperties{Content: textBlockText(node.Kids)}}), model.TextProperties{}, true
	default:
		return "", model.TextProperties{}, false
	}
}

func textFromNode(node model.TextNode) string {
	if strings.TrimSpace(node.TextProperties.Content) != "" {
		return strings.TrimSpace(node.TextProperties.Content)
	}
	return ""
}

func textBlockText(kids []model.ContentElement) string {
	parts := make([]string, 0, len(kids))
	for _, kid := range kids {
		if kid == nil {
			continue
		}
		text, _, ok := elementText(kid)
		if !ok || strings.TrimSpace(text) == "" {
			continue
		}
		parts = append(parts, strings.TrimSpace(text))
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func elementIndent(element model.ContentElement) float64 {
	return elementBounds(element).Left
}

func elementBounds(element model.ContentElement) model.Box {
	base := element.NodeBase()
	if base == nil {
		return model.Box{}
	}
	if !base.Bounds.IsZero() {
		return base.Bounds.Normalize()
	}
	if bounds, ok := base.Boxes.Bounds(); ok {
		return bounds.Normalize()
	}
	return model.Box{}
}

func samePage(frame *listFrame, candidate listCandidate) bool {
	return frame != nil && frame.pageIndex == candidate.base.PageIndex && frame.pageNumber == candidate.base.PageNumber
}

func parseListMarker(text string) (markerKind, string, string, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return markerUnknown, "", "", false
	}

	r, size := utf8.DecodeRuneInString(trimmed)
	if isBulletRune(r) {
		rest := strings.TrimSpace(trimmed[size:])
		if rest != "" {
			return markerUnordered, "unordered", rest, true
		}
	}

	if kind, style, rest, ok := parseOrderedMarker(trimmed); ok {
		return kind, style, rest, true
	}

	return markerUnknown, "", "", false
}

func parseOrderedMarker(text string) (markerKind, string, string, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return markerUnknown, "", "", false
	}

	start := 0
	if strings.HasPrefix(trimmed, "(") {
		start = 1
	}

	i := start
	for i < len(trimmed) {
		r, size := utf8.DecodeRuneInString(trimmed[i:])
		if !isOrderedTokenRune(r) {
			break
		}
		i += size
	}
	if i == start {
		return markerUnknown, "", "", false
	}

	token := trimmed[start:i]
	if strings.HasPrefix(trimmed, "(") {
		if i >= len(trimmed) || trimmed[i] != ')' {
			return markerUnknown, "", "", false
		}
		i++
	} else {
		if i >= len(trimmed) || (trimmed[i] != '.' && trimmed[i] != ')') {
			return markerUnknown, "", "", false
		}
		i++
	}

	if i < len(trimmed) && !unicode.IsSpace(rune(trimmed[i])) {
		return markerUnknown, "", "", false
	}

	rest := strings.TrimSpace(trimmed[i:])
	if rest == "" {
		return markerUnknown, "", "", false
	}

	return markerOrdered, numberingStyle(token), rest, true
}

func numberingStyle(token string) string {
	if token == "" {
		return ""
	}

	switch {
	case allDigits(token):
		return "arabic"
	case len(token) >= 2 && allRoman(token):
		return "roman"
	case allLetters(token):
		if token == strings.ToUpper(token) {
			return "upper-alpha"
		}
		if token == strings.ToLower(token) {
			return "lower-alpha"
		}
		return "alpha"
	default:
		return ""
	}
}

func isBulletRune(r rune) bool {
	switch r {
	case '-', '*', '+', '•', '‣', '◦', '∙', '·', '–', '—', '▪', '▫':
		return true
	default:
		return false
	}
}

func isOrderedTokenRune(r rune) bool {
	return unicode.IsDigit(r) || unicode.IsLetter(r)
}

func allDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return s != ""
}

func allLetters(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return s != ""
}

func allRoman(s string) bool {
	for _, r := range strings.ToLower(s) {
		switch r {
		case 'i', 'v', 'x', 'l', 'c', 'd', 'm':
			continue
		default:
			return false
		}
	}
	return s != ""
}

func floatAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
