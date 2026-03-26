// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// This package provides functionality equivalent to Apache PDFBox 3.0.4
// (https://pdfbox.apache.org/), implemented using pdfcpu.

package extractor

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

type pdfMatrix struct {
	A float64
	B float64
	C float64
	D float64
	E float64
	F float64
}

func identMatrix() pdfMatrix {
	return pdfMatrix{A: 1, D: 1}
}

func (m pdfMatrix) multiply(n pdfMatrix) pdfMatrix {
	return pdfMatrix{
		A: m.A*n.A + m.B*n.C,
		B: m.A*n.B + m.B*n.D,
		C: m.C*n.A + m.D*n.C,
		D: m.C*n.B + m.D*n.D,
		E: m.E*n.A + m.F*n.C + n.E,
		F: m.E*n.B + m.F*n.D + n.F,
	}
}

func (m pdfMatrix) transform(x, y float64) (float64, float64) {
	return x*m.A + y*m.C + m.E, x*m.B + y*m.D + m.F
}

func (m pdfMatrix) translate(tx, ty float64) pdfMatrix {
	return m.multiply(pdfMatrix{A: 1, D: 1, E: tx, F: ty})
}

type graphicsState struct {
	ctm         pdfMatrix
	lineWidth   float64
	fillColor   [3]float64
	strokeColor [3]float64
}

func defaultGraphicsState() graphicsState {
	return graphicsState{
		ctm:       identMatrix(),
		lineWidth: 1.0,
		fillColor: [3]float64{0, 0, 0},
	}
}

type textState struct {
	inText      bool
	fontName    string
	fontSize    float64
	leading     float64
	textMatrix  pdfMatrix
	lineMatrix  pdfMatrix
	fillColor   [3]float64
	charSpacing float64
	wordSpacing float64
}

type streamToken struct {
	kind  string
	value string
	items []streamToken
}

type streamScanner struct {
	data []byte
	pos  int
}

const tjSpaceThreshold = -100.0

func extractPageContentBytes(doc *model.PDDocument, pageIdx int) ([]byte, error) {
	if doc == nil || doc.Context == nil {
		return nil, fmt.Errorf("pdf document is not open")
	}
	if pageIdx < 0 || pageIdx >= doc.PageCount() {
		return nil, fmt.Errorf("page index out of range: %d", pageIdx)
	}

	r, err := pdfcpu.ExtractPageContent(doc.Context, pageIdx+1)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return io.ReadAll(r)
}

func tokenizeContent(data []byte) ([]streamToken, error) {
	s := &streamScanner{data: data}
	var tokens []streamToken
	for {
		tok, ok, err := s.nextToken()
		if err != nil {
			return nil, err
		}
		if !ok {
			return tokens, nil
		}
		tokens = append(tokens, tok)
	}
}

func (s *streamScanner) nextToken() (streamToken, bool, error) {
	s.skipWS()
	if s.pos >= len(s.data) {
		return streamToken{}, false, nil
	}

	switch ch := s.data[s.pos]; ch {
	case '(':
		v, err := s.readLiteralString()
		return streamToken{kind: "string", value: v}, true, err
	case '<':
		if s.peekString("<<") {
			s.pos += 2
			return streamToken{kind: "word", value: "<<"}, true, nil
		}
		v, err := s.readHexString()
		return streamToken{kind: "hex", value: v}, true, err
	case '>':
		if s.peekString(">>") {
			s.pos += 2
			return streamToken{kind: "word", value: ">>"}, true, nil
		}
	case '[':
		items, err := s.readArray()
		return streamToken{kind: "array", items: items}, true, err
	case '/':
		return streamToken{kind: "name", value: s.readName()}, true, nil
	}

	word := s.readWord()
	if _, err := strconv.ParseFloat(word, 64); err == nil {
		return streamToken{kind: "number", value: word}, true, nil
	}
	return streamToken{kind: "word", value: word}, true, nil
}

func (s *streamScanner) readArray() ([]streamToken, error) {
	s.pos++
	var items []streamToken
	for {
		s.skipWS()
		if s.pos >= len(s.data) {
			return items, nil
		}
		if s.data[s.pos] == ']' {
			s.pos++
			return items, nil
		}
		tok, ok, err := s.nextToken()
		if err != nil {
			return nil, err
		}
		if !ok {
			return items, nil
		}
		items = append(items, tok)
	}
}

func (s *streamScanner) readLiteralString() (string, error) {
	s.pos++
	var buf bytes.Buffer
	depth := 1
	for s.pos < len(s.data) {
		ch := s.data[s.pos]
		s.pos++
		if ch == '\\' {
			if s.pos >= len(s.data) {
				break
			}
			esc := s.data[s.pos]
			s.pos++
			switch esc {
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case '(', ')', '\\':
				buf.WriteByte(esc)
			case '\n':
			case '\r':
				if s.pos < len(s.data) && s.data[s.pos] == '\n' {
					s.pos++
				}
			default:
				if esc >= '0' && esc <= '7' {
					oct := []byte{esc}
					for len(oct) < 3 && s.pos < len(s.data) {
						b := s.data[s.pos]
						if b < '0' || b > '7' {
							break
						}
						oct = append(oct, b)
						s.pos++
					}
					v, err := strconv.ParseInt(string(oct), 8, 32)
					if err != nil {
						return "", err
					}
					buf.WriteByte(byte(v))
				} else {
					buf.WriteByte(esc)
				}
			}
			continue
		}
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
			if depth == 0 {
				return normalizePDFString(buf.Bytes()), nil
			}
		}
		buf.WriteByte(ch)
	}
	return normalizePDFString(buf.Bytes()), nil
}

func (s *streamScanner) readHexString() (string, error) {
	s.pos++
	start := s.pos
	for s.pos < len(s.data) && s.data[s.pos] != '>' {
		s.pos++
	}
	raw := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r', '\f', '\v':
			return -1
		default:
			return r
		}
	}, string(s.data[start:s.pos]))
	if len(raw)%2 == 1 {
		raw += "0"
	}
	if s.pos < len(s.data) {
		s.pos++
	}
	if raw == "" {
		return "", nil
	}
	b, err := hex.DecodeString(raw)
	if err != nil {
		return "", err
	}
	return normalizePDFString(b), nil
}

func (s *streamScanner) readName() string {
	start := s.pos
	s.pos++
	for s.pos < len(s.data) && !isDelimiter(s.data[s.pos]) {
		s.pos++
	}
	return string(s.data[start:s.pos])
}

func (s *streamScanner) readWord() string {
	start := s.pos
	for s.pos < len(s.data) && !isDelimiter(s.data[s.pos]) {
		s.pos++
	}
	return string(s.data[start:s.pos])
}

func (s *streamScanner) skipWS() {
	for s.pos < len(s.data) {
		if s.data[s.pos] == '%' {
			for s.pos < len(s.data) && s.data[s.pos] != '\n' && s.data[s.pos] != '\r' {
				s.pos++
			}
			continue
		}
		if !isWhiteSpace(s.data[s.pos]) {
			return
		}
		s.pos++
	}
}

func (s *streamScanner) peekString(v string) bool {
	return s.pos+len(v) <= len(s.data) && string(s.data[s.pos:s.pos+len(v)]) == v
}

func isWhiteSpace(ch byte) bool {
	switch ch {
	case 0x00, '\t', '\n', '\f', '\r', ' ':
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
		return isWhiteSpace(ch)
	}
}

func normalizePDFString(b []byte) string {
	if len(b) >= 2 {
		// PDF UTF-16 strings are BOM-prefixed.
		if b[0] == 0xFE && b[1] == 0xFF {
			runes := make([]rune, 0, (len(b)-2)/2)
			for i := 2; i+1 < len(b); i += 2 {
				runes = append(runes, rune(uint16(b[i])<<8|uint16(b[i+1])))
			}
			return string(runes)
		}
		if b[0] == 0xFF && b[1] == 0xFE {
			runes := make([]rune, 0, (len(b)-2)/2)
			for i := 2; i+1 < len(b); i += 2 {
				runes = append(runes, rune(uint16(b[i+1])<<8|uint16(b[i])))
			}
			return string(runes)
		}
	}
	// Most 8-bit PDF text content uses WinAnsiEncoding.
	return decodeWinAnsi(b)
}

var win1252Extras = map[byte]rune{
	0x80: '\u20AC',
	0x82: '\u201A',
	0x83: '\u0192',
	0x84: '\u201E',
	0x85: '\u2026',
	0x86: '\u2020',
	0x87: '\u2021',
	0x88: '\u02C6',
	0x89: '\u2030',
	0x8A: '\u0160',
	0x8B: '\u2039',
	0x8C: '\u0152',
	0x8E: '\u017D',
	0x91: '\u2018',
	0x92: '\u2019',
	0x93: '\u201C',
	0x94: '\u201D',
	0x95: '\u2022',
	0x96: '\u2013',
	0x97: '\u2014',
	0x98: '\u02DC',
	0x99: '\u2122',
	0x9A: '\u0161',
	0x9B: '\u203A',
	0x9C: '\u0153',
	0x9E: '\u017E',
	0x9F: '\u0178',
}

func decodeWinAnsi(b []byte) string {
	allASCII := true
	for _, ch := range b {
		if ch > 0x7F {
			allASCII = false
			break
		}
		if _, ok := ligatureExtras[ch]; ok {
			allASCII = false
			break
		}
	}
	if allASCII {
		return string(b)
	}

	var sb strings.Builder
	for idx, ch := range b {
		if ligature, ok := decodeLigatureByte(b, idx); ok {
			sb.WriteString(ligature)
			continue
		}
		if ch < 0x80 {
			sb.WriteRune(rune(ch))
			continue
		}
		if r, ok := win1252Extras[ch]; ok {
			sb.WriteRune(r)
			continue
		}
		sb.WriteRune(rune(ch))
	}
	return sb.String()
}

var ligatureExtras = map[byte]string{
	0x01: "ff",
	0x02: "fi",
	0x03: "fl",
	0x04: "ffi",
	0x05: "ffl",
}

func decodeLigatureByte(b []byte, idx int) (string, bool) {
	ch := b[idx]
	ligature, ok := ligatureExtras[ch]
	if !ok {
		return "", false
	}

	prevIsLetter := idx > 0 && isASCIILetter(b[idx-1])
	nextIsLetter := idx+1 < len(b) && isASCIILetter(b[idx+1])
	if prevIsLetter && nextIsLetter {
		return ligature, true
	}
	return "", false
}

func isASCIILetter(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
}

func parseFloatToken(tok streamToken) (float64, bool) {
	if tok.kind != "word" && tok.kind != "number" {
		return 0, false
	}
	v, err := strconv.ParseFloat(tok.value, 64)
	return v, err == nil
}

func parseFloats(tokens []streamToken) ([]float64, bool) {
	values := make([]float64, 0, len(tokens))
	for _, tok := range tokens {
		v, ok := parseFloatToken(tok)
		if !ok {
			return nil, false
		}
		values = append(values, v)
	}
	return values, true
}

func matrixFromOperands(tokens []streamToken) (pdfMatrix, bool) {
	values, ok := parseFloats(tokens)
	if !ok || len(values) != 6 {
		return pdfMatrix{}, false
	}
	return pdfMatrix{
		A: values[0],
		B: values[1],
		C: values[2],
		D: values[3],
		E: values[4],
		F: values[5],
	}, true
}

func colorFromOperands(tokens []streamToken, grayscale bool) ([3]float64, bool) {
	var rgb [3]float64
	values, ok := parseFloats(tokens)
	if !ok {
		return rgb, false
	}
	if grayscale {
		if len(values) != 1 {
			return rgb, false
		}
		rgb = [3]float64{values[0], values[0], values[0]}
		return clampColor(rgb), true
	}
	if len(values) == 3 {
		rgb = [3]float64{values[0], values[1], values[2]}
		return clampColor(rgb), true
	}
	if len(values) == 4 {
		c, m, y, k := values[0], values[1], values[2], values[3]
		rgb = [3]float64{
			1 - math.Min(1, c+k),
			1 - math.Min(1, m+k),
			1 - math.Min(1, y+k),
		}
		return clampColor(rgb), true
	}
	return rgb, false
}

func clampColor(c [3]float64) [3]float64 {
	for i := range c {
		if c[i] < 0 {
			c[i] = 0
		}
		if c[i] > 1 {
			c[i] = 1
		}
	}
	return c
}

func textWidthEstimate(text string, fontSize float64) float64 {
	if text == "" || fontSize <= 0 {
		return 0
	}
	return float64(len([]rune(text))) * fontSize * 0.5
}

func decodeTJText(tok streamToken) string {
	if tok.kind != "array" {
		return ""
	}
	var sb strings.Builder
	for _, item := range tok.items {
		if item.kind == "string" || item.kind == "hex" {
			sb.WriteString(item.value)
		} else if item.kind == "number" {
			if v, ok := parseFloatToken(item); ok && v <= tjSpaceThreshold {
				sb.WriteByte(' ')
			}
		}
	}
	return normalizeExtractedText(sb.String())
}

func normalizeExtractedText(text string) string {
	if text == "" {
		return text
	}
	b := []byte(text)
	var sb strings.Builder
	for idx, ch := range b {
		if ligature, ok := decodeLigatureByte(b, idx); ok {
			sb.WriteString(ligature)
			continue
		}
		sb.WriteByte(ch)
	}
	return sb.String()
}
