package extractor

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/model"
	pdfmodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"golang.org/x/text/encoding/charmap"
)

type fontDecoder struct {
	toUnicode map[uint32]string
	codeLens  []int
	encoding  [256]rune
	hasFont   bool
}

func loadPageFontDecoders(doc *model.PDDocument, pageIdx int) (map[string]*fontDecoder, error) {
	if doc == nil || doc.Context == nil {
		return nil, fmt.Errorf("pdf document is not open")
	}
	_, _, attrs, err := doc.Context.PageDict(pageIdx+1, false)
	if err != nil {
		return nil, err
	}
	if attrs == nil || attrs.Resources == nil {
		return map[string]*fontDecoder{}, nil
	}
	fontObj, ok := attrs.Resources.Find("Font")
	if !ok || fontObj == nil {
		return map[string]*fontDecoder{}, nil
	}
	fontDict, err := doc.Context.DereferenceDict(fontObj)
	if err != nil {
		return nil, err
	}
	decoders := map[string]*fontDecoder{}
	for name, obj := range fontDict {
		decodedName, err := types.DecodeName(name)
		if err != nil {
			decodedName = name
		}
		fd, err := loadFontDecoder(doc.Context, obj)
		if err != nil {
			continue
		}
		decoders["/"+decodedName] = fd
	}
	return decoders, nil
}

func loadFontDecoder(ctx *pdfmodel.Context, obj types.Object) (*fontDecoder, error) {
	fontDict, err := ctx.DereferenceDict(obj)
	if err != nil {
		return nil, err
	}
	decoder := &fontDecoder{hasFont: true, encoding: standardEncodingTable()}
	if cmap, codeLens, err := parseToUnicodeCMap(ctx, fontDict); err == nil && len(cmap) > 0 {
		decoder.toUnicode = cmap
		decoder.codeLens = codeLens
	}
	decoder.encoding = resolveFontEncoding(ctx, fontDict)
	return decoder, nil
}

func (d *fontDecoder) decode(raw []byte) string {
	if d == nil || !d.hasFont || len(raw) == 0 {
		return normalizePDFString(raw)
	}
	if len(d.toUnicode) > 0 {
		if s, ok := d.decodeToUnicode(raw); ok {
			return s
		}
	}
	if s, ok := d.decodeEncoding(raw); ok {
		return s
	}
	return normalizePDFString(raw)
}

func (d *fontDecoder) decodeToUnicode(raw []byte) (string, bool) {
	var sb strings.Builder
	for i := 0; i < len(raw); {
		matched := false
		for _, codeLen := range d.codeLens {
			if i+codeLen > len(raw) {
				continue
			}
			code := cmapCode(raw[i : i+codeLen])
			if mapped, ok := d.toUnicode[code]; ok {
				sb.WriteString(mapped)
				i += codeLen
				matched = true
				break
			}
		}
		if !matched {
			return "", false
		}
	}
	return sb.String(), true
}

func (d *fontDecoder) decodeEncoding(raw []byte) (string, bool) {
	if len(raw) == 0 {
		return "", true
	}
	var sb strings.Builder
	for _, b := range raw {
		r := d.encoding[b]
		if r == 0 {
			return "", false
		}
		sb.WriteRune(r)
	}
	return sb.String(), true
}

func resolveFontEncoding(ctx *pdfmodel.Context, fontDict types.Dict) [256]rune {
	table := standardEncodingTable()
	if subtype := fontDict.NameEntry("Subtype"); subtype != nil && *subtype == "TrueType" {
		table = winAnsiTable()
	}
	if baseFont := fontDict.NameEntry("BaseFont"); baseFont != nil {
		switch strings.TrimPrefix(*baseFont, "/") {
		case "Symbol":
			return symbolEncodingTable()
		case "ZapfDingbats":
			return zapfDingbatsTable()
		}
	}
	encObj, ok := fontDict.Find("Encoding")
	if !ok || encObj == nil {
		return table
	}
	obj, err := ctx.Dereference(encObj)
	if err != nil || obj == nil {
		return table
	}
	switch enc := obj.(type) {
	case types.Name:
		return encodingTableByName(enc.Value(), table)
	case types.Dict:
		if base := enc.NameEntry("BaseEncoding"); base != nil {
			table = encodingTableByName(*base, table)
		}
		applyDifferences(ctx, &table, enc.ArrayEntry("Differences"))
		return table
	}
	return table
}

func encodingTableByName(name string, fallback [256]rune) [256]rune {
	switch name {
	case "MacRomanEncoding":
		return macRomanTable()
	case "WinAnsiEncoding":
		return winAnsiTable()
	case "StandardEncoding":
		return standardEncodingTable()
	default:
		return fallback
	}
}

func applyDifferences(ctx *pdfmodel.Context, table *[256]rune, diffs types.Array) {
	if len(diffs) == 0 {
		return
	}
	current := -1
	for _, obj := range diffs {
		decoded, err := ctx.Dereference(obj)
		if err != nil || decoded == nil {
			continue
		}
		switch v := decoded.(type) {
		case types.Integer:
			current = v.Value()
		case types.Name:
			if current >= 0 && current < len(table) {
				if r, ok := glyphNameToRune(v.Value()); ok {
					table[current] = r
				}
			}
			current++
		}
	}
}

func parseToUnicodeCMap(ctx *pdfmodel.Context, fontDict types.Dict) (map[uint32]string, []int, error) {
	obj, ok := fontDict.Find("ToUnicode")
	if !ok || obj == nil {
		return nil, nil, nil
	}
	sd, _, err := ctx.DereferenceStreamDict(obj)
	if err != nil || sd == nil {
		return nil, nil, err
	}
	if err := sd.Decode(); err != nil {
		return nil, nil, err
	}
	return parseCMapContent(string(sd.Content))
}

var cmapHexPattern = regexp.MustCompile(`<([0-9A-Fa-f]+)>`)

func parseCMapContent(content string) (map[uint32]string, []int, error) {
	lines := strings.Split(content, "\n")
	mapping := map[uint32]string{}
	codeLensSet := map[int]struct{}{}
	mode := ""
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		switch {
		case strings.HasSuffix(line, "beginbfchar"):
			mode = "bfchar"
			continue
		case strings.HasSuffix(line, "endbfchar"):
			mode = ""
			continue
		case strings.HasSuffix(line, "beginbfrange"):
			mode = "bfrange"
			continue
		case strings.HasSuffix(line, "endbfrange"):
			mode = ""
			continue
		}
		matches := cmapHexPattern.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}
		switch mode {
		case "bfchar":
			if len(matches) < 2 {
				continue
			}
			src, err := hex.DecodeString(matches[0][1])
			if err != nil {
				return nil, nil, err
			}
			dst, err := decodeCMapUnicode(matches[1][1])
			if err != nil {
				return nil, nil, err
			}
			mapping[cmapCode(src)] = dst
			codeLensSet[len(src)] = struct{}{}
		case "bfrange":
			if len(matches) < 3 {
				continue
			}
			startBytes, err := hex.DecodeString(matches[0][1])
			if err != nil {
				return nil, nil, err
			}
			endBytes, err := hex.DecodeString(matches[1][1])
			if err != nil {
				return nil, nil, err
			}
			start := cmapCode(startBytes)
			end := cmapCode(endBytes)
			codeLensSet[len(startBytes)] = struct{}{}
			if strings.Contains(line, "[") {
				for i, match := range matches[2:] {
					dst, err := decodeCMapUnicode(match[1])
					if err != nil {
						return nil, nil, err
					}
					mapping[start+uint32(i)] = dst
				}
				continue
			}
			dstBytes, err := hex.DecodeString(matches[2][1])
			if err != nil {
				return nil, nil, err
			}
			for src := start; src <= end; src++ {
				mapping[src] = utf16BytesToString(dstBytes)
				incrementHexBytes(dstBytes)
			}
		}
	}
	codeLens := make([]int, 0, len(codeLensSet))
	for codeLen := range codeLensSet {
		codeLens = append(codeLens, codeLen)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(codeLens)))
	return mapping, codeLens, nil
}

func decodeCMapUnicode(hexValue string) (string, error) {
	b, err := hex.DecodeString(hexValue)
	if err != nil {
		return "", err
	}
	return utf16BytesToString(b), nil
}

func utf16BytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	if len(b)%2 == 1 {
		b = append([]byte(nil), b...)
		b = append(b, 0)
	}
	runes := make([]rune, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		runes = append(runes, rune(uint16(b[i])<<8|uint16(b[i+1])))
	}
	return string(runes)
}

func incrementHexBytes(b []byte) {
	for i := len(b) - 1; i >= 0; i-- {
		b[i]++
		if b[i] != 0 {
			return
		}
	}
}

func cmapCode(b []byte) uint32 {
	var code uint32
	for _, v := range b {
		code = (code << 8) | uint32(v)
	}
	return code
}

func decodeTextToken(tok streamToken, decoder *fontDecoder) string {
	if tok.kind != "string" && tok.kind != "hex" {
		return ""
	}
	if len(tok.raw) > 0 {
		return decoder.decode(tok.raw)
	}
	return tok.value
}

func decodeTJTextWithFont(tok streamToken, decoder *fontDecoder) string {
	if tok.kind != "array" {
		return ""
	}
	var sb strings.Builder
	for _, item := range tok.items {
		if item.kind == "string" || item.kind == "hex" {
			sb.WriteString(decodeTextToken(item, decoder))
		} else if item.kind == "number" {
			if v, ok := parseFloatToken(item); ok && v <= tjSpaceThreshold {
				sb.WriteByte(' ')
			}
		}
	}
	return normalizeExtractedText(sb.String())
}

func macRomanTable() [256]rune {
	var table [256]rune
	decoded, err := charmap.Macintosh.NewDecoder().String(string(bytes0to255()))
	if err != nil {
		return table
	}
	for i, r := range decoded {
		table[i] = r
	}
	return table
}

func winAnsiTable() [256]rune {
	var table [256]rune
	for i := 0; i < 256; i++ {
		table[i] = rune(i)
	}
	for b, r := range win1252Extras {
		table[int(b)] = r
	}
	return table
}

func standardEncodingTable() [256]rune {
	var table [256]rune
	for i := 32; i <= 126; i++ {
		table[i] = rune(i)
	}
	for k, v := range map[int]rune{
		161: '¡', 162: '¢', 163: '£', 164: '⁄', 165: '¥', 167: '§', 168: '¤',
		169: '\'', 170: '“', 171: '«', 172: '‹', 173: '›', 174: 'ﬁ', 175: 'ﬂ',
		177: '–', 178: '†', 179: '‡', 180: '·', 182: '¶', 183: '•', 184: '‚',
		185: '„', 186: '”', 187: '»', 188: '…', 189: '‰', 191: '¿',
		193: '`', 194: '´', 195: 'ˆ', 196: '˜', 197: '¯', 198: '˘', 199: '˙',
		200: '¨', 202: '˚', 203: '¸', 205: '˝', 206: '˛', 207: 'ˇ',
		225: 'Æ', 227: 'ª', 232: 'Ł', 233: 'Ø', 234: 'Œ', 235: 'º', 241: 'æ',
		245: 'ı', 248: 'ł', 249: 'ø', 250: 'œ', 251: 'ß',
	} {
		table[k] = v
	}
	return table
}

func symbolEncodingTable() [256]rune {
	return standardEncodingTable()
}

func zapfDingbatsTable() [256]rune {
	return standardEncodingTable()
}

func bytes0to255() []byte {
	b := make([]byte, 256)
	for i := 0; i < len(b); i++ {
		b[i] = byte(i)
	}
	return b
}

func glyphNameToRune(name string) (rune, bool) {
	name = strings.TrimPrefix(name, "/")
	if idx := strings.IndexByte(name, '.'); idx >= 0 {
		name = name[:idx]
	}
	if strings.HasPrefix(name, "uni") && len(name) >= 7 {
		v, err := strconv.ParseUint(name[3:7], 16, 32)
		if err == nil {
			return rune(v), true
		}
	}
	if strings.HasPrefix(name, "u") && len(name) >= 5 {
		v, err := strconv.ParseUint(name[1:], 16, 32)
		if err == nil {
			return rune(v), true
		}
	}
	if len(name) == 1 {
		return rune(name[0]), true
	}
	if r, ok := glyphNameRunes[name]; ok {
		return r, true
	}
	return 0, false
}

var glyphNameRunes = map[string]rune{
	"space": ' ', "exclam": '!', "quotedbl": '"', "numbersign": '#', "dollar": '$', "percent": '%',
	"ampersand": '&', "quotesingle": '\'', "parenleft": '(', "parenright": ')', "asterisk": '*',
	"plus": '+', "comma": ',', "hyphen": '-', "period": '.', "slash": '/',
	"zero": '0', "one": '1', "two": '2', "three": '3', "four": '4', "five": '5', "six": '6', "seven": '7', "eight": '8', "nine": '9',
	"colon": ':', "semicolon": ';', "less": '<', "equal": '=', "greater": '>', "question": '?', "at": '@',
	"bracketleft": '[', "backslash": '\\', "bracketright": ']', "asciicircum": '^', "underscore": '_', "grave": '`',
	"braceleft": '{', "bar": '|', "braceright": '}', "asciitilde": '~',
	"Euro": '€', "bullet": '•', "emdash": '—', "endash": '–', "ellipsis": '…',
	"quoteleft": '‘', "quoteright": '’', "quotedblleft": '“', "quotedblright": '”',
	"quotesinglbase": '‚', "quotedblbase": '„', "dagger": '†', "daggerdbl": '‡',
	"perthousand": '‰', "guilsinglleft": '‹', "guilsinglright": '›', "guillemotleft": '«', "guillemotright": '»',
	"fi": 'ﬁ', "fl": 'ﬂ', "fraction": '⁄', "minus": '−', "ring": '˚', "cedilla": '¸',
	"hungarumlaut": '˝', "ogonek": '˛', "caron": 'ˇ', "breve": '˘', "dotaccent": '˙',
	"circumflex": 'ˆ', "tilde": '˜', "macron": '¯', "Lslash": 'Ł', "lslash": 'ł',
	"OE": 'Œ', "oe": 'œ', "Scaron": 'Š', "scaron": 'š', "Ydieresis": 'Ÿ', "Zcaron": 'Ž', "zcaron": 'ž',
	"AE": 'Æ', "ae": 'æ', "Oslash": 'Ø', "oslash": 'ø', "ordfeminine": 'ª', "ordmasculine": 'º',
	"exclamdown": '¡', "questiondown": '¿', "sterling": '£', "yen": '¥', "section": '§',
	"currency": '¤', "paragraph": '¶', "copyright": '©', "registered": '®', "trademark": '™',
	"degree": '°', "plusminus": '±', "mu": 'µ', "logicalnot": '¬', "brokenbar": '¦',
}

func init() {
	for ch := 'A'; ch <= 'Z'; ch++ {
		glyphNameRunes[string(ch)] = ch
	}
	for ch := 'a'; ch <= 'z'; ch++ {
		glyphNameRunes[string(ch)] = ch
	}
}
