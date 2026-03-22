package nativepdf

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

type pdfRef struct {
	ObjectNumber int
	Generation   int
}

type pdfValue interface{}

type pdfName string

type pdfArray []pdfValue

type pdfDict map[string]pdfValue

type indirectObject struct {
	Ref   pdfRef
	Raw   []byte
	Value pdfValue
}

type objectGraph struct {
	objects map[pdfRef]*indirectObject
	trailer pdfDict
	rootRef pdfRef
	hasRoot bool
}

type pagePlan struct {
	Metadata    model.PageMetadata
	PageRef     pdfRef
	ContentRefs []pdfRef
	FontNames   map[string]string
	FontMetrics map[string]fontMetrics
	XObjectRefs map[string]pdfRef
}

func parsePageTreeMetadata(data []byte) ([]model.PageMetadata, bool) {
	plans, ok := parsePagePlans(data)
	if !ok {
		return nil, false
	}
	pages := make([]model.PageMetadata, 0, len(plans))
	for _, plan := range plans {
		pages = append(pages, plan.Metadata)
	}
	return pages, true
}

func parsePagePlans(data []byte) ([]pagePlan, bool) {
	graph, err := scanIndirectObjects(data)
	if err != nil || graph == nil || len(graph.objects) == 0 {
		return nil, false
	}

	rootRef, ok := graph.catalogPagesRef()
	if !ok {
		return nil, false
	}

	pages, ok := graph.resolvePagePlans(rootRef)
	if !ok || len(pages) == 0 {
		return nil, false
	}
	return pages, true
}

func scanIndirectObjects(data []byte) (*objectGraph, error) {
	graph := &objectGraph{
		objects: make(map[pdfRef]*indirectObject),
	}

	if info, ok := parseXrefInfo(data); ok {
		graph.trailer = info.Trailer
		if rootRef, ok := xrefInfoRoot(info); ok {
			graph.rootRef = rootRef
			graph.hasRoot = true
		}
		refs := sortEntriesByOffset(info.Entries)
		for _, ref := range refs {
			entry := info.Entries[ref]
			if !entry.InUse || entry.Offset < 0 || entry.Offset >= len(data) {
				continue
			}
			obj, ok := parseIndirectObjectAt(data, entry.Offset)
			if !ok {
				continue
			}
			graph.objects[obj.Ref] = obj
		}
	}

	if len(graph.objects) == 0 {
		offset := 0
		for {
			start, ref, ok := findNextObjectHeader(data, offset)
			if !ok {
				break
			}
			end, ok := findObjectBodyEnd(data, start)
			if !ok {
				break
			}

			raw := bytes.TrimSpace(data[start:end])
			value := firstObjectValue(raw)
			if value != nil {
				graph.objects[ref] = &indirectObject{
					Ref:   ref,
					Raw:   append([]byte(nil), raw...),
					Value: value,
				}
			}
			offset = end
		}
	}

	if len(graph.objects) == 0 {
		return nil, fmt.Errorf("no indirect objects scanned")
	}
	return graph, nil
}

func findNextObjectHeader(data []byte, offset int) (int, pdfRef, bool) {
	for i := offset; i < len(data); i++ {
		if !isDigit(data[i]) {
			continue
		}
		objNum, next, ok := parsePositiveInt(data, i)
		if !ok {
			continue
		}
		next = skipSpaces(data, next)
		genNum, next, ok := parsePositiveInt(data, next)
		if !ok {
			continue
		}
		next = skipSpaces(data, next)
		if !hasKeyword(data, next, "obj") {
			continue
		}
		next += len("obj")
		return next, pdfRef{ObjectNumber: objNum, Generation: genNum}, true
	}
	return 0, pdfRef{}, false
}

func (g *objectGraph) catalogPagesRef() (pdfRef, bool) {
	if g.hasRoot {
		if object := g.objects[g.rootRef]; object != nil {
			if dict, ok := object.Value.(pdfDict); ok && nameValue(dict["Type"]) == "Catalog" {
				if ref, ok := dict["Pages"].(pdfRef); ok {
					return ref, true
				}
			}
		}
	}
	for _, object := range g.objects {
		dict, ok := object.Value.(pdfDict)
		if !ok {
			continue
		}
		if nameValue(dict["Type"]) != "Catalog" {
			continue
		}
		ref, ok := dict["Pages"].(pdfRef)
		if ok {
			return ref, true
		}
	}
	return pdfRef{}, false
}

func (g *objectGraph) resolvePagePlans(root pdfRef) ([]pagePlan, bool) {
	rootObject := g.objects[root]
	if rootObject == nil {
		return nil, false
	}
	rootDict, ok := rootObject.Value.(pdfDict)
	if !ok || nameValue(rootDict["Type"]) != "Pages" {
		return nil, false
	}

	state := inheritedPageState{}
	if mediaBox, ok := arrayToBox(rootDict["MediaBox"]); ok {
		state.MediaBox = mediaBox
	}
	if cropBox, ok := arrayToBox(rootDict["CropBox"]); ok {
		state.CropBox = cropBox
	}
	if rotate, ok := intValue(rootDict["Rotate"]); ok {
		state.Rotate = rotate
	}
	if resources, ok := resolveDictValue(g, rootDict["Resources"]); ok {
		state.Resources = resources
	}

	var out []pagePlan
	if !g.walkPages(root, state, &out) {
		return nil, false
	}
	for i := range out {
		out[i].Metadata.Index = model.PageIndex(i)
		if out[i].Metadata.Number <= 0 {
			out[i].Metadata.Number = model.PageNumber(i + 1)
		}
	}
	return out, true
}

func (g *objectGraph) resolvePageTree(root pdfRef) ([]model.PageMetadata, bool) {
	plans, ok := g.resolvePagePlans(root)
	if !ok {
		return nil, false
	}
	pages := make([]model.PageMetadata, 0, len(plans))
	for _, plan := range plans {
		pages = append(pages, plan.Metadata)
	}
	return pages, true
}

type inheritedPageState struct {
	MediaBox  model.Box
	CropBox   model.Box
	Rotate    int
	Resources pdfDict
}

func (g *objectGraph) walkPages(ref pdfRef, inherited inheritedPageState, out *[]pagePlan) bool {
	object := g.objects[ref]
	if object == nil {
		return false
	}
	dict, ok := object.Value.(pdfDict)
	if !ok {
		return false
	}

	state := inherited
	if mediaBox, ok := arrayToBox(dict["MediaBox"]); ok {
		state.MediaBox = mediaBox
	}
	if cropBox, ok := arrayToBox(dict["CropBox"]); ok {
		state.CropBox = cropBox
	}
	if rotate, ok := intValue(dict["Rotate"]); ok {
		state.Rotate = rotate
	}
	if resources, ok := resolveDictValue(g, dict["Resources"]); ok {
		state.Resources = resources
	}

	switch nameValue(dict["Type"]) {
	case "Pages":
		kids, ok := dict["Kids"].(pdfArray)
		if !ok || len(kids) == 0 {
			return false
		}
		for _, kid := range kids {
			childRef, ok := kid.(pdfRef)
			if !ok {
				return false
			}
			if !g.walkPages(childRef, state, out) {
				return false
			}
		}
		return true
	case "Page":
		bounds := state.CropBox
		if bounds.IsZero() {
			bounds = state.MediaBox
		}
		bounds = bounds.Normalize()
		*out = append(*out, pagePlan{
			Metadata: model.PageMetadata{
				Bounds:   bounds,
				Size:     model.PageSize{Width: bounds.Width(), Height: bounds.Height()},
				Rotation: state.Rotate,
			},
			PageRef:     ref,
			ContentRefs: contentRefsFromValue(dict["Contents"]),
			FontNames:   resolveFontNames(g, state.Resources),
			FontMetrics: resolveFontMetrics(g, state.Resources),
			XObjectRefs: resolveImageXObjectRefs(g, state.Resources),
		})
		return true
	default:
		return false
	}
}

func contentRefsFromValue(value pdfValue) []pdfRef {
	switch typed := value.(type) {
	case pdfRef:
		return []pdfRef{typed}
	case pdfArray:
		refs := make([]pdfRef, 0, len(typed))
		for _, item := range typed {
			if ref, ok := item.(pdfRef); ok {
				refs = append(refs, ref)
			}
		}
		if len(refs) == 0 {
			return nil
		}
		return refs
	default:
		return nil
	}
}

func (g *objectGraph) decodedStreamForRef(ref pdfRef) ([]byte, bool) {
	if g == nil {
		return nil, false
	}
	object := g.objects[ref]
	if object == nil {
		return nil, false
	}
	return decodePDFStreamObject(object.Raw)
}

func (g *objectGraph) decodedStreamsForRefs(refs []pdfRef) [][]byte {
	if g == nil || len(refs) == 0 {
		return nil
	}
	out := make([][]byte, 0, len(refs))
	for _, ref := range refs {
		if decoded, ok := g.decodedStreamForRef(ref); ok && len(decoded) > 0 {
			out = append(out, decoded)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func firstObjectValue(raw []byte) pdfValue {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil
	}
	if start, _, ok := findNextObjectHeader(raw, 0); ok && start < len(raw) {
		raw = bytes.TrimSpace(raw[start:])
	}
	if idx := bytes.Index(raw, []byte("stream")); idx >= 0 {
		raw = bytes.TrimSpace(raw[:idx])
	}
	parser := pdfValueParser{data: raw}
	value, ok := parser.parseValue()
	if !ok {
		return nil
	}
	return value
}

type pdfValueParser struct {
	data []byte
	pos  int
}

func (p *pdfValueParser) parseValue() (pdfValue, bool) {
	p.skipWS()
	if p.pos >= len(p.data) {
		return nil, false
	}

	switch p.data[p.pos] {
	case '<':
		if p.pos+1 < len(p.data) && p.data[p.pos+1] == '<' {
			return p.parseDict()
		}
	case '[':
		return p.parseArray()
	case '/':
		return p.parseName()
	case '(':
		return p.parseLiteralString()
	}

	if ref, ok := p.parseRef(); ok {
		return ref, true
	}
	if number, ok := p.parseNumber(); ok {
		return number, true
	}
	if keyword, ok := p.parseKeyword(); ok {
		switch keyword {
		case "true":
			return true, true
		case "false":
			return false, true
		case "null":
			return nil, true
		default:
			return keyword, true
		}
	}
	return nil, false
}

func (p *pdfValueParser) parseDict() (pdfValue, bool) {
	if !p.consumeString("<<") {
		return nil, false
	}
	dict := make(pdfDict)
	for {
		p.skipWS()
		if p.consumeString(">>") {
			return dict, true
		}
		keyValue, ok := p.parseName()
		if !ok {
			return nil, false
		}
		key, ok := keyValue.(pdfName)
		if !ok {
			return nil, false
		}
		value, ok := p.parseValue()
		if !ok {
			return nil, false
		}
		dict[string(key)] = value
	}
}

func (p *pdfValueParser) parseArray() (pdfValue, bool) {
	if p.pos >= len(p.data) || p.data[p.pos] != '[' {
		return nil, false
	}
	p.pos++
	var values pdfArray
	for {
		p.skipWS()
		if p.pos >= len(p.data) {
			return nil, false
		}
		if p.data[p.pos] == ']' {
			p.pos++
			return values, true
		}
		value, ok := p.parseValue()
		if !ok {
			return nil, false
		}
		values = append(values, value)
	}
}

func (p *pdfValueParser) parseName() (pdfValue, bool) {
	if p.pos >= len(p.data) || p.data[p.pos] != '/' {
		return nil, false
	}
	p.pos++
	start := p.pos
	for p.pos < len(p.data) && !isDelimiter(p.data[p.pos]) && !isWhitespace(p.data[p.pos]) {
		p.pos++
	}
	return pdfName(p.data[start:p.pos]), p.pos > start
}

func (p *pdfValueParser) parseLiteralString() (pdfValue, bool) {
	if p.pos >= len(p.data) || p.data[p.pos] != '(' {
		return nil, false
	}
	p.pos++
	start := p.pos
	depth := 1
	for p.pos < len(p.data) {
		ch := p.data[p.pos]
		if ch == '\\' {
			p.pos += 2
			continue
		}
		if ch == '(' {
			depth++
		}
		if ch == ')' {
			depth--
			if depth == 0 {
				text := string(p.data[start:p.pos])
				p.pos++
				return text, true
			}
		}
		p.pos++
	}
	return nil, false
}

func (p *pdfValueParser) parseRef() (pdfRef, bool) {
	save := p.pos
	first, next, ok := parseIntegerAt(p.data, p.pos)
	if !ok {
		return pdfRef{}, false
	}
	next = skipSpaces(p.data, next)
	second, next, ok := parseIntegerAt(p.data, next)
	if !ok {
		p.pos = save
		return pdfRef{}, false
	}
	next = skipSpaces(p.data, next)
	if !hasKeyword(p.data, next, "R") {
		p.pos = save
		return pdfRef{}, false
	}
	p.pos = next + 1
	return pdfRef{ObjectNumber: first, Generation: second}, true
}

func (p *pdfValueParser) parseNumber() (pdfValue, bool) {
	number, next, ok := parseNumericAt(p.data, p.pos)
	if !ok {
		return nil, false
	}
	p.pos = next
	return number, true
}

func (p *pdfValueParser) parseKeyword() (string, bool) {
	p.skipWS()
	start := p.pos
	for p.pos < len(p.data) && !isDelimiter(p.data[p.pos]) && !isWhitespace(p.data[p.pos]) {
		p.pos++
	}
	if p.pos == start {
		return "", false
	}
	return string(p.data[start:p.pos]), true
}

func (p *pdfValueParser) skipWS() {
	p.pos = skipSpaces(p.data, p.pos)
}

func (p *pdfValueParser) consumeString(s string) bool {
	p.skipWS()
	if !bytes.HasPrefix(p.data[p.pos:], []byte(s)) {
		return false
	}
	p.pos += len(s)
	return true
}

func parsePositiveInt(data []byte, pos int) (int, int, bool) {
	value, next, ok := parseIntegerAt(data, pos)
	if !ok || value < 0 {
		return 0, pos, false
	}
	return value, next, true
}

func parseIntegerAt(data []byte, pos int) (int, int, bool) {
	start := pos
	if pos < len(data) && (data[pos] == '+' || data[pos] == '-') {
		pos++
	}
	for pos < len(data) && isDigit(data[pos]) {
		pos++
	}
	if pos == start || (pos == start+1 && (data[start] == '+' || data[start] == '-')) {
		return 0, start, false
	}
	value, err := strconv.Atoi(string(data[start:pos]))
	if err != nil {
		return 0, start, false
	}
	return value, pos, true
}

func parseNumericAt(data []byte, pos int) (float64, int, bool) {
	start := pos
	if pos < len(data) && (data[pos] == '+' || data[pos] == '-') {
		pos++
	}
	dotCount := 0
	for pos < len(data) {
		switch {
		case isDigit(data[pos]):
			pos++
		case data[pos] == '.':
			dotCount++
			if dotCount > 1 {
				return 0, start, false
			}
			pos++
		default:
			goto done
		}
	}
done:
	if pos == start || (pos == start+1 && (data[start] == '+' || data[start] == '-')) {
		return 0, start, false
	}
	value, err := strconv.ParseFloat(string(data[start:pos]), 64)
	if err != nil {
		return 0, start, false
	}
	return value, pos, true
}

func arrayToBox(value pdfValue) (model.Box, bool) {
	items, ok := value.(pdfArray)
	if !ok || len(items) < 4 {
		return model.Box{}, false
	}
	left, ok := floatValue(items[0])
	if !ok {
		return model.Box{}, false
	}
	bottom, ok := floatValue(items[1])
	if !ok {
		return model.Box{}, false
	}
	right, ok := floatValue(items[2])
	if !ok {
		return model.Box{}, false
	}
	top, ok := floatValue(items[3])
	if !ok {
		return model.Box{}, false
	}
	return model.Box{Left: left, Bottom: bottom, Right: right, Top: top}, true
}

func floatValue(value pdfValue) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case string:
		number, err := strconv.ParseFloat(typed, 64)
		return number, err == nil
	default:
		return 0, false
	}
}

func intValue(value pdfValue) (int, bool) {
	switch typed := value.(type) {
	case float64:
		return int(typed), true
	case int:
		return typed, true
	case string:
		number, err := strconv.Atoi(typed)
		return number, err == nil
	default:
		return 0, false
	}
}

func nameValue(value pdfValue) string {
	switch typed := value.(type) {
	case pdfName:
		return string(typed)
	case string:
		return strings.TrimPrefix(typed, "/")
	default:
		return ""
	}
}

func skipSpaces(data []byte, pos int) int {
	for pos < len(data) {
		if isWhitespace(data[pos]) {
			pos++
			continue
		}
		if data[pos] == '%' {
			for pos < len(data) && data[pos] != '\n' && data[pos] != '\r' {
				pos++
			}
			continue
		}
		break
	}
	return pos
}

func hasKeyword(data []byte, pos int, keyword string) bool {
	if pos < 0 || pos+len(keyword) > len(data) {
		return false
	}
	if !bytes.Equal(data[pos:pos+len(keyword)], []byte(keyword)) {
		return false
	}
	beforeOK := pos == 0 || isWhitespace(data[pos-1]) || isDelimiter(data[pos-1])
	afterPos := pos + len(keyword)
	afterOK := afterPos >= len(data) || isWhitespace(data[afterPos]) || isDelimiter(data[afterPos])
	return beforeOK && afterOK
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isWhitespace(ch byte) bool {
	switch ch {
	case 0, 9, 10, 12, 13, 32:
		return true
	default:
		return false
	}
}

func isDelimiter(ch byte) bool {
	switch ch {
	case '(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	default:
		return false
	}
}
