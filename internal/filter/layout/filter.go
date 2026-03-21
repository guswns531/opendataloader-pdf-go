package layout

import (
	"fmt"
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

const (
	defaultTinyFontSize  = 1.0
	defaultTinyBoxExtent = 1.0
	defaultOffPageMargin = 0.01
)

// Options controls geometry-based filtering thresholds.
type Options struct {
	TinyFontSize  float64
	TinyBoxExtent float64
	OffPageMargin float64
}

// DefaultOptions returns the pragmatic defaults used by Apply.
func DefaultOptions() Options {
	return Options{
		TinyFontSize:  defaultTinyFontSize,
		TinyBoxExtent: defaultTinyBoxExtent,
		OffPageMargin: defaultOffPageMargin,
	}
}

// Apply filters a document in place, dropping obvious off-page and tiny text-bearing content.
func Apply(document *model.Document) error {
	return ApplyWithOptions(document, DefaultOptions())
}

// ApplyWithOptions filters a document in place using the supplied thresholds.
func ApplyWithOptions(document *model.Document, opts Options) error {
	if document == nil {
		return fmt.Errorf("layout filter requires a document")
	}
	opts = opts.withDefaults()

	pageLookup := buildPageLookup(document.Pages)

	document.Kids = filterContentElements(document.Kids, pageLookup, opts)
	document.Artifacts = filterArtifacts(document.Artifacts, pageLookup, opts)

	for _, page := range document.Pages {
		if page == nil {
			continue
		}
		pageBounds, pageBoundsOK := pageGeometry(page)
		pageCtx := filterContext{
			pageBounds:   pageBounds,
			pageBoundsOK: pageBoundsOK,
		}
		page.Kids = filterContentElementsWithContext(page.Kids, pageLookup, pageCtx, opts)
		page.Artifacts = filterArtifactsWithContext(page.Artifacts, pageLookup, pageCtx, opts)
	}

	return nil
}

type pageLookup struct {
	byNumber map[model.PageNumber]model.Box
	byIndex  map[model.PageIndex]model.Box
}

type filterContext struct {
	pageBounds   model.Box
	pageBoundsOK bool
}

func (o Options) withDefaults() Options {
	if o.TinyFontSize <= 0 {
		o.TinyFontSize = defaultTinyFontSize
	}
	if o.TinyBoxExtent <= 0 {
		o.TinyBoxExtent = defaultTinyBoxExtent
	}
	if o.OffPageMargin < 0 {
		o.OffPageMargin = defaultOffPageMargin
	}
	return o
}

func buildPageLookup(pages []*model.Page) pageLookup {
	lookup := pageLookup{
		byNumber: make(map[model.PageNumber]model.Box, len(pages)),
		byIndex:  make(map[model.PageIndex]model.Box, len(pages)),
	}
	for _, page := range pages {
		if page == nil {
			continue
		}
		bounds, ok := pageGeometry(page)
		if !ok {
			continue
		}
		lookup.byNumber[page.Metadata.Number] = bounds
		lookup.byIndex[page.Metadata.Index] = bounds
	}
	return lookup
}

func pageGeometry(page *model.Page) (model.Box, bool) {
	if page == nil {
		return model.Box{}, false
	}
	if !page.Metadata.Bounds.IsZero() {
		return page.Metadata.Bounds.Normalize(), true
	}
	if page.Metadata.Size.Width > 0 || page.Metadata.Size.Height > 0 {
		return model.NewBox(0, 0, page.Metadata.Size.Width, page.Metadata.Size.Height).Normalize(), true
	}
	return model.Box{}, false
}

func filterContentElements(elements []model.ContentElement, lookup pageLookup, opts Options) []model.ContentElement {
	return filterContentElementsWithContext(elements, lookup, filterContext{}, opts)
}

func filterContentElementsWithContext(elements []model.ContentElement, lookup pageLookup, ctx filterContext, opts Options) []model.ContentElement {
	if len(elements) == 0 {
		return nil
	}

	filtered := make([]model.ContentElement, 0, len(elements))
	for _, element := range elements {
		filteredElement := filterContentElement(element, lookup, ctx, opts)
		if filteredElement != nil {
			filtered = append(filtered, filteredElement)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func filterArtifacts(artifacts []*model.RawArtifact, lookup pageLookup, opts Options) []*model.RawArtifact {
	return filterArtifactsWithContext(artifacts, lookup, filterContext{}, opts)
}

func filterArtifactsWithContext(artifacts []*model.RawArtifact, lookup pageLookup, ctx filterContext, opts Options) []*model.RawArtifact {
	if len(artifacts) == 0 {
		return nil
	}

	filtered := make([]*model.RawArtifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifact == nil {
			continue
		}
		if shouldDropArtifact(artifact, lookup, ctx, opts) {
			continue
		}
		filtered = append(filtered, artifact)
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func filterContentElement(element model.ContentElement, lookup pageLookup, ctx filterContext, opts Options) model.ContentElement {
	if element == nil {
		return nil
	}

	base := element.NodeBase()
	if base != nil {
		if shouldDropOffPage(base, lookup, ctx, opts) {
			return nil
		}
		if shouldDropTinyText(element, base, opts) {
			return nil
		}
	}

	switch node := element.(type) {
	case *model.TextBlock:
		node.Kids = filterContentElementsWithContext(node.Kids, lookup, ctx, opts)
		if len(node.Kids) == 0 {
			return nil
		}
	case *model.List:
		node.ListItems = filterListItems(node.ListItems, lookup, ctx, opts)
		if len(node.ListItems) == 0 {
			return nil
		}
	case *model.ListItem:
		node.Kids = filterContentElementsWithContext(node.Kids, lookup, ctx, opts)
		if strings.TrimSpace(node.Content) == "" && len(node.Kids) == 0 {
			return nil
		}
	case *model.HeaderFooter:
		node.Kids = filterContentElementsWithContext(node.Kids, lookup, ctx, opts)
		if len(node.Kids) == 0 {
			return nil
		}
	case *model.Table:
		node.Rows = filterTableRows(node.Rows, lookup, ctx, opts)
		if len(node.Rows) == 0 {
			return nil
		}
	case *model.TableCell:
		node.Kids = filterContentElementsWithContext(node.Kids, lookup, ctx, opts)
		if len(node.Kids) == 0 {
			return nil
		}
	}

	if base != nil && isEmptyTextNode(element) {
		return nil
	}

	return element
}

func filterListItems(items []*model.ListItem, lookup pageLookup, ctx filterContext, opts Options) []*model.ListItem {
	if len(items) == 0 {
		return nil
	}

	filtered := make([]*model.ListItem, 0, len(items))
	for _, item := range items {
		filteredItem, ok := filterListItem(item, lookup, ctx, opts)
		if ok {
			filtered = append(filtered, filteredItem)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func filterListItem(item *model.ListItem, lookup pageLookup, ctx filterContext, opts Options) (*model.ListItem, bool) {
	if item == nil {
		return nil, false
	}
	filtered := filterContentElement(item, lookup, ctx, opts)
	if filtered == nil {
		return nil, false
	}
	listItem, ok := filtered.(*model.ListItem)
	if !ok {
		return nil, false
	}
	return listItem, true
}

func filterTableRows(rows []model.TableRow, lookup pageLookup, ctx filterContext, opts Options) []model.TableRow {
	if len(rows) == 0 {
		return nil
	}

	filtered := make([]model.TableRow, 0, len(rows))
	for _, row := range rows {
		row.Cells = filterTableCells(row.Cells, lookup, ctx, opts)
		if len(row.Cells) == 0 {
			continue
		}
		filtered = append(filtered, row)
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func filterTableCells(cells []*model.TableCell, lookup pageLookup, ctx filterContext, opts Options) []*model.TableCell {
	if len(cells) == 0 {
		return nil
	}

	filtered := make([]*model.TableCell, 0, len(cells))
	for _, cell := range cells {
		filteredCell, ok := filterTableCell(cell, lookup, ctx, opts)
		if ok {
			filtered = append(filtered, filteredCell)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func filterTableCell(cell *model.TableCell, lookup pageLookup, ctx filterContext, opts Options) (*model.TableCell, bool) {
	if cell == nil {
		return nil, false
	}
	filtered := filterContentElement(cell, lookup, ctx, opts)
	if filtered == nil {
		return nil, false
	}
	tableCell, ok := filtered.(*model.TableCell)
	if !ok {
		return nil, false
	}
	return tableCell, true
}

func shouldDropArtifact(artifact *model.RawArtifact, lookup pageLookup, ctx filterContext, opts Options) bool {
	if artifact == nil {
		return true
	}
	if shouldDropOffPageArtifact(artifact, lookup, ctx, opts) {
		return true
	}
	if isTextArtifact(artifact) {
		box, _ := artifactBounds(artifact)
		if isTinyTextBox(box, artifact.Style.FontSize, opts) {
			return true
		}
	}
	return false
}

func shouldDropOffPage(base *model.BaseNode, lookup pageLookup, ctx filterContext, opts Options) bool {
	box, ok := elementBounds(base)
	if !ok {
		return false
	}
	pageBounds, ok := resolvePageBounds(base, lookup, ctx)
	if !ok {
		return false
	}
	return !expanded(pageBounds, opts.OffPageMargin).Intersects(box)
}

func shouldDropOffPageArtifact(artifact *model.RawArtifact, lookup pageLookup, ctx filterContext, opts Options) bool {
	box, ok := artifactBounds(artifact)
	if !ok {
		return false
	}
	pageBounds, ok := resolveArtifactPageBounds(artifact, lookup, ctx)
	if !ok {
		return false
	}
	return !expanded(pageBounds, opts.OffPageMargin).Intersects(box)
}

func shouldDropTinyText(element model.ContentElement, base *model.BaseNode, opts Options) bool {
	if !isTextBearingNode(element) {
		return false
	}
	text := strings.TrimSpace(textContent(element))
	if text == "" {
		return false
	}
	return isTinyTextBox(elementBoundsOrZero(base), textFontSize(element), opts)
}

func isTextBearingNode(element model.ContentElement) bool {
	switch node := element.(type) {
	case *model.Paragraph:
		return true
	case *model.Heading:
		return true
	case *model.Caption:
		return true
	case *model.ListItem:
		return true
	default:
		_ = node
		return false
	}
}

func isTextArtifact(artifact *model.RawArtifact) bool {
	if artifact == nil {
		return false
	}
	return strings.TrimSpace(artifact.Text) != "" || artifact.Kind == model.ArtifactKindText
}

func textContent(element model.ContentElement) string {
	switch node := element.(type) {
	case *model.Paragraph:
		return node.Content
	case *model.Heading:
		return node.Content
	case *model.Caption:
		return node.Content
	case *model.ListItem:
		return node.Content
	default:
		return ""
	}
}

func textFontSize(element model.ContentElement) float64 {
	switch node := element.(type) {
	case *model.Paragraph:
		return node.FontSize
	case *model.Heading:
		return node.FontSize
	case *model.Caption:
		return node.FontSize
	case *model.ListItem:
		return node.FontSize
	default:
		return 0
	}
}

func elementBounds(base *model.BaseNode) (model.Box, bool) {
	if base == nil {
		return model.Box{}, false
	}
	box := base.Bounds
	hasBox := !box.IsZero()
	if multi, ok := base.Boxes.Bounds(); ok {
		if hasBox {
			box = box.Union(multi)
		} else {
			box = multi
			hasBox = true
		}
	}
	if !hasBox {
		return model.Box{}, false
	}
	return box.Normalize(), true
}

func elementBoundsOrZero(base *model.BaseNode) model.Box {
	box, ok := elementBounds(base)
	if !ok {
		return model.Box{}
	}
	return box
}

func artifactBounds(artifact *model.RawArtifact) (model.Box, bool) {
	if artifact == nil {
		return model.Box{}, false
	}
	box := artifact.Bounds
	hasBox := !box.IsZero()
	if multi, ok := artifact.Boxes.Bounds(); ok {
		if hasBox {
			box = box.Union(multi)
		} else {
			box = multi
			hasBox = true
		}
	}
	if !hasBox {
		return model.Box{}, false
	}
	return box.Normalize(), true
}

func resolvePageBounds(base *model.BaseNode, lookup pageLookup, ctx filterContext) (model.Box, bool) {
	if base == nil {
		return model.Box{}, false
	}
	if box, ok := lookup.byNumber[base.PageNumber]; ok {
		return box, true
	}
	if box, ok := lookup.byIndex[base.PageIndex]; ok {
		return box, true
	}
	if ctx.pageBoundsOK {
		return ctx.pageBounds, true
	}
	return model.Box{}, false
}

func resolveArtifactPageBounds(artifact *model.RawArtifact, lookup pageLookup, ctx filterContext) (model.Box, bool) {
	if artifact == nil {
		return model.Box{}, false
	}
	if box, ok := lookup.byNumber[artifact.PageNumber]; ok {
		return box, true
	}
	if box, ok := lookup.byIndex[artifact.PageIndex]; ok {
		return box, true
	}
	if ctx.pageBoundsOK {
		return ctx.pageBounds, true
	}
	return model.Box{}, false
}

func isTinyTextBox(box model.Box, fontSize float64, opts Options) bool {
	if fontSize > 0 && fontSize <= opts.TinyFontSize {
		return true
	}
	if box.IsZero() {
		return false
	}
	box = box.Normalize()
	return box.Width() <= opts.TinyBoxExtent && box.Height() <= opts.TinyBoxExtent
}

func expanded(box model.Box, margin float64) model.Box {
	if margin <= 0 {
		return box.Normalize()
	}
	box = box.Normalize()
	return model.Box{
		Left:   box.Left - margin,
		Bottom: box.Bottom - margin,
		Right:  box.Right + margin,
		Top:    box.Top + margin,
	}
}

func isEmptyTextNode(element model.ContentElement) bool {
	switch node := element.(type) {
	case *model.Paragraph:
		return strings.TrimSpace(node.Content) == ""
	case *model.Heading:
		return strings.TrimSpace(node.Content) == ""
	case *model.Caption:
		return strings.TrimSpace(node.Content) == ""
	case *model.ListItem:
		return strings.TrimSpace(node.Content) == "" && len(node.Kids) == 0
	default:
		return false
	}
}
