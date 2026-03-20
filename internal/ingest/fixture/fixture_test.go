package fixture

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestIngestorIngestFromReader(t *testing.T) {
	author := "Hancom"
	title := "Fixture Sample"
	pageNumber := model.PageNumber(1)
	pageIndex := model.PageIndex(0)
	nodeID := model.NodeID(11)
	cellID := model.NodeID(12)
	artifactID := model.ArtifactID(91)
	linkedID := model.NodeID(77)

	fixtureDoc := documentFixture{
		Metadata: documentMetadataFixture{
			Author:    &author,
			Title:     &title,
			PageCount: 1,
		},
		Pages: []pageFixture{
			{
				Metadata: pageMetadataFixture{
					Index:  pageIndex,
					Number: pageNumber,
					Label:  "1",
					Size:   model.PageSize{Width: 612, Height: 792},
					Bounds: model.Box{Left: 0, Bottom: 0, Right: 612, Top: 792},
				},
				Artifacts: []rawArtifactFixture{
					{
						ID:         &artifactID,
						Kind:       model.ArtifactKindText,
						PageIndex:  &pageIndex,
						PageNumber: &pageNumber,
						Text:       "page artifact",
						Data:       []byte{1, 2, 3},
						Style: textPropertiesFixture{
							Font:     "Helvetica",
							FontSize: 10,
							Content:  "artifact style",
						},
						Links: linkFixture{
							LinkedContentID: &linkedID,
						},
					},
				},
				Kids: []contentFixture{
					{
						Type:       model.ElementTypeParagraph,
						ID:         &nodeID,
						PageIndex:  &pageIndex,
						PageNumber: &pageNumber,
						Content:    "Hello fixture",
						Font:       "Times-Roman",
						FontSize:   12,
						Links: linkFixture{
							NextID: &linkedID,
						},
					},
					{
						Type:       model.ElementTypeTable,
						ID:         &linkedID,
						PageIndex:  &pageIndex,
						PageNumber: &pageNumber,
						Rows: []tableRowFixture{
							{
								Type:      model.ElementTypeTableRow,
								RowNumber: 1,
								Cells: []tableCellFixture{
									{
										Type:         model.ElementTypeTableCell,
										ID:           &cellID,
										PageIndex:    &pageIndex,
										PageNumber:   &pageNumber,
										RowNumber:    1,
										ColumnNumber: 1,
										RowSpan:      1,
										ColumnSpan:   1,
										Kids: []contentFixture{
											{
												Type:       model.ElementTypeCaption,
												PageIndex:  &pageIndex,
												PageNumber: &pageNumber,
												Content:    "Inside cell",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Kids: []contentFixture{
			{
				Type:    model.ElementTypeHeading,
				Content: "Document heading",
			},
		},
	}

	data, err := json.Marshal(fixtureDoc)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}

	ingestor := New()
	doc, err := ingestor.Ingest(nil, core.Source{
		Name:   "sample.pdf",
		Reader: bytes.NewReader(data),
	})
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}

	if doc.Metadata.FileName != "sample.pdf" {
		t.Fatalf("file name = %q, want sample.pdf", doc.Metadata.FileName)
	}
	if doc.Metadata.PageCount != 1 {
		t.Fatalf("page count = %d, want 1", doc.Metadata.PageCount)
	}
	if doc.Metadata.Title == nil || *doc.Metadata.Title != title {
		t.Fatalf("title = %v, want %q", doc.Metadata.Title, title)
	}
	if len(doc.Pages) != 1 {
		t.Fatalf("pages = %d, want 1", len(doc.Pages))
	}

	page := doc.Pages[0]
	if page.Metadata.Number != pageNumber {
		t.Fatalf("page number = %d, want %d", page.Metadata.Number, pageNumber)
	}
	if len(page.Artifacts) != 1 {
		t.Fatalf("page artifacts = %d, want 1", len(page.Artifacts))
	}
	if got := page.Artifacts[0].Data; !bytes.Equal(got, []byte{1, 2, 3}) {
		t.Fatalf("artifact data = %v, want [1 2 3]", got)
	}
	if len(page.Kids) != 2 {
		t.Fatalf("page kids = %d, want 2", len(page.Kids))
	}

	paragraph, ok := page.Kids[0].(*model.Paragraph)
	if !ok {
		t.Fatalf("kid 0 type = %T, want *model.Paragraph", page.Kids[0])
	}
	if paragraph.ContentType() != model.ElementTypeParagraph {
		t.Fatalf("paragraph type = %q, want paragraph", paragraph.ContentType())
	}
	if paragraph.Content != "Hello fixture" {
		t.Fatalf("paragraph content = %q, want Hello fixture", paragraph.Content)
	}

	table, ok := page.Kids[1].(*model.Table)
	if !ok {
		t.Fatalf("kid 1 type = %T, want *model.Table", page.Kids[1])
	}
	if table.NumberOfRows != 1 || table.NumberOfColumns != 1 {
		t.Fatalf("table size = %dx%d, want 1x1", table.NumberOfRows, table.NumberOfColumns)
	}
	if len(table.Rows) != 1 || len(table.Rows[0].Cells) != 1 {
		t.Fatalf("table rows/cells = %d/%d, want 1/1", len(table.Rows), len(table.Rows[0].Cells))
	}
	cell := table.Rows[0].Cells[0]
	if cell.RowSpan != 1 || cell.ColumnSpan != 1 {
		t.Fatalf("cell span = %dx%d, want 1x1", cell.RowSpan, cell.ColumnSpan)
	}
	if len(cell.Kids) != 1 {
		t.Fatalf("cell kids = %d, want 1", len(cell.Kids))
	}
	if caption, ok := cell.Kids[0].(*model.Caption); !ok || caption.Content != "Inside cell" {
		t.Fatalf("cell kid = %#v, want caption inside cell", cell.Kids[0])
	}
}

func TestIngestorIngestFromPathUsesFilenameFallback(t *testing.T) {
	fixtureDoc := documentFixture{
		Metadata: documentMetadataFixture{},
		Pages: []pageFixture{
			{
				Metadata: pageMetadataFixture{
					Number: 1,
				},
			},
		},
	}

	data, err := json.Marshal(fixtureDoc)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "from-path.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	doc, err := New().Ingest(nil, core.Source{Path: path})
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}

	if doc.Metadata.FileName != "from-path.json" {
		t.Fatalf("file name = %q, want from-path.json", doc.Metadata.FileName)
	}
	if len(doc.Pages) != 1 {
		t.Fatalf("pages = %d, want 1", len(doc.Pages))
	}
}

func TestIngestorRequiresReaderOrPath(t *testing.T) {
	if _, err := New().Ingest(nil, core.Source{}); err == nil {
		t.Fatal("expected error for empty source")
	}
}
