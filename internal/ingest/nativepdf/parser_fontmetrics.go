package nativepdf

import (
	"encoding/hex"
	"math"
	"strings"
)

type fontMetrics struct {
	widths       map[int]float64
	defaultWidth float64
	codeBytes    int
}

func resolveFontMetrics(g *objectGraph, resources pdfDict) map[string]fontMetrics {
	if g == nil || len(resources) == 0 {
		return nil
	}

	fontsValue, ok := resources["Font"]
	if !ok {
		return nil
	}
	fontsDict, ok := resolveDictValue(g, fontsValue)
	if !ok || len(fontsDict) == 0 {
		return nil
	}

	out := make(map[string]fontMetrics, len(fontsDict))
	for alias, value := range fontsDict {
		fontDict, ok := resolveDictValue(g, value)
		if !ok {
			continue
		}
		metrics, ok := fontMetricsFromDict(g, fontDict)
		if !ok {
			continue
		}
		out[alias] = metrics
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func fontMetricsFromDict(g *objectGraph, dict pdfDict) (fontMetrics, bool) {
	if metrics, ok := simpleFontMetricsFromDict(g, dict); ok {
		return metrics, true
	}
	return descendantFontMetricsFromDict(g, dict)
}

func simpleFontMetricsFromDict(g *objectGraph, dict pdfDict) (fontMetrics, bool) {
	metrics := fontMetrics{
		defaultWidth: defaultWidthFromDict(g, dict),
		codeBytes:    1,
	}
	if widthArray, ok := resolveArrayValue(g, dict["Widths"]); ok && len(widthArray) > 0 {
		firstChar, ok := intValue(dict["FirstChar"])
		if ok {
			metrics.widths = make(map[int]float64, len(widthArray))
			for i, item := range widthArray {
				width, ok := floatValue(item)
				if !ok || width <= 0 {
					continue
				}
				metrics.widths[firstChar+i] = width
			}
		}
	}
	if len(metrics.widths) == 0 && metrics.defaultWidth <= 0 {
		return standardFontMetricsFromDict(dict)
	}
	return metrics, true
}

func descendantFontMetricsFromDict(g *objectGraph, dict pdfDict) (fontMetrics, bool) {
	items, ok := resolveArrayValue(g, dict["DescendantFonts"])
	if !ok || len(items) == 0 {
		return fontMetrics{}, false
	}
	descendant, ok := resolveDictValue(g, items[0])
	if !ok {
		return fontMetrics{}, false
	}

	metrics := fontMetrics{
		widths:       make(map[int]float64),
		defaultWidth: defaultWidthFromDict(g, descendant),
		codeBytes:    2,
	}
	if metrics.defaultWidth <= 0 {
		metrics.defaultWidth = 1000
	}
	if dw, ok := floatValueResolved(g, descendant["DW"]); ok && dw > 0 {
		metrics.defaultWidth = dw
	}
	if widthSpec, ok := resolveArrayValue(g, descendant["W"]); ok {
		applyCIDWidths(&metrics, widthSpec)
	}
	if len(metrics.widths) == 0 && metrics.defaultWidth <= 0 {
		return fontMetrics{}, false
	}
	return metrics, true
}

func applyCIDWidths(metrics *fontMetrics, spec pdfArray) {
	for i := 0; i < len(spec); {
		start, ok := intValue(spec[i])
		if !ok || i+1 >= len(spec) {
			return
		}

		switch next := spec[i+1].(type) {
		case pdfArray:
			for offset, item := range next {
				width, ok := floatValue(item)
				if !ok || width <= 0 {
					continue
				}
				metrics.widths[start+offset] = width
			}
			i += 2
		default:
			end, ok := intValue(spec[i+1])
			if !ok || i+2 >= len(spec) {
				return
			}
			width, ok := floatValue(spec[i+2])
			if !ok || width <= 0 {
				i += 3
				continue
			}
			for code := start; code <= end; code++ {
				metrics.widths[code] = width
			}
			i += 3
		}
	}
}

func missingWidthFromDict(g *objectGraph, dict pdfDict) float64 {
	if descriptor, ok := fontDescriptorFromDict(g, dict); ok {
		if width, ok := floatValue(descriptor["MissingWidth"]); ok && width > 0 {
			return width
		}
	}
	return 0
}

func defaultWidthFromDict(g *objectGraph, dict pdfDict) float64 {
	if width := missingWidthFromDict(g, dict); width > 0 {
		return width
	}
	if descriptor, ok := fontDescriptorFromDict(g, dict); ok {
		if width, ok := floatValue(descriptor["AvgWidth"]); ok && width > 0 {
			return width
		}
	}
	return 0
}

func fontDescriptorFromDict(g *objectGraph, dict pdfDict) (pdfDict, bool) {
	if descriptor, ok := resolveDictValue(g, dict["FontDescriptor"]); ok {
		return descriptor, true
	}
	items, ok := resolveArrayValue(g, dict["DescendantFonts"])
	if !ok || len(items) == 0 {
		return nil, false
	}
	descendant, ok := resolveDictValue(g, items[0])
	if !ok {
		return nil, false
	}
	return resolveDictValue(g, descendant["FontDescriptor"])
}

func resolveArrayValue(g *objectGraph, value pdfValue) (pdfArray, bool) {
	switch typed := value.(type) {
	case pdfArray:
		return typed, true
	case pdfRef:
		if g == nil {
			return nil, false
		}
		object := g.objects[typed]
		if object == nil {
			return nil, false
		}
		items, ok := object.Value.(pdfArray)
		return items, ok
	default:
		return nil, false
	}
}

func floatValueResolved(g *objectGraph, value pdfValue) (float64, bool) {
	if number, ok := floatValue(value); ok {
		return number, true
	}
	ref, ok := value.(pdfRef)
	if !ok || g == nil {
		return 0, false
	}
	object := g.objects[ref]
	if object == nil {
		return 0, false
	}
	return floatValue(object.Value)
}

func (m fontMetrics) advanceForLiteralString(text string, fontSize float64) float64 {
	if len(m.widths) == 0 && m.defaultWidth <= 0 {
		return 0
	}
	codes := make([]int, 0, len(text))
	for i := 0; i < len(text); i++ {
		codes = append(codes, int(text[i]))
	}
	return m.advanceForCodes(codes, fontSize)
}

func (m fontMetrics) advanceForHexString(text string, fontSize float64) float64 {
	if len(m.widths) == 0 && m.defaultWidth <= 0 {
		return 0
	}
	raw, err := hex.DecodeString(padOddHex(text))
	if err != nil || len(raw) == 0 {
		return 0
	}
	codeBytes := m.codeBytes
	if codeBytes <= 0 {
		codeBytes = 1
	}
	codes := make([]int, 0, len(raw)/codeBytes)
	for i := 0; i+codeBytes <= len(raw); i += codeBytes {
		code := 0
		for j := 0; j < codeBytes; j++ {
			code = (code << 8) | int(raw[i+j])
		}
		codes = append(codes, code)
	}
	return m.advanceForCodes(codes, fontSize)
}

func (m fontMetrics) advanceForCodes(codes []int, fontSize float64) float64 {
	if len(codes) == 0 {
		return 0
	}
	if fontSize <= 0 {
		fontSize = 12
	}
	var total float64
	var used bool
	for _, code := range codes {
		width, ok := m.widths[code]
		if !ok || width <= 0 {
			width = m.defaultWidth
		}
		if width <= 0 {
			return 0
		}
		used = true
		total += width
	}
	if !used {
		return 0
	}
	return math.Max(fontSize*0.35, (total*fontSize)/1000.0)
}

func padOddHex(text string) string {
	if len(text)%2 == 0 {
		return text
	}
	return text + "0"
}

func standardFontMetricsFromDict(dict pdfDict) (fontMetrics, bool) {
	name := normalizeStandardFontName(fontNameFromDict(dict))
	switch name {
	case "Courier", "Courier-Bold", "Courier-Oblique", "Courier-BoldOblique":
		return fontMetrics{defaultWidth: 600, codeBytes: 1}, true
	case "Helvetica", "Helvetica-Bold", "Helvetica-Oblique", "Helvetica-BoldOblique":
		return fontMetrics{
			widths:       standardHelveticaWidths(),
			defaultWidth: 556,
			codeBytes:    1,
		}, true
	case "Times-Roman", "Times-Bold", "Times-Italic", "Times-BoldItalic":
		return fontMetrics{
			widths:       standardTimesRomanWidths(),
			defaultWidth: 500,
			codeBytes:    1,
		}, true
	case "Symbol", "ZapfDingbats":
		return fontMetrics{defaultWidth: 600, codeBytes: 1}, true
	default:
		return fontMetrics{}, false
	}
}

func normalizeStandardFontName(name string) string {
	name = strings.TrimSpace(name)
	if len(name) > 7 && name[6] == '+' {
		prefix := name[:6]
		upper := true
		for i := 0; i < len(prefix); i++ {
			if prefix[i] < 'A' || prefix[i] > 'Z' {
				upper = false
				break
			}
		}
		if upper {
			name = name[7:]
		}
	}
	return name
}

func standardHelveticaWidths() map[int]float64 {
	return map[int]float64{
		int(' '): 278,
		int('!'): 278,
		int('"'): 355,
		int('#'): 556,
		int('$'): 556,
		int('%'): 889,
		int('&'): 667,
		int('('): 333,
		int(')'): 333,
		int(','): 278,
		int('-'): 333,
		int('.'): 278,
		int('/'): 278,
		int('0'): 556,
		int('1'): 556,
		int('2'): 556,
		int('3'): 556,
		int('4'): 556,
		int('5'): 556,
		int('6'): 556,
		int('7'): 556,
		int('8'): 556,
		int('9'): 556,
		int(':'): 278,
		int(';'): 278,
		int('?'): 556,
		int('@'): 1015,
		int('A'): 667,
		int('B'): 667,
		int('C'): 722,
		int('D'): 722,
		int('E'): 667,
		int('F'): 611,
		int('G'): 778,
		int('H'): 722,
		int('I'): 278,
		int('J'): 500,
		int('K'): 667,
		int('L'): 556,
		int('M'): 833,
		int('N'): 722,
		int('O'): 778,
		int('P'): 667,
		int('Q'): 778,
		int('R'): 722,
		int('S'): 667,
		int('T'): 611,
		int('U'): 722,
		int('V'): 667,
		int('W'): 944,
		int('X'): 667,
		int('Y'): 667,
		int('Z'): 611,
		int('a'): 556,
		int('b'): 556,
		int('c'): 500,
		int('d'): 556,
		int('e'): 556,
		int('f'): 278,
		int('g'): 556,
		int('h'): 556,
		int('i'): 222,
		int('j'): 222,
		int('k'): 500,
		int('l'): 222,
		int('m'): 833,
		int('n'): 556,
		int('o'): 556,
		int('p'): 556,
		int('q'): 556,
		int('r'): 333,
		int('s'): 500,
		int('t'): 278,
		int('u'): 556,
		int('v'): 500,
		int('w'): 722,
		int('x'): 500,
		int('y'): 500,
		int('z'): 500,
	}
}

func standardTimesRomanWidths() map[int]float64 {
	return map[int]float64{
		int(' '): 250,
		int('!'): 333,
		int('"'): 408,
		int('#'): 500,
		int('$'): 500,
		int('%'): 833,
		int('&'): 778,
		int('('): 333,
		int(')'): 333,
		int(','): 250,
		int('-'): 333,
		int('.'): 250,
		int('/'): 278,
		int('0'): 500,
		int('1'): 500,
		int('2'): 500,
		int('3'): 500,
		int('4'): 500,
		int('5'): 500,
		int('6'): 500,
		int('7'): 500,
		int('8'): 500,
		int('9'): 500,
		int(':'): 278,
		int(';'): 278,
		int('?'): 444,
		int('@'): 921,
		int('A'): 722,
		int('B'): 667,
		int('C'): 667,
		int('D'): 722,
		int('E'): 611,
		int('F'): 556,
		int('G'): 722,
		int('H'): 722,
		int('I'): 333,
		int('J'): 389,
		int('K'): 722,
		int('L'): 611,
		int('M'): 889,
		int('N'): 722,
		int('O'): 722,
		int('P'): 556,
		int('Q'): 722,
		int('R'): 667,
		int('S'): 556,
		int('T'): 611,
		int('U'): 722,
		int('V'): 722,
		int('W'): 944,
		int('X'): 722,
		int('Y'): 722,
		int('Z'): 611,
		int('a'): 444,
		int('b'): 500,
		int('c'): 444,
		int('d'): 500,
		int('e'): 444,
		int('f'): 333,
		int('g'): 500,
		int('h'): 500,
		int('i'): 278,
		int('j'): 278,
		int('k'): 500,
		int('l'): 278,
		int('m'): 778,
		int('n'): 500,
		int('o'): 500,
		int('p'): 500,
		int('q'): 500,
		int('r'): 333,
		int('s'): 389,
		int('t'): 278,
		int('u'): 500,
		int('v'): 500,
		int('w'): 722,
		int('x'): 500,
		int('y'): 500,
		int('z'): 444,
	}
}
