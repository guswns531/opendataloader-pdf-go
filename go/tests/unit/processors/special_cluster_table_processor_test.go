// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/processors"
)

func TestSpecialTableProcessorDetectsAlignedBorderlessTable(t *testing.T) {
	processor := &processors.SpecialTableProcessor{}
	ctx := containers.NewProcessorContext()

	elements := []entities.IObject{
		lineFromChunks(0, chunk("A1", 50, 120, 20, 10), chunk("B1", 110, 120, 20, 10)),
		lineFromChunks(1, chunk("A2", 50, 100, 20, 10), chunk("B2", 110, 100, 20, 10)),
	}

	result := processor.Process(elements, ctx)

	require.Len(t, result, 1)
	table, ok := result[0].(*entities.SemanticTable)
	require.True(t, ok)
	require.Len(t, table.Rows, 2)
	require.Len(t, table.Rows[0].Cells, 2)
	assert.Equal(t, "A1", cellText(table.Rows[0].Cells[0]))
	assert.Equal(t, "B1", cellText(table.Rows[0].Cells[1]))
	assert.Equal(t, "A2", cellText(table.Rows[1].Cells[0]))
	assert.Equal(t, "B2", cellText(table.Rows[1].Cells[1]))
}

func TestClusterTableProcessorDetectsChunkGridTable(t *testing.T) {
	processor := &processors.ClusterTableProcessor{}
	ctx := containers.NewProcessorContext()

	elements := []entities.IObject{
		chunk("A1", 50, 120, 18, 10),
		chunk("B1", 100, 120, 18, 10),
		chunk("A2", 50, 102, 18, 10),
		chunk("B2", 100, 102, 18, 10),
	}

	result := processor.Process(elements, ctx)

	require.Len(t, result, 1)
	table, ok := result[0].(*entities.SemanticTable)
	require.True(t, ok)
	require.Len(t, table.Rows, 2)
	require.Len(t, table.Rows[0].Cells, 2)
	assert.Equal(t, "A1", cellText(table.Rows[0].Cells[0]))
	assert.Equal(t, "B1", cellText(table.Rows[0].Cells[1]))
	assert.Equal(t, "A2", cellText(table.Rows[1].Cells[0]))
	assert.Equal(t, "B2", cellText(table.Rows[1].Cells[1]))
}

func chunk(text string, x, y, width, height float64) *entities.TextChunk {
	return &entities.TextChunk{
		BaseObject: entities.BaseObject{
			ID: text,
			BBox: entities.BoundingBox{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
				Page:   0,
			},
		},
		Text:     text,
		Baseline: y,
		FontStyle: entities.FontStyle{
			FontSize: 10,
		},
	}
}

func lineFromChunks(id int, chunks ...*entities.TextChunk) *entities.TextLine {
	box := chunks[0].BBox
	for _, c := range chunks[1:] {
		if c.BBox.X < box.X {
			box.X = c.BBox.X
		}
		right := c.BBox.X + c.BBox.Width
		boxRight := box.X + box.Width
		if right > boxRight {
			box.Width = right - box.X
		}
		if c.BBox.Y < box.Y {
			box.Y = c.BBox.Y
		}
		top := c.BBox.Y + c.BBox.Height
		boxTop := box.Y + box.Height
		if top > boxTop {
			box.Height = top - box.Y
		}
	}
	return &entities.TextLine{
		BaseObject: entities.BaseObject{
			ID:   chunks[0].Text + "-line",
			BBox: box,
		},
		Chunks:   chunks,
		Baseline: chunks[0].Baseline,
	}
}
