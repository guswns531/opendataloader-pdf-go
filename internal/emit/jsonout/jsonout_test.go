package jsonout

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestEmitterEmitSerializesDocumentGraph(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{
		FileName:  "sample.pdf",
		PageCount: 1,
	})

	page := doc.AddPage(&model.Page{
		Metadata: model.PageMetadata{
			Index:    0,
			Number:   1,
			Label:    "p1",
			Size:     model.PageSize{Width: 612, Height: 792},
			Bounds:   model.NewBox(0, 0, 612, 792),
			Rotation: 90,
		},
		Artifacts: []*model.RawArtifact{
			{
				ID:         7,
				Kind:       model.ArtifactKindText,
				PageIndex:  0,
				PageNumber: 1,
				Bounds:     model.NewBox(10, 20, 30, 40),
				Boxes: model.MultiBox{
					model.NewBox(10, 20, 15, 25),
					model.NewBox(16, 26, 30, 40),
				},
				Text:   "artifact text",
				Format: model.ImageFormatPNG,
				Data:   []byte("artifact-bytes"),
				Style: model.TextProperties{
					Font:       "Arial",
					FontSize:   9,
					TextColor:  "#000000",
					Content:    "artifact content",
					HiddenText: true,
					Bold:       true,
					Italic:     true,
					Underline:  true,
				},
				Sequence: 3,
				Links: model.LinkField{
					ParentID: ptrNodeID(11),
				},
			},
		},
	})

	paragraph := &model.Paragraph{
		TextNode: model.TextNode{
			BaseNode: model.BaseNode{
				ID:         doc.NewNodeID(),
				Type:       model.ElementTypeParagraph,
				Index:      4,
				PageIndex:  0,
				PageNumber: 1,
				Bounds:     model.NewBox(100, 200, 300, 400),
				Boxes: model.MultiBox{
					model.NewBox(100, 200, 150, 240),
					model.NewBox(160, 250, 300, 400),
				},
				Links: model.LinkField{
					ParentID: ptrNodeID(1),
					NextID:   ptrNodeID(2),
				},
			},
			TextProperties: model.TextProperties{
				Font:      "Inter",
				FontSize:  12,
				TextColor: "#111111",
				Content:   "Hello world",
				Bold:      true,
			},
		},
	}

	list := &model.List{
		BaseNode: model.BaseNode{
			ID:         doc.NewNodeID(),
			Type:       model.ElementTypeList,
			Index:      5,
			PageIndex:  0,
			PageNumber: 1,
		},
		NumberingStyle: "arabic",
		NumberOfItems:  2,
		ListItems: []*model.ListItem{
			{
				TextNode: model.TextNode{
					BaseNode: model.BaseNode{
						ID:         doc.NewNodeID(),
						Type:       model.ElementTypeListItem,
						Index:      6,
						PageIndex:  0,
						PageNumber: 1,
					},
					TextProperties: model.TextProperties{
						Content: "First item",
					},
				},
				Kids: []model.ContentElement{
					&model.Caption{
						TextNode: model.TextNode{
							BaseNode: model.BaseNode{
								ID:         doc.NewNodeID(),
								Type:       model.ElementTypeCaption,
								Index:      7,
								PageIndex:  0,
								PageNumber: 1,
							},
							TextProperties: model.TextProperties{
								Content: "Nested caption",
							},
						},
					},
				},
			},
			{
				TextNode: model.TextNode{
					BaseNode: model.BaseNode{
						ID:         doc.NewNodeID(),
						Type:       model.ElementTypeListItem,
						Index:      8,
						PageIndex:  0,
						PageNumber: 1,
					},
					TextProperties: model.TextProperties{
						Content: "Second item",
					},
				},
			},
		},
	}

	cellParagraph := &model.Paragraph{
		TextNode: model.TextNode{
			BaseNode: model.BaseNode{
				ID:         doc.NewNodeID(),
				Type:       model.ElementTypeParagraph,
				Index:      9,
				PageIndex:  0,
				PageNumber: 1,
				Bounds:     model.NewBox(20, 30, 40, 50),
			},
			TextProperties: model.TextProperties{
				Content: "Cell text",
			},
		},
	}

	prevTableID := model.NodeID(21)
	nextTableID := model.NodeID(22)
	table := &model.Table{
		BaseNode: model.BaseNode{
			ID:         doc.NewNodeID(),
			Type:       model.ElementTypeTable,
			Index:      10,
			PageIndex:  0,
			PageNumber: 1,
			Bounds:     model.NewBox(50, 60, 500, 600),
		},
		NumberOfRows:    1,
		NumberOfColumns: 1,
		PreviousTableID: &prevTableID,
		NextTableID:     &nextTableID,
		Rows: []model.TableRow{
			{
				Type:      model.ElementTypeTableRow,
				RowNumber: 1,
				Cells: []*model.TableCell{
					{
						BaseNode: model.BaseNode{
							ID:         doc.NewNodeID(),
							Type:       model.ElementTypeTableCell,
							Index:      11,
							PageIndex:  0,
							PageNumber: 1,
							Bounds:     model.NewBox(55, 65, 150, 120),
							Links: model.LinkField{
								PreviousID: ptrNodeID(99),
							},
						},
						RowNumber:    1,
						ColumnNumber: 1,
						RowSpan:      2,
						ColumnSpan:   3,
						Kids: []model.ContentElement{
							cellParagraph,
						},
					},
				},
			},
		},
	}

	page.Kids = []model.ContentElement{paragraph, list, table}
	doc.Kids = append([]model.ContentElement(nil), page.Kids...)

	var buf bytes.Buffer
	emitter := New()
	if emitter.Name() != "json" {
		t.Fatalf("Name() = %q, want json", emitter.Name())
	}
	if emitter.Format() != core.OutputFormatJSON {
		t.Fatalf("Format() = %q, want json", emitter.Format())
	}

	if err := emitter.Emit(core.NewProcessingContext(doc, core.ProcessingOptions{}), doc, &buf); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid json: %v", err)
	}

	metadata := mustMap(t, got["metadata"])
	if metadata["file_name"] != "sample.pdf" {
		t.Fatalf("file_name = %v, want sample.pdf", metadata["file_name"])
	}
	if int(metadata["page_count"].(float64)) != 1 {
		t.Fatalf("page_count = %v, want 1", metadata["page_count"])
	}

	pages := mustSlice(t, got["pages"])
	if len(pages) != 1 {
		t.Fatalf("pages length = %d, want 1", len(pages))
	}

	page0 := mustMap(t, pages[0])
	pageMetadata := mustMap(t, page0["metadata"])
	if int(pageMetadata["index"].(float64)) != 0 || int(pageMetadata["number"].(float64)) != 1 {
		t.Fatalf("page metadata indices = %v", pageMetadata)
	}
	if pageMetadata["label"] != "p1" {
		t.Fatalf("page label = %v, want p1", pageMetadata["label"])
	}
	if int(pageMetadata["rotation"].(float64)) != 90 {
		t.Fatalf("page rotation = %v, want 90", pageMetadata["rotation"])
	}
	size := mustMap(t, pageMetadata["size"])
	if size["width"].(float64) != 612 || size["height"].(float64) != 792 {
		t.Fatalf("page size = %v, want 612x792", size)
	}
	assertBox(t, pageMetadata["bounds"], 0, 0, 612, 792)

	artifacts := mustSlice(t, page0["artifacts"])
	if len(artifacts) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(artifacts))
	}
	artifact := mustMap(t, artifacts[0])
	if int(artifact["id"].(float64)) != 7 || artifact["kind"] != string(model.ArtifactKindText) {
		t.Fatalf("artifact identity = %v", artifact)
	}
	if int(artifact["page_index"].(float64)) != 0 || int(artifact["page_number"].(float64)) != 1 {
		t.Fatalf("artifact page coordinates = %v", artifact)
	}
	if artifact["format"] != string(model.ImageFormatPNG) {
		t.Fatalf("artifact format = %v, want png", artifact["format"])
	}
	if artifact["data"] != base64.StdEncoding.EncodeToString([]byte("artifact-bytes")) {
		t.Fatalf("artifact data = %v", artifact["data"])
	}
	assertBox(t, artifact["bounds"], 10, 20, 30, 40)
	boxes := mustSlice(t, artifact["boxes"])
	if len(boxes) != 2 {
		t.Fatalf("artifact boxes length = %d, want 2", len(boxes))
	}
	assertBox(t, boxes[0], 10, 20, 15, 25)
	assertBox(t, boxes[1], 16, 26, 30, 40)
	style := mustMap(t, artifact["style"])
	if style["font"] != "Arial" || style["content"] != "artifact content" {
		t.Fatalf("artifact style = %v", style)
	}
	if style["hidden_text"] != true || style["bold"] != true || style["italic"] != true || style["underline"] != true {
		t.Fatalf("artifact style flags = %v", style)
	}
	links := mustMap(t, artifact["links"])
	if int(links["parent_id"].(float64)) != 11 {
		t.Fatalf("artifact links = %v", links)
	}

	kids := mustSlice(t, page0["kids"])
	if len(kids) != 3 {
		t.Fatalf("page kids length = %d, want 3", len(kids))
	}

	paragraphOut := mustMap(t, kids[0])
	if paragraphOut["type"] != string(model.ElementTypeParagraph) {
		t.Fatalf("paragraph type = %v", paragraphOut["type"])
	}
	if int(paragraphOut["id"].(float64)) != 1 {
		t.Fatalf("paragraph id = %v, want 1", paragraphOut["id"])
	}
	if paragraphOut["font"] != "Inter" || paragraphOut["content"] != "Hello world" {
		t.Fatalf("paragraph fields = %v", paragraphOut)
	}
	assertBox(t, paragraphOut["bounds"], 100, 200, 300, 400)
	paragraphLinks := mustMap(t, paragraphOut["links"])
	if int(paragraphLinks["parent_id"].(float64)) != 1 || int(paragraphLinks["next_id"].(float64)) != 2 {
		t.Fatalf("paragraph links = %v", paragraphLinks)
	}

	listOut := mustMap(t, kids[1])
	if listOut["type"] != string(model.ElementTypeList) || listOut["numbering_style"] != "arabic" {
		t.Fatalf("list = %v", listOut)
	}
	if int(listOut["number_of_items"].(float64)) != 2 {
		t.Fatalf("list number_of_items = %v", listOut["number_of_items"])
	}
	listItems := mustSlice(t, listOut["list_items"])
	if len(listItems) != 2 {
		t.Fatalf("list_items length = %d, want 2", len(listItems))
	}
	firstListItem := mustMap(t, listItems[0])
	if firstListItem["type"] != string(model.ElementTypeListItem) || firstListItem["content"] != "First item" {
		t.Fatalf("first list item = %v", firstListItem)
	}
	firstListItemKids := mustSlice(t, firstListItem["kids"])
	if len(firstListItemKids) != 1 {
		t.Fatalf("first list item kids length = %d, want 1", len(firstListItemKids))
	}
	nestedCaption := mustMap(t, firstListItemKids[0])
	if nestedCaption["type"] != string(model.ElementTypeCaption) || nestedCaption["content"] != "Nested caption" {
		t.Fatalf("nested caption = %v", nestedCaption)
	}

	tableOut := mustMap(t, kids[2])
	if tableOut["type"] != string(model.ElementTypeTable) {
		t.Fatalf("table type = %v", tableOut["type"])
	}
	if int(tableOut["number_of_rows"].(float64)) != 1 || int(tableOut["number_of_columns"].(float64)) != 1 {
		t.Fatalf("table dimensions = %v", tableOut)
	}
	if int(tableOut["previous_table_id"].(float64)) != 21 || int(tableOut["next_table_id"].(float64)) != 22 {
		t.Fatalf("table links = %v", tableOut)
	}
	rows := mustSlice(t, tableOut["rows"])
	if len(rows) != 1 {
		t.Fatalf("table rows length = %d, want 1", len(rows))
	}
	row0 := mustMap(t, rows[0])
	if row0["type"] != string(model.ElementTypeTableRow) || int(row0["row_number"].(float64)) != 1 {
		t.Fatalf("table row = %v", row0)
	}
	cells := mustSlice(t, row0["cells"])
	if len(cells) != 1 {
		t.Fatalf("table cells length = %d, want 1", len(cells))
	}
	cell0 := mustMap(t, cells[0])
	if cell0["type"] != string(model.ElementTypeTableCell) {
		t.Fatalf("table cell type = %v", cell0["type"])
	}
	if int(cell0["row_number"].(float64)) != 1 || int(cell0["column_number"].(float64)) != 1 {
		t.Fatalf("table cell coordinates = %v", cell0)
	}
	if int(cell0["row_span"].(float64)) != 2 || int(cell0["column_span"].(float64)) != 3 {
		t.Fatalf("table cell span = %v", cell0)
	}
	assertBox(t, cell0["bounds"], 55, 65, 150, 120)
	cellLinks := mustMap(t, cell0["links"])
	if int(cellLinks["previous_id"].(float64)) != 99 {
		t.Fatalf("table cell links = %v", cellLinks)
	}
	cellKids := mustSlice(t, cell0["kids"])
	if len(cellKids) != 1 {
		t.Fatalf("table cell kids length = %d, want 1", len(cellKids))
	}
	cellParagraphOut := mustMap(t, cellKids[0])
	if cellParagraphOut["content"] != "Cell text" {
		t.Fatalf("cell paragraph = %v", cellParagraphOut)
	}

	docKids := mustSlice(t, got["kids"])
	if len(docKids) != 3 {
		t.Fatalf("document kids length = %d, want 3", len(docKids))
	}
}

func TestEmitterRejectsNilInputs(t *testing.T) {
	emitter := New()

	if err := emitter.Emit(nil, nil, &bytes.Buffer{}); err == nil {
		t.Fatal("Emit(nil document) = nil error, want error")
	}
	if err := emitter.Emit(nil, model.NewDocument(model.DocumentMetadata{}), nil); err == nil {
		t.Fatal("Emit(nil writer) = nil error, want error")
	}
}

func ptrNodeID(id model.NodeID) *model.NodeID {
	return &id
}

func mustMap(t *testing.T, value any) map[string]any {
	t.Helper()
	m, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value is %T, want map[string]any", value)
	}
	return m
}

func mustSlice(t *testing.T, value any) []any {
	t.Helper()
	s, ok := value.([]any)
	if !ok {
		t.Fatalf("value is %T, want []any", value)
	}
	return s
}

func assertBox(t *testing.T, value any, left, bottom, right, top float64) {
	t.Helper()
	box := mustSlice(t, value)
	if len(box) != 4 {
		t.Fatalf("box length = %d, want 4", len(box))
	}
	got := []float64{
		box[0].(float64),
		box[1].(float64),
		box[2].(float64),
		box[3].(float64),
	}
	want := []float64{left, bottom, right, top}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("box[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}
