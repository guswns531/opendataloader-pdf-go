package nativepdf

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// Ingestor adapts a native PDF loader to the existing core.Ingestor seam.
type Ingestor struct {
	Loader Loader
}

// NewIngestor returns a native PDF ingestor backed by the provided loader.
func NewIngestor(loader Loader) *Ingestor {
	return &Ingestor{Loader: loader}
}

// Name identifies the ingestor.
func (i *Ingestor) Name() string {
	return "nativepdf"
}

// Ingest opens the source with the configured loader and converts the result
// into the current Go document model.
func (i *Ingestor) Ingest(ctx *core.ProcessingContext, source core.Source) (*core.Document, error) {
	if i == nil || i.Loader == nil {
		return nil, fmt.Errorf("nativepdf ingest requires loader")
	}

	loadCtx := context.Background()
	if ctx != nil {
		loadCtx = context.WithValue(loadCtx, contextKeyProcessingContext{}, ctx)
	}

	opts := openOptionsFromProcessingContext(ctx)
	handle, err := i.open(loadCtx, source, opts)
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	document, err := BuildDocumentFromHandle(handle, source)
	if err != nil {
		return nil, err
	}
	if err := writeRawDumpIfRequested(document); err != nil {
		return nil, err
	}
	return document, nil
}

// BuildDocumentFromHandle converts a native document handle into the current
// in-memory document model used by the rest of the pipeline.
func BuildDocumentFromHandle(handle DocumentHandle, source core.Source) (*core.Document, error) {
	if handle == nil {
		return nil, fmt.Errorf("nativepdf document handle is nil")
	}

	metadata := handle.Metadata()
	if metadata.FileName == "" {
		metadata.FileName = resolveSourceName(source)
	}

	pageCount := handle.PageCount()
	if metadata.PageCount == 0 {
		metadata.PageCount = pageCount
	}

	document := model.NewDocument(metadata)
	for pageIndex := 0; pageIndex < pageCount; pageIndex++ {
		pageHandle, err := handle.Page(pageIndex)
		if err != nil {
			return nil, fmt.Errorf("load page %d: %w", pageIndex, err)
		}
		page, err := buildPage(document, pageIndex, pageHandle)
		if err != nil {
			return nil, err
		}
		document.AddPage(page)
	}

	if document.Metadata.PageCount == 0 {
		document.Metadata.PageCount = len(document.Pages)
	}

	return document, nil
}

type contextKeyProcessingContext struct{}

func (i *Ingestor) open(ctx context.Context, source core.Source, opts OpenOptions) (DocumentHandle, error) {
	switch {
	case source.Reader != nil:
		return i.Loader.OpenReader(ctx, resolveSourceName(source), source.Reader, opts)
	case source.Path != "":
		return i.Loader.OpenPath(ctx, source.Path, opts)
	default:
		return nil, fmt.Errorf("nativepdf ingest requires source path or reader")
	}
}

func buildPage(document *model.Document, pageIndex int, pageHandle PageHandle) (*model.Page, error) {
	if pageHandle == nil {
		return nil, fmt.Errorf("nativepdf page handle is nil")
	}

	metadata := pageHandle.Metadata()
	metadata.Index = model.PageIndex(pageIndex)
	if metadata.Number <= 0 {
		metadata.Number = model.PageNumber(pageIndex + 1)
	}

	artifacts, err := pageHandle.Artifacts(context.Background(), ArtifactOptions{
		IncludeText:  true,
		IncludeImage: true,
		IncludeLine:  true,
		IncludePath:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("load artifacts for page %d: %w", pageIndex, err)
	}
	tableCandidates, err := pageHandle.TableCandidates(context.Background())
	if err != nil {
		return nil, fmt.Errorf("load table candidates for page %d: %w", pageIndex, err)
	}

	page := &model.Page{
		Metadata:  metadata,
		Artifacts: make([]*model.RawArtifact, 0, len(artifacts)+countTableCandidateArtifacts(tableCandidates)),
		Kids:      make([]model.ContentElement, 0),
	}
	maxSequence := -1
	for _, artifact := range artifacts {
		if artifact == nil {
			continue
		}
		cloned := cloneArtifact(artifact)
		if cloned.ID == 0 && document != nil {
			cloned.ID = document.NewArtifactID()
		}
		cloned.PageIndex = metadata.Index
		cloned.PageNumber = metadata.Number
		page.Artifacts = append(page.Artifacts, cloned)
		if cloned.Sequence > maxSequence {
			maxSequence = cloned.Sequence
		}
	}
	for _, artifact := range tableCandidateArtifacts(tableCandidates, metadata, maxSequence+1) {
		if artifact == nil {
			continue
		}
		if artifact.ID == 0 && document != nil {
			artifact.ID = document.NewArtifactID()
		}
		page.Artifacts = append(page.Artifacts, artifact)
	}

	return page, nil
}

func countTableCandidateArtifacts(set *TableCandidateSet) int {
	if set == nil {
		return 0
	}
	return len(set.HorizontalLines) + len(set.VerticalLines) + len(set.Rectangles)
}

func tableCandidateArtifacts(set *TableCandidateSet, page model.PageMetadata, sequenceStart int) []*model.RawArtifact {
	if set == nil {
		return nil
	}

	artifacts := make([]*model.RawArtifact, 0, countTableCandidateArtifacts(set))
	sequence := sequenceStart
	for _, segment := range set.HorizontalLines {
		artifact := lineArtifactFromSegment(segment, page, sequence, nil)
		artifact.Style.Content = ""
		artifacts = append(artifacts, artifact)
		sequence++
	}
	for _, segment := range set.VerticalLines {
		artifact := lineArtifactFromSegment(segment, page, sequence, nil)
		artifact.Style.Content = ""
		artifacts = append(artifacts, artifact)
		sequence++
	}
	for _, box := range set.Rectangles {
		normalized := box.Normalize()
		if normalized.IsZero() {
			continue
		}
		artifact := pathArtifactFromBox(normalized, page, sequence, nil)
		artifact.Style.Content = ""
		artifacts = append(artifacts, artifact)
		sequence++
	}
	return artifacts
}

func cloneArtifact(artifact *model.RawArtifact) *model.RawArtifact {
	cloned := *artifact
	if artifact.MarkedContentID != nil {
		mcid := *artifact.MarkedContentID
		cloned.MarkedContentID = &mcid
	}
	if len(artifact.Boxes) > 0 {
		cloned.Boxes = append(model.MultiBox(nil), artifact.Boxes...)
	}
	if len(artifact.Data) > 0 {
		cloned.Data = append([]byte(nil), artifact.Data...)
	}
	if len(artifact.Filters) > 0 {
		cloned.Filters = append([]string(nil), artifact.Filters...)
	}
	return &cloned
}

func openOptionsFromProcessingContext(ctx *core.ProcessingContext) OpenOptions {
	if ctx == nil {
		return OpenOptions{}
	}

	opts := OpenOptions{
		Strict: ctx.Options.Strict,
	}
	if password, ok := stringExtra(ctx.Options.Extras, "password"); ok {
		opts.Password = password
	}
	if pages, ok := intSliceExtra(ctx.Options.Extras, "pages"); ok {
		opts.Pages = pages
	}
	return opts
}

func resolveSourceName(source core.Source) string {
	if source.Name != "" {
		return source.Name
	}
	if source.Path != "" {
		return filepath.Base(source.Path)
	}
	return ""
}

func stringExtra(extras map[string]any, key string) (string, bool) {
	if len(extras) == 0 {
		return "", false
	}
	raw, ok := extras[key]
	if !ok {
		return "", false
	}
	value, ok := raw.(string)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}

func intSliceExtra(extras map[string]any, key string) ([]int, bool) {
	if len(extras) == 0 {
		return nil, false
	}
	raw, ok := extras[key]
	if !ok {
		return nil, false
	}

	switch pages := raw.(type) {
	case []int:
		return append([]int(nil), pages...), true
	case []model.PageNumber:
		out := make([]int, 0, len(pages))
		for _, page := range pages {
			out = append(out, int(page))
		}
		return out, true
	default:
		return nil, false
	}
}
