package nativepdf

import (
	"bytes"
	"strconv"
	"strings"
)

type contentTokenKind int

const (
	contentTokenUnknown contentTokenKind = iota
	contentTokenOperator
	contentTokenLiteralString
	contentTokenHexString
	contentTokenArray
	contentTokenDict
	contentTokenNumber
	contentTokenName
)

type contentToken struct {
	kind    contentTokenKind
	text    string
	number  float64
	items   []contentToken
	dict    map[string]contentToken
	isValid bool
}

type textShellExtractor struct {
	glyphMap map[string]string
	lines    []string
	current  strings.Builder
	inText   bool
}

func extractContentTextByOperators(data []byte, glyphMap map[string]string) []string {
	tokenizer := newContentTokenizer(data)
	extractor := textShellExtractor{glyphMap: glyphMap}
	stack := make([]contentToken, 0, 8)

	for {
		token, ok := tokenizer.next()
		if !ok {
			break
		}
		if token.kind != contentTokenOperator {
			stack = append(stack, token)
			continue
		}
		extractor.applyOperator(token.text, stack)
		stack = stack[:0]
	}
	extractor.flush()
	return dedupeStrings(extractor.lines)
}

func (e *textShellExtractor) applyOperator(op string, operands []contentToken) {
	switch op {
	case "BT":
		e.flush()
		e.inText = true
	case "ET":
		e.flush()
		e.inText = false
	case "Tj":
		if !e.inText || len(operands) == 0 {
			return
		}
		e.appendTokenText(operands[len(operands)-1])
	case "TJ":
		if !e.inText || len(operands) == 0 {
			return
		}
		e.appendArrayText(operands[len(operands)-1])
	case "'", "\"":
		if !e.inText {
			return
		}
		e.flush()
		if len(operands) > 0 {
			e.appendTokenText(operands[len(operands)-1])
		}
	case "Td", "TD", "T*":
		if !e.inText {
			return
		}
		e.flush()
	}
}

func (e *textShellExtractor) appendTokenText(token contentToken) {
	switch token.kind {
	case contentTokenLiteralString:
		e.current.WriteString(token.text)
	case contentTokenHexString:
		e.current.WriteString(decodeHexGlyphString([]byte(token.text), e.glyphMap))
	case contentTokenArray:
		e.appendArrayText(token)
	}
}

func (e *textShellExtractor) appendArrayText(token contentToken) {
	if token.kind != contentTokenArray {
		return
	}
	for _, item := range token.items {
		switch item.kind {
		case contentTokenLiteralString:
			e.current.WriteString(item.text)
		case contentTokenHexString:
			e.current.WriteString(decodeHexGlyphString([]byte(item.text), e.glyphMap))
		case contentTokenNumber:
			if item.number < -120 && e.current.Len() > 0 {
				e.current.WriteByte(' ')
			}
		}
	}
}

func (e *textShellExtractor) flush() {
	text := strings.TrimSpace(e.current.String())
	if text != "" {
		e.lines = append(e.lines, text)
	}
	e.current.Reset()
}

type contentTokenizer struct {
	data []byte
	pos  int
}

func newContentTokenizer(data []byte) *contentTokenizer {
	return &contentTokenizer{data: data}
}

func (t *contentTokenizer) next() (contentToken, bool) {
	t.skipWS()
	if t.pos >= len(t.data) {
		return contentToken{}, false
	}

	switch t.data[t.pos] {
	case '(':
		return t.readLiteralString()
	case '<':
		if t.pos+1 < len(t.data) && t.data[t.pos+1] == '<' {
			return t.readDict()
		}
		return t.readHexString()
	case '[':
		return t.readArray()
	case '/':
		return t.readName()
	default:
		if isNumberStart(t.data[t.pos]) {
			if token, ok := t.readNumber(); ok {
				return token, true
			}
		}
		return t.readOperatorishToken()
	}
}

func (t *contentTokenizer) readLiteralString() (contentToken, bool) {
	if t.data[t.pos] != '(' {
		return contentToken{}, false
	}
	t.pos++
	var builder strings.Builder
	depth := 1
	for t.pos < len(t.data) {
		ch := t.data[t.pos]
		if ch == '\\' {
			if t.pos+1 >= len(t.data) {
				break
			}
			next := t.data[t.pos+1]
			switch next {
			case 'n':
				builder.WriteByte('\n')
			case 'r':
				builder.WriteByte('\r')
			case 't':
				builder.WriteByte('\t')
			case 'b':
				builder.WriteByte('\b')
			case 'f':
				builder.WriteByte('\f')
			case '(', ')', '\\':
				builder.WriteByte(next)
			default:
				builder.WriteByte(next)
			}
			t.pos += 2
			continue
		}
		switch ch {
		case '(':
			depth++
			builder.WriteByte(ch)
		case ')':
			depth--
			if depth == 0 {
				t.pos++
				return contentToken{kind: contentTokenLiteralString, text: builder.String(), isValid: true}, true
			}
			builder.WriteByte(ch)
		default:
			builder.WriteByte(ch)
		}
		t.pos++
	}
	return contentToken{}, false
}

func (t *contentTokenizer) readHexString() (contentToken, bool) {
	if t.data[t.pos] != '<' {
		return contentToken{}, false
	}
	t.pos++
	start := t.pos
	for t.pos < len(t.data) && t.data[t.pos] != '>' {
		t.pos++
	}
	if t.pos >= len(t.data) {
		return contentToken{}, false
	}
	text := string(bytes.TrimSpace(t.data[start:t.pos]))
	t.pos++
	return contentToken{kind: contentTokenHexString, text: text, isValid: true}, true
}

func (t *contentTokenizer) readArray() (contentToken, bool) {
	if t.data[t.pos] != '[' {
		return contentToken{}, false
	}
	t.pos++
	items := make([]contentToken, 0, 8)
	for {
		t.skipWS()
		if t.pos >= len(t.data) {
			return contentToken{}, false
		}
		if t.data[t.pos] == ']' {
			t.pos++
			return contentToken{kind: contentTokenArray, items: items, isValid: true}, true
		}
		item, ok := t.next()
		if !ok {
			return contentToken{}, false
		}
		items = append(items, item)
	}
}

func (t *contentTokenizer) readDict() (contentToken, bool) {
	if t.pos+1 >= len(t.data) || t.data[t.pos] != '<' || t.data[t.pos+1] != '<' {
		return contentToken{}, false
	}
	t.pos += 2
	dict := make(map[string]contentToken)
	for {
		t.skipWS()
		if t.pos+1 < len(t.data) && t.data[t.pos] == '>' && t.data[t.pos+1] == '>' {
			t.pos += 2
			return contentToken{kind: contentTokenDict, dict: dict, isValid: true}, true
		}
		key, ok := t.readName()
		if !ok {
			return contentToken{}, false
		}
		value, ok := t.next()
		if !ok {
			return contentToken{}, false
		}
		dict[key.text] = value
	}
}

func (t *contentTokenizer) readName() (contentToken, bool) {
	if t.data[t.pos] != '/' {
		return contentToken{}, false
	}
	start := t.pos + 1
	t.pos++
	for t.pos < len(t.data) && !isWhitespace(t.data[t.pos]) && !isDelimiter(t.data[t.pos]) {
		t.pos++
	}
	return contentToken{kind: contentTokenName, text: string(t.data[start:t.pos]), isValid: true}, true
}

func (t *contentTokenizer) readNumber() (contentToken, bool) {
	start := t.pos
	if t.data[t.pos] == '+' || t.data[t.pos] == '-' {
		t.pos++
	}
	dots := 0
	for t.pos < len(t.data) {
		ch := t.data[t.pos]
		if ch == '.' {
			dots++
			if dots > 1 {
				return contentToken{}, false
			}
			t.pos++
			continue
		}
		if !isDigit(ch) {
			break
		}
		t.pos++
	}
	value, err := strconv.ParseFloat(string(t.data[start:t.pos]), 64)
	if err != nil {
		return contentToken{}, false
	}
	return contentToken{kind: contentTokenNumber, number: value, isValid: true}, true
}

func (t *contentTokenizer) readOperatorishToken() (contentToken, bool) {
	start := t.pos
	for t.pos < len(t.data) && !isWhitespace(t.data[t.pos]) && !isDelimiter(t.data[t.pos]) {
		t.pos++
	}
	if t.pos == start {
		t.pos++
		return contentToken{}, false
	}
	return contentToken{kind: contentTokenOperator, text: string(t.data[start:t.pos]), isValid: true}, true
}

func (t *contentTokenizer) skipWS() {
	for t.pos < len(t.data) {
		if isWhitespace(t.data[t.pos]) {
			t.pos++
			continue
		}
		if t.data[t.pos] == '%' {
			for t.pos < len(t.data) && t.data[t.pos] != '\n' && t.data[t.pos] != '\r' {
				t.pos++
			}
			continue
		}
		break
	}
}

func isNumberStart(ch byte) bool {
	return isDigit(ch) || ch == '+' || ch == '-' || ch == '.'
}
