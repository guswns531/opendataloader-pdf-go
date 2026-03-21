# Native Ingestion API Draft

This draft defines the minimum Pure Go ingestion boundary required to replace the
temporary `internal/ingest/pdftext` path and feed the existing document model.

## Design Goals

- no Java runtime dependency
- no expansion of `github.com/ledongthuc/pdf`
- stable raw artifacts for downstream fixture-driven processor work
- tagged-PDF hooks available in the shape of the API, but optional for MVP

## Scope For The First Native Ingestion Milestone

The first usable native ingestion layer only needs to support:

- digital PDFs
- document metadata
- page count, page size, page rotation
- raw text artifacts with stable geometry and style
- image and line/path artifacts required by layout heuristics
- password entry point in the API, even if encrypted PDFs are not solved on day one

It does not need to solve in the first milestone:

- full struct-tree extraction
- hybrid backend integration
- PDF writing
- formula/picture enrichments

## Proposed Package Boundary

- package: `internal/ingest/nativepdf`
- role: low-level Pure Go PDF loading and page artifact extraction
- output target: current `internal/model` types plus a small native-only raw layer where needed

## Proposed Interfaces

```go
package nativepdf

import (
    "context"
    "io"

    "github.com/guswns531/opendataloader-pdf-go/internal/model"
)

type OpenOptions struct {
    Password string
    Pages    []int
    Strict   bool
}

type Loader interface {
    OpenPath(ctx context.Context, path string, opts OpenOptions) (DocumentHandle, error)
    OpenReader(ctx context.Context, name string, r io.Reader, opts OpenOptions) (DocumentHandle, error)
}

type DocumentHandle interface {
    Metadata() model.DocumentMetadata
    PageCount() int
    Page(pageIndex int) (PageHandle, error)
    Close() error
}

type PageHandle interface {
    Metadata() model.PageMetadata
    Artifacts(ctx context.Context, opts ArtifactOptions) ([]*model.RawArtifact, error)
    TableCandidates(ctx context.Context) (*TableCandidateSet, error)
    StructTree(ctx context.Context) (*StructNode, error)
}

type ArtifactOptions struct {
    IncludeText  bool
    IncludeImage bool
    IncludeLine  bool
    IncludePath  bool
}
```

## Minimum Raw Signals Required

### Document Level

- file name
- author / title / creation / modification timestamps when available
- producer / creator / language / keywords where available
- total page count

### Page Level

- zero-based page index
- one-based page number
- page label if present
- width / height
- crop/media bounds
- rotation

### Artifact Level

- stable sequence order per page
- artifact kind: `text`, `image`, `line`, `path`, `shape`
- single bounding box and optional multi-box
- page index / page number
- text content for text artifacts
- style signals:
  - font name
  - font size
  - text color
  - bold / italic / underline when recoverable
  - hidden text flag when recoverable
- binary/image reference for images
- enough geometry for future table-border reconstruction

## Native-Only Types Needed Beside `model.RawArtifact`

The current model is close, but the ingestion layer will likely need temporary
native-only helper types before everything is collapsed into `model.RawArtifact`.

```go
type TableCandidateSet struct {
    HorizontalLines []LineSegment
    VerticalLines   []LineSegment
    Rectangles      []model.Box
}

type LineSegment struct {
    Start     model.Point
    End       model.Point
    Width     float64
    StrokeRGB string
    PageIndex model.PageIndex
}

type StructNode struct {
    Type       string
    PageIndex  *model.PageIndex
    Bounds     model.Box
    Kids       []*StructNode
    ArtifactIDs []model.ArtifactID
}
```

## Why This Boundary

- `internal/core.Ingestor` is already a good high-level seam for the pipeline.
- `internal/model` already has document, page, metadata, geometry, text properties, and raw artifacts.
- downstream heuristic work can keep targeting `model.Document` while native parser work matures underneath.
- table detection and tagged-PDF work need richer hooks than the current `pdftext` path provides.

## Reusable Existing Pieces

These existing packages should stay and be fed by the new ingestion layer:

- `internal/core/interfaces.go`
- `internal/model`
- `internal/pipeline/local`
- `internal/pipeline/control`
- `internal/emit/*`
- `internal/ingest/fixture`

## Immediate Implementation Order

1. Add the `internal/ingest/nativepdf` package with interfaces only.
2. Add fixture-backed tests for page metadata and raw artifact expectations.
3. Introduce a temporary adapter that converts native extraction output into `model.Document`.
4. Keep `pdftext` untouched except to prevent new feature work from landing there.
5. Once text artifacts are stable, move table-border and image extraction into the native path.
6. Remove `ledongthuc/pdf` from the active PDF ingestion path.
