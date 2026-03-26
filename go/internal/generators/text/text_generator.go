// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package text

import (
	"strconv"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

type TextGenerator struct {
	config *api.Config
}

func NewTextGenerator(config *api.Config) *TextGenerator {
	if config == nil {
		config = api.DefaultConfig()
	}
	return &TextGenerator{config: config}
}

func (g *TextGenerator) Generate(doc *entities.Document) (string, error) {
	if doc == nil {
		return "", nil
	}

	var b strings.Builder
	for pageIndex, page := range doc.Pages {
		if g.config.TextPageSeparator != "" && pageIndex > 0 {
			separator := g.config.TextPageSeparator
			if strings.Contains(separator, api.PageNumberString) {
				separator = strings.ReplaceAll(separator, api.PageNumberString, strconv.Itoa(pageIndex+1))
			}
			b.WriteString(separator)
			b.WriteString("\n")
		}

		for idx, element := range page.Elements {
			g.write(&b, element, 0)
			if idx < len(page.Elements)-1 {
				b.WriteString("\n")
			}
		}
		if pageIndex < len(doc.Pages)-1 {
			b.WriteString("\n")
		}
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func (g *TextGenerator) write(b *strings.Builder, object entities.IObject, indentLevel int) {
	switch v := object.(type) {
	case *entities.SemanticHeaderFooter:
		if g.config.IncludeHeaderFooter {
			for _, line := range v.Lines {
				g.writeLine(b, line.GetText(), indentLevel)
			}
		}
	case *entities.SemanticHeading:
		g.writeLine(b, joinLines(v.Lines, g.config.KeepLineBreaks), indentLevel)
	case *entities.SemanticParagraph:
		g.writeLine(b, joinLines(v.Lines, g.config.KeepLineBreaks), indentLevel)
	case *entities.TextLine:
		g.writeLine(b, v.GetText(), indentLevel)
	case *entities.PDFList:
		for _, item := range v.Items {
			g.writeLine(b, g.collectPlainText(item.Content), indentLevel)
			for _, content := range item.Content {
				switch content.(type) {
				case *entities.SemanticParagraph, *entities.SemanticHeading, *entities.TextLine:
				default:
					g.write(b, content, indentLevel+1)
				}
			}
		}
	case *entities.SemanticTable:
		for _, row := range v.Rows {
			var cells []string
			for _, cell := range row.Cells {
				text := strings.TrimSpace(g.collectPlainText(cell.Content))
				if text != "" {
					cells = append(cells, text)
				}
			}
			if len(cells) == 0 {
				continue
			}
			g.writeLine(b, strings.Join(cells, "\t"), indentLevel)
		}
	case *entities.SemanticCaption:
		g.writeLine(b, v.Text, indentLevel)
	case *entities.SemanticFormula:
		g.writeLine(b, v.LaTeX, indentLevel)
	}
}

func (g *TextGenerator) writeLine(b *strings.Builder, value string, indentLevel int) {
	if value == "" {
		return
	}
	lines := strings.Split(value, "\n")
	indent := strings.Repeat("  ", indentLevel)
	for _, line := range lines {
		line = strings.TrimSpace(strings.ReplaceAll(line, "\u0000", " "))
		if line == "" {
			continue
		}
		b.WriteString(indent)
		b.WriteString(line)
		b.WriteString("\n")
	}
}

func (g *TextGenerator) collectPlainText(contents []entities.IObject) string {
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
