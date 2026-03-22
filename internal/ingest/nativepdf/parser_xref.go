package nativepdf

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type xrefEntry struct {
	Offset     int
	Generation int
	InUse      bool
}

type xrefInfo struct {
	StartXref int
	Trailer   pdfDict
	Entries   map[pdfRef]xrefEntry
}

func parseXrefInfo(data []byte) (*xrefInfo, bool) {
	startXref, ok := findStartXref(data)
	if !ok {
		return nil, false
	}

	info := &xrefInfo{
		StartXref: startXref,
		Entries:   make(map[pdfRef]xrefEntry),
	}

	if entries, trailer, ok := parseTraditionalXrefTable(data, startXref); ok {
		info.Entries = entries
		info.Trailer = trailer
		return info, true
	}

	if trailer, ok := parseTrailerAfterKeyword(data, startXref); ok {
		info.Trailer = trailer
		return info, true
	}

	return info, true
}

func findStartXref(data []byte) (int, bool) {
	pos := bytes.LastIndex(data, []byte("startxref"))
	if pos < 0 {
		return 0, false
	}
	pos += len("startxref")
	pos = skipSpaces(data, pos)
	value, _, ok := parsePositiveInt(data, pos)
	if !ok {
		return 0, false
	}
	return value, true
}

func parseTraditionalXrefTable(data []byte, offset int) (map[pdfRef]xrefEntry, pdfDict, bool) {
	if offset < 0 || offset >= len(data) {
		return nil, nil, false
	}

	pos := skipSpaces(data, offset)
	if !hasKeyword(data, pos, "xref") {
		return nil, nil, false
	}
	pos += len("xref")

	entries := make(map[pdfRef]xrefEntry)
	for {
		pos = skipSpaces(data, pos)
		if pos >= len(data) {
			return nil, nil, false
		}
		if hasKeyword(data, pos, "trailer") {
			pos += len("trailer")
			trailer, ok := parseTrailerDict(data, pos)
			if !ok {
				return nil, nil, false
			}
			return entries, trailer, true
		}

		startObj, next, ok := parsePositiveInt(data, pos)
		if !ok {
			return nil, nil, false
		}
		next = skipSpaces(data, next)
		count, next, ok := parsePositiveInt(data, next)
		if !ok {
			return nil, nil, false
		}
		pos = next

		for i := 0; i < count; i++ {
			pos = skipSpaces(data, pos)
			line, nextLine := readLine(data, pos)
			if len(line) == 0 {
				return nil, nil, false
			}
			fields := strings.Fields(string(line))
			if len(fields) < 3 {
				return nil, nil, false
			}
			offsetValue, err := strconv.Atoi(fields[0])
			if err != nil {
				return nil, nil, false
			}
			generationValue, err := strconv.Atoi(fields[1])
			if err != nil {
				return nil, nil, false
			}
			inUse := fields[2] == "n"
			ref := pdfRef{
				ObjectNumber: startObj + i,
				Generation:   generationValue,
			}
			entries[ref] = xrefEntry{
				Offset:     offsetValue,
				Generation: generationValue,
				InUse:      inUse,
			}
			pos = nextLine
		}
	}
}

func parseTrailerAfterKeyword(data []byte, offset int) (pdfDict, bool) {
	pos := bytes.Index(data[offset:], []byte("trailer"))
	if pos < 0 {
		return nil, false
	}
	return parseTrailerDict(data, offset+pos+len("trailer"))
}

func parseTrailerDict(data []byte, pos int) (pdfDict, bool) {
	pos = skipSpaces(data, pos)
	parser := pdfValueParser{data: data, pos: pos}
	value, ok := parser.parseDict()
	if !ok {
		return nil, false
	}
	dict, ok := value.(pdfDict)
	if !ok {
		return nil, false
	}
	return dict, true
}

func readLine(data []byte, pos int) ([]byte, int) {
	if pos >= len(data) {
		return nil, pos
	}
	start := pos
	for pos < len(data) && data[pos] != '\n' && data[pos] != '\r' {
		pos++
	}
	line := bytes.TrimSpace(data[start:pos])
	for pos < len(data) && (data[pos] == '\n' || data[pos] == '\r') {
		pos++
	}
	return line, pos
}

func parseIndirectObjectAt(data []byte, offset int) (*indirectObject, bool) {
	if offset < 0 || offset >= len(data) {
		return nil, false
	}

	pos := skipSpaces(data, offset)
	objNum, next, ok := parsePositiveInt(data, pos)
	if !ok {
		return nil, false
	}
	next = skipSpaces(data, next)
	genNum, next, ok := parsePositiveInt(data, next)
	if !ok {
		return nil, false
	}
	next = skipSpaces(data, next)
	if !hasKeyword(data, next, "obj") {
		return nil, false
	}
	next += len("obj")

	end, ok := findObjectBodyEnd(data, next)
	if !ok {
		return nil, false
	}
	raw := bytes.TrimSpace(data[pos:end])
	value := firstObjectValue(raw)
	if value == nil {
		return nil, false
	}
	return &indirectObject{
		Ref: pdfRef{
			ObjectNumber: objNum,
			Generation:   genNum,
		},
		Raw:   append([]byte(nil), raw...),
		Value: value,
	}, true
}

func findObjectBodyEnd(data []byte, bodyStart int) (int, bool) {
	if bodyStart < 0 || bodyStart >= len(data) {
		return 0, false
	}
	if endRel := bytes.Index(data[bodyStart:], []byte("endobj")); endRel >= 0 {
		return bodyStart + endRel, true
	}
	if endRel := bytes.Index(data[bodyStart:], []byte("endstream")); endRel >= 0 {
		return bodyStart + endRel + len("endstream"), true
	}
	return len(data), true
}

func parseRefValue(value pdfValue) (pdfRef, bool) {
	ref, ok := value.(pdfRef)
	return ref, ok
}

func sortEntriesByOffset(entries map[pdfRef]xrefEntry) []pdfRef {
	refs := make([]pdfRef, 0, len(entries))
	for ref := range entries {
		refs = append(refs, ref)
	}
	sort.Slice(refs, func(i, j int) bool {
		left := entries[refs[i]]
		right := entries[refs[j]]
		if left.Offset != right.Offset {
			return left.Offset < right.Offset
		}
		if refs[i].ObjectNumber != refs[j].ObjectNumber {
			return refs[i].ObjectNumber < refs[j].ObjectNumber
		}
		return refs[i].Generation < refs[j].Generation
	})
	return refs
}

func xrefRootRef(trailer pdfDict) (pdfRef, bool) {
	if trailer == nil {
		return pdfRef{}, false
	}
	ref, ok := parseRefValue(trailer["Root"])
	return ref, ok
}

func xrefSize(trailer pdfDict) (int, bool) {
	if trailer == nil {
		return 0, false
	}
	return intValue(trailer["Size"])
}

func xrefEntryForObject(info *xrefInfo, ref pdfRef) (xrefEntry, bool) {
	if info == nil {
		return xrefEntry{}, false
	}
	entry, ok := info.Entries[ref]
	return entry, ok
}

func xrefInfoHasRoot(info *xrefInfo) bool {
	if info == nil {
		return false
	}
	_, ok := xrefRootRef(info.Trailer)
	return ok
}

func xrefInfoRoot(info *xrefInfo) (pdfRef, bool) {
	if info == nil {
		return pdfRef{}, false
	}
	return xrefRootRef(info.Trailer)
}

func xrefInfoSize(info *xrefInfo) (int, bool) {
	if info == nil {
		return 0, false
	}
	return xrefSize(info.Trailer)
}

func xrefInfoString(info *xrefInfo) string {
	if info == nil {
		return ""
	}
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("startxref=%d", info.StartXref))
	if size, ok := xrefInfoSize(info); ok {
		builder.WriteString(fmt.Sprintf(" size=%d", size))
	}
	if ref, ok := xrefInfoRoot(info); ok {
		builder.WriteString(fmt.Sprintf(" root=%d %d R", ref.ObjectNumber, ref.Generation))
	}
	return builder.String()
}
