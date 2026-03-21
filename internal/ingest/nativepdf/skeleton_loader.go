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
// sources and produces a real DocumentHandle, but page extraction is not
// implemented yet.
type SkeletonLoader struct{}

// NewSkeletonLoader returns the current native PDF loader skeleton.
func NewSkeletonLoader() *SkeletonLoader {
	return &SkeletonLoader{}
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
	return &skeletonDocumentHandle{
		metadata: model.DocumentMetadata{
			FileName:  name,
			PageCount: 0,
		},
		raw: append([]byte(nil), data...),
	}
}

type skeletonDocumentHandle struct {
	metadata model.DocumentMetadata
	raw      []byte
}

func (h *skeletonDocumentHandle) Metadata() model.DocumentMetadata {
	return h.metadata
}

func (h *skeletonDocumentHandle) PageCount() int {
	return 0
}

func (h *skeletonDocumentHandle) Page(pageIndex int) (PageHandle, error) {
	return nil, fmt.Errorf("native skeleton page %d is not implemented", pageIndex)
}

func (h *skeletonDocumentHandle) Close() error {
	h.raw = nil
	return nil
}
