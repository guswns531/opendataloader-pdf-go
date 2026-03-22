package nativepdf

import (
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

type textFragment struct {
	text            string
	bounds          model.Box
	fontName        string
	fontSize        float64
	markedContentID *int
}

type textState struct {
	inText              bool
	fontAliases         map[string]string
	fontMetrics         map[string]fontMetrics
	fontName            string
	fontSize            float64
	leading             float64
	cursorX             float64
	cursorY             float64
	lineOriginX         float64
	lineOriginY         float64
	fontAlias           string
	metrics             fontMetrics
	matrix              matrix2D
	stack               []textSnapshot
	sequence            int
	markedContentScopes []markedContentScope
}

type textSnapshot struct {
	matrix      matrix2D
	cursorX     float64
	cursorY     float64
	lineOriginX float64
	lineOriginY float64
	fontAlias   string
	fontName    string
	fontSize    float64
	leading     float64
	metrics     fontMetrics
}

func extractPositionedTextArtifacts(data []byte, pages []model.PageMetadata) [][]*model.RawArtifact {
	if len(pages) == 0 {
		return nil
	}

	glyphMap := extractToUnicodeMap(data)
	decodedStreams := extractDecodedStreams(data)
	if len(decodedStreams) == 0 {
		return nil
	}

	fragmentsByStream := make([][]textFragment, 0, len(decodedStreams))
	for _, decoded := range decodedStreams {
		fragments := extractTextFragmentsFromStream(decoded, glyphMap)
		if len(fragments) > 0 {
			fragmentsByStream = append(fragmentsByStream, fragments)
		}
	}
	if len(fragmentsByStream) == 0 {
		return nil
	}

	out := make([][]*model.RawArtifact, len(pages))
	chunks := distributeFragmentsAcrossPages(fragmentsByStream, len(pages))
	for pageIndex, fragments := range chunks {
		if len(fragments) == 0 {
			continue
		}
		pageMeta := pages[pageIndex]
		pageBounds, pageBoundsOK := pageBoxFromMetadata(pageMeta)
		artifacts := make([]*model.RawArtifact, 0, len(fragments))
		for i, fragment := range fragments {
			text := strings.TrimSpace(fragment.text)
			if text == "" {
				continue
			}
			bounds := fragment.bounds
			if bounds.IsZero() || (pageBoundsOK && !bounds.Intersects(pageBounds)) {
				bounds = shellBoundsForText(pageMeta, i, text)
			}
			artifacts = append(artifacts, &model.RawArtifact{
				ID:              model.ArtifactID(i + 1),
				Kind:            model.ArtifactKindText,
				PageIndex:       pageMeta.Index,
				PageNumber:      pageMeta.Number,
				MarkedContentID: cloneIntPointer(fragment.markedContentID),
				Sequence:        i,
				Bounds:          bounds,
				Text:            text,
				Style: model.TextProperties{
					Font:     fragment.fontName,
					FontSize: fragment.fontSize,
					Content:  text,
				},
			})
		}
		out[pageIndex] = artifacts
	}
	return out
}

func extractTextFragmentsFromStreams(streams [][]byte, glyphMap map[string]string) []textFragment {
	return extractTextFragmentsFromStreamsWithFonts(streams, glyphMap, nil, nil)
}

func extractTextFragmentsFromStreamsWithFonts(streams [][]byte, glyphMap map[string]string, fontAliases map[string]string, fontMetrics map[string]fontMetrics) []textFragment {
	if len(streams) == 0 {
		return nil
	}
	fragments := make([]textFragment, 0, len(streams))
	for _, stream := range streams {
		fragments = append(fragments, extractTextFragmentsFromStreamWithFonts(stream, glyphMap, fontAliases, fontMetrics)...)
	}
	if len(fragments) == 0 {
		return nil
	}
	return fragments
}

func pageBoxFromMetadata(page model.PageMetadata) (model.Box, bool) {
	if !page.Bounds.IsZero() {
		return page.Bounds.Normalize(), true
	}
	if page.Size.Width > 0 || page.Size.Height > 0 {
		return model.Box{Left: 0, Bottom: 0, Right: page.Size.Width, Top: page.Size.Height}, true
	}
	return model.Box{}, false
}

func extractTextFragmentsFromStream(data []byte, glyphMap map[string]string) []textFragment {
	return extractTextFragmentsFromStreamWithFonts(data, glyphMap, nil, nil)
}

func extractTextFragmentsFromStreamWithFonts(data []byte, glyphMap map[string]string, fontAliases map[string]string, fontMetrics map[string]fontMetrics) []textFragment {
	tokenizer := newContentTokenizer(data)
	state := textState{
		fontAliases: fontAliases,
		fontMetrics: fontMetrics,
		matrix:      identityMatrix(),
		fontSize:    12,
	}

	fragments := make([]textFragment, 0, 16)
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
		switch token.text {
		case "BT":
			state.inText = true
			state.cursorX = 0
			state.cursorY = 0
			state.lineOriginX = 0
			state.lineOriginY = 0
			state.sequence = 0
		case "ET":
			state.inText = false
			stack = stack[:0]
		default:
			fragments = append(fragments, state.applyOperator(token.text, stack, glyphMap)...)
		}
		stack = stack[:0]
	}
	return fragments
}

func (s *textState) applyOperator(op string, operands []contentToken, glyphMap map[string]string) []textFragment {
	switch op {
	case "q":
		s.pushState()
		return nil
	case "Q":
		s.popState()
		return nil
	case "cm":
		s.applyTransform(operands)
		return nil
	case "Tf":
		s.applyTf(operands)
		return nil
	case "BMC":
		s.beginMarkedContent(operands)
		return nil
	case "BDC":
		s.beginMarkedContent(operands)
		return nil
	case "EMC":
		s.endMarkedContent()
		return nil
	case "Tm":
		s.applyTm(operands)
		return nil
	case "Td":
		s.applyTd(operands)
		return nil
	case "TD":
		s.applyTD(operands)
		return nil
	case "T*":
		s.applyTStar()
		return nil
	case "Tj":
		if !s.inText || len(operands) == 0 {
			return nil
		}
		return s.emitText(operands[len(operands)-1], glyphMap)
	case "TJ":
		if !s.inText || len(operands) == 0 {
			return nil
		}
		return s.emitText(operands[len(operands)-1], glyphMap)
	case "'", "\"":
		if !s.inText {
			return nil
		}
		s.applyTStar()
		if len(operands) > 0 {
			return s.emitText(operands[len(operands)-1], glyphMap)
		}
		return nil
	default:
		return nil
	}
}

func (s *textState) applyTf(operands []contentToken) {
	if len(operands) < 2 {
		return
	}
	if name, ok := operands[len(operands)-2].asName(); ok {
		s.fontAlias = strings.TrimPrefix(name, "/")
		s.fontName = s.fontAlias
		if resolved := s.resolveFontName(name); resolved != "" {
			s.fontName = resolved
		}
		s.metrics = s.resolveFontMetrics(s.fontAlias)
	}
	if size, ok := operands[len(operands)-1].asNumber(); ok {
		s.fontSize = math.Max(1, size)
	}
}

func (s *textState) applyTm(operands []contentToken) {
	if len(operands) < 6 {
		return
	}
	if x, ok := operands[len(operands)-2].asNumber(); ok {
		s.cursorX = x
		s.lineOriginX = x
	}
	if y, ok := operands[len(operands)-1].asNumber(); ok {
		s.cursorY = y
		s.lineOriginY = y
	}
}

func (s *textState) applyTransform(operands []contentToken) {
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

func (s *textState) pushState() {
	s.stack = append(s.stack, textSnapshot{
		matrix:      s.matrix,
		cursorX:     s.cursorX,
		cursorY:     s.cursorY,
		lineOriginX: s.lineOriginX,
		lineOriginY: s.lineOriginY,
		fontAlias:   s.fontAlias,
		fontName:    s.fontName,
		fontSize:    s.fontSize,
		leading:     s.leading,
		metrics:     s.metrics,
	})
}

func (s *textState) popState() {
	if len(s.stack) == 0 {
		return
	}
	last := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	s.matrix = last.matrix
	s.cursorX = last.cursorX
	s.cursorY = last.cursorY
	s.lineOriginX = last.lineOriginX
	s.lineOriginY = last.lineOriginY
	s.fontAlias = last.fontAlias
	s.fontName = last.fontName
	s.fontSize = last.fontSize
	s.leading = last.leading
	s.metrics = last.metrics
}

func (s *textState) resolveFontName(alias string) string {
	alias = strings.TrimPrefix(alias, "/")
	if s == nil || len(s.fontAliases) == 0 {
		return alias
	}
	if resolved, ok := s.fontAliases[alias]; ok && resolved != "" {
		return resolved
	}
	return alias
}

func (s *textState) resolveFontMetrics(alias string) fontMetrics {
	alias = strings.TrimPrefix(alias, "/")
	if s == nil || len(s.fontMetrics) == 0 {
		return fontMetrics{}
	}
	return s.fontMetrics[alias]
}

func (s *textState) applyTd(operands []contentToken) {
	if len(operands) < 2 {
		return
	}
	if tx, ok := operands[len(operands)-2].asNumber(); ok {
		s.cursorX += tx
		s.lineOriginX = s.cursorX
	}
	if ty, ok := operands[len(operands)-1].asNumber(); ok {
		s.cursorY += ty
		s.lineOriginY = s.cursorY
	}
}

func (s *textState) applyTD(operands []contentToken) {
	if len(operands) < 2 {
		return
	}
	if ty, ok := operands[len(operands)-1].asNumber(); ok {
		s.leading = -ty
	}
	s.applyTd(operands)
}

func (s *textState) applyTStar() {
	leading := s.leading
	if leading == 0 {
		leading = s.fontSize * 1.2
	}
	s.cursorX = s.lineOriginX
	s.cursorY -= leading
}

func (s *textState) emitText(token contentToken, glyphMap map[string]string) []textFragment {
	text := token.asText(glyphMap)
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	width := estimateTokenAdvance(token, s.fontSize, s.fontName, glyphMap, s.metrics)
	if width <= 0 {
		width = estimateTextWidth(text, s.fontSize, s.fontName)
	}
	height := math.Max(1, s.fontSize)
	localBounds := model.Box{
		Left:   s.cursorX,
		Bottom: s.cursorY,
		Right:  s.cursorX + width,
		Top:    s.cursorY + height,
	}
	bounds := s.matrix.transformBox(localBounds)
	s.cursorX += width
	s.sequence++
	return []textFragment{{
		text:            text,
		bounds:          bounds,
		fontName:        s.fontName,
		fontSize:        s.fontSize,
		markedContentID: cloneIntPointer(s.currentMarkedContentID()),
	}}
}

func (s *textState) beginMarkedContent(operands []contentToken) {
	s.markedContentScopes = append(s.markedContentScopes, markedContentScopeFromOperands(operands))
}

func (s *textState) endMarkedContent() {
	if len(s.markedContentScopes) == 0 {
		return
	}
	s.markedContentScopes = s.markedContentScopes[:len(s.markedContentScopes)-1]
}

func (s *textState) currentMarkedContentID() *int {
	for i := len(s.markedContentScopes) - 1; i >= 0; i-- {
		if s.markedContentScopes[i].mcid != nil {
			return s.markedContentScopes[i].mcid
		}
	}
	return nil
}

func estimateTokenAdvance(token contentToken, fontSize float64, fontName string, glyphMap map[string]string, metrics fontMetrics) float64 {
	switch token.kind {
	case contentTokenLiteralString:
		if width := metrics.advanceForLiteralString(token.text, fontSize); width > 0 {
			return width
		}
		return estimateTextWidth(token.text, fontSize, fontName)
	case contentTokenHexString:
		if width := metrics.advanceForHexString(token.text, fontSize); width > 0 {
			return width
		}
		return estimateTextWidth(decodeHexGlyphString([]byte(token.text), glyphMap), fontSize, fontName)
	case contentTokenArray:
		var total float64
		for _, item := range token.items {
			switch item.kind {
			case contentTokenLiteralString:
				if width := metrics.advanceForLiteralString(item.text, fontSize); width > 0 {
					total += width
					continue
				}
				total += estimateTextWidth(item.text, fontSize, fontName)
			case contentTokenHexString:
				if width := metrics.advanceForHexString(item.text, fontSize); width > 0 {
					total += width
					continue
				}
				total += estimateTextWidth(decodeHexGlyphString([]byte(item.text), glyphMap), fontSize, fontName)
			case contentTokenNumber:
				total += estimateKerningAdjustment(item.number, fontSize)
			}
		}
		return total
	default:
		return estimateTextWidth(token.text, fontSize, fontName)
	}
}

func estimateTextWidth(text string, fontSize float64, fontName string) float64 {
	if fontSize <= 0 {
		fontSize = 12
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}

	profile := fontAdvanceProfileForName(fontName)
	var total float64
	for _, r := range text {
		total += estimateRuneAdvance(r, profile)
	}
	if total <= 0 {
		total = float64(len([]rune(text))) * 0.55
	}
	return math.Max(fontSize*0.35, total*fontSize)
}

type fontAdvanceProfile struct {
	monospace bool
	serif     bool
	sans      bool
}

func fontAdvanceProfileForName(fontName string) fontAdvanceProfile {
	name := strings.ToLower(strings.TrimSpace(fontName))
	if name == "" {
		return fontAdvanceProfile{}
	}
	switch {
	case strings.Contains(name, "courier"), strings.Contains(name, "mono"), strings.Contains(name, "code"), strings.Contains(name, "terminal"):
		return fontAdvanceProfile{monospace: true}
	case strings.Contains(name, "times"), strings.Contains(name, "serif"):
		return fontAdvanceProfile{serif: true}
	case strings.Contains(name, "helvetica"), strings.Contains(name, "arial"), strings.Contains(name, "sans"):
		return fontAdvanceProfile{sans: true}
	default:
		return fontAdvanceProfile{}
	}
}

func estimateRuneAdvance(r rune, profile fontAdvanceProfile) float64 {
	switch {
	case unicode.IsSpace(r):
		if profile.monospace {
			return 0.6
		}
		return 0.28
	case unicode.In(r, unicode.Han, unicode.Hangul, unicode.Hiragana, unicode.Katakana):
		return 1.0
	case unicode.IsDigit(r):
		if profile.monospace {
			return 0.6
		}
		return 0.52
	case unicode.IsUpper(r):
		if profile.monospace {
			return 0.6
		}
		if profile.serif {
			return 0.63
		}
		return 0.65
	case unicode.IsLower(r):
		if profile.monospace {
			return 0.6
		}
		if profile.serif {
			return 0.48
		}
		return 0.54
	case strings.ContainsRune(",.;:'`", r):
		return 0.25
	case strings.ContainsRune("!|/\\", r):
		return 0.3
	case strings.ContainsRune("-_()[]{}<>", r):
		return 0.35
	default:
		if profile.monospace {
			return 0.6
		}
		return 0.55
	}
}

func estimateKerningAdjustment(value float64, fontSize float64) float64 {
	if fontSize <= 0 {
		fontSize = 12
	}
	return (-value * fontSize) / 1000.0
}

func distributeFragmentsAcrossPages(fragmentsByStream [][]textFragment, pageCount int) [][]textFragment {
	out := make([][]textFragment, pageCount)
	if pageCount == 0 || len(fragmentsByStream) == 0 {
		return out
	}
	for i, fragments := range fragmentsByStream {
		pageIndex := i * pageCount / len(fragmentsByStream)
		if pageIndex >= pageCount {
			pageIndex = pageCount - 1
		}
		out[pageIndex] = append(out[pageIndex], fragments...)
	}
	return out
}

func (t contentToken) asName() (string, bool) {
	switch value := t.text; t.kind {
	case contentTokenName, contentTokenOperator:
		if value == "" {
			return "", false
		}
		return value, true
	default:
		return "", false
	}
}

func (t contentToken) asNumber() (float64, bool) {
	switch t.kind {
	case contentTokenNumber:
		return t.number, true
	case contentTokenLiteralString, contentTokenHexString, contentTokenOperator:
		v, err := strconv.ParseFloat(strings.TrimSpace(t.text), 64)
		if err == nil {
			return v, true
		}
	}
	return 0, false
}

func (t contentToken) asText(glyphMap map[string]string) string {
	switch t.kind {
	case contentTokenLiteralString:
		return t.text
	case contentTokenHexString:
		return decodeHexGlyphString([]byte(t.text), glyphMap)
	case contentTokenArray:
		var builder strings.Builder
		for _, item := range t.items {
			switch item.kind {
			case contentTokenLiteralString:
				builder.WriteString(item.text)
			case contentTokenHexString:
				builder.WriteString(decodeHexGlyphString([]byte(item.text), glyphMap))
			case contentTokenNumber:
				if item.number < -120 && builder.Len() > 0 {
					builder.WriteByte(' ')
				}
			}
		}
		return builder.String()
	default:
		return ""
	}
}

func (t contentToken) asDict() (map[string]contentToken, bool) {
	if t.kind != contentTokenDict || len(t.dict) == 0 {
		return nil, false
	}
	return t.dict, true
}

type markedContentScope struct {
	tag  string
	mcid *int
}

func markedContentScopeFromOperands(operands []contentToken) markedContentScope {
	var scope markedContentScope
	if len(operands) == 0 {
		return scope
	}
	if len(operands) >= 2 {
		scope.tag, _ = operands[len(operands)-2].asName()
		scope.mcid = markedContentIDFromToken(operands[len(operands)-1])
		return scope
	}
	scope.tag, _ = operands[len(operands)-1].asName()
	return scope
}

func markedContentIDFromToken(token contentToken) *int {
	if dict, ok := token.asDict(); ok {
		if value, ok := dict["MCID"]; ok {
			if mcid, ok := value.asNumber(); ok {
				parsed := int(mcid)
				return &parsed
			}
		}
	}
	return nil
}

func cloneIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
