// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package html

import (
	stdhtml "html"
	"strconv"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/utils"
)

type HtmlGenerator struct {
	config       *api.Config
	tableNesting int
}

func NewHtmlGenerator(config *api.Config) *HtmlGenerator {
	if config == nil {
		config = api.DefaultConfig()
	}
	return &HtmlGenerator{config: config}
}

func (g *HtmlGenerator) Generate(doc *entities.Document) (string, error) {
	if doc == nil {
		return "", nil
	}

	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"und\">\n<head>\n<meta charset=\"utf-8\">\n<title>")
	b.WriteString(stdhtml.EscapeString(doc.Metadata.Title))
	b.WriteString("</title>\n</head>\n<body>\n")

	for pageIndex, page := range doc.Pages {
		g.writePageSeparator(&b, pageIndex)
		for _, element := range page.Elements {
			if err := g.write(&b, element); err != nil {
				return "", err
			}
		}
	}

	b.WriteString("</body>\n</html>")
	return b.String(), nil
}

func (g *HtmlGenerator) writePageSeparator(b *strings.Builder, pageIndex int) {
	if g.config.HTMLPageSeparator == "" || pageIndex <= 0 {
		return
	}
	separator := g.config.HTMLPageSeparator
	if strings.Contains(separator, api.PageNumberString) {
		separator = strings.ReplaceAll(separator, api.PageNumberString, strconv.Itoa(pageIndex+1))
	}
	b.WriteString(separator)
	b.WriteString(HTMLLineBreak)
}

func (g *HtmlGenerator) write(b *strings.Builder, object entities.IObject) error {
	switch v := object.(type) {
	case *entities.SemanticHeaderFooter:
		if g.config.IncludeHeaderFooter {
			for _, line := range v.Lines {
				g.writeTextTag(b, HTMLParagraphTag, HTMLParagraphCloseTag, line.GetText(), false)
			}
		}
	case *entities.SemanticHeading:
		g.writeHeading(b, v)
	case *entities.SemanticParagraph:
		g.writeParagraph(b, v)
	case *entities.SemanticTable:
		g.writeTable(b, v)
	case *entities.PDFList:
		g.writeList(b, v)
	case *entities.SemanticImage:
		g.writeImage(b, v)
	case *entities.SemanticFormula:
		g.writeFormula(b, v)
	case *entities.SemanticCaption:
		g.writeCaption(b, v)
	case *entities.TextLine:
		g.writeTextTag(b, HTMLParagraphTag, HTMLParagraphCloseTag, v.GetText(), false)
	}
	return nil
}

func (g *HtmlGenerator) writeHeading(b *strings.Builder, heading *entities.SemanticHeading) {
	level := heading.Level
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	tag := "h" + strconv.Itoa(level)
	g.writeTextTag(b, "<"+tag+">", "</"+tag+">", joinLines(heading.Lines, false), false)
}

func (g *HtmlGenerator) writeParagraph(b *strings.Builder, paragraph *entities.SemanticParagraph) {
	g.writeTextTag(b, HTMLParagraphTag, HTMLParagraphCloseTag, joinLines(paragraph.Lines, g.config.KeepLineBreaks), g.config.KeepLineBreaks)
}

func (g *HtmlGenerator) writeTable(b *strings.Builder, table *entities.SemanticTable) {
	g.tableNesting++
	defer func() { g.tableNesting-- }()

	b.WriteString(HTMLTableTag)
	b.WriteString(HTMLLineBreak)
	for rowIndex, row := range table.Rows {
		b.WriteString(HTMLTableRowTag)
		b.WriteString(HTMLLineBreak)
		for colIndex, cell := range row.Cells {
			if cell == nil {
				continue
			}
			if !cell.IsOrigin(rowIndex, colIndex) {
				continue
			}
			openTag := HTMLTableCellTag
			closeTag := HTMLTableCellCloseTag
			if rowIndex == 0 {
				openTag = HTMLTableHeaderTag
				closeTag = HTMLTableHeaderCloseTag
			}
			if cell.EffectiveColSpan() > 1 || cell.EffectiveRowSpan() > 1 {
				tagName := "td"
				if rowIndex == 0 {
					tagName = "th"
				}
				var attrs strings.Builder
				attrs.WriteString("<")
				attrs.WriteString(tagName)
				if cell.EffectiveColSpan() > 1 {
					attrs.WriteString(" colspan=\"")
					attrs.WriteString(strconv.Itoa(cell.EffectiveColSpan()))
					attrs.WriteString("\"")
				}
				if cell.EffectiveRowSpan() > 1 {
					attrs.WriteString(" rowspan=\"")
					attrs.WriteString(strconv.Itoa(cell.EffectiveRowSpan()))
					attrs.WriteString("\"")
				}
				attrs.WriteString(">")
				openTag = attrs.String()
				closeTag = "</" + tagName + ">"
			}
			b.WriteString(openTag)
			b.WriteString(g.renderCellContents(cell.Content))
			b.WriteString(closeTag)
			b.WriteString(HTMLLineBreak)
		}
		b.WriteString(HTMLTableRowCloseTag)
		b.WriteString(HTMLLineBreak)
	}
	b.WriteString(HTMLTableCloseTag)
	b.WriteString(HTMLLineBreak)
}

func (g *HtmlGenerator) writeList(b *strings.Builder, list *entities.PDFList) {
	if list.IsOrdered {
		b.WriteString(HTMLOrderedListTag)
	} else {
		b.WriteString(HTMLUnorderedListTag)
	}
	b.WriteString(HTMLLineBreak)
	for _, item := range list.Items {
		b.WriteString(HTMLListItemTag)
		itemText := strings.TrimSpace(g.collectPlainText(item.Content))
		if item.BulletText != "" && itemText == "" {
			itemText = item.BulletText
		}
		if itemText != "" {
			g.writeTextTag(b, HTMLParagraphTag, HTMLParagraphCloseTag, itemText, false)
		}
		for _, content := range item.Content {
			switch content.(type) {
			case *entities.SemanticHeading, *entities.SemanticParagraph, *entities.TextLine:
				continue
			default:
				g.write(b, content)
			}
		}
		b.WriteString(HTMLListItemCloseTag)
		b.WriteString(HTMLLineBreak)
	}
	if list.IsOrdered {
		b.WriteString(HTMLOrderedListCloseTag)
	} else {
		b.WriteString(HTMLUnorderedListCloseTag)
	}
	b.WriteString(HTMLLineBreak)
}

func (g *HtmlGenerator) writeImage(b *strings.Builder, image *entities.SemanticImage) {
	if g.config.ImageOutput == api.ImageOutputOff {
		return
	}
	src := image.ExternalPath
	if g.config.ImageOutput == api.ImageOutputEmbedded && len(image.Data) > 0 {
		src = utils.EncodeImageBase64(image.Data, g.config.ImageFormat)
	}
	if src == "" {
		return
	}
	b.WriteString("<img src=\"")
	b.WriteString(escapeAttribute(src))
	b.WriteString("\" alt=\"")
	b.WriteString(escapeAttribute(imageAlt(image)))
	b.WriteString("\">")
	b.WriteString(HTMLLineBreak)
}

func (g *HtmlGenerator) writeFormula(b *strings.Builder, formula *entities.SemanticFormula) {
	b.WriteString(HTMLMathDisplayTag)
	b.WriteString("\\[")
	b.WriteString(sanitizeString(formula.LaTeX))
	b.WriteString("\\]")
	b.WriteString(HTMLMathDisplayCloseTag)
	b.WriteString(HTMLLineBreak)
}

func (g *HtmlGenerator) writeCaption(b *strings.Builder, caption *entities.SemanticCaption) {
	b.WriteString(HTMLFigureCaptionTag)
	b.WriteString(sanitizeString(caption.Text))
	b.WriteString(HTMLFigureCaptionCloseTag)
	b.WriteString(HTMLLineBreak)
}

func (g *HtmlGenerator) renderCellContents(contents []entities.IObject) string {
	var parts []string
	for _, content := range contents {
		switch v := content.(type) {
		case *entities.SemanticHeading:
			parts = append(parts, sanitizeString(joinLines(v.Lines, false)))
		case *entities.SemanticParagraph:
			parts = append(parts, escapeWithBreaks(joinLines(v.Lines, g.config.KeepLineBreaks)))
		case *entities.TextLine:
			parts = append(parts, sanitizeString(v.GetText()))
		case *entities.SemanticCaption:
			parts = append(parts, sanitizeString(v.Text))
		case *entities.SemanticFormula:
			parts = append(parts, sanitizeString(v.LaTeX))
		case *entities.SemanticImage:
			src := v.ExternalPath
			if g.config.ImageOutput == api.ImageOutputEmbedded && len(v.Data) > 0 {
				src = utils.EncodeImageBase64(v.Data, g.config.ImageFormat)
			}
			if src != "" {
				parts = append(parts, "<img src=\""+escapeAttribute(src)+"\" alt=\""+escapeAttribute(imageAlt(v))+"\">")
			}
		}
	}
	return strings.Join(filterEmpty(parts), HTMLLineBreakTag)
}

func (g *HtmlGenerator) writeTextTag(b *strings.Builder, openTag, closeTag, value string, preserveBreaks bool) {
	if value == "" {
		return
	}
	b.WriteString(openTag)
	if preserveBreaks {
		b.WriteString(escapeWithBreaks(value))
	} else {
		b.WriteString(sanitizeString(value))
	}
	b.WriteString(closeTag)
	b.WriteString(HTMLLineBreak)
}

func (g *HtmlGenerator) collectPlainText(contents []entities.IObject) string {
	var parts []string
	for _, content := range contents {
		switch v := content.(type) {
		case *entities.SemanticHeading:
			parts = append(parts, joinLines(v.Lines, false))
		case *entities.SemanticParagraph:
			parts = append(parts, joinLines(v.Lines, false))
		case *entities.TextLine:
			parts = append(parts, v.GetText())
		case *entities.SemanticCaption:
			parts = append(parts, v.Text)
		case *entities.SemanticFormula:
			parts = append(parts, v.LaTeX)
		case *entities.PDFList:
			for _, item := range v.Items {
				parts = append(parts, g.collectPlainText(item.Content))
			}
		case *entities.SemanticTable:
			for _, row := range v.Rows {
				for _, cell := range row.Cells {
					if cell == nil {
						continue
					}
					parts = append(parts, g.collectPlainText(cell.Content))
				}
			}
		case *entities.SemanticHeaderFooter:
			if g.config.IncludeHeaderFooter {
				for _, line := range v.Lines {
					parts = append(parts, line.GetText())
				}
			}
		}
	}
	return strings.Join(filterEmpty(parts), " ")
}

func joinLines(lines []*entities.TextLine, keepLineBreaks bool) string {
	var parts []string
	for _, line := range lines {
		if line == nil {
			continue
		}
		text := strings.TrimSpace(line.GetText())
		if text == "" {
			continue
		}
		parts = append(parts, text)
	}
	if keepLineBreaks {
		return strings.Join(parts, "\n")
	}
	return strings.Join(parts, " ")
}

func escapeWithBreaks(value string) string {
	return strings.ReplaceAll(sanitizeString(value), "\n", HTMLLineBreakTag)
}

func imageAlt(image *entities.SemanticImage) string {
	if image.Alt != "" {
		return image.Alt
	}
	return "image"
}

func sanitizeString(value string) string {
	return stdhtml.EscapeString(strings.ReplaceAll(value, "\u0000", ""))
}

func escapeAttribute(value string) string {
	value = strings.ReplaceAll(value, "\u0000", "")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", "")
	return stdhtml.EscapeString(value)
}

func filterEmpty(values []string) []string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		filtered = append(filtered, value)
	}
	return filtered
}
