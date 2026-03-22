package nativepdf

import (
	"fmt"
	"math"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

type pageGraphics struct {
	artifacts       []*model.RawArtifact
	tableCandidates *TableCandidateSet
}

type graphicsSnapshot struct {
	matrix        matrix2D
	inPath        bool
	cursor        model.Point
	subpathStart  model.Point
	haveCursor    bool
	lineWidth     float64
	strokeRGB     string
	pathBounds    model.Box
	pathBoundsSet bool
	pathHasCurves bool
	segments      []LineSegment
	rectangles    []model.Box
}

type graphicsState struct {
	page            model.PageMetadata
	graph           *objectGraph
	xobjectRefs     map[string]pdfRef
	matrix          matrix2D
	stack           []graphicsSnapshot
	inPath          bool
	cursor          model.Point
	subpathStart    model.Point
	haveCursor      bool
	lineWidth       float64
	strokeRGB       string
	pathBounds      model.Box
	pathBoundsSet   bool
	pathHasCurves   bool
	segments        []LineSegment
	rectangles      []model.Box
	artifacts       []*model.RawArtifact
	tableCandidates TableCandidateSet
	sequence        int
	markedContent   []markedContentScope
}

func extractGraphicsData(data []byte, pages []model.PageMetadata, plans []pagePlan, graph *objectGraph) []pageGraphics {
	if len(pages) == 0 {
		return nil
	}

	if len(plans) == len(pages) && len(plans) > 0 {
		out := make([]pageGraphics, len(pages))
		for i, plan := range plans {
			streams := decodedStreamsForPagePlan(data, graph, plan)
			if len(streams) == 0 {
				continue
			}
			payload := extractGraphicsFromStreams(streams, pages[i], graph, plan.XObjectRefs)
			if len(payload.artifacts) == 0 && payload.tableCandidates == nil && len(pages) == 1 {
				payload = extractGraphicsFromStreams(extractDecodedStreams(data), pages[i], graph, plan.XObjectRefs)
			}
			if len(payload.artifacts) == 0 && payload.tableCandidates == nil {
				continue
			}
			out[i].artifacts = append(out[i].artifacts, payload.artifacts...)
			out[i].tableCandidates = mergeTableCandidateSets(out[i].tableCandidates, payload.tableCandidates)
		}
		return out
	}

	decodedStreams := extractDecodedStreams(data)
	if len(decodedStreams) == 0 {
		return make([]pageGraphics, len(pages))
	}

	out := make([]pageGraphics, len(pages))
	for streamIndex, decoded := range decodedStreams {
		pageIndex := streamIndex * len(pages) / len(decodedStreams)
		if pageIndex >= len(pages) {
			pageIndex = len(pages) - 1
		}
		payload := extractGraphicsFromStream(decoded, pages[pageIndex], graph, nil)
		if len(payload.artifacts) == 0 && payload.tableCandidates == nil {
			continue
		}
		out[pageIndex].artifacts = append(out[pageIndex].artifacts, payload.artifacts...)
		out[pageIndex].tableCandidates = mergeTableCandidateSets(out[pageIndex].tableCandidates, payload.tableCandidates)
	}
	return out
}

func extractGraphicsFromStreams(streams [][]byte, page model.PageMetadata, graph *objectGraph, xobjectRefs map[string]pdfRef) pageGraphics {
	if len(streams) == 0 {
		return pageGraphics{}
	}

	var combined pageGraphics
	for _, stream := range streams {
		payload := extractGraphicsFromStream(stream, page, graph, xobjectRefs)
		if len(payload.artifacts) > 0 {
			combined.artifacts = append(combined.artifacts, payload.artifacts...)
		}
		combined.tableCandidates = mergeTableCandidateSets(combined.tableCandidates, payload.tableCandidates)
	}
	return combined
}

func extractGraphicsFromStream(data []byte, page model.PageMetadata, graph *objectGraph, xobjectRefs map[string]pdfRef) pageGraphics {
	tokenizer := newContentTokenizer(data)
	state := graphicsState{
		page:        page,
		graph:       graph,
		xobjectRefs: xobjectRefs,
		matrix:      identityMatrix(),
		lineWidth:   1,
	}

	stack := make([]contentToken, 0, 16)
	for {
		token, ok := tokenizer.next()
		if !ok {
			break
		}
		if token.kind != contentTokenOperator {
			stack = append(stack, token)
			continue
		}
		state.applyOperator(token.text, stack)
		stack = stack[:0]
	}
	state.flushPath(true, false)
	return pageGraphics{
		artifacts:       state.artifacts,
		tableCandidates: cloneTableCandidateSet(&state.tableCandidates),
	}
}

func (s *graphicsState) applyOperator(op string, operands []contentToken) {
	switch op {
	case "q":
		s.pushState()
	case "Q":
		s.popState()
	case "cm":
		s.applyTransform(operands)
	case "BMC":
		s.beginMarkedContent(operands)
	case "BDC":
		s.beginMarkedContent(operands)
	case "EMC":
		s.endMarkedContent()
	case "w":
		s.applyLineWidth(operands)
	case "G", "RG":
		s.applyStrokeColor(op, operands)
	case "m":
		s.moveTo(operands)
	case "l":
		s.lineTo(operands)
	case "re":
		s.addRectangle(operands)
	case "c":
		s.curveTo(operands)
	case "v":
		s.curveToV(operands)
	case "y":
		s.curveToY(operands)
	case "h":
		s.closePath()
	case "Do":
		s.applyXObject(operands)
	case "S":
		s.flushPath(true, false)
	case "s":
		s.closePath()
		s.flushPath(true, false)
	case "f", "F", "f*":
		s.flushPath(false, true)
	case "B", "B*":
		s.flushPath(true, true)
	case "b", "b*":
		s.closePath()
		s.flushPath(true, true)
	case "n":
		s.clearPath()
	}
}

func (s *graphicsState) pushState() {
	s.stack = append(s.stack, graphicsSnapshot{
		matrix:        s.matrix,
		inPath:        s.inPath,
		cursor:        s.cursor,
		subpathStart:  s.subpathStart,
		haveCursor:    s.haveCursor,
		lineWidth:     s.lineWidth,
		strokeRGB:     s.strokeRGB,
		pathBounds:    s.pathBounds,
		pathBoundsSet: s.pathBoundsSet,
		pathHasCurves: s.pathHasCurves,
		segments:      append([]LineSegment(nil), s.segments...),
		rectangles:    append([]model.Box(nil), s.rectangles...),
	})
}

func (s *graphicsState) popState() {
	if len(s.stack) == 0 {
		return
	}
	last := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	s.matrix = last.matrix
	s.inPath = last.inPath
	s.cursor = last.cursor
	s.subpathStart = last.subpathStart
	s.haveCursor = last.haveCursor
	s.lineWidth = last.lineWidth
	s.strokeRGB = last.strokeRGB
	s.pathBounds = last.pathBounds
	s.pathBoundsSet = last.pathBoundsSet
	s.pathHasCurves = last.pathHasCurves
	s.segments = append(s.segments[:0], last.segments...)
	s.rectangles = append(s.rectangles[:0], last.rectangles...)
}

func (s *graphicsState) applyTransform(operands []contentToken) {
	if len(operands) < 6 {
		return
	}
	a, ok := operands[len(operands)-6].asNumber()
	if !ok {
		return
	}
	b, ok := operands[len(operands)-5].asNumber()
	if !ok {
		return
	}
	c, ok := operands[len(operands)-4].asNumber()
	if !ok {
		return
	}
	d, ok := operands[len(operands)-3].asNumber()
	if !ok {
		return
	}
	e, ok := operands[len(operands)-2].asNumber()
	if !ok {
		return
	}
	f, ok := operands[len(operands)-1].asNumber()
	if !ok {
		return
	}
	s.matrix = s.matrix.multiply(matrix2D{a: a, b: b, c: c, d: d, e: e, f: f})
}

func (s *graphicsState) applyLineWidth(operands []contentToken) {
	if len(operands) == 0 {
		return
	}
	if width, ok := operands[len(operands)-1].asNumber(); ok && width > 0 {
		s.lineWidth = width
	}
}

func (s *graphicsState) applyStrokeColor(op string, operands []contentToken) {
	switch op {
	case "G":
		if len(operands) == 0 {
			return
		}
		if gray, ok := operands[len(operands)-1].asNumber(); ok {
			s.strokeRGB = fmt.Sprintf("gray(%.3f)", gray)
		}
	case "RG":
		if len(operands) < 3 {
			return
		}
		r, ok := operands[len(operands)-3].asNumber()
		if !ok {
			return
		}
		g, ok := operands[len(operands)-2].asNumber()
		if !ok {
			return
		}
		b, ok := operands[len(operands)-1].asNumber()
		if !ok {
			return
		}
		s.strokeRGB = fmt.Sprintf("rgb(%.3f,%.3f,%.3f)", r, g, b)
	}
}

func (s *graphicsState) moveTo(operands []contentToken) {
	if len(operands) < 2 {
		return
	}
	x, ok := operands[len(operands)-2].asNumber()
	if !ok {
		return
	}
	y, ok := operands[len(operands)-1].asNumber()
	if !ok {
		return
	}
	point := s.matrix.transformPoint(model.Point{X: x, Y: y})
	s.cursor = point
	s.subpathStart = point
	s.haveCursor = true
	s.inPath = true
	s.includePathPoint(point)
}

func (s *graphicsState) lineTo(operands []contentToken) {
	if !s.haveCursor || len(operands) < 2 {
		return
	}
	x, ok := operands[len(operands)-2].asNumber()
	if !ok {
		return
	}
	y, ok := operands[len(operands)-1].asNumber()
	if !ok {
		return
	}
	next := s.matrix.transformPoint(model.Point{X: x, Y: y})
	s.includePathPoint(s.cursor)
	s.includePathPoint(next)
	segment := LineSegment{
		Start:     s.cursor,
		End:       next,
		Width:     s.lineWidth,
		StrokeRGB: s.strokeRGB,
		PageIndex: s.page.Index,
	}
	s.segments = append(s.segments, segment)
	s.cursor = next
	s.inPath = true
}

func (s *graphicsState) curveTo(operands []contentToken) {
	if !s.haveCursor || len(operands) < 6 {
		return
	}
	p1, ok := s.curvePoint(operands[len(operands)-6], operands[len(operands)-5])
	if !ok {
		return
	}
	p2, ok := s.curvePoint(operands[len(operands)-4], operands[len(operands)-3])
	if !ok {
		return
	}
	p3, ok := s.curvePoint(operands[len(operands)-2], operands[len(operands)-1])
	if !ok {
		return
	}
	s.includePathPoint(s.cursor)
	s.includePathPoint(p1)
	s.includePathPoint(p2)
	s.includePathPoint(p3)
	s.cursor = p3
	s.subpathStart = p3
	s.pathHasCurves = true
	s.inPath = true
}

func (s *graphicsState) curveToV(operands []contentToken) {
	if !s.haveCursor || len(operands) < 4 {
		return
	}
	p1 := s.cursor
	p2, ok := s.curvePoint(operands[len(operands)-4], operands[len(operands)-3])
	if !ok {
		return
	}
	p3, ok := s.curvePoint(operands[len(operands)-2], operands[len(operands)-1])
	if !ok {
		return
	}
	s.includePathPoint(p1)
	s.includePathPoint(p2)
	s.includePathPoint(p3)
	s.cursor = p3
	s.subpathStart = p3
	s.pathHasCurves = true
	s.inPath = true
}

func (s *graphicsState) curveToY(operands []contentToken) {
	if !s.haveCursor || len(operands) < 4 {
		return
	}
	p1, ok := s.curvePoint(operands[len(operands)-4], operands[len(operands)-3])
	if !ok {
		return
	}
	p2 := p1
	p3, ok := s.curvePoint(operands[len(operands)-2], operands[len(operands)-1])
	if !ok {
		return
	}
	s.includePathPoint(s.cursor)
	s.includePathPoint(p1)
	s.includePathPoint(p2)
	s.includePathPoint(p3)
	s.cursor = p3
	s.subpathStart = p3
	s.pathHasCurves = true
	s.inPath = true
}

func (s *graphicsState) addRectangle(operands []contentToken) {
	if len(operands) < 4 {
		return
	}
	x, ok := operands[len(operands)-4].asNumber()
	if !ok {
		return
	}
	y, ok := operands[len(operands)-3].asNumber()
	if !ok {
		return
	}
	w, ok := operands[len(operands)-2].asNumber()
	if !ok {
		return
	}
	h, ok := operands[len(operands)-1].asNumber()
	if !ok {
		return
	}
	rect := s.matrix.transformBox(model.Box{
		Left:   x,
		Bottom: y,
		Right:  x + w,
		Top:    y + h,
	})
	s.rectangles = append(s.rectangles, rect.Normalize())
	s.includePathBox(rect)
	s.inPath = true
}

func (s *graphicsState) closePath() {
	if !s.haveCursor {
		return
	}
	if s.cursor != s.subpathStart {
		s.includePathPoint(s.cursor)
		s.includePathPoint(s.subpathStart)
		segment := LineSegment{
			Start:     s.cursor,
			End:       s.subpathStart,
			Width:     s.lineWidth,
			StrokeRGB: s.strokeRGB,
			PageIndex: s.page.Index,
		}
		s.segments = append(s.segments, segment)
	}
	s.cursor = s.subpathStart
	s.inPath = true
}

func (s *graphicsState) applyXObject(operands []contentToken) {
	if len(operands) == 0 || len(s.xobjectRefs) == 0 {
		return
	}
	name, ok := operands[len(operands)-1].asName()
	if !ok {
		return
	}
	ref, ok := s.xobjectRefs[name]
	if !ok {
		return
	}
	artifact, ok := imageArtifactFromXObject(s.graph, ref, name, s.page, s.matrix, s.sequence)
	if !ok {
		return
	}
	artifact.MarkedContentID = cloneIntPointer(s.currentMarkedContentID())
	s.artifacts = append(s.artifacts, artifact)
	s.sequence++
}

func (s *graphicsState) flushPath(stroke, fill bool) {
	if !s.inPath && len(s.segments) == 0 && len(s.rectangles) == 0 && !s.pathHasCurves {
		return
	}

	if stroke {
		for _, segment := range s.segments {
			artifact := lineArtifactFromSegment(segment, s.page, s.sequence, s.currentMarkedContentID())
			s.sequence++
			s.artifacts = append(s.artifacts, artifact)
			s.tableCandidates.HorizontalLines, s.tableCandidates.VerticalLines = appendCandidateLines(s.tableCandidates.HorizontalLines, s.tableCandidates.VerticalLines, segment)
		}
	}
	if fill || s.pathHasCurves || len(s.rectangles) > 0 {
		box := s.pathBounds
		if !s.pathBoundsSet && len(s.rectangles) > 0 {
			box = s.rectangles[0]
			for _, rect := range s.rectangles[1:] {
				box = box.Union(rect)
			}
		}
		artifact := pathArtifactFromBox(box.Normalize(), s.page, s.sequence, s.currentMarkedContentID())
		s.sequence++
		s.artifacts = append(s.artifacts, artifact)
		for _, rect := range s.rectangles {
			s.tableCandidates.Rectangles = append(s.tableCandidates.Rectangles, rect)
		}
	}

	s.clearPath()
}

func (s *graphicsState) clearPath() {
	s.inPath = false
	s.haveCursor = false
	s.pathBounds = model.Box{}
	s.pathBoundsSet = false
	s.pathHasCurves = false
	s.segments = s.segments[:0]
	s.rectangles = s.rectangles[:0]
}

func (s *graphicsState) curvePoint(xTok, yTok contentToken) (model.Point, bool) {
	x, ok := xTok.asNumber()
	if !ok {
		return model.Point{}, false
	}
	y, ok := yTok.asNumber()
	if !ok {
		return model.Point{}, false
	}
	return s.matrix.transformPoint(model.Point{X: x, Y: y}), true
}

func (s *graphicsState) includePathPoint(point model.Point) {
	s.includePathCoord(point.X, point.Y)
}

func (s *graphicsState) includePathBox(box model.Box) {
	box = box.Normalize()
	s.includePathCoord(box.Left, box.Bottom)
	s.includePathCoord(box.Left, box.Top)
	s.includePathCoord(box.Right, box.Bottom)
	s.includePathCoord(box.Right, box.Top)
}

func (s *graphicsState) includePathCoord(x, y float64) {
	if !s.pathBoundsSet {
		s.pathBounds = model.Box{Left: x, Bottom: y, Right: x, Top: y}
		s.pathBoundsSet = true
		return
	}
	if x < s.pathBounds.Left {
		s.pathBounds.Left = x
	}
	if x > s.pathBounds.Right {
		s.pathBounds.Right = x
	}
	if y < s.pathBounds.Bottom {
		s.pathBounds.Bottom = y
	}
	if y > s.pathBounds.Top {
		s.pathBounds.Top = y
	}
}

func lineArtifactFromSegment(segment LineSegment, page model.PageMetadata, sequence int, markedContentID *int) *model.RawArtifact {
	return &model.RawArtifact{
		Kind:            model.ArtifactKindLine,
		PageIndex:       page.Index,
		PageNumber:      page.Number,
		MarkedContentID: cloneIntPointer(markedContentID),
		Sequence:        sequence,
		Bounds:          lineBounds(segment),
		Style: model.TextProperties{
			Content: "line",
		},
	}
}

func pathArtifactFromBox(box model.Box, page model.PageMetadata, sequence int, markedContentID *int) *model.RawArtifact {
	return &model.RawArtifact{
		Kind:            model.ArtifactKindPath,
		PageIndex:       page.Index,
		PageNumber:      page.Number,
		MarkedContentID: cloneIntPointer(markedContentID),
		Sequence:        sequence,
		Bounds:          box,
		Boxes:           model.MultiBox{box},
		Style: model.TextProperties{
			Content: "path",
		},
	}
}

func (s *graphicsState) beginMarkedContent(operands []contentToken) {
	s.markedContent = append(s.markedContent, markedContentScopeFromOperands(operands))
}

func (s *graphicsState) endMarkedContent() {
	if len(s.markedContent) == 0 {
		return
	}
	s.markedContent = s.markedContent[:len(s.markedContent)-1]
}

func (s *graphicsState) currentMarkedContentID() *int {
	for i := len(s.markedContent) - 1; i >= 0; i-- {
		if s.markedContent[i].mcid != nil {
			return s.markedContent[i].mcid
		}
	}
	return nil
}

func lineBounds(segment LineSegment) model.Box {
	width := math.Max(0.5, segment.Width/2)
	left := math.Min(segment.Start.X, segment.End.X) - width
	right := math.Max(segment.Start.X, segment.End.X) + width
	bottom := math.Min(segment.Start.Y, segment.End.Y) - width
	top := math.Max(segment.Start.Y, segment.End.Y) + width
	return model.Box{Left: left, Bottom: bottom, Right: right, Top: top}
}

func appendCandidateLines(horizontal, vertical []LineSegment, segment LineSegment) ([]LineSegment, []LineSegment) {
	const tolerance = 0.75
	dx := math.Abs(segment.End.X - segment.Start.X)
	dy := math.Abs(segment.End.Y - segment.Start.Y)
	if dx >= dy && dy <= tolerance {
		return append(horizontal, segment), vertical
	}
	if dy > dx && dx <= tolerance {
		return horizontal, append(vertical, segment)
	}
	return horizontal, vertical
}

func mergeTableCandidateSets(existing, incoming *TableCandidateSet) *TableCandidateSet {
	if existing == nil {
		return cloneTableCandidateSet(incoming)
	}
	if incoming == nil {
		return cloneTableCandidateSet(existing)
	}
	merged := cloneTableCandidateSet(existing)
	merged.HorizontalLines = append(merged.HorizontalLines, incoming.HorizontalLines...)
	merged.VerticalLines = append(merged.VerticalLines, incoming.VerticalLines...)
	merged.Rectangles = append(merged.Rectangles, incoming.Rectangles...)
	return merged
}

func cloneTableCandidateSet(set *TableCandidateSet) *TableCandidateSet {
	if set == nil {
		return nil
	}
	cloned := &TableCandidateSet{
		HorizontalLines: append([]LineSegment(nil), set.HorizontalLines...),
		VerticalLines:   append([]LineSegment(nil), set.VerticalLines...),
		Rectangles:      append([]model.Box(nil), set.Rectangles...),
	}
	return cloned
}
