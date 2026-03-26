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

func TestListProcessorBuildsOrderedList(t *testing.T) {
	processor := &processors.ListProcessor{}
	ctx := containers.NewProcessorContext()
	elements := []entities.IObject{
		listLine("1. first item", 50, 700, 120),
		listLine("2. second item", 50, 680, 140),
	}

	result := processor.Process(elements, ctx)

	require.Len(t, result, 1)
	list, ok := result[0].(*entities.PDFList)
	require.True(t, ok)
	require.Len(t, list.Items, 2)
	assert.True(t, list.IsOrdered)
	assert.Equal(t, "1.", list.Items[0].BulletText)
	assert.Equal(t, "2.", list.Items[1].BulletText)
}

func TestListProcessorMergesIndentedContinuationLine(t *testing.T) {
	processor := &processors.ListProcessor{}
	ctx := containers.NewProcessorContext()
	elements := []entities.IObject{
		listLine("• bullet item", 50, 700, 120),
		plainLine("continuation", 80, 688, 110),
	}

	result := processor.Process(elements, ctx)

	require.Len(t, result, 1)
	list, ok := result[0].(*entities.PDFList)
	require.True(t, ok)
	require.Len(t, list.Items, 1)
	require.Len(t, list.Items[0].Content, 2)
	assert.Equal(t, "•", list.Items[0].BulletText)
	assert.False(t, list.Items[0].IsOrdered)
}

func TestMergeListsAcrossPagesMergesAdjacentListsWithSamePattern(t *testing.T) {
	prevList := &entities.PDFList{
		BaseObject: entities.BaseObject{ID: "prev"},
		IsOrdered:  true,
		Items: []*entities.ListItem{
			{BulletText: "1.", IsOrdered: true},
		},
	}
	currList := &entities.PDFList{
		BaseObject: entities.BaseObject{ID: "curr"},
		IsOrdered:  true,
		Items: []*entities.ListItem{
			{BulletText: "2.", IsOrdered: true},
		},
	}
	pages := []*entities.Page{
		{Elements: []entities.IObject{plainLine("body", 50, 700, 100), prevList}},
		{Elements: []entities.IObject{currList, plainLine("next body", 50, 680, 100)}},
	}

	processors.MergeListsAcrossPages(pages)

	require.Len(t, prevList.Items, 2)
	assert.Equal(t, "1.", prevList.Items[0].BulletText)
	assert.Equal(t, "2.", prevList.Items[1].BulletText)
	require.Len(t, pages[1].Elements, 1)
	_, ok := pages[1].Elements[0].(*entities.PDFList)
	assert.False(t, ok)
}

func TestMergeListsAcrossPagesDoesNotMergeDifferentPatterns(t *testing.T) {
	prevList := &entities.PDFList{
		BaseObject: entities.BaseObject{ID: "prev"},
		IsOrdered:  false,
		Items: []*entities.ListItem{
			{BulletText: "•", IsOrdered: false},
		},
	}
	currList := &entities.PDFList{
		BaseObject: entities.BaseObject{ID: "curr"},
		IsOrdered:  false,
		Items: []*entities.ListItem{
			{BulletText: "-", IsOrdered: false},
		},
	}
	pages := []*entities.Page{
		{Elements: []entities.IObject{prevList}},
		{Elements: []entities.IObject{currList}},
	}

	processors.MergeListsAcrossPages(pages)

	require.Len(t, prevList.Items, 1)
	require.Len(t, pages[1].Elements, 1)
	_, ok := pages[1].Elements[0].(*entities.PDFList)
	assert.True(t, ok)
}

func listLine(text string, x, y, width float64) *entities.TextLine {
	return plainLine(text, x, y, width)
}

func plainLine(text string, x, y, width float64) *entities.TextLine {
	chunk := &entities.TextChunk{
		BaseObject: entities.BaseObject{
			BBox: entities.BoundingBox{X: x, Y: y, Width: width, Height: 12, Page: 0},
		},
		Text:      text,
		Baseline:  y,
		FontStyle: entities.FontStyle{FontSize: 12},
	}
	return &entities.TextLine{
		BaseObject: entities.BaseObject{
			BBox: chunk.BBox,
		},
		Chunks:   []*entities.TextChunk{chunk},
		Baseline: y,
	}
}
