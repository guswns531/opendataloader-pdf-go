// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/processors"
)

func TestTableBorderProcessorDetectsSimpleRectangleTable(t *testing.T) {
	ctx := containers.NewProcessorContext()
	processor := &processors.TableBorderProcessor{}

	elements := []entities.IObject{
		textChunkInBox("A1", 20, 20, 20, 10, 0),
	}
	lineArts := rectangleLineArts(10, 10, 60, 40, 0)

	result := processor.Process(elements, lineArts, ctx)

	require.Len(t, result, 1)
	table, ok := firstSemanticTableInObjects(result)
	require.True(t, ok)
	require.Len(t, table.Rows, 1)
	require.Len(t, table.Rows[0].Cells, 1)
	assert.Equal(t, "A1", cellText(table.Rows[0].Cells[0]))
}

func TestTableBorderProcessorProcessesNestedTablesTwoLevels(t *testing.T) {
	ctx := containers.NewProcessorContext()
	processor := &processors.TableBorderProcessor{}

	elements, lineArts := nestedTableFixture(2, 0, 0, 220, 180, 0)
	result := processor.Process(elements, lineArts, ctx)

	require.Len(t, result, 1)
	topTable, ok := firstSemanticTableInObjects(result)
	require.True(t, ok)
	assert.Equal(t, 2, countNestedTables(topTable))

	nested, ok := firstNestedTable(topTable)
	require.True(t, ok)
	deepest := deepestCell(nested)
	require.NotNil(t, deepest)
	assert.Equal(t, "leaf", cellText(deepest))
}

func TestTableBorderProcessorStopsAtDepthLimit(t *testing.T) {
	ctx := containers.NewProcessorContext()
	processor := &processors.TableBorderProcessor{}

	elements, lineArts := nestedTableFixture(11, 0, 0, 660, 660, 0)
	result := processor.Process(elements, lineArts, ctx)

	require.Len(t, result, 1)
	topTable, ok := firstSemanticTableInObjects(result)
	require.True(t, ok, "expected a semantic table in processor output")

	depth := countNestedTables(topTable)
	assert.Equal(t, 10, depth)

	deepest := deepestCell(topTable)
	require.NotNil(t, deepest)
	require.NotEmpty(t, deepest.Content)
	_, nested := deepest.Content[0].(*entities.SemanticTable)
	assert.False(t, nested, "processing should stop when depth reaches 10")
}

func nestedTableFixture(levels int, x, y, width, height float64, page int) ([]entities.IObject, []*entities.LineArtChunk) {
	if levels == 0 {
		text := &entities.TextChunk{
			BaseObject: entities.BaseObject{
				ID: fmt.Sprintf("text-%0.f-%0.f", x, y),
				BBox: entities.BoundingBox{
					X:      x + 8,
					Y:      y + 8,
					Width:  width - 16,
					Height: height - 16,
					Page:   page,
				},
			},
			Text:     "leaf",
			Baseline: y + height/2,
		}
		return []entities.IObject{text}, nil
	}

	lines := rectangleLineArts(x, y, width, height, page)
	innerElements, innerLines := nestedTableFixture(levels-1, x+20, y+20, width-40, height-40, page)
	elements := append([]entities.IObject{}, innerElements...)
	return elements, append(lines, innerLines...)
}

func rectangleLineArts(x, y, width, height float64, page int) []*entities.LineArtChunk {
	return []*entities.LineArtChunk{
		lineArt(x, y+height, width, 0.5, page, true, false),
		lineArt(x, y, width, 0.5, page, true, false),
		lineArt(x, y, 0.5, height, page, false, true),
		lineArt(x+width, y, 0.5, height, page, false, true),
	}
}

func lineArt(x, y, width, height float64, page int, horizontal, vertical bool) *entities.LineArtChunk {
	return &entities.LineArtChunk{
		BaseObject: entities.BaseObject{
			BBox: entities.BoundingBox{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
				Page:   page,
			},
		},
		IsHorizontal: horizontal,
		IsVertical:   vertical,
	}
}

func textChunkInBox(text string, x, y, width, height float64, page int) *entities.TextChunk {
	return &entities.TextChunk{
		BaseObject: entities.BaseObject{
			ID: text,
			BBox: entities.BoundingBox{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
				Page:   page,
			},
		},
		Text:     text,
		Baseline: y + height/2,
	}
}

func cellText(cell *entities.TableCell) string {
	if cell == nil {
		return ""
	}
	result := ""
	for _, content := range cell.Content {
		switch typed := content.(type) {
		case *entities.TextChunk:
			result += typed.Text
		case *entities.TextLine:
			result += typed.GetText()
		}
	}
	return result
}

func countNestedTables(table *entities.SemanticTable) int {
	depth := 1
	current := table
	for {
		next, ok := firstNestedTable(current)
		if !ok {
			return depth
		}
		depth++
		current = next
	}
}

func firstNestedTable(table *entities.SemanticTable) (*entities.SemanticTable, bool) {
	for _, row := range table.Rows {
		for _, cell := range row.Cells {
			for _, content := range cell.Content {
				nested, ok := content.(*entities.SemanticTable)
				if ok {
					return nested, true
				}
			}
		}
	}
	return nil, false
}

func firstSemanticTableInObjects(objects []entities.IObject) (*entities.SemanticTable, bool) {
	for _, object := range objects {
		table, ok := object.(*entities.SemanticTable)
		if ok {
			return table, true
		}
	}
	return nil, false
}

func deepestCell(table *entities.SemanticTable) *entities.TableCell {
	current := table
	var last *entities.TableCell
	for {
		var next *entities.SemanticTable
		for _, row := range current.Rows {
			for _, cell := range row.Cells {
				last = cell
				for _, content := range cell.Content {
					nested, ok := content.(*entities.SemanticTable)
					if ok {
						next = nested
						break
					}
				}
				if next != nil {
					break
				}
			}
			if next != nil {
				break
			}
		}
		if next == nil {
			return last
		}
		current = next
	}
}
