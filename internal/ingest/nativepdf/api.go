package nativepdf

import (
	"context"
	"io"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// OpenOptions describes the minimum loader knobs required by the Pure Go parser.
type OpenOptions struct {
	Password string
	Pages    []int
	Strict   bool
}

// Loader opens PDFs from a path or stream and returns a document handle.
type Loader interface {
	OpenPath(ctx context.Context, path string, opts OpenOptions) (DocumentHandle, error)
	OpenReader(ctx context.Context, name string, r io.Reader, opts OpenOptions) (DocumentHandle, error)
}

// DocumentHandle exposes document metadata and per-page access.
type DocumentHandle interface {
	Metadata() model.DocumentMetadata
	PageCount() int
	Page(pageIndex int) (PageHandle, error)
	Close() error
}

// PageHandle exposes raw page artifacts plus future table and struct-tree hooks.
type PageHandle interface {
	Metadata() model.PageMetadata
	Artifacts(ctx context.Context, opts ArtifactOptions) ([]*model.RawArtifact, error)
	TableCandidates(ctx context.Context) (*TableCandidateSet, error)
	StructTree(ctx context.Context) (*StructNode, error)
}

// ArtifactOptions selects which low-level artifact families to extract.
type ArtifactOptions struct {
	IncludeText  bool
	IncludeImage bool
	IncludeLine  bool
	IncludePath  bool
}

// TableCandidateSet carries the low-level geometry needed for future table reconstruction.
type TableCandidateSet struct {
	HorizontalLines []LineSegment
	VerticalLines   []LineSegment
	Rectangles      []model.Box
}

// LineSegment represents a low-level stroked segment on a page.
type LineSegment struct {
	Start     model.Point
	End       model.Point
	Width     float64
	StrokeRGB string
	PageIndex model.PageIndex
}

// StructNode is the minimal tagged-PDF hook shape for future struct-tree support.
type StructNode struct {
	Type             string
	PageIndex        *model.PageIndex
	Bounds           model.Box
	Kids             []*StructNode
	ArtifactIDs      []model.ArtifactID
	MarkedContentIDs []int
}
