package htmlout

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

var _ core.Emitter = (*Emitter)(nil)

// Emitter renders a deterministic semantic HTML view of the document graph.
type Emitter struct{}

// New returns an HTML emitter.
func New() *Emitter {
	return &Emitter{}
}

// Name identifies the emitter.
func (e *Emitter) Name() string {
	return "html"
}

// Format reports the output format.
func (e *Emitter) Format() core.OutputFormat {
	return core.OutputFormatHTML
}

// Emit writes a deterministic HTML view of the document graph.
func (e *Emitter) Emit(_ *core.ProcessingContext, document *core.Document, w io.Writer) error {
	if e == nil {
		return fmt.Errorf("html emitter is nil")
	}
	if document == nil {
		return fmt.Errorf("html emitter requires a document")
	}
	if w == nil {
		return fmt.Errorf("html emitter requires a writer")
	}

	var b strings.Builder
	writeLine(&b, 0, "<!doctype html>")
	writeLine(&b, 0, "<html lang=\"en\">")
	writeLine(&b, 1, "<head>")
	writeLine(&b, 2, "<meta charset=\"utf-8\">")
	writeLine(&b, 2, "<title>%s</title>", escapeHTML(documentTitle(document)))
	writeLine(&b, 1, "</head>")
	writeLine(&b, 1, "<body>")
	writeLine(&b, 2, "<main>")

	roots := document.Kids
	if len(roots) == 0 {
		roots = collectPageContent(document.Pages)
	}
	for _, element := range roots {
		renderContent(&b, element, 3)
	}

	writeLine(&b, 2, "</main>")
	writeLine(&b, 1, "</body>")
	writeLine(&b, 0, "</html>")

	_, err := io.WriteString(w, b.String())
	return err
}

func documentTitle(document *core.Document) string {
	if document == nil {
		return "Document"
	}
	if title := cleanTextPtr(document.Metadata.Title); title != "" {
		return title
	}
	if fileName := cleanText(document.Metadata.FileName); fileName != "" {
		return fileName
	}
	return "Document"
}

func collectPageContent(pages []*model.Page) []model.ContentElement {
	if len(pages) == 0 {
		return nil
	}

	out := make([]model.ContentElement, 0)
	for _, page := range pages {
		if page == nil || len(page.Kids) == 0 {
			continue
		}
		out = append(out, page.Kids...)
	}
	return out
}

func renderContent(b *strings.Builder, element model.ContentElement, indent int) {
	if element == nil {
		return
	}

	switch node := element.(type) {
	case *model.Heading:
		renderHeading(b, node, indent)
	case *model.Paragraph:
		renderParagraph(b, node.TextProperties, indent)
	case *model.Caption:
		renderParagraph(b, node.TextProperties, indent)
	case *model.List:
		renderList(b, node, indent)
	case *model.ListItem:
		renderListItem(b, node, indent)
	case *model.Table:
		renderTable(b, node, indent)
	case *model.TableCell:
		renderTableCell(b, node, indent)
	case *model.TextBlock:
		renderChildren(b, node.Kids, indent)
	case *model.HeaderFooter:
		renderChildren(b, node.Kids, indent)
	default:
		if text := collectInlineText(element); text != "" {
			writeLine(b, indent, "<p>%s</p>", escapeHTML(text))
		}
	}
}

func renderHeading(b *strings.Builder, node *model.Heading, indent int) {
	if node == nil {
		return
	}

	text := cleanTextValue(node.TextProperties.Content)
	if text == "" {
		return
	}

	level := node.HeadingLevel
	if level <= 0 {
		level = parseHeadingLevel(node.Level)
	}
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	writeLine(b, indent, "<h%d>%s</h%d>", level, escapeHTML(text), level)
}

func parseHeadingLevel(level model.Level) int {
	if level == "" {
		return 1
	}
	n, err := strconv.Atoi(string(level))
	if err != nil || n <= 0 {
		return 1
	}
	return n
}

func renderParagraph(b *strings.Builder, props model.TextProperties, indent int) {
	if text := cleanTextValue(props.Content); text != "" {
		writeLine(b, indent, "<p>%s</p>", escapeHTML(text))
	}
}

func renderList(b *strings.Builder, node *model.List, indent int) {
	if node == nil {
		return
	}

	tag := "ul"
	if node.NumberingStyle != "" && node.NumberingStyle != "unordered" {
		tag = "ol"
	}

	writeLine(b, indent, "<%s>", tag)
	for _, item := range node.ListItems {
		renderListItem(b, item, indent+1)
	}
	writeLine(b, indent, "</%s>", tag)
}

func renderListItem(b *strings.Builder, item *model.ListItem, indent int) {
	if item == nil {
		return
	}

	text := cleanTextValue(item.TextProperties.Content)
	writeLine(b, indent, "<li>")
	if text != "" {
		writeLine(b, indent+1, "<p>%s</p>", escapeHTML(text))
	}
	renderChildren(b, item.Kids, indent+1)
	writeLine(b, indent, "</li>")
}

func renderChildren(b *strings.Builder, elements []model.ContentElement, indent int) {
	for _, element := range elements {
		renderContent(b, element, indent)
	}
}

func renderTable(b *strings.Builder, node *model.Table, indent int) {
	if node == nil || len(node.Rows) == 0 {
		return
	}

	columns := 0
	for _, row := range node.Rows {
		if len(row.Cells) > columns {
			columns = len(row.Cells)
		}
	}
	if columns == 0 {
		return
	}

	writeLine(b, indent, "<table>")
	writeLine(b, indent+1, "<tbody>")
	for _, row := range node.Rows {
		writeLine(b, indent+2, "<tr>")
		for i := 0; i < columns; i++ {
			if i >= len(row.Cells) || row.Cells[i] == nil {
				writeLine(b, indent+3, "<td></td>")
				continue
			}
			renderTableCell(b, row.Cells[i], indent+3)
		}
		writeLine(b, indent+2, "</tr>")
	}
	writeLine(b, indent+1, "</tbody>")
	writeLine(b, indent, "</table>")
}

func renderTableCell(b *strings.Builder, cell *model.TableCell, indent int) {
	if cell == nil {
		writeLine(b, indent, "<td></td>")
		return
	}

	writeLine(b, indent, "<td>")
	renderChildren(b, cell.Kids, indent+1)
	writeLine(b, indent, "</td>")
}

func cleanTextValue(value string) string {
	return cleanText(value)
}

func cleanTextPtr(value *string) string {
	if value == nil {
		return ""
	}
	return cleanTextValue(*value)
}

func collectInlineText(element model.ContentElement) string {
	if element == nil {
		return ""
	}

	switch node := element.(type) {
	case *model.Heading:
		return cleanTextValue(node.TextProperties.Content)
	case *model.Paragraph:
		return cleanTextValue(node.TextProperties.Content)
	case *model.Caption:
		return cleanTextValue(node.TextProperties.Content)
	case *model.ListItem:
		parts := make([]string, 0, 1+len(node.Kids))
		if text := cleanTextValue(node.TextProperties.Content); text != "" {
			parts = append(parts, text)
		}
		for _, kid := range node.Kids {
			if text := collectInlineText(kid); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, " ")
	case *model.TextBlock:
		return collectInlineChildren(node.Kids)
	case *model.List:
		parts := make([]string, 0, len(node.ListItems))
		for _, item := range node.ListItems {
			if text := collectInlineText(item); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, " ")
	case *model.Table:
		parts := make([]string, 0, len(node.Rows))
		for _, row := range node.Rows {
			for _, cell := range row.Cells {
				if text := collectInlineText(cell); text != "" {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, " ")
	case *model.TableCell:
		return collectInlineChildren(node.Kids)
	case *model.HeaderFooter:
		return collectInlineChildren(node.Kids)
	case *model.Image:
		return cleanTextValue(node.Source)
	default:
		return ""
	}
}

func collectInlineChildren(elements []model.ContentElement) string {
	parts := make([]string, 0, len(elements))
	for _, element := range elements {
		if text := collectInlineText(element); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func cleanText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func escapeHTML(text string) string {
	if text == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func writeLine(b *strings.Builder, indent int, format string, args ...any) {
	for i := 0; i < indent; i++ {
		b.WriteString("  ")
	}
	fmt.Fprintf(b, format+"\n", args...)
}
