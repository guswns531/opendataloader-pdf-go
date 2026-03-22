package nativepdf

import (
	"bytes"
	"compress/zlib"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// SkeletonLoader is the first real native PDF loader shape. It accepts `.pdf`
// sources and produces a real DocumentHandle shell. Page extraction is still
// skeletal, but page metadata shells can already be injected for tests and
// future backend bring-up work.
type SkeletonLoader struct {
	pageMetadata []model.PageMetadata
	parser       *DocumentParser
}

// NewSkeletonLoader returns the current native PDF loader skeleton.
func NewSkeletonLoader() *SkeletonLoader {
	return &SkeletonLoader{
		parser: NewDocumentParser(),
	}
}

// NewSkeletonLoaderWithPages returns a skeleton loader preloaded with page
// metadata shells for tests and future parser bring-up work.
func NewSkeletonLoaderWithPages(pages []model.PageMetadata) *SkeletonLoader {
	return &SkeletonLoader{
		pageMetadata: append([]model.PageMetadata(nil), pages...),
		parser:       NewDocumentParser(),
	}
}

// OpenPath reads the PDF bytes from disk and returns a document handle shell.
func (l *SkeletonLoader) OpenPath(_ context.Context, path string, opts OpenOptions) (DocumentHandle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read pdf source: %w", err)
	}
	return l.open(filepath.Base(path), data, opts)
}

// OpenReader reads the PDF bytes from a stream and returns a document handle shell.
func (l *SkeletonLoader) OpenReader(_ context.Context, name string, r io.Reader, opts OpenOptions) (DocumentHandle, error) {
	if r == nil {
		return nil, fmt.Errorf("native skeleton reader is nil")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read pdf source: %w", err)
	}
	return l.open(name, data, opts)
}

func (l *SkeletonLoader) open(name string, data []byte, opts OpenOptions) (DocumentHandle, error) {
	parser := l.parser
	if parser == nil {
		parser = NewDocumentParser()
	}
	result, err := parser.Parse(name, data, opts, l.pageMetadata)
	if err != nil {
		return nil, err
	}

	pages := make([]model.PageMetadata, 0, len(result.Pages))
	artifactsByPage := make([][]*model.RawArtifact, 0, len(result.Pages))
	for _, page := range result.Pages {
		pages = append(pages, page.Metadata)
		artifactsByPage = append(artifactsByPage, cloneArtifacts(page.Artifacts))
	}

	return &skeletonDocumentHandle{
		metadata:        result.Metadata,
		raw:             append([]byte(nil), data...),
		pages:           pages,
		artifactsByPage: artifactsByPage,
		tableByPage:     tableByPageFromParsedPages(result.Pages),
		structByPage:    structByPageFromParsedPages(result.Pages),
	}, nil
}

type skeletonDocumentHandle struct {
	metadata        model.DocumentMetadata
	raw             []byte
	pages           []model.PageMetadata
	artifactsByPage [][]*model.RawArtifact
	tableByPage     []*TableCandidateSet
	structByPage    []*StructNode
}

func (h *skeletonDocumentHandle) Metadata() model.DocumentMetadata {
	return h.metadata
}

func (h *skeletonDocumentHandle) PageCount() int {
	return len(h.pages)
}

func (h *skeletonDocumentHandle) Page(pageIndex int) (PageHandle, error) {
	if pageIndex < 0 || pageIndex >= len(h.pages) {
		return nil, fmt.Errorf("native skeleton page %d is not implemented", pageIndex)
	}
	var artifacts []*model.RawArtifact
	if pageIndex < len(h.artifactsByPage) {
		artifacts = cloneArtifacts(h.artifactsByPage[pageIndex])
	}
	return &skeletonPageHandle{
		metadata:        h.pages[pageIndex],
		artifacts:       artifacts,
		tableCandidates: tableCandidateAt(h.tableByPage, pageIndex),
		structTree:      structTreeAt(h.structByPage, pageIndex),
	}, nil
}

func (h *skeletonDocumentHandle) Close() error {
	h.raw = nil
	return nil
}

type skeletonPageHandle struct {
	metadata        model.PageMetadata
	artifacts       []*model.RawArtifact
	tableCandidates *TableCandidateSet
	structTree      *StructNode
}

func (h *skeletonPageHandle) Metadata() model.PageMetadata {
	return h.metadata
}

func (h *skeletonPageHandle) Artifacts(_ context.Context, opts ArtifactOptions) ([]*model.RawArtifact, error) {
	if len(h.artifacts) == 0 {
		return nil, nil
	}
	if includeAllArtifacts(opts) {
		return cloneArtifacts(h.artifacts), nil
	}

	filtered := make([]*model.RawArtifact, 0, len(h.artifacts))
	for _, artifact := range h.artifacts {
		if artifact == nil {
			continue
		}
		if includeArtifactKind(artifact.Kind, opts) {
			filtered = append(filtered, cloneArtifact(artifact))
		}
	}
	return filtered, nil
}

func (h *skeletonPageHandle) TableCandidates(_ context.Context) (*TableCandidateSet, error) {
	return cloneTableCandidateSet(h.tableCandidates), nil
}

func (h *skeletonPageHandle) StructTree(_ context.Context) (*StructNode, error) {
	return cloneStructNode(h.structTree), nil
}

func structByPageFromParsedPages(pages []ParsedPage) []*StructNode {
	out := make([]*StructNode, 0, len(pages))
	for _, page := range pages {
		out = append(out, cloneStructNode(page.StructTree))
	}
	return out
}

func structTreeAt(trees []*StructNode, pageIndex int) *StructNode {
	if pageIndex < 0 || pageIndex >= len(trees) {
		return nil
	}
	return cloneStructNode(trees[pageIndex])
}

var (
	pageTypePattern       = regexp.MustCompile(`/Type\s*/Page\b`)
	mediaBoxPattern       = regexp.MustCompile(`/MediaBox\s*\[\s*([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)\s*\]`)
	literalStringPattern  = regexp.MustCompile(`\((?:\\.|[^\\()])*\)`)
	tjPattern             = regexp.MustCompile(`\((?:\\.|[^\\()])*\)\s*Tj`)
	tjArrayPattern        = regexp.MustCompile(`\[(?s:.*?)\]\s*TJ`)
	hexTjPattern          = regexp.MustCompile(`<([0-9A-Fa-f]+)>\s*Tj`)
	arrayTextTokenPattern = regexp.MustCompile(`\((?:\\.|[^\\()])*\)|<([0-9A-Fa-f]+)>`)
	bfcharBlockPattern    = regexp.MustCompile(`(?s)\d+\s+beginbfchar(.*?)endbfchar`)
	bfrangeBlockPattern   = regexp.MustCompile(`(?s)\d+\s+beginbfrange(.*?)endbfrange`)
	bfcharMappingPattern  = regexp.MustCompile(`<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>`)
	bfrangeMappingPattern = regexp.MustCompile(`<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>\s*(\[[^\]]+\]|<([0-9A-Fa-f]+)>)`)
	hexValuePattern       = regexp.MustCompile(`<([0-9A-Fa-f]+)>`)
)

func shellPagesFromPDF(data []byte) []model.PageMetadata {
	if len(data) == 0 {
		return nil
	}

	pageCount := len(pageTypePattern.FindAll(data, -1))
	sizes := extractMediaBoxSizes(data)
	if pageCount == 0 || len(sizes) == 0 {
		for _, decoded := range extractDecodedStreams(data) {
			if pageCount == 0 {
				pageCount += len(pageTypePattern.FindAll(decoded, -1))
			}
			if len(sizes) == 0 {
				sizes = append(sizes, extractMediaBoxSizes(decoded)...)
			}
		}
	}
	if pageCount == 0 {
		return nil
	}
	pages := make([]model.PageMetadata, 0, pageCount)
	for i := 0; i < pageCount; i++ {
		size := model.PageSize{}
		bounds := model.Box{}
		if i < len(sizes) {
			size = sizes[i]
			if size.Width > 0 || size.Height > 0 {
				bounds = model.Box{Left: 0, Bottom: 0, Right: size.Width, Top: size.Height}
			}
		} else if len(sizes) > 0 {
			size = sizes[len(sizes)-1]
			if size.Width > 0 || size.Height > 0 {
				bounds = model.Box{Left: 0, Bottom: 0, Right: size.Width, Top: size.Height}
			}
		}

		pages = append(pages, model.PageMetadata{
			Index:  model.PageIndex(i),
			Number: model.PageNumber(i + 1),
			Size:   size,
			Bounds: bounds,
		})
	}
	return pages
}

func extractMediaBoxSizes(data []byte) []model.PageSize {
	matches := mediaBoxPattern.FindAllSubmatch(data, -1)
	if len(matches) == 0 {
		return nil
	}

	out := make([]model.PageSize, 0, len(matches))
	for _, match := range matches {
		if len(match) != 5 {
			continue
		}
		left, ok := parseFloat(match[1])
		if !ok {
			continue
		}
		bottom, ok := parseFloat(match[2])
		if !ok {
			continue
		}
		right, ok := parseFloat(match[3])
		if !ok {
			continue
		}
		top, ok := parseFloat(match[4])
		if !ok {
			continue
		}
		out = append(out, model.PageSize{
			Width:  math.Max(0, right-left),
			Height: math.Max(0, top-bottom),
		})
	}
	return out
}

func parseFloat(raw []byte) (float64, bool) {
	value, err := strconv.ParseFloat(string(raw), 64)
	if err != nil {
		return 0, false
	}
	return value, true
}

func shellArtifactsFromPDF(data []byte, pages []model.PageMetadata) [][]*model.RawArtifact {
	if len(pages) == 0 {
		return nil
	}
	if positioned := extractPositionedTextArtifacts(data, pages); hasUsablePositionedArtifacts(positioned, pages) {
		return positioned
	}
	texts := extractTextShellStrings(data)
	artifactsByPage := make([][]*model.RawArtifact, len(pages))
	if len(texts) == 0 {
		return artifactsByPage
	}

	chunks := distributeTextAcrossPages(texts, len(pages))
	for pageIndex, pageTexts := range chunks {
		pageMeta := pages[pageIndex]
		artifacts := make([]*model.RawArtifact, 0, len(pageTexts))
		for i, text := range pageTexts {
			if strings.TrimSpace(text) == "" {
				continue
			}
			artifacts = append(artifacts, &model.RawArtifact{
				ID:         model.ArtifactID(i + 1),
				Kind:       model.ArtifactKindText,
				PageIndex:  pageMeta.Index,
				PageNumber: pageMeta.Number,
				Sequence:   i,
				Bounds:     shellBoundsForText(pageMeta, i, text),
				Text:       text,
				Style: model.TextProperties{
					Font:     "skeleton",
					FontSize: 12,
					Content:  text,
				},
			})
		}
		artifactsByPage[pageIndex] = artifacts
	}
	if hasArtifacts(artifactsByPage) {
		return artifactsByPage
	}
	if len(pages) == 1 {
		if texts := extractLiteralStrings(data); len(texts) > 0 {
			pageMeta := pages[0]
			artifacts := make([]*model.RawArtifact, 0, len(texts))
			for i, text := range texts {
				artifacts = append(artifacts, &model.RawArtifact{
					ID:         model.ArtifactID(i + 1),
					Kind:       model.ArtifactKindText,
					PageIndex:  pageMeta.Index,
					PageNumber: pageMeta.Number,
					Sequence:   i,
					Bounds:     shellBoundsForText(pageMeta, i, text),
					Text:       text,
					Style: model.TextProperties{
						Font:     "skeleton",
						FontSize: 12,
						Content:  text,
					},
				})
			}
			return [][]*model.RawArtifact{artifacts}
		}
	}
	return artifactsByPage
}

func tableByPageFromParsedPages(pages []ParsedPage) []*TableCandidateSet {
	if len(pages) == 0 {
		return nil
	}
	out := make([]*TableCandidateSet, len(pages))
	for i, page := range pages {
		out[i] = cloneTableCandidateSet(page.TableCandidates)
	}
	return out
}

func tableCandidateAt(tables []*TableCandidateSet, pageIndex int) *TableCandidateSet {
	if pageIndex < 0 || pageIndex >= len(tables) {
		return nil
	}
	return cloneTableCandidateSet(tables[pageIndex])
}

func hasArtifacts(pages [][]*model.RawArtifact) bool {
	for _, page := range pages {
		if len(page) > 0 {
			return true
		}
	}
	return false
}

func hasUsablePositionedArtifacts(pages [][]*model.RawArtifact, metadata []model.PageMetadata) bool {
	if len(pages) == 0 {
		return false
	}

	for pageIndex, artifacts := range pages {
		if len(artifacts) == 0 {
			continue
		}
		pageMeta := model.PageMetadata{}
		if pageIndex >= 0 && pageIndex < len(metadata) {
			pageMeta = metadata[pageIndex]
		}
		pageBounds, ok := pageBoxFromMetadata(pageMeta)
		if !ok {
			if len(artifacts) > 0 {
				return true
			}
			continue
		}
		for _, artifact := range artifacts {
			if artifact == nil || strings.TrimSpace(artifact.Text) == "" {
				continue
			}
			if !artifact.Bounds.IsZero() && artifact.Bounds.Intersects(pageBounds) {
				return true
			}
		}
	}

	return false
}

func extractTextShellStrings(data []byte) []string {
	glyphMap := extractToUnicodeMap(data)
	streamTexts := extractStringsFromStreams(data, glyphMap)
	if len(streamTexts) > 0 {
		return dedupeStrings(streamTexts)
	}

	texts := extractContentStreamStrings(data, glyphMap)
	if len(texts) > 0 {
		return dedupeStrings(texts)
	}

	if texts := extractLiteralStringsFromDecodedStreams(data); len(texts) > 0 {
		return dedupeStrings(texts)
	}

	// Raw literal fallback is only safe for tiny inline pseudo-PDF test inputs.
	// Real PDFs usually carry metadata and object-stream strings that would
	// overwhelm the shell with garbage if we fell back indiscriminately.
	if bytes.Contains(data, []byte("stream")) {
		return nil
	}
	return dedupeStrings(extractLiteralStrings(data))
}

func extractLiteralStrings(data []byte) []string {
	matches := literalStringPattern.FindAll(data, -1)
	if len(matches) == 0 {
		return nil
	}

	out := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		text := string(match[1 : len(match)-1])
		text = strings.ReplaceAll(text, `\(`, `(`)
		text = strings.ReplaceAll(text, `\)`, `)`)
		text = strings.ReplaceAll(text, `\\`, `\`)
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		out = append(out, text)
	}
	return out
}

func extractContentStreamStrings(data []byte, glyphMap map[string]string) []string {
	if len(data) == 0 {
		return nil
	}
	if out := extractContentTextByOperators(data, glyphMap); len(out) > 0 {
		return out
	}
	if out := extractContentTextByRegex(data, glyphMap); len(out) > 0 {
		return out
	}
	return nil
}

func extractContentTextByRegex(data []byte, glyphMap map[string]string) []string {
	if len(data) == 0 {
		return nil
	}

	out := make([]string, 0)
	for _, match := range tjPattern.FindAll(data, -1) {
		for _, text := range extractLiteralStrings(match) {
			out = append(out, text)
		}
	}
	for _, match := range tjArrayPattern.FindAll(data, -1) {
		if text := extractTJArrayString(match, glyphMap); text != "" {
			out = append(out, text)
		}
	}
	for _, match := range hexTjPattern.FindAllSubmatch(data, -1) {
		if len(match) < 2 {
			continue
		}
		if decoded := decodeHexGlyphString(match[1], glyphMap); decoded != "" {
			out = append(out, decoded)
		}
	}
	return out
}

func extractTJArrayString(data []byte, glyphMap map[string]string) string {
	tokens := arrayTextTokenPattern.FindAll(data, -1)
	if len(tokens) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, token := range tokens {
		if len(token) == 0 {
			continue
		}
		switch token[0] {
		case '(':
			literals := extractLiteralStrings(token)
			for _, text := range literals {
				builder.WriteString(text)
			}
		case '<':
			builder.WriteString(decodeHexGlyphString(token[1:len(token)-1], glyphMap))
		}
	}
	return strings.TrimSpace(builder.String())
}

func extractLiteralStringsFromDecodedStreams(data []byte) []string {
	decodedStreams := extractDecodedStreams(data)
	if len(decodedStreams) == 0 {
		return nil
	}
	out := make([]string, 0)
	for _, decoded := range decodedStreams {
		out = append(out, extractLiteralStrings(decoded)...)
	}
	return out
}

func extractStringsFromStreams(data []byte, glyphMap map[string]string) []string {
	if len(data) == 0 {
		return nil
	}

	out := make([]string, 0)
	for _, decoded := range extractDecodedStreams(data) {
		out = append(out, extractContentStreamStrings(decoded, glyphMap)...)
	}
	return out
}

func extractDecodedStreams(data []byte) [][]byte {
	if len(data) == 0 {
		return nil
	}

	out := make([][]byte, 0)
	search := data
	offset := 0
	for {
		streamIndex := bytes.Index(search, []byte("stream"))
		if streamIndex < 0 {
			break
		}
		streamIndex += offset
		streamStart := streamIndex + len("stream")
		streamStart = skipStreamNewline(data, streamStart)
		endRel := bytes.Index(data[streamStart:], []byte("endstream"))
		if endRel < 0 {
			break
		}
		streamEnd := streamStart + endRel
		streamData := trimStreamData(data[streamStart:streamEnd])
		dict := surroundingDictionary(data, streamIndex)

		decoded := streamData
		if bytes.Contains(dict, []byte("/FlateDecode")) {
			inflated, ok := inflateStream(streamData)
			if ok {
				decoded = inflated
			}
		}
		out = append(out, decoded)

		offset = streamEnd + len("endstream")
		if offset >= len(data) {
			break
		}
		search = data[offset:]
	}
	return out
}

func skipStreamNewline(data []byte, index int) int {
	if index >= len(data) {
		return index
	}
	if data[index] == '\r' {
		index++
	}
	if index < len(data) && data[index] == '\n' {
		index++
	}
	return index
}

func trimStreamData(data []byte) []byte {
	data = bytes.TrimPrefix(data, []byte("\r"))
	data = bytes.TrimPrefix(data, []byte("\n"))
	data = bytes.TrimSuffix(data, []byte("\r"))
	data = bytes.TrimSuffix(data, []byte("\n"))
	return data
}

func surroundingDictionary(data []byte, streamIndex int) []byte {
	start := streamIndex - 512
	if start < 0 {
		start = 0
	}
	window := data[start:streamIndex]
	dictStart := bytes.LastIndex(window, []byte("<<"))
	if dictStart < 0 {
		return nil
	}
	return window[dictStart:]
}

func inflateStream(data []byte) ([]byte, bool) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, false
	}
	defer reader.Close()
	decoded, err := readAllBounded(reader, maxDecodedStreamBytes)
	if err != nil {
		return nil, false
	}
	return decoded, true
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func extractToUnicodeMap(data []byte) map[string]string {
	decodedStreams := extractDecodedStreams(data)
	if len(decodedStreams) == 0 {
		return nil
	}

	out := make(map[string]string)
	for _, decoded := range decodedStreams {
		for _, block := range bfcharBlockPattern.FindAllSubmatch(decoded, -1) {
			if len(block) < 2 {
				continue
			}
			mergeUnicodeMappings(out, block[1])
		}
		for _, block := range bfrangeBlockPattern.FindAllSubmatch(decoded, -1) {
			if len(block) < 2 {
				continue
			}
			mergeUnicodeRangeMappings(out, block[1])
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func decodeHexGlyphString(hexBytes []byte, glyphMap map[string]string) string {
	hex := strings.ToUpper(strings.TrimSpace(string(hexBytes)))
	if hex == "" {
		return ""
	}
	if glyphMap != nil {
		widths := glyphCodeWidths(glyphMap)
		var builder strings.Builder
		for i := 0; i < len(hex); {
			matched := false
			for _, width := range widths {
				if width <= 0 || i+width > len(hex) {
					continue
				}
				if mapped, ok := glyphMap[hex[i:i+width]]; ok {
					builder.WriteString(mapped)
					i += width
					matched = true
					break
				}
			}
			if !matched {
				break
			}
		}
		if builder.Len() > 0 {
			return builder.String()
		}
	}
	return decodeUnicodeHex(hexBytes)
}

func decodeUnicodeHex(hexBytes []byte) string {
	hex := strings.TrimSpace(string(hexBytes))
	if hex == "" || len(hex)%4 != 0 {
		return ""
	}
	runes := make([]rune, 0, len(hex)/4)
	for i := 0; i+4 <= len(hex); i += 4 {
		value, err := strconv.ParseUint(hex[i:i+4], 16, 16)
		if err != nil {
			return ""
		}
		if value == 0xFFFF || value == 0x0000 {
			continue
		}
		runes = append(runes, rune(value))
	}
	return string(runes)
}

func mergeUnicodeMappings(out map[string]string, block []byte) {
	for _, mapping := range bfcharMappingPattern.FindAllSubmatch(block, -1) {
		if len(mapping) < 3 {
			continue
		}
		src := strings.ToUpper(string(mapping[1]))
		dst := decodeUnicodeHex(mapping[2])
		if src == "" || dst == "" {
			continue
		}
		out[src] = dst
	}
}

func mergeUnicodeRangeMappings(out map[string]string, block []byte) {
	for _, mapping := range bfrangeMappingPattern.FindAllSubmatch(block, -1) {
		if len(mapping) < 4 {
			continue
		}
		startHex := strings.ToUpper(string(mapping[1]))
		endHex := strings.ToUpper(string(mapping[2]))
		targetSpec := strings.TrimSpace(string(mapping[3]))
		if startHex == "" || endHex == "" || targetSpec == "" {
			continue
		}
		startValue, err := strconv.ParseUint(startHex, 16, 64)
		if err != nil {
			continue
		}
		endValue, err := strconv.ParseUint(endHex, 16, 64)
		if err != nil || endValue < startValue {
			continue
		}

		if strings.HasPrefix(targetSpec, "[") {
			matched := hexValuePattern.FindAllSubmatch([]byte(targetSpec), -1)
			if len(matched) == 0 {
				continue
			}
			for i, entry := range matched {
				if len(entry) < 2 {
					continue
				}
				key := strings.ToUpper(padHex(startValue+uint64(i), len(startHex)))
				out[key] = decodeUnicodeHex(entry[1])
			}
			continue
		}

		targetHex := strings.Trim(targetSpec, "<>")
		if targetHex == "" {
			continue
		}
		targetStart, err := strconv.ParseUint(targetHex, 16, 64)
		if err != nil {
			continue
		}
		for code := startValue; code <= endValue; code++ {
			key := strings.ToUpper(padHex(code, len(startHex)))
			out[key] = string(rune(targetStart + (code - startValue)))
		}
	}
}

func padHex(value uint64, width int) string {
	if width <= 0 {
		width = 2
	}
	return fmt.Sprintf("%0*X", width, value)
}

func glyphCodeWidths(glyphMap map[string]string) []int {
	if len(glyphMap) == 0 {
		return nil
	}
	widthSet := make(map[int]struct{})
	for key := range glyphMap {
		if key == "" {
			continue
		}
		widthSet[len(key)] = struct{}{}
	}
	if len(widthSet) == 0 {
		return nil
	}
	widths := make([]int, 0, len(widthSet))
	for width := range widthSet {
		widths = append(widths, width)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(widths)))
	return widths
}

func distributeTextAcrossPages(texts []string, pageCount int) [][]string {
	out := make([][]string, pageCount)
	if pageCount == 0 || len(texts) == 0 {
		return out
	}

	for i, text := range texts {
		pageIndex := i * pageCount / len(texts)
		if pageIndex >= pageCount {
			pageIndex = pageCount - 1
		}
		out[pageIndex] = append(out[pageIndex], text)
	}
	return out
}

func shellBoundsForText(page model.PageMetadata, index int, text string) model.Box {
	height := 12.0
	lineGap := 18.0
	left := 48.0
	top := page.Size.Height - 48 - float64(index)*lineGap
	if top <= 0 {
		top = page.Bounds.Top - 48 - float64(index)*lineGap
	}
	if top <= 0 {
		top = 800 - 48 - float64(index)*lineGap
	}
	bottom := top - height
	width := math.Max(48, float64(6*len([]rune(text))))
	return model.Box{
		Left:   left,
		Bottom: bottom,
		Right:  left + width,
		Top:    top,
	}
}
