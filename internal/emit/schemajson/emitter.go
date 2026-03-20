package schemajson

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

var _ core.Emitter = (*Emitter)(nil)

// Emitter serializes the document graph into the repository's public schema.json shape.
type Emitter struct{}

// New returns a schema-aligned JSON emitter.
func New() *Emitter {
	return &Emitter{}
}

// Name identifies the emitter.
func (e *Emitter) Name() string {
	return "schemajson"
}

// Format reports the output format.
func (e *Emitter) Format() core.OutputFormat {
	return core.OutputFormatJSON
}

// Emit writes schema-aligned JSON to w.
func (e *Emitter) Emit(_ *core.ProcessingContext, document *core.Document, w io.Writer) error {
	if e == nil {
		return fmt.Errorf("schemajson emitter is nil")
	}
	if document == nil {
		return fmt.Errorf("schemajson emitter requires a document")
	}
	if w == nil {
		return fmt.Errorf("schemajson emitter requires a writer")
	}

	out := convertDocument(document)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(out)
}

type documentOutput struct {
	FileName         string  `json:"file name"`
	NumberOfPages    int     `json:"number of pages"`
	Author           *string `json:"author"`
	Title            *string `json:"title"`
	CreationDate     *string `json:"creation date"`
	ModificationDate *string `json:"modification date"`
	Kids             []any   `json:"kids"`
}

type baseElementOutput struct {
	Type        string     `json:"type"`
	ID          *int64     `json:"id,omitempty"`
	Level       string     `json:"level,omitempty"`
	PageNumber  int        `json:"page number"`
	BoundingBox [4]float64 `json:"bounding box"`
}

type textPropertiesOutput struct {
	Font       string  `json:"font"`
	FontSize   float64 `json:"font size"`
	TextColor  string  `json:"text color"`
	Content    string  `json:"content"`
	HiddenText bool    `json:"hidden text,omitempty"`
}

type paragraphOutput struct {
	baseElementOutput
	textPropertiesOutput
}

type headingOutput struct {
	baseElementOutput
	textPropertiesOutput
	HeadingLevel int `json:"heading level"`
}

type captionOutput struct {
	baseElementOutput
	textPropertiesOutput
	LinkedContentID *int64 `json:"linked content id,omitempty"`
}

type textBlockOutput struct {
	baseElementOutput
	Kids []any `json:"kids"`
}

type listOutput struct {
	baseElementOutput
	NumberingStyle    string `json:"numbering style"`
	NumberOfListItems int    `json:"number of list items"`
	PreviousListID    *int64 `json:"previous list id,omitempty"`
	NextListID        *int64 `json:"next list id,omitempty"`
	ListItems         []any  `json:"list items"`
}

type listItemOutput struct {
	baseElementOutput
	textPropertiesOutput
	Kids []any `json:"kids"`
}

type imageOutput struct {
	baseElementOutput
	Source string `json:"source,omitempty"`
	Data   string `json:"data,omitempty"`
	Format string `json:"format,omitempty"`
}

type headerFooterOutput struct {
	baseElementOutput
	Kids []any `json:"kids"`
}

type tableOutput struct {
	baseElementOutput
	NumberOfRows    int              `json:"number of rows"`
	NumberOfColumns int              `json:"number of columns"`
	PreviousTableID *int64           `json:"previous table id,omitempty"`
	NextTableID     *int64           `json:"next table id,omitempty"`
	Rows            []tableRowOutput `json:"rows"`
}

type tableRowOutput struct {
	Type      string `json:"type"`
	RowNumber int    `json:"row number"`
	Cells     []any  `json:"cells"`
}

type tableCellOutput struct {
	baseElementOutput
	RowNumber    int   `json:"row number"`
	ColumnNumber int   `json:"column number"`
	RowSpan      int   `json:"row span"`
	ColumnSpan   int   `json:"column span"`
	Kids         []any `json:"kids"`
}

func convertDocument(document *core.Document) documentOutput {
	kids := convertRootKids(document)
	return documentOutput{
		FileName:         document.Metadata.FileName,
		NumberOfPages:    resolvedPageCount(document, kids),
		Author:           document.Metadata.Author,
		Title:            document.Metadata.Title,
		CreationDate:     document.Metadata.CreationDate,
		ModificationDate: document.Metadata.ModificationDate,
		Kids:             kids,
	}
}

func convertRootKids(document *core.Document) []any {
	if document == nil {
		return make([]any, 0)
	}
	if len(document.Kids) > 0 {
		return convertContentElements(document.Kids, 0)
	}

	kids := make([]any, 0)
	for _, page := range document.Pages {
		if page == nil || len(page.Kids) == 0 {
			continue
		}
		kids = append(kids, convertContentElements(page.Kids, int(page.Metadata.Number))...)
	}
	return kids
}

func resolvedPageCount(document *core.Document, kids []any) int {
	count := document.Metadata.PageCount
	if n := len(document.Pages); n > count {
		count = n
	}
	if count == 0 {
		for _, kid := range kids {
			if pageNumber := pageNumberFromAny(kid); pageNumber > count {
				count = pageNumber
			}
		}
	}
	return count
}

func pageNumberFromAny(value any) int {
	switch node := value.(type) {
	case *paragraphOutput:
		return node.PageNumber
	case *headingOutput:
		return node.PageNumber
	case *captionOutput:
		return node.PageNumber
	case *textBlockOutput:
		return node.PageNumber
	case *listOutput:
		return node.PageNumber
	case *listItemOutput:
		return node.PageNumber
	case *imageOutput:
		return node.PageNumber
	case *headerFooterOutput:
		return node.PageNumber
	case *tableOutput:
		return node.PageNumber
	case *tableCellOutput:
		return node.PageNumber
	default:
		return 0
	}
}

func convertContentElements(elements []model.ContentElement, inheritedPageNumber int) []any {
	out := make([]any, 0, len(elements))
	for _, element := range elements {
		if element == nil {
			continue
		}
		if converted := convertContentElement(element, inheritedPageNumber); converted != nil {
			out = append(out, converted)
		}
	}
	return out
}

func convertContentElement(element model.ContentElement, inheritedPageNumber int) any {
	switch node := element.(type) {
	case *model.Paragraph:
		base := baseElement(node.NodeBase(), inheritedPageNumber)
		return &paragraphOutput{
			baseElementOutput:    base,
			textPropertiesOutput: textProps(node.TextProperties),
		}
	case *model.Heading:
		base := baseElement(node.NodeBase(), inheritedPageNumber)
		headingLevel := node.HeadingLevel
		if headingLevel <= 0 {
			if parsed, err := strconv.Atoi(string(node.Level)); err == nil && parsed > 0 {
				headingLevel = parsed
			}
		}
		return &headingOutput{
			baseElementOutput:    base,
			textPropertiesOutput: textProps(node.TextProperties),
			HeadingLevel:         headingLevel,
		}
	case *model.Caption:
		base := baseElement(node.NodeBase(), inheritedPageNumber)
		return &captionOutput{
			baseElementOutput:    base,
			textPropertiesOutput: textProps(node.TextProperties),
			LinkedContentID:      nodeIDPtrValue(node.NodeBase().Links.LinkedContentID),
		}
	case *model.TextBlock:
		base := baseElement(node.NodeBase(), inheritedPageNumber)
		nextPage := base.PageNumber
		return &textBlockOutput{
			baseElementOutput: base,
			Kids:              convertContentElements(node.Kids, nextPage),
		}
	case *model.List:
		base := baseElement(node.NodeBase(), inheritedPageNumber)
		nextPage := base.PageNumber
		previousListID, nextListID := listLinkIDs(node.NodeBase())
		return &listOutput{
			baseElementOutput: base,
			NumberingStyle:    node.NumberingStyle,
			NumberOfListItems: resolvedListCount(node),
			PreviousListID:    previousListID,
			NextListID:        nextListID,
			ListItems:         convertListItems(node.ListItems, nextPage),
		}
	case *model.ListItem:
		base := baseElement(node.NodeBase(), inheritedPageNumber)
		nextPage := base.PageNumber
		return &listItemOutput{
			baseElementOutput:    base,
			textPropertiesOutput: textProps(node.TextProperties),
			Kids:                 convertContentElements(node.Kids, nextPage),
		}
	case *model.Image:
		base := baseElement(node.NodeBase(), inheritedPageNumber)
		return &imageOutput{
			baseElementOutput: base,
			Source:            node.Source,
			Data:              node.Data,
			Format:            string(node.Format),
		}
	case *model.HeaderFooter:
		base := baseElement(node.NodeBase(), inheritedPageNumber)
		nextPage := base.PageNumber
		return &headerFooterOutput{
			baseElementOutput: base,
			Kids:              convertContentElements(node.Kids, nextPage),
		}
	case *model.Table:
		base := baseElement(node.NodeBase(), inheritedPageNumber)
		nextPage := base.PageNumber
		previousTableID, nextTableID := tableLinkIDs(node)
		return &tableOutput{
			baseElementOutput: base,
			NumberOfRows:      resolvedTableRowCount(node),
			NumberOfColumns:   resolvedTableColumnCount(node),
			PreviousTableID:   previousTableID,
			NextTableID:       nextTableID,
			Rows:              convertTableRows(node.Rows, nextPage),
		}
	case *model.TableCell:
		base := baseElement(node.NodeBase(), inheritedPageNumber)
		nextPage := base.PageNumber
		return &tableCellOutput{
			baseElementOutput: base,
			RowNumber:         resolvedTableCellRowNumber(node),
			ColumnNumber:      resolvedTableCellColumnNumber(node),
			RowSpan:           resolvedSpan(node.RowSpan),
			ColumnSpan:        resolvedSpan(node.ColumnSpan),
			Kids:              convertContentElements(node.Kids, nextPage),
		}
	default:
		return nil
	}
}

func convertListItems(items []*model.ListItem, inheritedPageNumber int) []any {
	out := make([]any, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, convertContentElement(item, inheritedPageNumber))
	}
	return out
}

func convertTableRows(rows []model.TableRow, inheritedPageNumber int) []tableRowOutput {
	out := make([]tableRowOutput, 0, len(rows))
	for i, row := range rows {
		rowNumber := row.RowNumber
		if rowNumber <= 0 {
			rowNumber = i + 1
		}
		out = append(out, tableRowOutput{
			Type:      resolvedElementType(string(row.Type), string(model.ElementTypeTableRow)),
			RowNumber: rowNumber,
			Cells:     convertTableCells(row.Cells, inheritedPageNumber, rowNumber),
		})
	}
	return out
}

func convertTableCells(cells []*model.TableCell, inheritedPageNumber, rowNumber int) []any {
	out := make([]any, 0, len(cells))
	for i, cell := range cells {
		if cell == nil {
			continue
		}
		converted := convertContentElement(cell, inheritedPageNumber)
		if typed, ok := converted.(*tableCellOutput); ok {
			if typed.RowNumber <= 0 {
				typed.RowNumber = rowNumber
			}
			if typed.ColumnNumber <= 0 {
				typed.ColumnNumber = i + 1
			}
		}
		out = append(out, converted)
	}
	return out
}

func baseElement(node *model.BaseNode, inheritedPageNumber int) baseElementOutput {
	if node == nil {
		return baseElementOutput{BoundingBox: zeroBoundingBox()}
	}
	pageNumber := int(node.PageNumber)
	if pageNumber <= 0 {
		pageNumber = inheritedPageNumber
	}
	return baseElementOutput{
		Type:        resolvedElementType(string(node.Type), ""),
		ID:          nodeIDPtr(node.ID),
		Level:       string(node.Level),
		PageNumber:  pageNumber,
		BoundingBox: boundingBox(node.Bounds, node.Boxes),
	}
}

func textProps(props model.TextProperties) textPropertiesOutput {
	return textPropertiesOutput{
		Font:       props.Font,
		FontSize:   props.FontSize,
		TextColor:  props.TextColor,
		Content:    props.Content,
		HiddenText: props.HiddenText,
	}
}

func nodeIDPtr(id model.NodeID) *int64 {
	if id == 0 {
		return nil
	}
	value := int64(id)
	return &value
}

func tableLinkIDs(node *model.Table) (*int64, *int64) {
	if node == nil {
		return nil, nil
	}
	return nodeIDPtrValue(node.PreviousTableID), nodeIDPtrValue(node.NextTableID)
}

func listLinkIDs(node *model.BaseNode) (*int64, *int64) {
	if node == nil {
		return nil, nil
	}
	return nodeIDPtrValue(node.Links.PreviousID), nodeIDPtrValue(node.Links.NextID)
}

func nodeIDPtrValue(id *model.NodeID) *int64 {
	if id == nil || *id == 0 {
		return nil
	}
	value := int64(*id)
	return &value
}

func boundingBox(box model.Box, boxes model.MultiBox) [4]float64 {
	if !box.IsZero() {
		box = box.Normalize()
		return [4]float64{box.Left, box.Bottom, box.Right, box.Top}
	}
	if bounds, ok := boxes.Bounds(); ok && !bounds.IsZero() {
		bounds = bounds.Normalize()
		return [4]float64{bounds.Left, bounds.Bottom, bounds.Right, bounds.Top}
	}
	return zeroBoundingBox()
}

func zeroBoundingBox() [4]float64 {
	return [4]float64{}
}

func resolvedElementType(current string, fallback string) string {
	if current != "" {
		return current
	}
	return fallback
}

func resolvedListCount(list *model.List) int {
	if list == nil {
		return 0
	}
	if list.NumberOfItems > 0 {
		return list.NumberOfItems
	}
	return len(list.ListItems)
}

func resolvedTableRowCount(table *model.Table) int {
	if table == nil {
		return 0
	}
	if table.NumberOfRows > 0 {
		return table.NumberOfRows
	}
	return len(table.Rows)
}

func resolvedTableColumnCount(table *model.Table) int {
	if table == nil {
		return 0
	}
	if table.NumberOfColumns > 0 {
		return table.NumberOfColumns
	}
	maxColumns := 0
	for _, row := range table.Rows {
		if cells := len(row.Cells); cells > maxColumns {
			maxColumns = cells
		}
	}
	return maxColumns
}

func resolvedTableCellRowNumber(cell *model.TableCell) int {
	if cell == nil {
		return 0
	}
	return cell.RowNumber
}

func resolvedTableCellColumnNumber(cell *model.TableCell) int {
	if cell == nil {
		return 0
	}
	return cell.ColumnNumber
}

func resolvedSpan(span int) int {
	if span > 0 {
		return span
	}
	return 1
}
