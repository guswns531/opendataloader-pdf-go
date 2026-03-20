package fixture

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

// Ingestor loads document fixtures from JSON.
type Ingestor struct{}

// New returns a fixture ingestor.
func New() *Ingestor {
	return &Ingestor{}
}

// Name identifies the ingestor.
func (i *Ingestor) Name() string {
	return "fixture"
}

// Ingest loads a document fixture from the provided source.
func (i *Ingestor) Ingest(ctx *core.ProcessingContext, source core.Source) (*core.Document, error) {
	_ = ctx
	data, err := readSource(source)
	if err != nil {
		return nil, err
	}

	var fixture documentFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		return nil, fmt.Errorf("decode fixture json: %w", err)
	}

	document, err := fixture.toDocument(source)
	if err != nil {
		return nil, err
	}
	return document, nil
}

func readSource(source core.Source) ([]byte, error) {
	if source.Reader != nil {
		return io.ReadAll(source.Reader)
	}
	if source.Path != "" {
		return os.ReadFile(source.Path)
	}
	return nil, fmt.Errorf("fixture ingest requires source reader or path")
}

type documentFixture struct {
	Metadata  documentMetadataFixture `json:"metadata"`
	Pages     []pageFixture           `json:"pages,omitempty"`
	Kids      []contentFixture        `json:"kids,omitempty"`
	Artifacts []rawArtifactFixture    `json:"artifacts,omitempty"`
}

func (f documentFixture) toDocument(source core.Source) (*core.Document, error) {
	metadata := f.Metadata.toModel()
	if metadata.FileName == "" {
		metadata.FileName = source.Name
	}
	if metadata.FileName == "" && source.Path != "" {
		metadata.FileName = filepath.Base(source.Path)
	}

	document := model.NewDocument(metadata)
	for _, pageFixture := range f.Pages {
		page, err := pageFixture.toModel()
		if err != nil {
			return nil, err
		}
		document.AddPage(page)
	}

	if len(f.Artifacts) > 0 {
		document.Artifacts = make([]*model.RawArtifact, 0, len(f.Artifacts))
		for _, artifactFixture := range f.Artifacts {
			artifact, err := artifactFixture.toModel()
			if err != nil {
				return nil, err
			}
			document.Artifacts = append(document.Artifacts, artifact)
		}
	}

	if len(f.Kids) > 0 {
		document.Kids = make([]model.ContentElement, 0, len(f.Kids))
		for _, kidFixture := range f.Kids {
			kid, err := kidFixture.toModel()
			if err != nil {
				return nil, err
			}
			document.Kids = append(document.Kids, kid)
		}
	}

	if document.Metadata.PageCount == 0 {
		document.Metadata.PageCount = len(document.Pages)
	}

	return document, nil
}

type documentMetadataFixture struct {
	FileName         string   `json:"file_name,omitempty"`
	Author           *string  `json:"author,omitempty"`
	Title            *string  `json:"title,omitempty"`
	CreationDate     *string  `json:"creation_date,omitempty"`
	ModificationDate *string  `json:"modification_date,omitempty"`
	Producer         *string  `json:"producer,omitempty"`
	Creator          *string  `json:"creator,omitempty"`
	Subject          *string  `json:"subject,omitempty"`
	Language         *string  `json:"language,omitempty"`
	Keywords         []string `json:"keywords,omitempty"`
	PageCount        int      `json:"page_count,omitempty"`
}

func (f documentMetadataFixture) toModel() model.DocumentMetadata {
	return model.DocumentMetadata{
		FileName:         f.FileName,
		Author:           f.Author,
		Title:            f.Title,
		CreationDate:     f.CreationDate,
		ModificationDate: f.ModificationDate,
		Producer:         f.Producer,
		Creator:          f.Creator,
		Subject:          f.Subject,
		Language:         f.Language,
		Keywords:         append([]string(nil), f.Keywords...),
		PageCount:        f.PageCount,
	}
}

type pageFixture struct {
	Metadata  pageMetadataFixture  `json:"metadata"`
	Artifacts []rawArtifactFixture `json:"artifacts,omitempty"`
	Kids      []contentFixture     `json:"kids,omitempty"`
}

func (f pageFixture) toModel() (*model.Page, error) {
	page := &model.Page{
		Metadata:  f.Metadata.toModel(),
		Artifacts: make([]*model.RawArtifact, 0, len(f.Artifacts)),
		Kids:      make([]model.ContentElement, 0, len(f.Kids)),
	}

	for _, artifactFixture := range f.Artifacts {
		artifact, err := artifactFixture.toModel()
		if err != nil {
			return nil, err
		}
		page.Artifacts = append(page.Artifacts, artifact)
	}

	for _, kidFixture := range f.Kids {
		kid, err := kidFixture.toModel()
		if err != nil {
			return nil, err
		}
		page.Kids = append(page.Kids, kid)
	}

	return page, nil
}

type pageMetadataFixture struct {
	Index    model.PageIndex  `json:"index,omitempty"`
	Number   model.PageNumber `json:"number,omitempty"`
	Label    string           `json:"label,omitempty"`
	Size     model.PageSize   `json:"size,omitempty"`
	Bounds   model.Box        `json:"bounds,omitempty"`
	Rotation int              `json:"rotation,omitempty"`
}

func (f pageMetadataFixture) toModel() model.PageMetadata {
	return model.PageMetadata{
		Index:    f.Index,
		Number:   f.Number,
		Label:    f.Label,
		Size:     f.Size,
		Bounds:   f.Bounds,
		Rotation: f.Rotation,
	}
}

type rawArtifactFixture struct {
	ID         *model.ArtifactID     `json:"id,omitempty"`
	Kind       model.ArtifactKind    `json:"kind,omitempty"`
	PageIndex  *model.PageIndex      `json:"page_index,omitempty"`
	PageNumber *model.PageNumber     `json:"page_number,omitempty"`
	Bounds     model.Box             `json:"bounds,omitempty"`
	Boxes      model.MultiBox        `json:"boxes,omitempty"`
	Text       string                `json:"text,omitempty"`
	Format     model.ImageFormat     `json:"format,omitempty"`
	Data       []byte                `json:"data,omitempty"`
	Style      textPropertiesFixture `json:"style,omitempty"`
	Sequence   int                   `json:"sequence,omitempty"`
	Links      linkFixture           `json:"links,omitempty"`
}

func (f rawArtifactFixture) toModel() (*model.RawArtifact, error) {
	artifact := &model.RawArtifact{
		Kind:     f.Kind,
		Bounds:   f.Bounds,
		Boxes:    append(model.MultiBox(nil), f.Boxes...),
		Text:     f.Text,
		Format:   f.Format,
		Data:     append([]byte(nil), f.Data...),
		Style:    f.Style.toModel(),
		Sequence: f.Sequence,
		Links:    f.Links.toModel(),
	}
	if f.ID != nil {
		artifact.ID = *f.ID
	}
	if f.PageIndex != nil {
		artifact.PageIndex = *f.PageIndex
	}
	if f.PageNumber != nil {
		artifact.PageNumber = *f.PageNumber
	}
	return artifact, nil
}

type contentFixture struct {
	Type           model.ElementType `json:"type"`
	ID             *model.NodeID     `json:"id,omitempty"`
	Level          model.Level       `json:"level,omitempty"`
	Index          int               `json:"index,omitempty"`
	PageIndex      *model.PageIndex  `json:"page_index,omitempty"`
	PageNumber     *model.PageNumber `json:"page_number,omitempty"`
	Bounds         model.Box         `json:"bounds,omitempty"`
	Boxes          model.MultiBox    `json:"boxes,omitempty"`
	Links          linkFixture       `json:"links,omitempty"`
	Font           string            `json:"font,omitempty"`
	FontSize       float64           `json:"font_size,omitempty"`
	TextColor      string            `json:"text_color,omitempty"`
	Content        string            `json:"content,omitempty"`
	HiddenText     bool              `json:"hidden_text,omitempty"`
	Bold           bool              `json:"bold,omitempty"`
	Italic         bool              `json:"italic,omitempty"`
	Underline      bool              `json:"underline,omitempty"`
	HeadingLevel   int               `json:"heading_level,omitempty"`
	NumberingStyle string            `json:"numbering_style,omitempty"`
	NumberOfItems  int               `json:"number_of_items,omitempty"`
	Source         string            `json:"source,omitempty"`
	Data           string            `json:"data,omitempty"`
	Format         model.ImageFormat `json:"format,omitempty"`
	Rows           []tableRowFixture `json:"rows,omitempty"`
	Kids           []contentFixture  `json:"kids,omitempty"`
}

func (f contentFixture) toModel() (model.ContentElement, error) {
	base := model.BaseNode{
		Type:   f.Type,
		Level:  f.Level,
		Index:  f.Index,
		Bounds: f.Bounds,
		Boxes:  append(model.MultiBox(nil), f.Boxes...),
		Links:  f.Links.toModel(),
	}
	if f.ID != nil {
		base.ID = *f.ID
	}
	if f.PageIndex != nil {
		base.PageIndex = *f.PageIndex
	}
	if f.PageNumber != nil {
		base.PageNumber = *f.PageNumber
	}

	textNode := model.TextNode{
		BaseNode: base,
		TextProperties: model.TextProperties{
			Font:       f.Font,
			FontSize:   f.FontSize,
			TextColor:  f.TextColor,
			Content:    f.Content,
			HiddenText: f.HiddenText,
			Bold:       f.Bold,
			Italic:     f.Italic,
			Underline:  f.Underline,
		},
	}

	switch f.Type {
	case model.ElementTypeParagraph:
		return &model.Paragraph{TextNode: textNode}, nil
	case model.ElementTypeHeading:
		return &model.Heading{TextNode: textNode, HeadingLevel: f.HeadingLevel}, nil
	case model.ElementTypeCaption:
		return &model.Caption{TextNode: textNode}, nil
	case model.ElementTypeTextBlock:
		kids, err := convertKids(f.Kids)
		if err != nil {
			return nil, err
		}
		return &model.TextBlock{BaseNode: base, Kids: kids}, nil
	case model.ElementTypeList:
		items, err := convertListItems(f.Kids)
		if err != nil {
			return nil, err
		}
		return &model.List{
			BaseNode:       base,
			NumberingStyle: f.NumberingStyle,
			NumberOfItems:  f.NumberOfItems,
			ListItems:      items,
		}, nil
	case model.ElementTypeListItem:
		kids, err := convertKids(f.Kids)
		if err != nil {
			return nil, err
		}
		return &model.ListItem{TextNode: textNode, Kids: kids}, nil
	case model.ElementTypeImage:
		return &model.Image{
			BaseNode: base,
			Source:   f.Source,
			Data:     f.Data,
			Format:   f.Format,
		}, nil
	case model.ElementTypeHeader:
		kids, err := convertKids(f.Kids)
		if err != nil {
			return nil, err
		}
		return &model.HeaderFooter{BaseNode: base, Kids: kids}, nil
	case model.ElementTypeFooter:
		kids, err := convertKids(f.Kids)
		if err != nil {
			return nil, err
		}
		return &model.HeaderFooter{BaseNode: base, Kids: kids}, nil
	case model.ElementTypeTable:
		rows, err := convertRows(f.Rows)
		if err != nil {
			return nil, err
		}
		return &model.Table{
			BaseNode:        base,
			NumberOfRows:    len(rows),
			NumberOfColumns: maxColumns(rows),
			Rows:            rows,
		}, nil
	case model.ElementTypeTableCell:
		kids, err := convertKids(f.Kids)
		if err != nil {
			return nil, err
		}
		return &model.TableCell{
			BaseNode:     base,
			Kids:         kids,
			RowSpan:      1,
			ColumnSpan:   1,
			RowNumber:    0,
			ColumnNumber: 0,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported content type %q", f.Type)
	}
}

type tableRowFixture struct {
	Type      model.ElementType  `json:"type,omitempty"`
	RowNumber int                `json:"row_number,omitempty"`
	Cells     []tableCellFixture `json:"cells,omitempty"`
}

func (f tableRowFixture) toModel() (model.TableRow, error) {
	cells := make([]*model.TableCell, 0, len(f.Cells))
	for _, cellFixture := range f.Cells {
		cell, err := cellFixture.toModel()
		if err != nil {
			return model.TableRow{}, err
		}
		cells = append(cells, cell)
	}
	rowType := f.Type
	if rowType == "" {
		rowType = model.ElementTypeTableRow
	}
	return model.TableRow{
		Type:      rowType,
		RowNumber: f.RowNumber,
		Cells:     cells,
	}, nil
}

type tableCellFixture struct {
	Type         model.ElementType `json:"type,omitempty"`
	ID           *model.NodeID     `json:"id,omitempty"`
	Level        model.Level       `json:"level,omitempty"`
	Index        int               `json:"index,omitempty"`
	PageIndex    *model.PageIndex  `json:"page_index,omitempty"`
	PageNumber   *model.PageNumber `json:"page_number,omitempty"`
	Bounds       model.Box         `json:"bounds,omitempty"`
	Boxes        model.MultiBox    `json:"boxes,omitempty"`
	Links        linkFixture       `json:"links,omitempty"`
	RowNumber    int               `json:"row_number,omitempty"`
	ColumnNumber int               `json:"column_number,omitempty"`
	RowSpan      int               `json:"row_span,omitempty"`
	ColumnSpan   int               `json:"column_span,omitempty"`
	Kids         []contentFixture  `json:"kids,omitempty"`
}

func (f tableCellFixture) toModel() (*model.TableCell, error) {
	kids, err := convertKids(f.Kids)
	if err != nil {
		return nil, err
	}
	base := model.BaseNode{
		Type:   f.Type,
		Level:  f.Level,
		Index:  f.Index,
		Bounds: f.Bounds,
		Boxes:  append(model.MultiBox(nil), f.Boxes...),
		Links:  f.Links.toModel(),
	}
	if base.Type == "" {
		base.Type = model.ElementTypeTableCell
	}
	if f.ID != nil {
		base.ID = *f.ID
	}
	if f.PageIndex != nil {
		base.PageIndex = *f.PageIndex
	}
	if f.PageNumber != nil {
		base.PageNumber = *f.PageNumber
	}

	cell := &model.TableCell{
		BaseNode:     base,
		RowNumber:    f.RowNumber,
		ColumnNumber: f.ColumnNumber,
		RowSpan:      f.RowSpan,
		ColumnSpan:   f.ColumnSpan,
		Kids:         kids,
	}
	if cell.RowSpan == 0 {
		cell.RowSpan = 1
	}
	if cell.ColumnSpan == 0 {
		cell.ColumnSpan = 1
	}
	return cell, nil
}

type textPropertiesFixture struct {
	Font       string  `json:"font,omitempty"`
	FontSize   float64 `json:"font_size,omitempty"`
	TextColor  string  `json:"text_color,omitempty"`
	Content    string  `json:"content,omitempty"`
	HiddenText bool    `json:"hidden_text,omitempty"`
	Bold       bool    `json:"bold,omitempty"`
	Italic     bool    `json:"italic,omitempty"`
	Underline  bool    `json:"underline,omitempty"`
}

func (f textPropertiesFixture) toModel() model.TextProperties {
	return model.TextProperties{
		Font:       f.Font,
		FontSize:   f.FontSize,
		TextColor:  f.TextColor,
		Content:    f.Content,
		HiddenText: f.HiddenText,
		Bold:       f.Bold,
		Italic:     f.Italic,
		Underline:  f.Underline,
	}
}

type linkFixture struct {
	ParentID        *model.NodeID `json:"parent_id,omitempty"`
	PreviousID      *model.NodeID `json:"previous_id,omitempty"`
	NextID          *model.NodeID `json:"next_id,omitempty"`
	LinkedContentID *model.NodeID `json:"linked_content_id,omitempty"`
}

func (f linkFixture) toModel() model.LinkField {
	return model.LinkField{
		ParentID:        f.ParentID,
		PreviousID:      f.PreviousID,
		NextID:          f.NextID,
		LinkedContentID: f.LinkedContentID,
	}
}

func convertKids(fixtures []contentFixture) ([]model.ContentElement, error) {
	kids := make([]model.ContentElement, 0, len(fixtures))
	for _, kidFixture := range fixtures {
		kid, err := kidFixture.toModel()
		if err != nil {
			return nil, err
		}
		kids = append(kids, kid)
	}
	return kids, nil
}

func convertListItems(fixtures []contentFixture) ([]*model.ListItem, error) {
	items := make([]*model.ListItem, 0, len(fixtures))
	for _, itemFixture := range fixtures {
		kid, err := itemFixture.toModel()
		if err != nil {
			return nil, err
		}
		item, ok := kid.(*model.ListItem)
		if !ok {
			return nil, fmt.Errorf("list item must use type %q", model.ElementTypeListItem)
		}
		items = append(items, item)
	}
	return items, nil
}

func convertRows(fixtures []tableRowFixture) ([]model.TableRow, error) {
	rows := make([]model.TableRow, 0, len(fixtures))
	for _, rowFixture := range fixtures {
		row, err := rowFixture.toModel()
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func maxColumns(rows []model.TableRow) int {
	max := 0
	for _, row := range rows {
		if n := len(row.Cells); n > max {
			max = n
		}
	}
	return max
}
