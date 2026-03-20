package local

import (
	"fmt"
	"io"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/heuristics/headerfooter"
	"github.com/guswns531/opendataloader-pdf-go/internal/heuristics/heading"
	"github.com/guswns531/opendataloader-pdf-go/internal/heuristics/lists"
	"github.com/guswns531/opendataloader-pdf-go/internal/heuristics/paragraph"
	"github.com/guswns531/opendataloader-pdf-go/internal/heuristics/readingorder"
	"github.com/guswns531/opendataloader-pdf-go/internal/heuristics/table"
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
	applyParagraphAssembly(document)
	applyHeaderFooterFiltering(document)
	applyHeadingDetection(document)
	applyTableDetection(document)
	applyListDetection(document)
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

func applyParagraphAssembly(document *model.Document) {
	if document == nil {
		return
	}

	for _, page := range document.Pages {
		if page == nil || len(page.Kids) > 0 || len(page.Artifacts) == 0 {
			continue
		}

		grouped := textheur.GroupArtifactsToParagraphs(page.Artifacts)
		paragraphs := paragraph.Assemble(document, grouped)
		page.Kids = make([]model.ContentElement, 0, len(paragraphs))
		for _, node := range paragraphs {
			page.Kids = append(page.Kids, node)
		}
	}
}

func applyHeadingDetection(document *model.Document) {
	if document == nil {
		return
	}

	for _, page := range document.Pages {
		if page == nil || len(page.Kids) == 0 {
			continue
		}

		detections := heading.Detect(page.Kids)
		if len(detections) == 0 {
			continue
		}

		replacements := make(map[model.ContentElement]model.ContentElement, len(detections))
		for _, detection := range detections {
			if detection.Source == nil || detection.Heading == nil {
				continue
			}
			replacements[detection.Source] = detection.Heading
		}

		for i, element := range page.Kids {
			if replacement, ok := replacements[element]; ok {
				page.Kids[i] = replacement
			}
		}
	}
}

func applyHeaderFooterFiltering(document *model.Document) {
	if document == nil {
		return
	}
	detector := headerfooter.NewDetector()
	detector.Apply(document, headerfooter.ModeRemove)
}

func applyTableDetection(document *model.Document) {
	if document == nil {
		return
	}

	for _, page := range document.Pages {
		if page == nil || len(page.Kids) == 0 {
			continue
		}
		page.Kids = table.Detect(page.Kids)
	}
}

func applyListDetection(document *model.Document) {
	if document == nil {
		return
	}

	for _, page := range document.Pages {
		if page == nil || len(page.Kids) == 0 {
			continue
		}
		page.Kids = lists.Detect(page.Kids)
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
