package nativepdf

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	for i, page := range l.pageMetadata {
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
