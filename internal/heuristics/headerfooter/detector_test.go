package headerfooter

import (
	"strconv"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestDetectFindsRepeatedHeadersAndFooters(t *testing.T) {
	doc := buildDocument()

	detections := Detect(doc)
	if got, want := len(detections), 9; got != want {
		t.Fatalf("len(Detect()) = %d, want %d", got, want)
	}

	if containsText(detections, "Executive Summary") {
		t.Fatalf("unique top-of-page text was incorrectly detected as a header/footer")
	}

	if got := countByRoleAndText(detections, RoleHeader, "Quarterly Report"); got != 3 {
		t.Fatalf("header detections for %q = %d, want 3", "Quarterly Report", got)
	}
	if got := countByRoleAndText(detections, RoleHeader, "Confidential"); got != 3 {
		t.Fatalf("header detections for %q = %d, want 3", "Confidential", got)
	}
	if got := countByNormalizedTextAndRole(detections, RoleFooter, "page # of #"); got != 3 {
		t.Fatalf("footer detections for normalized text %q = %d, want 3", "page # of #", got)
	}
}

func TestApplyRemoveDropsDetectedHeaderFooterNodes(t *testing.T) {
	doc := buildDocument()

	detections := NewDetector().Apply(doc, ModeRemove)
	if got, want := len(detections), 9; got != want {
		t.Fatalf("len(Apply(remove)) = %d, want %d", got, want)
	}

	if got, want := len(doc.Pages[0].Kids), 2; got != want {
		t.Fatalf("page 1 kids = %d, want %d", got, want)
	}
	assertParagraphText(t, doc.Pages[0].Kids[0], "Executive Summary")
	assertParagraphText(t, doc.Pages[0].Kids[1], "Body content on page 1.")

	for pageIndex := 1; pageIndex < len(doc.Pages); pageIndex++ {
		if got, want := len(doc.Pages[pageIndex].Kids), 1; got != want {
			t.Fatalf("page %d kids = %d, want %d", pageIndex+1, got, want)
		}
		assertParagraphText(t, doc.Pages[pageIndex].Kids[0], bodyText(pageIndex+1))
	}
}

func TestApplyMarkWrapsDetectedRunsInHeaderFooterNodes(t *testing.T) {
	doc := buildDocument()

	detections := NewDetector().Apply(doc, ModeMark)
	if got, want := len(detections), 9; got != want {
		t.Fatalf("len(Apply(mark)) = %d, want %d", got, want)
	}

	page1 := doc.Pages[0]
	if got, want := len(page1.Kids), 4; got != want {
		t.Fatalf("page 1 kids = %d, want %d", got, want)
	}

	header, ok := page1.Kids[0].(*model.HeaderFooter)
	if !ok {
		t.Fatalf("page 1 kid 0 = %T, want *model.HeaderFooter", page1.Kids[0])
	}
	if got, want := header.Type, model.ElementTypeHeader; got != want {
		t.Fatalf("page 1 header type = %q, want %q", got, want)
	}
	if got, want := len(header.Kids), 2; got != want {
		t.Fatalf("page 1 header kids = %d, want %d", got, want)
	}
	assertParagraphText(t, header.Kids[0], "Quarterly Report")
	assertParagraphText(t, header.Kids[1], "Confidential")

	assertParagraphText(t, page1.Kids[1], "Executive Summary")
	assertParagraphText(t, page1.Kids[2], "Body content on page 1.")

	footer, ok := page1.Kids[3].(*model.HeaderFooter)
	if !ok {
		t.Fatalf("page 1 kid 3 = %T, want *model.HeaderFooter", page1.Kids[3])
	}
	if got, want := footer.Type, model.ElementTypeFooter; got != want {
		t.Fatalf("page 1 footer type = %q, want %q", got, want)
	}
	if got, want := len(footer.Kids), 1; got != want {
		t.Fatalf("page 1 footer kids = %d, want %d", got, want)
	}
	assertParagraphText(t, footer.Kids[0], "Page 1 of 3")
}

func buildDocument() *model.Document {
	doc := model.NewDocument(model.DocumentMetadata{FileName: "fixture.pdf"})
	for pageNum := 1; pageNum <= 3; pageNum++ {
		doc.AddPage(&model.Page{
			Metadata: model.PageMetadata{
				Index:  model.PageIndex(pageNum - 1),
				Number: model.PageNumber(pageNum),
				Size: model.PageSize{
					Width:  595,
					Height: 842,
				},
			},
			Kids: []model.ContentElement{
				paragraphNode(int64(pageNum*10+1), pageNum-1, pageNum, "Quarterly Report", model.NewBox(40, 768, 240, 786)),
				paragraphNode(int64(pageNum*10+2), pageNum-1, pageNum, "Confidential", model.NewBox(40, 744, 180, 758)),
			},
		})

		page := doc.Pages[pageNum-1]
		if pageNum == 1 {
			page.Kids = append(page.Kids, paragraphNode(100, 0, 1, "Executive Summary", model.NewBox(40, 690, 190, 706)))
		}
		page.Kids = append(page.Kids, paragraphNode(int64(pageNum*10+3), pageNum-1, pageNum, bodyText(pageNum), model.NewBox(48, 460, 520, 520)))
		page.Kids = append(page.Kids, paragraphNode(int64(pageNum*10+4), pageNum-1, pageNum, footerText(pageNum), model.NewBox(40, 18, 170, 34)))
	}
	doc.Kids = append(doc.Kids, doc.Pages[0].Kids...)
	doc.Kids = append(doc.Kids, doc.Pages[1].Kids...)
	doc.Kids = append(doc.Kids, doc.Pages[2].Kids...)
	return doc
}

func paragraphNode(id int64, pageIndex int, pageNumber int, text string, bounds model.Box) *model.Paragraph {
	return &model.Paragraph{
		TextNode: model.TextNode{
			BaseNode: model.BaseNode{
				ID:         model.NodeID(id),
				Type:       model.ElementTypeParagraph,
				PageIndex:  model.PageIndex(pageIndex),
				PageNumber: model.PageNumber(pageNumber),
				Bounds:     bounds,
			},
			TextProperties: model.TextProperties{
				Content: text,
			},
		},
	}
}

func bodyText(pageNumber int) string {
	return "Body content on page " + strconv.Itoa(pageNumber) + "."
}

func footerText(pageNumber int) string {
	return "Page " + strconv.Itoa(pageNumber) + " of 3"
}

func containsText(detections []Detection, text string) bool {
	for _, detection := range detections {
		if detection.Text == text {
			return true
		}
	}
	return false
}

func countByRoleAndText(detections []Detection, role Role, text string) int {
	count := 0
	for _, detection := range detections {
		if detection.Role == role && detection.Text == text {
			count++
		}
	}
	return count
}

func countByNormalizedTextAndRole(detections []Detection, role Role, text string) int {
	count := 0
	for _, detection := range detections {
		if detection.Role == role && detection.NormalizedText == text {
			count++
		}
	}
	return count
}

func assertParagraphText(t *testing.T, element model.ContentElement, want string) {
	t.Helper()

	switch node := element.(type) {
	case *model.Paragraph:
		if got := node.Content; got != want {
			t.Fatalf("paragraph content = %q, want %q", got, want)
		}
	case *model.HeaderFooter:
		if len(node.Kids) != 1 {
			t.Fatalf("header/footer kids = %d, want 1", len(node.Kids))
		}
		assertParagraphText(t, node.Kids[0], want)
	default:
		t.Fatalf("element type = %T, want *model.Paragraph or *model.HeaderFooter", element)
	}
}
