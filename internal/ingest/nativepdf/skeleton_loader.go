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
}

// NewSkeletonLoader returns the current native PDF loader skeleton.
func NewSkeletonLoader() *SkeletonLoader { return &SkeletonLoader{} }

// NewSkeletonLoaderWithPages returns a skeleton loader preloaded with page
// metadata shells for tests and future parser bring-up work.
func NewSkeletonLoaderWithPages(pages []model.PageMetadata) *SkeletonLoader {
	return &SkeletonLoader{
		pageMetadata: append([]model.PageMetadata(nil), pages...),
	}
}

// OpenPath reads the PDF bytes from disk and returns a document handle shell.
func (l *SkeletonLoader) OpenPath(_ context.Context, path string, _ OpenOptions) (DocumentHandle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read pdf source: %w", err)
	}
	return l.open(filepath.Base(path), data), nil
}

// OpenReader reads the PDF bytes from a stream and returns a document handle shell.
func (l *SkeletonLoader) OpenReader(_ context.Context, name string, r io.Reader, _ OpenOptions) (DocumentHandle, error) {
	if r == nil {
		return nil, fmt.Errorf("native skeleton reader is nil")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read pdf source: %w", err)
	}
	return l.open(name, data), nil
}

func (l *SkeletonLoader) open(name string, data []byte) DocumentHandle {
	pages := make([]model.PageMetadata, 0, len(l.pageMetadata))
	sourcePages := l.pageMetadata
	if len(sourcePages) == 0 {
		sourcePages = shellPagesFromPDF(data)
	}
	for i, page := range sourcePages {
		if page.Index == 0 && i > 0 {
			page.Index = model.PageIndex(i)
		}
		if page.Number <= 0 {
			page.Number = model.PageNumber(i + 1)
		}
		pages = append(pages, page)
	}
	artifactsByPage := shellArtifactsFromPDF(data, pages)
	return &skeletonDocumentHandle{
		metadata: model.DocumentMetadata{
			FileName:  name,
			PageCount: len(pages),
		},
		raw:             append([]byte(nil), data...),
		pages:           pages,
		artifactsByPage: artifactsByPage,
	}
}

type skeletonDocumentHandle struct {
	metadata        model.DocumentMetadata
	raw             []byte
	pages           []model.PageMetadata
	artifactsByPage [][]*model.RawArtifact
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
		metadata:  h.pages[pageIndex],
		artifacts: artifacts,
	}, nil
}

func (h *skeletonDocumentHandle) Close() error {
	h.raw = nil
	return nil
}

type skeletonPageHandle struct {
	metadata  model.PageMetadata
	artifacts []*model.RawArtifact
}

func (h *skeletonPageHandle) Metadata() model.PageMetadata {
	return h.metadata
}

func (h *skeletonPageHandle) Artifacts(_ context.Context, _ ArtifactOptions) ([]*model.RawArtifact, error) {
	return cloneArtifacts(h.artifacts), nil
}

func (h *skeletonPageHandle) TableCandidates(_ context.Context) (*TableCandidateSet, error) {
	return nil, nil
}

func (h *skeletonPageHandle) StructTree(_ context.Context) (*StructNode, error) {
	return nil, nil
}

var (
	pageTypePattern = regexp.MustCompile(`/Type\s*/Page\b`)
	mediaBoxPattern = regexp.MustCompile(`/MediaBox\s*\[\s*([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)\s*\]`)
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
	return artifactsByPage
}

var literalStringPattern = regexp.MustCompile(`\((?:\\.|[^\\()])*\)`)
var tjPattern = regexp.MustCompile(`\((?:\\.|[^\\()])*\)\s*Tj`)
var tjArrayPattern = regexp.MustCompile(`\[(?s:.*?)\]\s*TJ`)

func extractTextShellStrings(data []byte) []string {
	streamTexts := extractStringsFromStreams(data)
	if len(streamTexts) > 0 {
		return dedupeStrings(streamTexts)
	}

	texts := extractContentStreamStrings(data)
	if len(texts) > 0 {
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

func extractContentStreamStrings(data []byte) []string {
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
		for _, text := range extractLiteralStrings(match) {
			out = append(out, text)
		}
	}
	return out
}

func extractStringsFromStreams(data []byte) []string {
	if len(data) == 0 {
		return nil
	}

	out := make([]string, 0)
	for _, decoded := range extractDecodedStreams(data) {
		out = append(out, extractContentStreamStrings(decoded)...)
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
	decoded, err := io.ReadAll(reader)
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
