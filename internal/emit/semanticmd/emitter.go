package semanticmd

import (
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

var _ core.Emitter = (*Emitter)(nil)

// Emitter renders a document graph as semantic Markdown blocks.
type Emitter struct{}

// New returns a semantic Markdown emitter.
func New() *Emitter {
	return &Emitter{}
}

// Name identifies the emitter.
func (e *Emitter) Name() string {
	return "semanticmd"
}

// Format reports the output format.
func (e *Emitter) Format() core.OutputFormat {
	return core.OutputFormatMarkdown
}

// Emit writes a deterministic semantic Markdown view of the document graph.
func (e *Emitter) Emit(_ *core.ProcessingContext, document *core.Document, w io.Writer) error {
	if e == nil {
		return fmt.Errorf("semanticmd emitter is nil")
	}
	if document == nil {
		return fmt.Errorf("semanticmd emitter requires a document")
	}
	if w == nil {
		return fmt.Errorf("semanticmd emitter requires a writer")
	}

	output := renderDocument(document)
	_, err := io.WriteString(w, output)
	return err
}

func renderDocument(document *core.Document) string {
	roots := document.Kids
	if len(roots) == 0 {
		roots = collectPageContent(document.Pages)
	}

	blocks := make([]string, 0, len(roots))
	for _, element := range roots {
		block := renderContent(element, 0)
		if block == "" {
			continue
		}
		blocks = append(blocks, block)
	}
	return strings.Join(blocks, "\n\n")
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

func renderContent(element model.ContentElement, indent int) string {
	if element == nil {
		return ""
	}

	switch node := element.(type) {
	case *model.Heading:
		return renderHeading(node, indent)
	case *model.Paragraph:
		return renderParagraphLike(node.TextProperties, indent)
	case *model.Caption:
		return renderParagraphLike(node.TextProperties, indent)
	case *model.List:
		return renderList(node, indent)
	case *model.ListItem:
		return renderListItem(node, indent, false)
	case *model.Table:
		return renderTable(node, indent)
	case *model.TextBlock:
		return renderBlocks(node.Kids, indent)
	case *model.HeaderFooter:
		return renderBlocks(node.Kids, indent)
	default:
		text := collectInlineText(element)
		if text == "" {
			return ""
		}
		return indentLines(text, indent)
	}
}

func renderHeading(node *model.Heading, indent int) string {
	if node == nil {
		return ""
	}

	text := renderTextProperties(node.TextProperties)
	if text == "" {
		return ""
	}

	level := node.HeadingLevel
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	return indentLines(strings.Repeat("#", level)+" "+text, indent)
}

func renderParagraphLike(props model.TextProperties, indent int) string {
	text := renderTextProperties(props)
	if text == "" {
		return ""
	}
	return indentLines(text, indent)
}

func renderList(node *model.List, indent int) string {
	if node == nil || len(node.ListItems) == 0 {
		return ""
	}

	ordered := node.NumberingStyle != "" && node.NumberingStyle != "unordered"
	blocks := make([]string, 0, len(node.ListItems))
	for _, item := range node.ListItems {
		block := renderListItem(item, indent, ordered)
		if block == "" {
			continue
		}
		blocks = append(blocks, block)
	}
	return strings.Join(blocks, "\n\n")
}

func renderListItem(item *model.ListItem, indent int, ordered bool) string {
	if item == nil {
		return ""
	}

	marker := "-"
	if ordered {
		marker = "1."
	}

	text := renderTextProperties(item.TextProperties)
	line := marker
	if text != "" {
		line += " " + text
	}

	block := indentLines(line, indent)
	children := renderBlocks(item.Kids, indent+4)
	if children == "" {
		return block
	}
	return block + "\n\n" + children
}

func renderTable(node *model.Table, indent int) string {
	if node == nil || len(node.Rows) == 0 {
		return ""
	}

	columns := 0
	for _, row := range node.Rows {
		if len(row.Cells) > columns {
			columns = len(row.Cells)
		}
	}
	if columns == 0 {
		return ""
	}

	rows := make([][]string, 0, len(node.Rows))
	for _, row := range node.Rows {
		cells := make([]string, columns)
		for i := 0; i < columns; i++ {
			if i >= len(row.Cells) || row.Cells[i] == nil {
				continue
			}
			cells[i] = renderTableCell(row.Cells[i])
		}
		rows = append(rows, cells)
	}

	lines := make([]string, 0, len(rows)+1)
	lines = append(lines, renderMarkdownTableRow(rows[0]))
	lines = append(lines, renderMarkdownTableSeparator(columns))
	for _, row := range rows[1:] {
		lines = append(lines, renderMarkdownTableRow(row))
	}
	return indentLines(strings.Join(lines, "\n"), indent)
}

func renderMarkdownTableRow(cells []string) string {
	rendered := make([]string, len(cells))
	for i, cell := range cells {
		rendered[i] = cell
	}
	return "| " + strings.Join(rendered, " | ") + " |"
}

func renderMarkdownTableSeparator(columns int) string {
	cells := make([]string, columns)
	for i := range cells {
		cells[i] = "---"
	}
	return renderMarkdownTableRow(cells)
}

func renderTableCell(cell *model.TableCell) string {
	if cell == nil {
		return ""
	}
	text := strings.TrimSpace(collectInlineText(cell))
	if text == "" {
		return ""
	}
	return escapeMarkdownText(text, true)
}

func renderBlocks(elements []model.ContentElement, indent int) string {
	blocks := make([]string, 0, len(elements))
	for _, element := range elements {
		block := renderContent(element, indent)
		if block == "" {
			continue
		}
		blocks = append(blocks, block)
	}
	return strings.Join(blocks, "\n\n")
}

func renderTextProperties(props model.TextProperties) string {
	text := cleanText(props.Content)
	if text == "" {
		return ""
	}

	escaped := escapeMarkdownText(text, false)
	switch {
	case props.Bold && props.Italic:
		return "***" + escaped + "***"
	case props.Bold:
		return "**" + escaped + "**"
	case props.Italic:
		return "*" + escaped + "*"
	default:
		return escaped
	}
}

func collectInlineText(element model.ContentElement) string {
	if element == nil {
		return ""
	}

	switch node := element.(type) {
	case *model.Heading:
		return cleanText(node.TextProperties.Content)
	case *model.Paragraph:
		return cleanText(node.TextProperties.Content)
	case *model.Caption:
		return cleanText(node.TextProperties.Content)
	case *model.ListItem:
		parts := make([]string, 0, 1+len(node.Kids))
		if text := cleanText(node.TextProperties.Content); text != "" {
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
		return cleanText(node.Source)
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

var markdownEscaper = strings.NewReplacer(
	"\\", "\\\\",
	"`", "\\`",
	"*", `\*`,
	"_", `\_`,
	"[", `\[`,
	"]", `\]`,
	"#", `\#`,
	"|", `\|`,
	">", `\>`,
)

func escapeMarkdownText(text string, tableCell bool) string {
	if text == "" {
		return ""
	}

	escaped := markdownEscaper.Replace(text)
	if tableCell {
		return escaped
	}

	if needsLeadingListEscape(text) {
		return `\` + escaped
	}
	return escaped
}

func needsLeadingListEscape(text string) bool {
	if text == "" {
		return false
	}

	r, size := utf8.DecodeRuneInString(text)
	switch r {
	case '-', '+':
		if len(text) == size {
			return true
		}
		next, _ := utf8.DecodeRuneInString(text[size:])
		return unicode.IsSpace(next)
	default:
		if !unicode.IsDigit(r) {
			return false
		}
	}

	i := size
	for i < len(text) {
		r, sz := utf8.DecodeRuneInString(text[i:])
		if !unicode.IsDigit(r) {
			break
		}
		i += sz
	}
	if i >= len(text) {
		return false
	}

	switch text[i] {
	case '.', ')':
		if i+1 >= len(text) {
			return true
		}
		next, _ := utf8.DecodeRuneInString(text[i+1:])
		return unicode.IsSpace(next)
	default:
		return false
	}
}

func indentLines(text string, indent int) string {
	if text == "" || indent <= 0 {
		return text
	}

	prefix := strings.Repeat(" ", indent)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
