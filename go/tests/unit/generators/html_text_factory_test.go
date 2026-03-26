// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package generators_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/generators"
	htmlgen "github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/html"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/markdown"
	textgen "github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/text"
)

func TestHtmlGeneratorWrapsDocumentAndSkipsFirstPageSeparator(t *testing.T) {
	cfg := api.DefaultConfig()
	cfg.HTMLPageSeparator = "<hr data-page=\"%page-number%\">"

	doc := &entities.Document{
		Metadata: entities.DocumentMetadata{Title: "Doc"},
		Pages: []*entities.Page{
			{Elements: []entities.IObject{&entities.SemanticParagraph{Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "first"}}}}}}},
			{Elements: []entities.IObject{&entities.SemanticParagraph{Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "second"}}}}}}},
		},
	}

	got, err := htmlgen.NewHtmlGenerator(cfg).Generate(doc)

	require.NoError(t, err)
	assert.Contains(t, got, "<!DOCTYPE html>")
	assert.Contains(t, got, "<body>")
	assert.Contains(t, got, "<p>first</p>")
	assert.Contains(t, got, "<p>second</p>")
	assert.Contains(t, got, "<hr data-page=\"2\">")
	assert.NotContains(t, got, "<hr data-page=\"1\">")
}

func TestTextGeneratorSkipsFirstPageSeparator(t *testing.T) {
	cfg := api.DefaultConfig()
	cfg.TextPageSeparator = "--- page %page-number% ---"

	doc := &entities.Document{
		Pages: []*entities.Page{
			{Elements: []entities.IObject{&entities.SemanticParagraph{Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "first"}}}}}}},
			{Elements: []entities.IObject{&entities.SemanticParagraph{Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "second"}}}}}}},
		},
	}

	got, err := textgen.NewTextGenerator(cfg).Generate(doc)

	require.NoError(t, err)
	assert.Contains(t, got, "first")
	assert.Contains(t, got, "second")
	assert.Contains(t, got, "--- page 2 ---")
	assert.NotContains(t, got, "--- page 1 ---")
}

func TestMarkdownHTMLGeneratorUsesHTMLTableForComplexTable(t *testing.T) {
	cfg := api.DefaultConfig()
	doc := &entities.Document{
		Pages: []*entities.Page{
			{
				Elements: []entities.IObject{
					&entities.SemanticTable{
						Rows: []*entities.TableRow{
							{Cells: []*entities.TableCell{{Colspan: 2, Content: []entities.IObject{&entities.SemanticParagraph{Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "header"}}}}}}}}},
							{Cells: []*entities.TableCell{
								{Content: []entities.IObject{&entities.SemanticParagraph{Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "a"}}}}}}},
								{Content: []entities.IObject{&entities.SemanticParagraph{Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "b"}}}}}}},
							}},
						},
					},
				},
			},
		},
	}

	got, err := markdown.NewMarkdownHTMLGenerator(cfg).Generate(doc)

	require.NoError(t, err)
	assert.Contains(t, got, "<table>")
	assert.Contains(t, got, "<th colspan=\"2\">header</th>")
	assert.NotContains(t, got, "| --- |")
}

func TestGeneratorFactoryReturnsExpectedGenerators(t *testing.T) {
	cfg := api.DefaultConfig()

	assert.IsType(t, &textgen.TextGenerator{}, generators.GetGenerator(api.FormatText, cfg))
	assert.IsType(t, &htmlgen.HtmlGenerator{}, generators.GetGenerator(api.FormatHTML, cfg))
	assert.IsType(t, &markdown.MarkdownGenerator{}, generators.GetGenerator(api.FormatMarkdown, cfg))
	assert.IsType(t, &markdown.MarkdownGenerator{}, generators.GetGenerator(api.FormatMarkdownWithHTML, cfg))
	assert.IsType(t, &markdown.MarkdownGenerator{}, generators.GetGenerator(api.FormatMarkdownWithImages, cfg))
	assert.Nil(t, generators.GetGenerator(api.FormatJSON, cfg))
}
