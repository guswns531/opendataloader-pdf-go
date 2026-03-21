package pagefilter

import (
	"fmt"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// Apply filters the document to the selected 1-based page numbers.
//
// The document is updated in place. Page ordering follows the selection order,
// document metadata is preserved, and document-level kids and artifacts are
// kept in the same page order.
func Apply(document *model.Document, selections []model.PageNumber) error {
	if document == nil {
		return fmt.Errorf("page filter requires a document")
	}
	if len(selections) == 0 {
		return nil
	}

	pagesByNumber := make(map[model.PageNumber]*model.Page, len(document.Pages))
	for _, page := range document.Pages {
		if page == nil {
			continue
		}
		pagesByNumber[page.Metadata.Number] = page
	}

	filteredPages := make([]*model.Page, 0, len(selections))
	for _, number := range selections {
		if number < 1 {
			return fmt.Errorf("invalid page selection %d: page numbers are 1-based", number)
		}
		page, ok := pagesByNumber[number]
		if !ok {
			return fmt.Errorf("page %d is not present in the document", number)
		}
		filteredPages = append(filteredPages, page)
	}

	filteredKids := filterContentElements(document.Kids, selections)
	filteredArtifacts := filterArtifacts(document.Artifacts, selections)

	document.Pages = filteredPages
	document.Kids = filteredKids
	document.Artifacts = filteredArtifacts
	document.Metadata.PageCount = len(filteredPages)

	return nil
}

func filterContentElements(elements []model.ContentElement, selections []model.PageNumber) []model.ContentElement {
	if len(elements) == 0 {
		return nil
	}

	byPage := make(map[model.PageNumber][]model.ContentElement)
	for _, element := range elements {
		if element == nil {
			continue
		}
		base := element.NodeBase()
		if base == nil {
			continue
		}
		byPage[base.PageNumber] = append(byPage[base.PageNumber], element)
	}

	filtered := make([]model.ContentElement, 0, len(elements))
	for _, number := range selections {
		filtered = append(filtered, byPage[number]...)
	}
	return filtered
}

func filterArtifacts(artifacts []*model.RawArtifact, selections []model.PageNumber) []*model.RawArtifact {
	if len(artifacts) == 0 {
		return nil
	}

	byPage := make(map[model.PageNumber][]*model.RawArtifact)
	for _, artifact := range artifacts {
		if artifact == nil {
			continue
		}
		byPage[artifact.PageNumber] = append(byPage[artifact.PageNumber], artifact)
	}

	filtered := make([]*model.RawArtifact, 0, len(artifacts))
	for _, number := range selections {
		filtered = append(filtered, byPage[number]...)
	}
	return filtered
}
