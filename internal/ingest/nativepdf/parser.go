package nativepdf

import (
	"fmt"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// ParsePhase names the highest parser stage that produced the current result.
type ParsePhase string

const (
	ParsePhaseShell      ParsePhase = "shell"
	ParsePhaseContainer  ParsePhase = "container"
	ParsePhaseContent    ParsePhase = "content"
	ParsePhaseStructured ParsePhase = "structured"
)

// ParseResult is the canonical output of the native PDF parser.
//
// The immediate implementation still uses shell extraction for some fields, but
// the result shape is intentionally richer than the current implementation so
// future work can replace the internals without disturbing loader callers.
type ParseResult struct {
	Phase    ParsePhase
	Metadata model.DocumentMetadata
	Pages    []ParsedPage
}

// ParsedPage carries all page-level parser output before adaptation to the
// broader document model.
type ParsedPage struct {
	Metadata        model.PageMetadata
	Artifacts       []*model.RawArtifact
	TableCandidates *TableCandidateSet
	StructTree      *StructNode
}

// DocumentParser is the phased entry point for the future Pure Go PDF engine.
type DocumentParser struct{}

// NewDocumentParser returns the current native parser entry point.
func NewDocumentParser() *DocumentParser {
	return &DocumentParser{}
}

// Parse converts PDF bytes into the intermediate native representation.
//
// Today this method still relies on shell extraction helpers. Its purpose is to
// establish the stable seam where the real container/content parser will land.
func (p *DocumentParser) Parse(name string, data []byte, opts OpenOptions, injectedPages []model.PageMetadata) (*ParseResult, error) {
	if p == nil {
		return nil, fmt.Errorf("native pdf parser is nil")
	}

	phase := ParsePhaseShell
	var pagePlans []pagePlan
	graph, _ := scanIndirectObjects(data)
	pages := append([]model.PageMetadata(nil), injectedPages...)
	if len(pages) == 0 {
		if graph != nil {
			if rootRef, ok := graph.catalogPagesRef(); ok {
				if plans, ok := graph.resolvePagePlans(rootRef); ok && len(plans) > 0 {
					pagePlans = plans
					pages = make([]model.PageMetadata, 0, len(plans))
					for _, plan := range plans {
						pages = append(pages, plan.Metadata)
					}
					phase = ParsePhaseContainer
				}
			}
		}
		if len(pages) == 0 {
			if plans, ok := parsePagePlans(data); ok && len(plans) > 0 {
				pagePlans = plans
				pages = make([]model.PageMetadata, 0, len(plans))
				for _, plan := range plans {
					pages = append(pages, plan.Metadata)
				}
				phase = ParsePhaseContainer
			} else {
				pages = shellPagesFromPDF(data)
			}
		}
	}
	for i := range pages {
		if pages[i].Index == 0 && i > 0 {
			pages[i].Index = model.PageIndex(i)
		}
		if pages[i].Number <= 0 {
			pages[i].Number = model.PageNumber(i + 1)
		}
	}

	artifactsByPage := shellArtifactsFromPDF(data, pages)
	graphicsByPage := extractGraphicsData(data, pages, pagePlans, graph)
	structTreesByPage := resolveStructTreesByPage(graph, pagePlans)
	if len(pagePlans) > 0 {
		if plannedArtifacts := shellArtifactsFromPagePlans(data, pagePlans); hasArtifacts(plannedArtifacts) {
			artifactsByPage = plannedArtifacts
		}
	}
	parsedPages := make([]ParsedPage, 0, len(pages))
	for i, page := range pages {
		pageArtifacts := cloneArtifacts(artifactsByPageForIndex(artifactsByPage, i))
		if i < len(graphicsByPage) {
			pageArtifacts = append(pageArtifacts, cloneArtifacts(graphicsByPage[i].artifacts)...)
		}
		parsed := ParsedPage{
			Metadata:  page,
			Artifacts: pageArtifacts,
		}
		if i < len(graphicsByPage) {
			parsed.TableCandidates = cloneTableCandidateSet(graphicsByPage[i].tableCandidates)
		}
		if i < len(structTreesByPage) {
			parsed.StructTree = cloneStructNode(structTreesByPage[i])
			if parsed.StructTree != nil {
				phase = ParsePhaseStructured
			}
		}
		if includePage(opts.Pages, page.Number) {
			parsedPages = append(parsedPages, parsed)
		}
	}
	renumberParsedArtifacts(parsedPages)
	linkParsedStructTrees(parsedPages)

	return &ParseResult{
		Phase: phase,
		Metadata: model.DocumentMetadata{
			FileName:  name,
			PageCount: len(parsedPages),
		},
		Pages: parsedPages,
	}, nil
}

func artifactsByPageForIndex(pages [][]*model.RawArtifact, index int) []*model.RawArtifact {
	if index < 0 || index >= len(pages) {
		return nil
	}
	return pages[index]
}

func includePage(selected []int, number model.PageNumber) bool {
	if len(selected) == 0 {
		return true
	}
	for _, page := range selected {
		if page > 0 && model.PageNumber(page) == number {
			return true
		}
	}
	return false
}

func renumberParsedArtifacts(pages []ParsedPage) {
	for i := range pages {
		nextID := model.ArtifactID(1)
		for j, artifact := range pages[i].Artifacts {
			if artifact == nil {
				continue
			}
			if artifact.ID == 0 {
				artifact.ID = nextID
			}
			if artifact.ID >= nextID {
				nextID = artifact.ID + 1
			}
			artifact.Sequence = j
		}
	}
}

func linkParsedStructTrees(pages []ParsedPage) {
	for i := range pages {
		if pages[i].StructTree == nil || len(pages[i].Artifacts) == 0 {
			continue
		}
		mcidToArtifactIDs := make(map[int][]model.ArtifactID)
		for _, artifact := range pages[i].Artifacts {
			if artifact == nil || artifact.ID == 0 || artifact.MarkedContentID == nil {
				continue
			}
			mcid := *artifact.MarkedContentID
			mcidToArtifactIDs[mcid] = appendUniqueArtifactID(mcidToArtifactIDs[mcid], artifact.ID)
		}
		populateStructNodeArtifactIDs(pages[i].StructTree, mcidToArtifactIDs)
	}
}

func populateStructNodeArtifactIDs(node *StructNode, mcidToArtifactIDs map[int][]model.ArtifactID) []model.ArtifactID {
	if node == nil {
		return nil
	}
	ids := append([]model.ArtifactID(nil), node.ArtifactIDs...)
	for _, mcid := range node.MarkedContentIDs {
		for _, id := range mcidToArtifactIDs[mcid] {
			ids = appendUniqueArtifactID(ids, id)
		}
	}
	for _, kid := range node.Kids {
		for _, id := range populateStructNodeArtifactIDs(kid, mcidToArtifactIDs) {
			ids = appendUniqueArtifactID(ids, id)
		}
	}
	node.ArtifactIDs = ids
	return ids
}

func appendUniqueArtifactID(ids []model.ArtifactID, id model.ArtifactID) []model.ArtifactID {
	if id == 0 {
		return ids
	}
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}
