// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package markdown

import (
	"fmt"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/utils"
)

type MarkdownGenerator struct {
	config       *api.Config
	withHTML     bool
	withImages   bool
	tableNesting int
}

func NewMarkdownGenerator(config *api.Config, withHTML, withImages bool) *MarkdownGenerator {
	if config == nil {
		config = api.DefaultConfig()
	}
	return &MarkdownGenerator{config: config, withHTML: withHTML, withImages: withImages}
}

func (g *MarkdownGenerator) Generate(doc *entities.Document) (string, error) {
	if doc == nil {
		return "", nil
	}

	var b strings.Builder
	for pageIndex, page := range doc.Pages {
		if err := g.writePageSeparator(&b, pageIndex); err != nil {
			return "", err
		}

		for _, element := range page.Elements {
			if !g.isSupportedContent(element) {
				continue
			}
			if err := g.write(&b, element); err != nil {
				return "", err
			}
			g.writeContentsSeparator(&b)
		}
	}

	return strings.TrimRight(b.String(), "\n"), nil
}

func (g *MarkdownGenerator) writePageSeparator(b *strings.Builder, pageIndex int) error {
	if g.config.MarkdownPageSeparator == "" || pageIndex <= 0 {
		return nil
	}

	separator := g.config.MarkdownPageSeparator
	if strings.Contains(separator, api.PageNumberString) {
		separator = strings.ReplaceAll(separator, api.PageNumberString, fmt.Sprintf("%d", pageIndex+1))
	}
	b.WriteString(separator)
	g.writeContentsSeparator(b)
	return nil
}

func (g *MarkdownGenerator) writeContentsSeparator(b *strings.Builder) {
	b.WriteString(DoubleLineBreak)
}

func (g *MarkdownGenerator) write(b *strings.Builder, object entities.IObject) error {
	switch v := object.(type) {
	case *entities.SemanticHeaderFooter:
		if !g.config.IncludeHeaderFooter {
			return nil
		}
		return g.writeHeaderFooter(b, v)
	case *entities.SemanticHeading:
		return g.writeHeading(b, v)
	case *entities.SemanticParagraph:
		return g.writeParagraph(b, v)
	case *entities.SemanticTable:
		return g.writeTable(b, v)
	case *entities.PDFList:
		return g.writeList(b, v)
	case *entities.SemanticImage:
		return g.writeImage(b, v)
	case *entities.SemanticFormula:
		return g.writeFormula(b, v)
	case *entities.SemanticCaption:
		return g.writeCaption(b, v)
	case *entities.TextLine:
		b.WriteString(g.correctString(g.renderTextLine(v)))
	}

	return nil
}

func (g *MarkdownGenerator) writeHeaderFooter(b *strings.Builder, hf *entities.SemanticHeaderFooter) error {
	for _, line := range hf.Lines {
		if err := g.write(b, line); err != nil {
			return err
		}
		g.writeContentsSeparator(b)
	}
	return nil
}

func (g *MarkdownGenerator) writeHeading(b *strings.Builder, heading *entities.SemanticHeading) error {
	if !g.isInsideTable() {
		level := heading.Level
		if level < 1 {
			level = 1
		}
		if level > 6 {
			level = 6
		}
		b.WriteString(strings.Repeat(HeadingLevel, level))
		b.WriteString(Space)
	}

	b.WriteString(g.correctString(g.renderLines(heading.Lines, false)))
	return nil
}

func (g *MarkdownGenerator) writeParagraph(b *strings.Builder, paragraph *entities.SemanticParagraph) error {
	b.WriteString(g.correctString(g.renderLines(paragraph.Lines, g.config.KeepLineBreaks)))
	return nil
}

func (g *MarkdownGenerator) writeTable(b *strings.Builder, table *entities.SemanticTable) error {
	if g.withHTML && g.tableNeedsHTML(table) {
		return g.writeHTMLTable(b, table)
	}

	g.enterTable()
	defer g.leaveTable()

	columnCount := g.tableColumnCount(table)
	for rowIndex, row := range table.Rows {
		b.WriteString(TableColumnSeparator)
		for colIndex := 0; colIndex < columnCount; colIndex++ {
			b.WriteString(Space)
			if colIndex < len(row.Cells) && row.Cells[colIndex] != nil {
				b.WriteString(g.renderTableCell(row.Cells[colIndex]))
			}
			b.WriteString(Space)
			b.WriteString(TableColumnSeparator)
		}
		b.WriteString(LineBreak)
		if rowIndex == 0 {
			b.WriteString(TableColumnSeparator)
			for i := 0; i < columnCount; i++ {
				b.WriteString(Space)
				b.WriteString(TableHeaderSeparator)
				b.WriteString(Space)
				b.WriteString(TableColumnSeparator)
			}
			b.WriteString(LineBreak)
		}
	}
	return nil
}

func (g *MarkdownGenerator) writeList(b *strings.Builder, list *entities.PDFList) error {
	for idx, item := range list.Items {
		if !g.isInsideTable() {
			prefix := ListItem
			if item.IsOrdered || list.IsOrdered {
				prefix = fmt.Sprintf("%d.", idx+1)
			}
			b.WriteString(strings.Repeat(Indent, max(item.Level, 0)))
			b.WriteString(prefix)
			b.WriteString(Space)
		}

		text := g.collectPlainText(item.Content, g.isInsideTable())
		if item.BulletText != "" && text == "" {
			text = item.BulletText
		}
		b.WriteString(g.correctString(text))

		if len(item.Content) > 0 && !g.isInsideTable() {
			b.WriteString(LineBreak)
		}

		if !g.isInsideTable() {
			for _, content := range item.Content {
				if !g.isSupportedContent(content) {
					continue
				}
				if _, ok := content.(*entities.TextLine); ok {
					continue
				}
				if _, ok := content.(*entities.SemanticParagraph); ok {
					continue
				}
				if err := g.write(b, content); err != nil {
					return err
				}
				g.writeContentsSeparator(b)
			}
		}
		if idx < len(list.Items)-1 {
			b.WriteString(LineBreak)
		}
	}
	return nil
}

func (g *MarkdownGenerator) writeImage(b *strings.Builder, image *entities.SemanticImage) error {
	if !g.withImages || g.config.ImageOutput == api.ImageOutputOff {
		return nil
	}

	imageSource := ""
	switch g.config.ImageOutput {
	case api.ImageOutputEmbedded:
		if len(image.Data) == 0 {
			return nil
		}
		imageSource = utils.EncodeImageBase64(image.Data, g.config.ImageFormat)
	case api.ImageOutputExternal:
		imageSource = image.ExternalPath
	default:
		return nil
	}

	if imageSource == "" {
		return nil
	}

	alt := image.Alt
	if alt == "" {
		alt = "image"
	}
	b.WriteString(fmt.Sprintf(ImageFormat, g.correctString(alt), imageSource))
	return nil
}

func (g *MarkdownGenerator) writeFormula(b *strings.Builder, formula *entities.SemanticFormula) error {
	b.WriteString(MathBlockStart)
	b.WriteString(LineBreak)
	b.WriteString(g.correctString(formula.LaTeX))
	b.WriteString(LineBreak)
	b.WriteString(MathBlockEnd)
	return nil
}

func (g *MarkdownGenerator) writeCaption(b *strings.Builder, caption *entities.SemanticCaption) error {
	b.WriteString("*")
	b.WriteString(g.correctString(caption.Text))
	b.WriteString("*")
	return nil
}

func (g *MarkdownGenerator) writeHTMLTable(b *strings.Builder, table *entities.SemanticTable) error {
	g.enterTable()
	defer g.leaveTable()

	b.WriteString(HTMLTableTag)
	b.WriteString(LineBreak)
	for rowIndex, row := range table.Rows {
		b.WriteString(Indent)
		b.WriteString(HTMLTableRowTag)
		b.WriteString(LineBreak)
		for _, cell := range row.Cells {
			if cell == nil {
				continue
			}
			tag := "td"
			if rowIndex == 0 {
				tag = "th"
			}
			b.WriteString(Indent)
			b.WriteString(Indent)
			b.WriteString("<")
			b.WriteString(tag)
			if cell.Colspan > 1 {
				b.WriteString(fmt.Sprintf(" colspan=\"%d\"", cell.Colspan))
			}
			if cell.Rowspan > 1 {
				b.WriteString(fmt.Sprintf(" rowspan=\"%d\"", cell.Rowspan))
			}
			b.WriteString(">")
			b.WriteString(g.renderTableCell(cell))
			b.WriteString(fmt.Sprintf("</%s>", tag))
			b.WriteString(LineBreak)
		}
		b.WriteString(Indent)
		b.WriteString(HTMLTableRowCloseTag)
		b.WriteString(LineBreak)
	}
	b.WriteString(HTMLTableCloseTag)
	b.WriteString(LineBreak)
	return nil
}

func (g *MarkdownGenerator) renderTableCell(cell *entities.TableCell) string {
	if cell == nil {
		return Space
	}

	value := strings.TrimSpace(g.collectPlainText(cell.Content, true))
	if value == "" {
		return Space
	}
	if g.withHTML && g.tableNeedsHTMLContent(cell.Content) {
		return value
	}
	return value
}

func (g *MarkdownGenerator) collectPlainText(contents []entities.IObject, forTable bool) string {
	var parts []string
	for _, content := range contents {
		if !g.isSupportedContent(content) {
			continue
		}
		switch v := content.(type) {
		case *entities.SemanticHeading:
			parts = append(parts, g.renderLines(v.Lines, false))
		case *entities.SemanticParagraph:
			parts = append(parts, g.renderLines(v.Lines, g.config.KeepLineBreaks))
		case *entities.TextLine:
			parts = append(parts, g.renderTextLine(v))
		case *entities.SemanticCaption:
			parts = append(parts, v.Text)
		case *entities.SemanticFormula:
			parts = append(parts, v.LaTeX)
		case *entities.PDFList:
			var itemParts []string
			for _, item := range v.Items {
				itemParts = append(itemParts, g.collectPlainText(item.Content, forTable))
			}
			parts = append(parts, strings.Join(itemParts, Space))
		case *entities.SemanticImage:
			if forTable {
				parts = append(parts, imagePlaceholder(v))
			}
		case *entities.SemanticTable:
			if g.withHTML {
				var nested strings.Builder
				_ = g.writeHTMLTable(&nested, v)
				parts = append(parts, strings.TrimSpace(nested.String()))
			}
		case *entities.SemanticHeaderFooter:
			if g.config.IncludeHeaderFooter {
				for _, line := range v.Lines {
					parts = append(parts, g.renderTextLine(line))
				}
			}
		}
	}

	separator := Space
	if forTable && g.config.KeepLineBreaks {
		separator = HTMLLineBreakTag
	}
	value := strings.Join(filterEmpty(parts), separator)
	if forTable && !g.config.KeepLineBreaks {
		value = strings.ReplaceAll(value, LineBreak, Space)
	}
	return g.correctString(value)
}

func (g *MarkdownGenerator) renderLines(lines []*entities.TextLine, keepLineBreaks bool) string {
	var parts []string
	for _, line := range lines {
		text := g.renderTextLine(line)
		if text == "" {
			continue
		}
		if keepLineBreaks {
			parts = append(parts, text)
		} else {
			parts = append(parts, strings.TrimSpace(text))
		}
	}
	if keepLineBreaks {
		joined := strings.Join(parts, LineBreak)
		if joined != "" {
			joined += LineBreak
		}
		if g.isInsideTable() {
			joined = strings.ReplaceAll(joined, LineBreak, HTMLLineBreakTag)
		}
		return joined
	}
	separator := Space
	if g.isInsideTable() {
		separator = Space
	}
	return strings.Join(parts, separator)
}

func (g *MarkdownGenerator) renderTextLine(line *entities.TextLine) string {
	if line == nil {
		return ""
	}
	return g.correctString(line.GetText())
}

func (g *MarkdownGenerator) correctString(value string) string {
	if value == "" {
		return value
	}
	replacement := g.config.ReplaceInvalidChars
	if replacement == "" {
		replacement = Space
	}
	return strings.ReplaceAll(value, "\u0000", replacement)
}

func (g *MarkdownGenerator) isSupportedContent(content entities.IObject) bool {
	switch content.(type) {
	case *entities.SemanticHeaderFooter:
		return g.config.IncludeHeaderFooter
	case *entities.SemanticHeading, *entities.SemanticParagraph, *entities.TextLine,
		*entities.SemanticFormula, *entities.SemanticImage, *entities.SemanticTable,
		*entities.PDFList, *entities.SemanticCaption:
		return true
	default:
		return false
	}
}

func (g *MarkdownGenerator) enterTable() {
	g.tableNesting++
}

func (g *MarkdownGenerator) leaveTable() {
	if g.tableNesting > 0 {
		g.tableNesting--
	}
}

func (g *MarkdownGenerator) isInsideTable() bool {
	return g.tableNesting > 0
}

func (g *MarkdownGenerator) tableNeedsHTML(table *entities.SemanticTable) bool {
	for _, row := range table.Rows {
		for _, cell := range row.Cells {
			if cell != nil && (cell.Colspan > 1 || cell.Rowspan > 1) {
				return true
			}
		}
	}
	return false
}

func (g *MarkdownGenerator) tableNeedsHTMLContent(contents []entities.IObject) bool {
	for _, content := range contents {
		if _, ok := content.(*entities.SemanticTable); ok {
			return true
		}
	}
	return false
}

func (g *MarkdownGenerator) tableColumnCount(table *entities.SemanticTable) int {
	maxColumns := 0
	for _, row := range table.Rows {
		if len(row.Cells) > maxColumns {
			maxColumns = len(row.Cells)
		}
	}
	if maxColumns == 0 {
		return 1
	}
	return maxColumns
}

func filterEmpty(values []string) []string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		filtered = append(filtered, trimmed)
	}
	return filtered
}

func imagePlaceholder(image *entities.SemanticImage) string {
	if image.Alt != "" {
		return image.Alt
	}
	return "image"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
