package headerfooter

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// Mode controls how detected header/footer elements are applied.
type Mode string

const (
	// ModeRemove removes detected elements from the page kids.
	ModeRemove Mode = "remove"
	// ModeMark wraps detected elements in model.HeaderFooter containers.
	ModeMark Mode = "mark"
)

// Role identifies whether a detection is a header or footer candidate.
type Role string

const (
	// RoleHeader marks content near the top of a page.
	RoleHeader Role = "header"
	// RoleFooter marks content near the bottom of a page.
	RoleFooter Role = "footer"
)

// Detector applies lightweight repeated header/footer heuristics across pages.
//
// The detector is intentionally pragmatic:
//   - It only considers text-bearing nodes near the top/bottom of the page.
//   - It normalizes text by lowercasing, collapsing whitespace, stripping
//     punctuation, and masking digit runs so page numbers still match.
//   - It only keeps candidates that repeat on at least MinRepeatPages pages.
type Detector struct {
	MinRepeatPages       int
	TopBandShare         float64
	BottomBandShare      float64
	MaxHeightShare       float64
	MaxWords             int
	MaxChars             int
	MinTextSimilarity    float64
	MaxEdgeOffsetDelta   float64
	MaxCenterOffsetDelta float64
	MaxWidthDelta        float64
}

// Detection describes one repeated header/footer element.
type Detection struct {
	Element        model.ContentElement
	Role           Role
	PageIndex      model.PageIndex
	PageNumber     model.PageNumber
	Order          int
	Text           string
	NormalizedText string
	Bounds         model.Box
	Score          float64
}

type candidate struct {
	element        model.ContentElement
	role           Role
	pageIndex      model.PageIndex
	pageNumber     model.PageNumber
	order          int
	text           string
	normalizedText string
	bounds         model.Box
	pageWidth      float64
	pageHeight     float64
	edgeOffset     float64
	centerXShare   float64
	widthShare     float64
	heightShare    float64
}

type cluster struct {
	role      Role
	prototype candidate
	items     []candidate
	pages     map[model.PageIndex]struct{}
}

// NewDetector returns a detector with conservative defaults.
func NewDetector() Detector {
	return Detector{
		MinRepeatPages:       2,
		TopBandShare:         0.20,
		BottomBandShare:      0.20,
		MaxHeightShare:       0.16,
		MaxWords:             24,
		MaxChars:             160,
		MinTextSimilarity:    0.88,
		MaxEdgeOffsetDelta:   0.08,
		MaxCenterOffsetDelta: 0.22,
		MaxWidthDelta:        0.30,
	}
}

// Detect scans the document and returns repeated header/footer detections.
func Detect(document *model.Document) []Detection {
	return NewDetector().Detect(document)
}

// Detect scans the document and returns repeated header/footer detections.
func (d Detector) Detect(document *model.Document) []Detection {
	d = d.withDefaults()
	candidates := d.collectCandidates(document)
	if len(candidates) == 0 {
		return nil
	}

	clusters := d.clusterCandidates(candidates)
	if len(clusters) == 0 {
		return nil
	}

	detections := make([]Detection, 0, len(candidates))
	for _, cl := range clusters {
		for _, item := range cl.items {
			detections = append(detections, Detection{
				Element:        item.element,
				Role:           item.role,
				PageIndex:      item.pageIndex,
				PageNumber:     item.pageNumber,
				Order:          item.order,
				Text:           item.text,
				NormalizedText: item.normalizedText,
				Bounds:         item.bounds,
				Score:          textSimilarity(item.normalizedText, cl.prototype.normalizedText),
			})
		}
	}

	sort.SliceStable(detections, func(i, j int) bool {
		if detections[i].PageIndex != detections[j].PageIndex {
			return detections[i].PageIndex < detections[j].PageIndex
		}
		if detections[i].Role != detections[j].Role {
			return detections[i].Role < detections[j].Role
		}
		if detections[i].Bounds.Top != detections[j].Bounds.Top {
			return detections[i].Bounds.Top > detections[j].Bounds.Top
		}
		if detections[i].Bounds.Bottom != detections[j].Bounds.Bottom {
			return detections[i].Bounds.Bottom < detections[j].Bounds.Bottom
		}
		return detections[i].Order < detections[j].Order
	})

	return detections
}

// Apply detects repeated headers/footers and mutates the document in place.
//
// ModeRemove drops detected nodes from the page kids.
// ModeMark wraps detected nodes in model.HeaderFooter containers.
func (d Detector) Apply(document *model.Document, mode Mode) []Detection {
	detections := d.Detect(document)
	switch mode {
	case ModeMark:
		applyMarked(document, detections)
	case ModeRemove:
		applyRemoved(document, detections)
	}
	return detections
}

func (d Detector) withDefaults() Detector {
	if d.MinRepeatPages <= 0 {
		d.MinRepeatPages = 2
	}
	if d.TopBandShare <= 0 || d.TopBandShare >= 0.5 {
		d.TopBandShare = 0.20
	}
	if d.BottomBandShare <= 0 || d.BottomBandShare >= 0.5 {
		d.BottomBandShare = 0.20
	}
	if d.MaxHeightShare <= 0 || d.MaxHeightShare >= 0.5 {
		d.MaxHeightShare = 0.16
	}
	if d.MaxWords <= 0 {
		d.MaxWords = 24
	}
	if d.MaxChars <= 0 {
		d.MaxChars = 160
	}
	if d.MinTextSimilarity <= 0 || d.MinTextSimilarity > 1 {
		d.MinTextSimilarity = 0.88
	}
	if d.MaxEdgeOffsetDelta <= 0 || d.MaxEdgeOffsetDelta >= 1 {
		d.MaxEdgeOffsetDelta = 0.08
	}
	if d.MaxCenterOffsetDelta <= 0 || d.MaxCenterOffsetDelta >= 1 {
		d.MaxCenterOffsetDelta = 0.22
	}
	if d.MaxWidthDelta <= 0 || d.MaxWidthDelta >= 1 {
		d.MaxWidthDelta = 0.30
	}
	return d
}

func (d Detector) collectCandidates(document *model.Document) []candidate {
	if document == nil {
		return nil
	}

	out := make([]candidate, 0)
	for _, page := range document.Pages {
		if page == nil || len(page.Kids) == 0 {
			continue
		}
		pageWidth, pageHeight, ok := pageMetrics(page)
		if !ok || pageHeight <= 0 {
			continue
		}

		for order, element := range page.Kids {
			cand, ok := buildCandidate(element, page, order, pageWidth, pageHeight, d)
			if !ok {
				continue
			}
			out = append(out, cand)
		}
	}
	return out
}

func (d Detector) clusterCandidates(candidates []candidate) []cluster {
	clusters := make([]cluster, 0, len(candidates))
	for _, cand := range candidates {
		best := -1
		bestScore := 0.0
		for i := range clusters {
			score, ok := d.matchCluster(clusters[i], cand)
			if !ok {
				continue
			}
			if score > bestScore {
				bestScore = score
				best = i
			}
		}

		if best < 0 {
			clusters = append(clusters, cluster{
				role:      cand.role,
				prototype: cand,
				items:     []candidate{cand},
				pages:     map[model.PageIndex]struct{}{cand.pageIndex: struct{}{}},
			})
			continue
		}

		clusters[best].items = append(clusters[best].items, cand)
		clusters[best].pages[cand.pageIndex] = struct{}{}
	}

	filtered := clusters[:0]
	for _, cl := range clusters {
		if len(cl.pages) >= d.MinRepeatPages {
			filtered = append(filtered, cl)
		}
	}
	return filtered
}

func (d Detector) matchCluster(cl cluster, cand candidate) (float64, bool) {
	if cl.role != cand.role {
		return 0, false
	}
	if len(cl.items) == 0 {
		return 0, false
	}

	textScore := textSimilarity(cl.prototype.normalizedText, cand.normalizedText)
	if textScore < d.MinTextSimilarity {
		return 0, false
	}

	if math.Abs(cl.prototype.edgeOffset-cand.edgeOffset) > d.MaxEdgeOffsetDelta {
		return 0, false
	}

	regionScore := 1.0
	if cl.prototype.pageWidth > 0 && cand.pageWidth > 0 {
		centerDelta := math.Abs(cl.prototype.centerXShare - cand.centerXShare)
		widthDelta := math.Abs(cl.prototype.widthShare - cand.widthShare)
		if centerDelta > d.MaxCenterOffsetDelta || widthDelta > d.MaxWidthDelta {
			return 0, false
		}
		centerScore := 1 - (centerDelta / d.MaxCenterOffsetDelta)
		widthScore := 1 - (widthDelta / d.MaxWidthDelta)
		regionScore = (centerScore + widthScore) / 2
	}

	return (textScore * 0.75) + (regionScore * 0.25), true
}

func buildCandidate(element model.ContentElement, page *model.Page, order int, pageWidth, pageHeight float64, d Detector) (candidate, bool) {
	if element == nil || page == nil {
		return candidate{}, false
	}

	text, ok := elementText(element)
	if !ok {
		return candidate{}, false
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return candidate{}, false
	}

	words := len(strings.Fields(text))
	if words == 0 || words > d.MaxWords || len(text) > d.MaxChars {
		return candidate{}, false
	}

	base := element.NodeBase()
	if base == nil {
		return candidate{}, false
	}

	bounds := nodeBounds(base)
	if bounds.IsZero() || pageHeight <= 0 {
		return candidate{}, false
	}

	heightShare := bounds.Height() / pageHeight
	if heightShare <= 0 || heightShare > d.MaxHeightShare {
		return candidate{}, false
	}

	topShare := bounds.Top / pageHeight
	bottomShare := bounds.Bottom / pageHeight
	headerEligible := topShare >= 1-d.TopBandShare
	footerEligible := bottomShare <= d.BottomBandShare
	if !headerEligible && !footerEligible {
		return candidate{}, false
	}

	role := RoleHeader
	edgeOffset := 1 - topShare
	if footerEligible && (!headerEligible || bottomShare < edgeOffset) {
		role = RoleFooter
		edgeOffset = bottomShare
	}

	cand := candidate{
		element:        element,
		role:           role,
		pageIndex:      base.PageIndex,
		pageNumber:     base.PageNumber,
		order:          order,
		text:           text,
		normalizedText: normalizeText(text),
		bounds:         bounds,
		pageWidth:      pageWidth,
		pageHeight:     pageHeight,
		edgeOffset:     edgeOffset,
		heightShare:    heightShare,
	}

	if pageWidth > 0 {
		cand.centerXShare = ((bounds.Left + bounds.Right) / 2) / pageWidth
		cand.widthShare = bounds.Width() / pageWidth
	}

	return cand, true
}

func applyRemoved(document *model.Document, detections []Detection) {
	if document == nil || len(detections) == 0 {
		return
	}

	remove := make(map[model.ContentElement]struct{}, len(detections))
	for _, detection := range detections {
		if detection.Element == nil {
			continue
		}
		remove[detection.Element] = struct{}{}
	}

	for _, page := range document.Pages {
		if page == nil || len(page.Kids) == 0 {
			continue
		}
		filtered := page.Kids[:0]
		for _, element := range page.Kids {
			if element == nil {
				continue
			}
			if _, ok := remove[element]; ok {
				continue
			}
			filtered = append(filtered, element)
		}
		page.Kids = append([]model.ContentElement(nil), filtered...)
	}

	rebuildDocumentKids(document)
}

func applyMarked(document *model.Document, detections []Detection) {
	if document == nil || len(detections) == 0 {
		return
	}

	perPage := make(map[model.PageIndex][]Detection)
	for _, detection := range detections {
		perPage[detection.PageIndex] = append(perPage[detection.PageIndex], detection)
	}

	for _, page := range document.Pages {
		if page == nil || len(page.Kids) == 0 {
			continue
		}

		pageDetections := perPage[page.Metadata.Index]
		if len(pageDetections) == 0 {
			continue
		}

		newKids := make([]model.ContentElement, 0, len(page.Kids))
		headerGroup := groupDetections(pageDetections, RoleHeader)
		footerGroup := groupDetections(pageDetections, RoleFooter)

		headerEmitted := false
		footerEmitted := false
		headerMembers := elementSet(headerGroup.elements)
		footerMembers := elementSet(footerGroup.elements)

		for _, element := range page.Kids {
			if element == nil {
				continue
			}

			if !headerEmitted {
				if _, ok := headerMembers[element]; ok {
					newKids = append(newKids, buildHeaderFooterNode(document, headerGroup, RoleHeader))
					headerEmitted = true
					continue
				}
			}
			if !footerEmitted {
				if _, ok := footerMembers[element]; ok {
					newKids = append(newKids, buildHeaderFooterNode(document, footerGroup, RoleFooter))
					footerEmitted = true
					continue
				}
			}
			if _, ok := headerMembers[element]; ok {
				continue
			}
			if _, ok := footerMembers[element]; ok {
				continue
			}
			newKids = append(newKids, element)
		}

		if len(newKids) > 0 {
			page.Kids = newKids
		}
	}

	rebuildDocumentKids(document)
}

type groupedDetections struct {
	elements []model.ContentElement
	bounds   model.Box
}

func groupDetections(detections []Detection, role Role) groupedDetections {
	filtered := make([]Detection, 0, len(detections))
	for _, detection := range detections {
		if detection.Role == role && detection.Element != nil {
			filtered = append(filtered, detection)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].Order != filtered[j].Order {
			return filtered[i].Order < filtered[j].Order
		}
		return filtered[i].Bounds.Top > filtered[j].Bounds.Top
	})

	group := groupedDetections{
		elements: make([]model.ContentElement, 0, len(filtered)),
	}
	for _, detection := range filtered {
		group.elements = append(group.elements, detection.Element)
		if group.bounds.IsZero() {
			group.bounds = detection.Bounds
		} else {
			group.bounds = group.bounds.Union(detection.Bounds)
		}
	}
	return group
}

func buildHeaderFooterNode(document *model.Document, group groupedDetections, role Role) model.ContentElement {
	if len(group.elements) == 0 {
		return nil
	}

	base := groupBase(group, document, role)
	return &model.HeaderFooter{
		BaseNode: *base,
		Kids:     append([]model.ContentElement(nil), group.elements...),
	}
}

func groupBase(group groupedDetections, document *model.Document, role Role) *model.BaseNode {
	var base *model.BaseNode
	for _, element := range group.elements {
		if element == nil {
			continue
		}
		nodeBase := element.NodeBase()
		if nodeBase == nil {
			continue
		}
		copyBase := copyBaseNode(nodeBase)
		if copyBase != nil {
			base = copyBase
			break
		}
	}
	if base == nil {
		base = &model.BaseNode{}
	}

	base.Type = elementTypeForRole(role)
	base.Bounds = group.bounds
	if document != nil && base.ID == 0 {
		base.ID = document.NewNodeID()
	}
	return base
}

func elementTypeForRole(role Role) model.ElementType {
	switch role {
	case RoleFooter:
		return model.ElementTypeFooter
	case RoleHeader:
		return model.ElementTypeHeader
	default:
		return model.ElementTypeUnknown
	}
}

func elementSet(elements []model.ContentElement) map[model.ContentElement]struct{} {
	out := make(map[model.ContentElement]struct{}, len(elements))
	for _, element := range elements {
		if element == nil {
			continue
		}
		out[element] = struct{}{}
	}
	return out
}

func rebuildDocumentKids(document *model.Document) {
	if document == nil {
		return
	}
	document.Kids = document.Kids[:0]
	for _, page := range document.Pages {
		if page == nil || len(page.Kids) == 0 {
			continue
		}
		document.Kids = append(document.Kids, page.Kids...)
	}
}

func (g groupedDetections) empty() bool {
	return len(g.elements) == 0
}

func pageMetrics(page *model.Page) (float64, float64, bool) {
	if page == nil {
		return 0, 0, false
	}

	if page.Metadata.Size.Width > 0 || page.Metadata.Size.Height > 0 {
		width := page.Metadata.Size.Width
		height := page.Metadata.Size.Height
		if width > 0 && height > 0 {
			return width, height, true
		}
	}

	if !page.Metadata.Bounds.IsZero() {
		width := page.Metadata.Bounds.Width()
		height := page.Metadata.Bounds.Height()
		if width > 0 && height > 0 {
			return width, height, true
		}
	}

	var maxRight, maxTop float64
	minLeft := math.MaxFloat64
	minBottom := math.MaxFloat64
	for _, element := range page.Kids {
		if element == nil {
			continue
		}
		box := nodeBounds(element.NodeBase())
		if box.IsZero() {
			continue
		}
		if box.Right > maxRight {
			maxRight = box.Right
		}
		if box.Top > maxTop {
			maxTop = box.Top
		}
		if box.Left < minLeft {
			minLeft = box.Left
		}
		if box.Bottom < minBottom {
			minBottom = box.Bottom
		}
	}

	if maxRight <= minLeft || maxTop <= minBottom {
		return 0, 0, false
	}
	return maxRight - minLeft, maxTop - minBottom, true
}

func nodeBounds(base *model.BaseNode) model.Box {
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

func elementText(element model.ContentElement) (string, bool) {
	switch node := element.(type) {
	case *model.Paragraph:
		return textFromNode(node.TextNode), true
	case *model.Heading:
		return textFromNode(node.TextNode), true
	case *model.Caption:
		return textFromNode(node.TextNode), true
	case *model.ListItem:
		return textFromNode(node.TextNode), true
	case *model.TextBlock:
		return textBlockText(node.Kids), true
	default:
		return "", false
	}
}

func textFromNode(node model.TextNode) string {
	return strings.TrimSpace(node.TextProperties.Content)
}

func textBlockText(kids []model.ContentElement) string {
	parts := make([]string, 0, len(kids))
	for _, kid := range kids {
		if kid == nil {
			continue
		}
		text, ok := elementText(kid)
		if !ok {
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		parts = append(parts, text)
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func normalizeText(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(text))

	lastWasSpace := false
	lastWasHash := false
	for _, r := range text {
		switch {
		case unicode.IsDigit(r):
			if !lastWasHash {
				b.WriteRune('#')
				lastWasHash = true
			}
			lastWasSpace = false
		case unicode.IsLetter(r):
			b.WriteRune(r)
			lastWasSpace = false
			lastWasHash = false
		case unicode.IsSpace(r):
			if !lastWasSpace && b.Len() > 0 {
				b.WriteByte(' ')
			}
			lastWasSpace = true
			lastWasHash = false
		default:
			if !lastWasSpace && b.Len() > 0 {
				b.WriteByte(' ')
			}
			lastWasSpace = true
			lastWasHash = false
		}
	}

	return strings.Join(strings.Fields(b.String()), " ")
}

func textSimilarity(a, b string) float64 {
	a = normalizeText(a)
	b = normalizeText(b)
	if a == "" || b == "" {
		return 0
	}
	if a == b {
		return 1
	}

	ar := []rune(a)
	br := []rune(b)
	distance := levenshtein(ar, br)
	maxLen := len(ar)
	if len(br) > maxLen {
		maxLen = len(br)
	}
	if maxLen == 0 {
		return 0
	}
	score := 1 - (float64(distance) / float64(maxLen))
	if score < 0 {
		return 0
	}
	return score
}

func levenshtein(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}

	for i, ra := range a {
		curr[0] = i + 1
		for j, rb := range b {
			cost := 0
			if ra != rb {
				cost = 1
			}

			deletion := prev[j+1] + 1
			insertion := curr[j] + 1
			substitution := prev[j] + cost
			curr[j+1] = minInt(deletion, minInt(insertion, substitution))
		}
		prev, curr = curr, prev
	}

	return prev[len(b)]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func copyBaseNode(base *model.BaseNode) *model.BaseNode {
	if base == nil {
		return nil
	}
	copy := *base
	if len(base.Boxes) > 0 {
		copy.Boxes = append(model.MultiBox(nil), base.Boxes...)
	}
	return &copy
}
