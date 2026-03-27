// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/processors"
)

func TestTextLineProcessorGroupsChunksOnSameBaseline(t *testing.T) {
	processor := &processors.TextLineProcessor{}
	ctx := containers.NewProcessorContext()
	chunks := []*entities.TextChunk{
		newTextChunk("Hello", 10, 100, 20, 10, 12, 100),
		newTextChunk("world", 30, 100, 25, 10, 12, 100),
	}

	lines := processor.Process(chunks, nil, ctx)

	assert.Len(t, lines, 1)
	assert.Len(t, lines[0].Chunks, 2)
	assert.Equal(t, "Helloworld", lines[0].GetText())
}

func TestTextLineProcessorSeparatesDifferentBaselines(t *testing.T) {
	processor := &processors.TextLineProcessor{}
	ctx := containers.NewProcessorContext()
	chunks := []*entities.TextChunk{
		newTextChunk("Top", 10, 120, 20, 10, 12, 120),
		newTextChunk("Bottom", 10, 90, 35, 10, 12, 90),
	}

	lines := processor.Process(chunks, nil, ctx)

	assert.Len(t, lines, 2)
	assert.Equal(t, "Top", lines[0].GetText())
	assert.Equal(t, "Bottom", lines[1].GetText())
}

func TestTextLineProcessorInsertsSpaceAndLinksLineArtBullet(t *testing.T) {
	processor := &processors.TextLineProcessor{}
	ctx := containers.NewProcessorContext()
	chunks := []*entities.TextChunk{
		newTextChunk("Item", 40, 100, 24, 10, 12, 100),
		newTextChunk("one", 70, 100, 18, 10, 12, 100),
	}
	lineArts := []*entities.LineArtChunk{
		{
			BaseObject: entities.BaseObject{
				BBox: entities.BoundingBox{X: 28, Y: 100, Width: 8, Height: 4},
			},
			IsHorizontal: true,
		},
	}

	lines := processor.Process(chunks, lineArts, ctx)

	assert.Len(t, lines, 1)
	assert.Equal(t, "Item one", lines[0].GetText())
	assert.NotNil(t, lines[0].LineArtBullet)
}

func TestTextLineProcessorInsertsSpaceForSmallPositiveWordBoundaryGap(t *testing.T) {
	processor := &processors.TextLineProcessor{}
	ctx := containers.NewProcessorContext()
	chunks := []*entities.TextChunk{
		newTextChunk("In", 10, 100, 8, 10, 12, 100),
		newTextChunk("Proceedings", 19, 100, 55, 10, 12, 100),
	}

	lines := processor.Process(chunks, nil, ctx)

	assert.Len(t, lines, 1)
	assert.Equal(t, "In Proceedings", lines[0].GetText())
}

func TestTextLineProcessorSuppressesIntrawordSpaceAfterSingleLowercaseFragment(t *testing.T) {
	processor := &processors.TextLineProcessor{}
	ctx := containers.NewProcessorContext()
	chunks := []*entities.TextChunk{
		newTextChunk("we thus propose a", 50.112, 100, 94.855, 10, 12, 100),
		newTextChunk("m", 145.835, 100, 5.455, 10, 12, 100),
		newTextChunk("ulti-", 154.322, 100, 16.185, 10, 12, 100),
		newTextChunk("o", 172.507, 100, 5.455, 10, 12, 100),
		newTextChunk("bject", 177.962, 100, 27.273, 10, 12, 100),
		newTextChunk("r", 205.495, 100, 2.243, 10, 12, 100),
		newTextChunk("ectiﬁed", 209.738, 100, 35.832, 10, 12, 100),
		newTextChunk("a", 247.570, 100, 3.454, 10, 12, 100),
		newTextChunk("ttention", 253.024, 100, 43.636, 10, 12, 100),
	}

	lines := processor.Process(chunks, nil, ctx)

	assert.Len(t, lines, 1)
	assert.Equal(t, "we thus propose a multi-objectrectiﬁed attention", lines[0].GetText())
	assert.NotContains(t, lines[0].GetText(), "m ulti-")
	assert.NotContains(t, lines[0].GetText(), "multi- object")
	assert.NotContains(t, lines[0].GetText(), "objectr ectiﬁed")
	assert.NotContains(t, lines[0].GetText(), "a ttention")
}

func newTextChunk(text string, x, y, width, height, fontSize, baseline float64) *entities.TextChunk {
	return &entities.TextChunk{
		BaseObject: entities.BaseObject{
			BBox: entities.BoundingBox{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
			},
		},
		Text:      text,
		Baseline:  baseline,
		FontStyle: entities.FontStyle{FontSize: fontSize},
	}
}
