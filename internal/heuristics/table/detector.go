package table

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// Detector applies a lightweight table heuristic over ordered content nodes.
//
// The detector is intentionally conservative: it only converts consecutive
// text-bearing nodes on the same page when they form a stable row/column grid.
type Detector struct {
	MinRows         int
	MinColumns      int
	RowTolerance    float64
	ColumnTolerance float64
}

const (
	defaultMinRows         = 2
	defaultMinColumns      = 2
	defaultRowTolerance    = 6.0
	defaultColumnTolerance = 10.0
)

// Detect scans elements and replaces simple aligned text runs with tables.
func Detect(elements []model.ContentElement) []model.ContentElement {
	return Detector{}.Detect(elements)
}

// Detect scans elements and replaces simple aligned text runs with tables.
func (d Detector) Detect(elements []model.ContentElement) []model.ContentElement {
	d = d.withDefaults()

	out := make([]model.ContentElement, 0, len(elements))
	run := make([]cellCandidate, 0, 8)
	raw := make([]model.ContentElement, 0, 8)
	var page pageKey
	havePage := false

	flush := func() {
		if len(run) == 0 {
			return
		}
		out = append(out, d.transformRun(raw, run)...)
		run = run[:0]
		raw = raw[:0]
		havePage = false
	}

	for _, element := range elements {
		if element == nil {
			continue
		}

		candidate, ok := d.candidateFromElement(element)
		if !ok {
			flush()
			out = append(out, element)
			continue
		}

		if havePage && candidate.page != page {
			flush()
		}
		if !havePage {
			page = candidate.page
			havePage = true
		}

		candidate.position = len(raw)
		run = append(run, candidate)
		raw = append(raw, element)
	}

	flush()
	return out
}

type pageKey struct {
	index  model.PageIndex
	number model.PageNumber
}

type cellCandidate struct {
	element    model.ContentElement
	base       model.BaseNode
	bounds     model.Box
	page       pageKey
	index      int
	position   int
	text       string
	hasContent bool
}

type rowCluster struct {
	items    []cellCandidate
	bounds   model.Box
	startPos int
	endPos   int
}

func (d Detector) withDefaults() Detector {
	if d.MinRows <= 0 {
		d.MinRows = defaultMinRows
	}
	if d.MinColumns <= 0 {
		d.MinColumns = defaultMinColumns
	}
	if d.RowTolerance <= 0 {
		d.RowTolerance = defaultRowTolerance
	}
	if d.ColumnTolerance <= 0 {
		d.ColumnTolerance = defaultColumnTolerance
	}
	return d
}

func (d Detector) candidateFromElement(element model.ContentElement) (cellCandidate, bool) {
	base := element.NodeBase()
	if base == nil {
		return cellCandidate{}, false
	}

	text, hasText, ok := elementText(element)
	if !ok {
		return cellCandidate{}, false
	}

	bounds := elementBounds(element)
	if bounds.IsZero() {
		return cellCandidate{}, false
	}

	copyBase := *base
	copyBase.Bounds = bounds
	if copyBase.Type == "" {
		copyBase.Type = element.ContentType()
	}

	return cellCandidate{
		element:    element,
		base:       copyBase,
		bounds:     bounds,
		page:       pageKey{index: copyBase.PageIndex, number: copyBase.PageNumber},
		index:      copyBase.Index,
		text:       text,
		hasContent: hasText,
	}, true
}

func (d Detector) transformRun(raw []model.ContentElement, run []cellCandidate) []model.ContentElement {
	if len(run) < d.MinRows*d.MinColumns {
		return append([]model.ContentElement(nil), raw...)
	}

	rowTol := d.rowTolerance(run)
	rows := groupRows(run, rowTol)
	if len(rows) < d.MinRows {
		return append([]model.ContentElement(nil), raw...)
	}

	for i := range rows {
		sortCells(rows[i].items)
	}

	columnTol := d.columnTolerance(run)
	start, end, ok := d.bestTableSpan(rows, columnTol)
	if !ok {
		return append([]model.ContentElement(nil), raw...)
	}

	if !looksTableLike(rows[start : end+1]) {
		return append([]model.ContentElement(nil), raw...)
	}

	table := d.buildTableFromRows(rows[start : end+1])
	if table == nil {
		return append([]model.ContentElement(nil), raw...)
	}

	out := make([]model.ContentElement, 0, len(raw)+1)
	out = append(out, raw[:rows[start].startPos]...)
	out = append(out, table)
	out = append(out, raw[rows[end].endPos+1:]...)
	return out
}

func groupRows(candidates []cellCandidate, rowTol float64) []rowCluster {
	sorted := append([]cellCandidate(nil), candidates...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].page.index != sorted[j].page.index {
			return sorted[i].page.index < sorted[j].page.index
		}
		if sorted[i].page.number != sorted[j].page.number {
			return sorted[i].page.number < sorted[j].page.number
		}
		if !floatEqual(sorted[i].bounds.Top, sorted[j].bounds.Top) {
			return sorted[i].bounds.Top > sorted[j].bounds.Top
		}
		if !floatEqual(sorted[i].bounds.Left, sorted[j].bounds.Left) {
			return sorted[i].bounds.Left < sorted[j].bounds.Left
		}
		return sorted[i].index < sorted[j].index
	})

	rows := make([]rowCluster, 0, len(sorted))
	for _, candidate := range sorted {
		if len(rows) == 0 || !sameRow(rows[len(rows)-1], candidate, rowTol) {
			rows = append(rows, rowCluster{
				items:    []cellCandidate{candidate},
				bounds:   candidate.bounds,
				startPos: candidate.position,
				endPos:   candidate.position,
			})
			continue
		}
		last := &rows[len(rows)-1]
		last.items = append(last.items, candidate)
		last.bounds = last.bounds.Union(candidate.bounds)
		if candidate.position < last.startPos {
			last.startPos = candidate.position
		}
		if candidate.position > last.endPos {
			last.endPos = candidate.position
		}
	}
	return rows
}

func (d Detector) bestTableSpan(rows []rowCluster, columnTol float64) (int, int, bool) {
	bestStart := -1
	bestEnd := -1
	bestLength := 0

	for start := 0; start < len(rows); start++ {
		if len(rows[start].items) < d.MinColumns {
			continue
		}
		if !isRowSeparated(rows[start].items, d.ColumnTolerance) {
			continue
		}

		end := start
		for next := start + 1; next < len(rows); next++ {
			if len(rows[next].items) != len(rows[start].items) {
				break
			}
			if !isRowSeparated(rows[next].items, d.ColumnTolerance) {
				break
			}
			if !rowsAligned(rows[start].items, rows[next].items, columnTol) {
				break
			}
			end = next
		}

		length := end - start + 1
		if length >= d.MinRows && length > bestLength {
			bestStart = start
			bestEnd = end
			bestLength = length
		}
	}

	if bestStart < 0 {
		return 0, 0, false
	}
	return bestStart, bestEnd, true
}

func (d Detector) buildTableFromRows(rows []rowCluster) *model.Table {
	if len(rows) < d.MinRows {
		return nil
	}
	columnCount := len(rows[0].items)
	if columnCount < d.MinColumns {
		return nil
	}

	table := &model.Table{
		BaseNode: model.BaseNode{
			Type:       model.ElementTypeTable,
			Index:      rows[0].items[0].index,
			PageIndex:  rows[0].items[0].base.PageIndex,
			PageNumber: rows[0].items[0].base.PageNumber,
		},
		NumberOfRows:    len(rows),
		NumberOfColumns: columnCount,
		Rows:            make([]model.TableRow, 0, len(rows)),
	}

	for rowIndex, row := range rows {
		tableRow := model.TableRow{
			Type:      model.ElementTypeTableRow,
			RowNumber: rowIndex + 1,
			Cells:     make([]*model.TableCell, 0, len(row.items)),
		}
		for colIndex, item := range row.items {
			cell := &model.TableCell{
				BaseNode: model.BaseNode{
					Type:       model.ElementTypeTableCell,
					Index:      item.index,
					PageIndex:  item.base.PageIndex,
					PageNumber: item.base.PageNumber,
					Bounds:     item.bounds,
				},
				RowNumber:    rowIndex + 1,
				ColumnNumber: colIndex + 1,
				RowSpan:      1,
				ColumnSpan:   1,
				Kids:         []model.ContentElement{item.element},
			}
			tableRow.Cells = append(tableRow.Cells, cell)
			table.BaseNode.Bounds = table.BaseNode.Bounds.Union(item.bounds)
		}
		table.Rows = append(table.Rows, tableRow)
	}

	return table
}

func sameRow(row rowCluster, candidate cellCandidate, rowTol float64) bool {
	if len(row.items) == 0 {
		return false
	}
	if row.items[0].page != candidate.page {
		return false
	}
	if row.bounds.IsZero() || candidate.bounds.IsZero() {
		return false
	}

	rowCenter := boxCenterY(row.bounds)
	candidateCenter := boxCenterY(candidate.bounds)
	if math.Abs(rowCenter-candidateCenter) <= rowTol {
		return true
	}
	return boxVerticalGap(row.bounds, candidate.bounds) <= rowTol
}

func rowsAligned(anchor, candidate []cellCandidate, tolerance float64) bool {
	if len(anchor) != len(candidate) {
		return false
	}
	for i := range anchor {
		if !columnAligned(candidate[i].bounds, anchor[i].bounds, tolerance) {
			return false
		}
	}
	return true
}

func sortCells(items []cellCandidate) {
	sort.SliceStable(items, func(i, j int) bool {
		if !floatEqual(items[i].bounds.Left, items[j].bounds.Left) {
			return items[i].bounds.Left < items[j].bounds.Left
		}
		if !floatEqual(items[i].bounds.Right, items[j].bounds.Right) {
			return items[i].bounds.Right < items[j].bounds.Right
		}
		return items[i].index < items[j].index
	})
}

func isRowSeparated(items []cellCandidate, tolerance float64) bool {
	if len(items) < 2 {
		return false
	}
	for i := 1; i < len(items); i++ {
		prev := items[i-1].bounds.Normalize()
		next := items[i].bounds.Normalize()
		if prev.Right > next.Left+tolerance {
			return false
		}
	}
	return true
}

func (d Detector) rowTolerance(candidates []cellCandidate) float64 {
	heights := make([]float64, 0, len(candidates))
	for _, candidate := range candidates {
		if height := candidate.bounds.Height(); height > 0 {
			heights = append(heights, height)
		}
	}
	tol := d.RowTolerance
	if median := medianFloat(heights); median > 0 {
		tol = math.Max(tol, median*0.6)
	}
	return tol
}

func (d Detector) columnTolerance(candidates []cellCandidate) float64 {
	widths := make([]float64, 0, len(candidates))
	for _, candidate := range candidates {
		if width := candidate.bounds.Width(); width > 0 {
			widths = append(widths, width)
		}
	}
	tol := d.ColumnTolerance
	if median := medianFloat(widths); median > 0 {
		tol = math.Max(tol, median*0.25)
	}
	return tol
}

func columnAligned(candidate, anchor model.Box, tolerance float64) bool {
	if candidate.IsZero() || anchor.IsZero() {
		return false
	}
	if math.Abs(candidate.Left-anchor.Left) <= tolerance {
		return true
	}
	if math.Abs(candidate.Right-anchor.Right) <= tolerance {
		return true
	}
	return math.Abs(boxCenterX(candidate)-boxCenterX(anchor)) <= tolerance
}

func elementText(element model.ContentElement) (string, bool, bool) {
	switch node := element.(type) {
	case *model.Paragraph:
		return strings.TrimSpace(node.TextProperties.Content), node.TextProperties.Content != "", true
	case *model.Caption:
		return strings.TrimSpace(node.TextProperties.Content), node.TextProperties.Content != "", true
	case *model.ListItem:
		return strings.TrimSpace(node.TextProperties.Content), node.TextProperties.Content != "", true
	case *model.TextBlock:
		text := textBlockText(node.Kids)
		return text, text != "", true
	default:
		return "", false, false
	}
}

func textBlockText(kids []model.ContentElement) string {
	parts := make([]string, 0, len(kids))
	for _, kid := range kids {
		if kid == nil {
			continue
		}
		text, _, ok := elementText(kid)
		if !ok {
			continue
		}
		if trimmed := strings.TrimSpace(text); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func elementBounds(element model.ContentElement) model.Box {
	base := element.NodeBase()
	if base == nil {
		return model.Box{}
	}
	if !base.Bounds.IsZero() {
		return base.Bounds.Normalize()
	}
	if bounds, ok := base.Boxes.Bounds(); ok {
		return bounds.Normalize()
	}
	return model.Box{}
}

func boxCenterX(box model.Box) float64 {
	box = box.Normalize()
	return (box.Left + box.Right) / 2
}

func boxCenterY(box model.Box) float64 {
	box = box.Normalize()
	return (box.Bottom + box.Top) / 2
}

func boxVerticalGap(a, b model.Box) float64 {
	a = a.Normalize()
	b = b.Normalize()
	if a.Top < b.Bottom {
		return b.Bottom - a.Top
	}
	if b.Top < a.Bottom {
		return a.Bottom - b.Top
	}
	return 0
}

func medianFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	ordered := append([]float64(nil), values...)
	sort.Float64s(ordered)
	middle := len(ordered) / 2
	if len(ordered)%2 == 1 {
		return ordered[middle]
	}
	return (ordered[middle-1] + ordered[middle]) / 2
}

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) <= 0.01
}

func looksTableLike(rows []rowCluster) bool {
	cellCount := 0
	shortCellCount := 0
	sentenceLikeCount := 0
	totalWords := 0

	for _, row := range rows {
		for _, item := range row.items {
			text := strings.TrimSpace(item.text)
			if text == "" {
				continue
			}

			words := len(strings.Fields(text))
			totalWords += words
			cellCount++

			if isShortTableCell(text, words) {
				shortCellCount++
			}
			if isSentenceLikeCell(text, words) {
				sentenceLikeCount++
			}
		}
	}

	if cellCount == 0 {
		return false
	}

	avgWords := float64(totalWords) / float64(cellCount)
	if avgWords > 8 {
		return false
	}

	if sentenceLikeCount >= max(1, cellCount/2) {
		return false
	}

	requiredShortCells := 1
	if len(rows) <= 2 {
		requiredShortCells = 2
	} else if cellCount >= 6 {
		requiredShortCells = cellCount / 3
	}
	if shortCellCount < requiredShortCells {
		return false
	}

	return true
}

func isShortTableCell(text string, wordCount int) bool {
	if wordCount <= 3 {
		return true
	}
	if len(text) <= 24 {
		return true
	}
	return containsDigit(text)
}

func isSentenceLikeCell(text string, wordCount int) bool {
	if wordCount >= 8 {
		return true
	}
	if text == "" {
		return false
	}
	last := text[len(text)-1]
	switch last {
	case '.', '!', '?':
		return true
	}
	return false
}

func containsDigit(text string) bool {
	for _, r := range text {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
