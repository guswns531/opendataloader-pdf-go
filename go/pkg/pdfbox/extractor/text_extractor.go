// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// This package provides functionality equivalent to Apache PDFBox 3.0.4
// (https://pdfbox.apache.org/), implemented using pdfcpu.

package extractor

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/model"
)

type ExtractedText struct {
	Text     string
	X        float64
	Y        float64
	Width    float64
	Height   float64
	FontName string
	FontSize float64
	Bold     bool
	Italic   bool
	Color    [3]float64
	Baseline float64
	Page     int
}

func ExtractTextChunks(doc *model.PDDocument, pageIdx int) ([]*ExtractedText, error) {
	if _, err := doc.GetPage(pageIdx); err != nil {
		return nil, err
	}
	fontDecoders, err := loadPageFontDecoders(doc, pageIdx)
	if err != nil {
		return nil, err
	}

	content, err := extractPageContentBytes(doc, pageIdx)
	if err != nil {
		return nil, err
	}

	tokens, err := tokenizeContent(content)
	if err != nil {
		return nil, err
	}

	gs := defaultGraphicsState()
	ts := textState{
		textMatrix: identMatrix(),
		lineMatrix: identMatrix(),
		fillColor:  gs.fillColor,
	}
	var stack []graphicsState
	var operands []streamToken
	var out []*ExtractedText

	var appendTextToken func(string, []byte, *float64)
	appendTextToken = func(text string, raw []byte, advanceOverride *float64) {
		text = normalizeExtractedText(text)
		if text == "" || !ts.inText {
			return
		}
		x, y := gs.ctm.transform(ts.textMatrix.E, ts.textMatrix.F)
		if last := lastExtractedTextOnBaseline(out, pageIdx, y); last != nil {
			text = foldLeadingBoundaryWhitespace(last, text)
		}
		// Effective font size = Tf size × text matrix horizontal scale.
		// PDF stores the actual rendered size via the text matrix (e.g. "10 0 0 10 x y Tm"
		// with "/F1 1 Tf" means rendered size = 1 × 10 = 10pt). Without this, all
		// chunks appear 1pt tall and get filtered as tiny text.
		effectiveSize := ts.fontSize * math.Sqrt(ts.textMatrix.A*ts.textMatrix.A+ts.textMatrix.B*ts.textMatrix.B)
		if effectiveSize <= 0 {
			effectiveSize = ts.fontSize
		}
		advance := fallbackTextAdvance(text, ts.fontSize)
		if advanceOverride != nil {
			advance = *advanceOverride
		} else if decoder := fontDecoders[ts.fontName]; decoder != nil {
			advance = decoder.textAdvance(raw, text, ts.fontSize, ts.charSpacing, ts.wordSpacing)
		}
		if strings.TrimSpace(text) == "" {
			if last := lastExtractedTextOnBaseline(out, pageIdx, y); last != nil {
				appendBoundaryWhitespace(last, text)
			}
			ts.textMatrix = ts.textMatrix.translate(advance, 0)
			return
		}
		horizontalScale := math.Sqrt(ts.textMatrix.A*ts.textMatrix.A + ts.textMatrix.B*ts.textMatrix.B)
		if horizontalScale <= 0 {
			horizontalScale = 1
		}
		width := advance * horizontalScale
		if last := lastExtractedTextOnBaseline(out, pageIdx, y); last != nil {
			clampExtractedChunkWidth(last, x)
		}
		out = append(out, &ExtractedText{
			Text:     text,
			X:        x,
			Y:        y,
			Width:    width,
			Height:   effectiveSize,
			FontName: strings.TrimPrefix(ts.fontName, "/"),
			FontSize: effectiveSize,
			Bold:     strings.Contains(strings.ToLower(ts.fontName), "bold"),
			Italic:   strings.Contains(strings.ToLower(ts.fontName), "italic") || strings.Contains(strings.ToLower(ts.fontName), "oblique"),
			Color:    ts.fillColor,
			Baseline: y,
			Page:     pageIdx,
		})
		ts.textMatrix = ts.textMatrix.translate(advance, 0)
	}

	appendText := func(text string) {
		appendTextToken(text, nil, nil)
	}

	for _, tok := range tokens {
		if tok.kind != "word" {
			operands = append(operands, tok)
			continue
		}

		switch tok.value {
		case "q":
			stack = append(stack, gs)
		case "Q":
			if len(stack) > 0 {
				gs = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				ts.fillColor = gs.fillColor
			}
		case "cm":
			if m, ok := matrixFromOperands(operands); ok {
				gs.ctm = gs.ctm.multiply(m)
			}
		case "rg":
			if c, ok := colorFromOperands(operands, false); ok {
				gs.fillColor = c
				ts.fillColor = c
			}
		case "g":
			if c, ok := colorFromOperands(operands, true); ok {
				gs.fillColor = c
				ts.fillColor = c
			}
		case "k":
			if c, ok := colorFromOperands(operands, false); ok {
				gs.fillColor = c
				ts.fillColor = c
			}
		case "BT":
			ts.inText = true
			ts.textMatrix = identMatrix()
			ts.lineMatrix = identMatrix()
			ts.fillColor = gs.fillColor
		case "ET":
			ts.inText = false
		case "Tf":
			if len(operands) >= 2 {
				if operands[len(operands)-2].kind == "name" {
					ts.fontName = operands[len(operands)-2].value
				}
				if v, ok := parseFloatToken(operands[len(operands)-1]); ok {
					ts.fontSize = v
				}
			}
		case "Tm":
			if m, ok := matrixFromOperands(operands); ok {
				ts.textMatrix = m
				ts.lineMatrix = m
			}
		case "Td":
			if values, ok := parseFloats(operands); ok && len(values) == 2 {
				ts.lineMatrix = ts.lineMatrix.translate(values[0], values[1])
				ts.textMatrix = ts.lineMatrix
			}
		case "TD":
			if values, ok := parseFloats(operands); ok && len(values) == 2 {
				ts.leading = -values[1]
				ts.lineMatrix = ts.lineMatrix.translate(values[0], values[1])
				ts.textMatrix = ts.lineMatrix
			}
		case "T*":
			ts.lineMatrix = ts.lineMatrix.translate(0, -ts.leading)
			ts.textMatrix = ts.lineMatrix
		case "TL":
			if len(operands) == 1 {
				if v, ok := parseFloatToken(operands[0]); ok {
					ts.leading = v
				}
			}
		case "Tc":
			if len(operands) == 1 {
				if v, ok := parseFloatToken(operands[0]); ok {
					ts.charSpacing = v
				}
			}
		case "Tw":
			if len(operands) == 1 {
				if v, ok := parseFloatToken(operands[0]); ok {
					ts.wordSpacing = v
				}
			}
		case "Tj":
			if len(operands) == 1 && (operands[0].kind == "string" || operands[0].kind == "hex") {
				appendTextToken(decodeTextToken(operands[0], fontDecoders[ts.fontName]), operands[0].raw, nil)
			}
		case "TJ":
			if len(operands) == 1 && operands[0].kind == "array" {
				text := decodeTJTextWithFont(operands[0], fontDecoders[ts.fontName])
				if decoder := fontDecoders[ts.fontName]; decoder != nil {
					advance := decoder.tjTextAdvance(operands[0], ts.fontSize, ts.charSpacing, ts.wordSpacing)
					appendTextToken(text, nil, &advance)
					break
				}
				appendText(text)
			}
		case "'":
			ts.lineMatrix = ts.lineMatrix.translate(0, -ts.leading)
			ts.textMatrix = ts.lineMatrix
			if len(operands) == 1 && (operands[0].kind == "string" || operands[0].kind == "hex") {
				appendTextToken(decodeTextToken(operands[0], fontDecoders[ts.fontName]), operands[0].raw, nil)
			}
		case "\"":
			ts.lineMatrix = ts.lineMatrix.translate(0, -ts.leading)
			ts.textMatrix = ts.lineMatrix
			if len(operands) >= 1 {
				last := operands[len(operands)-1]
				if last.kind == "string" || last.kind == "hex" {
					appendTextToken(decodeTextToken(last, fontDecoders[ts.fontName]), last.raw, nil)
				}
			}
		default:
		}

		operands = operands[:0]
	}

	if out == nil {
		return []*ExtractedText{}, nil
	}
	if pageIdx < 0 {
		return nil, fmt.Errorf("invalid page index")
	}
	return out, nil
}

func clampExtractedChunkWidth(prev *ExtractedText, nextX float64) {
	if prev == nil || nextX <= prev.X {
		return
	}
	endX := prev.X + prev.Width
	if endX <= nextX {
		return
	}
	boundaryEpsilon := math.Min(prev.Height*0.2, 2.0)
	width := nextX - prev.X - boundaryEpsilon
	if width < 0 {
		width = 0
	}
	prev.Width = width
}

func lastExtractedTextOnBaseline(out []*ExtractedText, pageIdx int, baseline float64) *ExtractedText {
	if len(out) == 0 {
		return nil
	}
	last := out[len(out)-1]
	if last == nil || last.Page != pageIdx {
		return nil
	}
	if math.Abs(last.Baseline-baseline) > 0.5 {
		return nil
	}
	return last
}

func foldLeadingBoundaryWhitespace(last *ExtractedText, text string) string {
	if last == nil || text == "" {
		return text
	}
	trimmed := strings.TrimLeftFunc(text, unicode.IsSpace)
	leading := text[:len(text)-len(trimmed)]
	if leading == "" || len(leading) == len(text) {
		return text
	}
	boundary := normalizeBoundaryWhitespace(leading)
	if boundary == "" {
		return trimmed
	}
	if tail, _ := utf8LastRuneInString(last.Text); !unicode.IsSpace(tail) {
		last.Text += boundary
		last.Width = textWidthEstimate(last.Text, last.FontSize)
	}
	return trimmed
}

func appendBoundaryWhitespace(last *ExtractedText, text string) {
	if last == nil || text == "" {
		return
	}
	boundary := normalizeBoundaryWhitespace(text)
	if boundary == "" {
		return
	}
	if tail, _ := utf8LastRuneInString(last.Text); unicode.IsSpace(tail) {
		return
	}
	last.Text += boundary
	last.Width = textWidthEstimate(last.Text, last.FontSize)
}

func normalizeBoundaryWhitespace(text string) string {
	for _, r := range text {
		if !unicode.IsSpace(r) {
			return ""
		}
	}
	if text == "" {
		return ""
	}
	return " "
}
