package local

import (
	"fmt"
	"io"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/heuristics/readingorder"
	textheur "github.com/guswns531/opendataloader-pdf-go/internal/heuristics/text"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// Pipeline is a minimal local pipeline that can ingest document fixtures,
// derive paragraphs from raw text artifacts, and stabilize page reading order.
type Pipeline struct {
	ingestor core.Ingestor
}

// New creates a local pipeline with the provided ingestor.
func New(ingestor core.Ingestor) *Pipeline {
	return &Pipeline{ingestor: ingestor}
}

// Run executes the minimal local pipeline.
func (p *Pipeline) Run(ctx *core.ProcessingContext, source core.Source, emitter core.Emitter) (*core.Document, error) {
	if p == nil || p.ingestor == nil {
		return nil, fmt.Errorf("local pipeline requires an ingestor")
	}
	if ctx == nil {
		return nil, fmt.Errorf("local pipeline requires a processing context")
	}

	ctx.SetStage(core.StageIngestion)
	document, err := p.ingestor.Ingest(ctx, source)
	if err != nil {
		return nil, err
	}
	ctx.SetDocument(document)

	ctx.SetStage(core.StageHeuristics)
	applyTextGrouping(document)
	applyReadingOrder(document)
	rebuildDocumentKids(document)

	if emitter != nil {
		ctx.SetStage(core.StageEmission)
		if err := emitter.Emit(ctx, document, io.Discard); err != nil {
			return nil, err
		}
	}

	return document, nil
}

func applyTextGrouping(document *model.Document) {
	if document == nil {
		return
	}

	for _, page := range document.Pages {
		if page == nil || len(page.Kids) > 0 || len(page.Artifacts) == 0 {
			continue
		}

		paragraphs := textheur.GroupArtifactsToParagraphs(page.Artifacts)
		page.Kids = make([]model.ContentElement, 0, len(paragraphs))
		for _, paragraph := range paragraphs {
			id := document.NewNodeID()
			page.Kids = append(page.Kids, &model.Paragraph{
				TextNode: model.TextNode{
					BaseNode: model.BaseNode{
						ID:         id,
						Type:       model.ElementTypeParagraph,
						PageIndex:  paragraph.PageIndex,
						PageNumber: paragraph.PageNumber,
						Bounds:     paragraph.Bounds,
					},
					TextProperties: model.TextProperties{
						Content: paragraph.Text,
					},
				},
			})
		}
	}
}

func applyReadingOrder(document *model.Document) {
	if document == nil {
		return
	}
	for _, page := range document.Pages {
		if page == nil || len(page.Kids) == 0 {
			continue
		}
		page.Kids = readingorder.Sort(page.Kids)
	}
}

func rebuildDocumentKids(document *model.Document) {
	if document == nil {
		return
	}
	document.Kids = document.Kids[:0]
	for _, page := range document.Pages {
		if page == nil || len(page.Kids) == 0 {
			continue
		}
		document.Kids = append(document.Kids, page.Kids...)
	}
}
