package jsonout

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

var _ core.Emitter = (*Emitter)(nil)

// Emitter serializes the current document graph as stable JSON.
type Emitter struct{}

// New returns a JSON emitter.
func New() *Emitter {
	return &Emitter{}
}

// Name identifies the emitter.
func (e *Emitter) Name() string {
	return "json"
}

// Format reports the output format.
func (e *Emitter) Format() core.OutputFormat {
	return core.OutputFormatJSON
}

// Emit writes the document graph to w as indented JSON.
func (e *Emitter) Emit(ctx *core.ProcessingContext, document *core.Document, w io.Writer) error {
	_ = ctx
	if e == nil {
		return fmt.Errorf("json emitter is nil")
	}
	if document == nil {
		return fmt.Errorf("json emitter requires a document")
	}
	if w == nil {
		return fmt.Errorf("json emitter requires a writer")
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(convertDocument(document))
}

type documentOutput struct {
	Metadata  documentMetadataOutput `json:"metadata"`
	Pages     []pageOutput           `json:"pages,omitempty"`
	Kids      []contentOutput        `json:"kids,omitempty"`
	Artifacts []rawArtifactOutput    `json:"artifacts,omitempty"`
}

type documentMetadataOutput struct {
	FileName         string   `json:"file_name"`
	Author           *string  `json:"author,omitempty"`
	Title            *string  `json:"title,omitempty"`
	CreationDate     *string  `json:"creation_date,omitempty"`
	ModificationDate *string  `json:"modification_date,omitempty"`
	Producer         *string  `json:"producer,omitempty"`
	Creator          *string  `json:"creator,omitempty"`
	Subject          *string  `json:"subject,omitempty"`
	Language         *string  `json:"language,omitempty"`
	Keywords         []string `json:"keywords,omitempty"`
	PageCount        int      `json:"page_count"`
}

type pageOutput struct {
	Metadata  pageMetadataOutput  `json:"metadata"`
	Artifacts []rawArtifactOutput `json:"artifacts,omitempty"`
	Kids      []contentOutput     `json:"kids,omitempty"`
}

type pageMetadataOutput struct {
	Index    model.PageIndex  `json:"index"`
	Number   model.PageNumber `json:"number"`
	Label    string           `json:"label,omitempty"`
	Size     *pageSizeOutput  `json:"size,omitempty"`
	Bounds   *boxValue        `json:"bounds,omitempty"`
	Rotation int              `json:"rotation"`
}

type pageSizeOutput struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type rawArtifactOutput struct {
	ID         model.ArtifactID      `json:"id"`
	Kind       model.ArtifactKind    `json:"kind"`
	PageIndex  model.PageIndex       `json:"page_index"`
	PageNumber model.PageNumber      `json:"page_number"`
	Bounds     *boxValue             `json:"bounds,omitempty"`
	Boxes      []boxValue            `json:"boxes,omitempty"`
	Text       string                `json:"text,omitempty"`
	Format     model.ImageFormat     `json:"format,omitempty"`
	Data       []byte                `json:"data,omitempty"`
	Style      *textPropertiesOutput `json:"style,omitempty"`
	Sequence   int                   `json:"sequence"`
	Links      *linkOutput           `json:"links,omitempty"`
}

type contentOutput struct {
	Type            model.ElementType `json:"type"`
	ID              model.NodeID      `json:"id"`
	Level           model.Level       `json:"level,omitempty"`
	Index           int               `json:"index"`
	PageIndex       model.PageIndex   `json:"page_index"`
	PageNumber      model.PageNumber  `json:"page_number"`
	Bounds          *boxValue         `json:"bounds,omitempty"`
	Boxes           []boxValue        `json:"boxes,omitempty"`
	Links           *linkOutput       `json:"links,omitempty"`
	Font            string            `json:"font,omitempty"`
	FontSize        float64           `json:"font_size,omitempty"`
	TextColor       string            `json:"text_color,omitempty"`
	Content         string            `json:"content,omitempty"`
	HiddenText      bool              `json:"hidden_text,omitempty"`
	Bold            bool              `json:"bold,omitempty"`
	Italic          bool              `json:"italic,omitempty"`
	Underline       bool              `json:"underline,omitempty"`
	HeadingLevel    int               `json:"heading_level,omitempty"`
	NumberingStyle  string            `json:"numbering_style,omitempty"`
	NumberOfItems   int               `json:"number_of_items,omitempty"`
	NumberOfRows    int               `json:"number_of_rows,omitempty"`
	NumberOfColumns int               `json:"number_of_columns,omitempty"`
	PreviousTableID *model.NodeID     `json:"previous_table_id,omitempty"`
	NextTableID     *model.NodeID     `json:"next_table_id,omitempty"`
	RowNumber       int               `json:"row_number,omitempty"`
	ColumnNumber    int               `json:"column_number,omitempty"`
	RowSpan         int               `json:"row_span,omitempty"`
	ColumnSpan      int               `json:"column_span,omitempty"`
	Source          string            `json:"source,omitempty"`
	Data            string            `json:"data,omitempty"`
	Format          model.ImageFormat `json:"format,omitempty"`
	Rows            []tableRowOutput  `json:"rows,omitempty"`
	ListItems       []contentOutput   `json:"list_items,omitempty"`
	Kids            []contentOutput   `json:"kids,omitempty"`
}

type tableRowOutput struct {
	Type      model.ElementType `json:"type"`
	RowNumber int               `json:"row_number"`
	Cells     []contentOutput   `json:"cells,omitempty"`
}

type textPropertiesOutput struct {
	Font       string  `json:"font,omitempty"`
	FontSize   float64 `json:"font_size,omitempty"`
	TextColor  string  `json:"text_color,omitempty"`
	Content    string  `json:"content,omitempty"`
	HiddenText bool    `json:"hidden_text,omitempty"`
	Bold       bool    `json:"bold,omitempty"`
	Italic     bool    `json:"italic,omitempty"`
	Underline  bool    `json:"underline,omitempty"`
}

type linkOutput struct {
	ParentID        *model.NodeID `json:"parent_id,omitempty"`
	PreviousID      *model.NodeID `json:"previous_id,omitempty"`
	NextID          *model.NodeID `json:"next_id,omitempty"`
	LinkedContentID *model.NodeID `json:"linked_content_id,omitempty"`
}

type boxValue struct {
	Box model.Box
}

func (b boxValue) MarshalJSON() ([]byte, error) {
	return json.Marshal([]float64{b.Box.Left, b.Box.Bottom, b.Box.Right, b.Box.Top})
}

func (b *boxValue) UnmarshalJSON(data []byte) error {
	var coords []float64
	if err := json.Unmarshal(data, &coords); err != nil {
		return err
	}
	if len(coords) != 4 {
		return fmt.Errorf("box must contain four coordinates, got %d", len(coords))
	}
	b.Box = model.Box{
		Left:   coords[0],
		Bottom: coords[1],
		Right:  coords[2],
		Top:    coords[3],
	}
	return nil
}

func convertDocument(document *core.Document) documentOutput {
	output := documentOutput{
		Metadata: convertDocumentMetadata(document.Metadata),
	}

	if len(document.Pages) > 0 {
		output.Pages = make([]pageOutput, 0, len(document.Pages))
		for _, page := range document.Pages {
			if page == nil {
				continue
			}
			output.Pages = append(output.Pages, convertPage(page))
		}
	}

	if len(document.Kids) > 0 {
		output.Kids = convertContentElements(document.Kids)
	}

	if len(document.Artifacts) > 0 {
		output.Artifacts = make([]rawArtifactOutput, 0, len(document.Artifacts))
		for _, artifact := range document.Artifacts {
			if artifact == nil {
				continue
			}
			output.Artifacts = append(output.Artifacts, convertRawArtifact(artifact))
		}
	}

	return output
}

func convertDocumentMetadata(metadata model.DocumentMetadata) documentMetadataOutput {
	return documentMetadataOutput{
		FileName:         metadata.FileName,
		Author:           metadata.Author,
		Title:            metadata.Title,
		CreationDate:     metadata.CreationDate,
		ModificationDate: metadata.ModificationDate,
		Producer:         metadata.Producer,
		Creator:          metadata.Creator,
		Subject:          metadata.Subject,
		Language:         metadata.Language,
		Keywords:         append([]string(nil), metadata.Keywords...),
		PageCount:        metadata.PageCount,
	}
}

func convertPage(page *model.Page) pageOutput {
	output := pageOutput{
		Metadata: convertPageMetadata(page.Metadata),
	}

	if len(page.Artifacts) > 0 {
		output.Artifacts = make([]rawArtifactOutput, 0, len(page.Artifacts))
		for _, artifact := range page.Artifacts {
			if artifact == nil {
				continue
			}
			output.Artifacts = append(output.Artifacts, convertRawArtifact(artifact))
		}
	}

	if len(page.Kids) > 0 {
		output.Kids = convertContentElements(page.Kids)
	}

	return output
}

func convertPageMetadata(metadata model.PageMetadata) pageMetadataOutput {
	return pageMetadataOutput{
		Index:    metadata.Index,
		Number:   metadata.Number,
		Label:    metadata.Label,
		Size:     pageSizePtr(metadata.Size),
		Bounds:   boxPtr(metadata.Bounds),
		Rotation: metadata.Rotation,
	}
}

func convertRawArtifact(artifact *model.RawArtifact) rawArtifactOutput {
	output := rawArtifactOutput{
		ID:         artifact.ID,
		Kind:       artifact.Kind,
		PageIndex:  artifact.PageIndex,
		PageNumber: artifact.PageNumber,
		Bounds:     boxPtr(artifact.Bounds),
		Text:       artifact.Text,
		Format:     artifact.Format,
		Data:       append([]byte(nil), artifact.Data...),
		Style:      textPropertiesPtr(artifact.Style),
		Sequence:   artifact.Sequence,
		Links:      linkPtr(artifact.Links),
	}
	if len(artifact.Boxes) > 0 {
		output.Boxes = make([]boxValue, 0, len(artifact.Boxes))
		for _, box := range artifact.Boxes {
			output.Boxes = append(output.Boxes, boxValue{Box: box})
		}
	}
	return output
}

func convertContentElements(elements []model.ContentElement) []contentOutput {
	output := make([]contentOutput, 0, len(elements))
	for _, element := range elements {
		if element == nil {
			continue
		}
		output = append(output, convertContentElement(element))
	}
	return output
}

func convertContentElement(element model.ContentElement) contentOutput {
	base := element.NodeBase()
	output := contentOutput{}
	if base != nil {
		output = baseContentOutput(base)
	}

	switch node := element.(type) {
	case *model.Paragraph:
		output.Font = node.Font
		output.FontSize = node.FontSize
		output.TextColor = node.TextColor
		output.Content = node.Content
		output.HiddenText = node.HiddenText
		output.Bold = node.Bold
		output.Italic = node.Italic
		output.Underline = node.Underline
	case *model.Heading:
		output.Font = node.Font
		output.FontSize = node.FontSize
		output.TextColor = node.TextColor
		output.Content = node.Content
		output.HiddenText = node.HiddenText
		output.Bold = node.Bold
		output.Italic = node.Italic
		output.Underline = node.Underline
		output.HeadingLevel = node.HeadingLevel
	case *model.Caption:
		output.Font = node.Font
		output.FontSize = node.FontSize
		output.TextColor = node.TextColor
		output.Content = node.Content
		output.HiddenText = node.HiddenText
		output.Bold = node.Bold
		output.Italic = node.Italic
		output.Underline = node.Underline
	case *model.TextBlock:
		output.Kids = convertContentElements(node.Kids)
	case *model.List:
		output.NumberingStyle = node.NumberingStyle
		output.NumberOfItems = node.NumberOfItems
		output.ListItems = convertListItems(node.ListItems)
	case *model.ListItem:
		output.Font = node.Font
		output.FontSize = node.FontSize
		output.TextColor = node.TextColor
		output.Content = node.Content
		output.HiddenText = node.HiddenText
		output.Bold = node.Bold
		output.Italic = node.Italic
		output.Underline = node.Underline
		output.Kids = convertContentElements(node.Kids)
	case *model.Image:
		output.Source = node.Source
		output.Data = node.Data
		output.Format = node.Format
	case *model.HeaderFooter:
		output.Kids = convertContentElements(node.Kids)
	case *model.Table:
		output.NumberOfRows = node.NumberOfRows
		output.NumberOfColumns = node.NumberOfColumns
		if node.PreviousTableID != nil {
			output.PreviousTableID = node.PreviousTableID
		}
		if node.NextTableID != nil {
			output.NextTableID = node.NextTableID
		}
		output.Rows = convertTableRows(node.Rows)
	case *model.TableCell:
		output.RowNumber = node.RowNumber
		output.ColumnNumber = node.ColumnNumber
		output.RowSpan = node.RowSpan
		output.ColumnSpan = node.ColumnSpan
		output.Kids = convertContentElements(node.Kids)
	default:
		// Preserve the common node fields even if new content element types appear.
	}

	return output
}

func baseContentOutput(base *model.BaseNode) contentOutput {
	output := contentOutput{
		Type:       base.Type,
		ID:         base.ID,
		Level:      base.Level,
		Index:      base.Index,
		PageIndex:  base.PageIndex,
		PageNumber: base.PageNumber,
		Bounds:     boxPtr(base.Bounds),
		Links:      linkPtr(base.Links),
	}
	if len(base.Boxes) > 0 {
		output.Boxes = make([]boxValue, 0, len(base.Boxes))
		for _, box := range base.Boxes {
			output.Boxes = append(output.Boxes, boxValue{Box: box})
		}
	}
	return output
}

func convertTableRows(rows []model.TableRow) []tableRowOutput {
	output := make([]tableRowOutput, 0, len(rows))
	for _, row := range rows {
		output = append(output, tableRowOutput{
			Type:      row.Type,
			RowNumber: row.RowNumber,
			Cells:     convertTableCells(row.Cells),
		})
	}
	return output
}

func convertTableCells(cells []*model.TableCell) []contentOutput {
	output := make([]contentOutput, 0, len(cells))
	for _, cell := range cells {
		if cell == nil {
			continue
		}
		output = append(output, convertContentElement(cell))
	}
	return output
}

func convertListItems(items []*model.ListItem) []contentOutput {
	output := make([]contentOutput, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		output = append(output, convertContentElement(item))
	}
	return output
}

func pageSizePtr(size model.PageSize) *pageSizeOutput {
	if size.Width == 0 && size.Height == 0 {
		return nil
	}
	return &pageSizeOutput{
		Width:  size.Width,
		Height: size.Height,
	}
}

func boxPtr(box model.Box) *boxValue {
	if box.IsZero() {
		return nil
	}
	return &boxValue{Box: box}
}

func textPropertiesPtr(props model.TextProperties) *textPropertiesOutput {
	if !props.HasContent() && props.Font == "" && props.FontSize == 0 && props.TextColor == "" && !props.HiddenText && !props.Bold && !props.Italic && !props.Underline {
		return nil
	}
	return &textPropertiesOutput{
		Font:       props.Font,
		FontSize:   props.FontSize,
		TextColor:  props.TextColor,
		Content:    props.Content,
		HiddenText: props.HiddenText,
		Bold:       props.Bold,
		Italic:     props.Italic,
		Underline:  props.Underline,
	}
}

func linkPtr(links model.LinkField) *linkOutput {
	if links.IsZero() {
		return nil
	}
	return &linkOutput{
		ParentID:        links.ParentID,
		PreviousID:      links.PreviousID,
		NextID:          links.NextID,
		LinkedContentID: links.LinkedContentID,
	}
}
