// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package generators_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/markdown"
)

func TestMarkdownGeneratorHeading(t *testing.T) {
	doc := &entities.Document{
		Pages: []*entities.Page{
			{
				Elements: []entities.IObject{
					&entities.SemanticHeading{
						Level: 1,
						Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "title"}}}},
					},
					&entities.SemanticHeading{
						Level: 2,
						Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "subtitle"}}}},
					},
				},
			},
		},
	}

	got, err := markdown.NewMarkdownGenerator(api.DefaultConfig(), false, false).Generate(doc)

	assert.NoError(t, err)
	assert.Contains(t, got, "# title")
	assert.Contains(t, got, "## subtitle")
}

func TestMarkdownGeneratorTable(t *testing.T) {
	doc := &entities.Document{
		Pages: []*entities.Page{
			{
				Elements: []entities.IObject{
					&entities.SemanticTable{
						Rows: []*entities.TableRow{
							{Cells: []*entities.TableCell{
								{Content: []entities.IObject{&entities.SemanticParagraph{Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "col1"}}}}}}},
								{Content: []entities.IObject{&entities.SemanticParagraph{Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "col2"}}}}}}},
							}},
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

	got, err := markdown.NewMarkdownGenerator(api.DefaultConfig(), false, false).Generate(doc)

	assert.NoError(t, err)
	assert.Contains(t, got, "| col1 | col2 |")
	assert.Contains(t, got, "| --- | --- |")
	assert.Contains(t, got, "| a | b |")
}

func TestMarkdownGeneratorTableTextChunks(t *testing.T) {
	doc := &entities.Document{
		Pages: []*entities.Page{
			{
				Elements: []entities.IObject{
					&entities.SemanticTable{
						Rows: []*entities.TableRow{
							{Cells: []*entities.TableCell{
								{Content: []entities.IObject{&entities.TextChunk{Text: "No."}}},
								{Content: []entities.IObject{&entities.TextChunk{Text: "Name"}}},
								{Content: []entities.IObject{&entities.TextChunk{Text: "Count"}}},
							}},
							{Cells: []*entities.TableCell{
								{Content: []entities.IObject{&entities.TextChunk{Text: "1"}}},
								{Content: []entities.IObject{&entities.TextChunk{Text: "Union of Youth Federations"}}},
								{Content: []entities.IObject{&entities.TextChunk{Text: "17,266"}}},
							}},
						},
					},
				},
			},
		},
	}

	got, err := markdown.NewMarkdownGenerator(api.DefaultConfig(), false, false).Generate(doc)

	assert.NoError(t, err)
	assert.Contains(t, got, "| No. | Name | Count |")
	assert.Contains(t, got, "| --- | --- | --- |")
	assert.Contains(t, got, "| 1 | Union of Youth Federations | 17,266 |")
}

func TestMarkdownGeneratorEscapesPipeCharactersInTableCells(t *testing.T) {
	doc := &entities.Document{
		Pages: []*entities.Page{
			{
				Elements: []entities.IObject{
					&entities.SemanticTable{
						Rows: []*entities.TableRow{
							{Cells: []*entities.TableCell{
								{Content: []entities.IObject{&entities.TextChunk{Text: "A|B"}}},
								{Content: []entities.IObject{&entities.TextChunk{Text: "Value"}}},
							}},
							{Cells: []*entities.TableCell{
								{Content: []entities.IObject{&entities.TextChunk{Text: "1|2"}}},
								{Content: []entities.IObject{&entities.TextChunk{Text: "3"}}},
							}},
						},
					},
				},
			},
		},
	}

	got, err := markdown.NewMarkdownGenerator(api.DefaultConfig(), false, false).Generate(doc)

	assert.NoError(t, err)
	assert.Contains(t, got, "| A\\|B | Value |")
	assert.Contains(t, got, "| 1\\|2 | 3 |")
}

func TestMarkdownGeneratorUsesHTMLForMergedCells(t *testing.T) {
	doc := &entities.Document{
		Pages: []*entities.Page{
			{
				Elements: []entities.IObject{
					&entities.SemanticTable{
						Rows: []*entities.TableRow{
							{Cells: []*entities.TableCell{
								{Content: []entities.IObject{&entities.TextChunk{Text: "Header"}}, Colspan: 2},
							}},
							{Cells: []*entities.TableCell{
								{Content: []entities.IObject{&entities.TextChunk{Text: "A"}}},
								{Content: []entities.IObject{&entities.TextChunk{Text: "B"}}},
							}},
						},
					},
				},
			},
		},
	}

	got, err := markdown.NewMarkdownGenerator(api.DefaultConfig(), true, false).Generate(doc)

	assert.NoError(t, err)
	assert.Contains(t, got, "<table>")
	assert.Contains(t, got, "<th colspan=\"2\">Header</th>")
	assert.NotContains(t, got, "| --- |")
}

func TestMarkdownGeneratorUsesHTMLForNestedTables(t *testing.T) {
	nested := &entities.SemanticTable{
		Rows: []*entities.TableRow{
			{Cells: []*entities.TableCell{{Content: []entities.IObject{&entities.TextChunk{Text: "inner"}}}}},
		},
	}
	doc := &entities.Document{
		Pages: []*entities.Page{
			{
				Elements: []entities.IObject{
					&entities.SemanticTable{
						Rows: []*entities.TableRow{
							{Cells: []*entities.TableCell{
								{Content: []entities.IObject{nested}},
							}},
						},
					},
				},
			},
		},
	}

	got, err := markdown.NewMarkdownGenerator(api.DefaultConfig(), true, false).Generate(doc)

	assert.NoError(t, err)
	assert.Contains(t, got, "<table>")
	assert.NotContains(t, got, "| --- |")
}

func TestMarkdownGeneratorPageSeparator(t *testing.T) {
	cfg := api.DefaultConfig()
	cfg.MarkdownPageSeparator = "--- page %page-number% ---"

	doc := &entities.Document{
		Pages: []*entities.Page{
			{Elements: []entities.IObject{&entities.SemanticParagraph{Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "first"}}}}}}},
			{Elements: []entities.IObject{&entities.SemanticParagraph{Lines: []*entities.TextLine{{Chunks: []*entities.TextChunk{{Text: "second"}}}}}}},
		},
	}

	got, err := markdown.NewMarkdownGenerator(cfg, false, false).Generate(doc)

	assert.NoError(t, err)
	assert.Contains(t, got, "--- page 2 ---")
	assert.NotContains(t, got, "--- page 1 ---")
}

func TestMarkdownGeneratorEmbeddedImage(t *testing.T) {
	cfg := api.DefaultConfig()
	cfg.ImageOutput = api.ImageOutputEmbedded
	cfg.ImageFormat = api.ImageFormatPNG

	doc := &entities.Document{
		Pages: []*entities.Page{
			{
				Elements: []entities.IObject{
					&entities.SemanticImage{
						Alt:  "diagram",
						Data: []byte{0x89, 0x50, 0x4e, 0x47},
					},
				},
			},
		},
	}

	got, err := markdown.NewMarkdownGenerator(cfg, false, true).Generate(doc)

	assert.NoError(t, err)
	assert.Contains(t, got, "![diagram](data:image/png;base64,")
}
