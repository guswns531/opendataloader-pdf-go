package nativepdf

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

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
	return &skeletonDocumentHandle{
		metadata: model.DocumentMetadata{
			FileName:  name,
			PageCount: len(pages),
		},
		raw:   append([]byte(nil), data...),
		pages: pages,
	}
}

type skeletonDocumentHandle struct {
	metadata model.DocumentMetadata
	raw      []byte
	pages    []model.PageMetadata
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
	return &skeletonPageHandle{metadata: h.pages[pageIndex]}, nil
}

func (h *skeletonDocumentHandle) Close() error {
	h.raw = nil
	return nil
}

type skeletonPageHandle struct {
	metadata model.PageMetadata
}

func (h *skeletonPageHandle) Metadata() model.PageMetadata {
	return h.metadata
}

func (h *skeletonPageHandle) Artifacts(_ context.Context, _ ArtifactOptions) ([]*model.RawArtifact, error) {
	return nil, nil
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
	if pageCount == 0 {
		return nil
	}

	sizes := extractMediaBoxSizes(data)
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
