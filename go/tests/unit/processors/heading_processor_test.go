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

func TestHeadingProcessorDetectsLargeBoldTitle(t *testing.T) {
	processor := &processors.HeadingProcessor{}
	ctx := containers.NewProcessorContext()
	lines := []*entities.TextLine{
		headingLine("Document Title", 50, 700, 180, 16, true),
		headingLine("body paragraph text", 50, 660, 220, 12, false),
		headingLine("more body text", 50, 640, 220, 12, false),
	}

	headings := processor.Process(lines, ctx)

	require.Len(t, headings, 1)
	assert.Equal(t, "Document Title", headings[0].Lines[0].GetText())
	assert.Equal(t, 16.0, headings[0].FontSize)
	assert.True(t, headings[0].IsBold)
}

func TestHeadingProcessorIgnoresLongBodyLine(t *testing.T) {
	processor := &processors.HeadingProcessor{}
	ctx := containers.NewProcessorContext()
	lines := []*entities.TextLine{
		headingLine("short body", 50, 700, 180, 12, false),
		headingLine("this is a very long line that should not become a heading because it looks like normal body text even if it is alone on the page and has many words in it", 50, 660, 500, 12, false),
	}

	headings := processor.Process(lines, ctx)

	assert.Empty(t, headings)
}

func headingLine(text string, x, y, width, fontSize float64, bold bool) *entities.TextLine {
	chunk := &entities.TextChunk{
		BaseObject: entities.BaseObject{
			BBox: entities.BoundingBox{X: x, Y: y, Width: width, Height: fontSize, Page: 0},
		},
		Text:     text,
		Baseline: y,
		FontStyle: entities.FontStyle{
			FontSize:   fontSize,
			FontWeight: map[bool]float64{true: 700, false: 400}[bold],
			Bold:       bold,
		},
	}
	return &entities.TextLine{
		BaseObject: entities.BaseObject{
			BBox: chunk.BBox,
		},
		Chunks:   []*entities.TextChunk{chunk},
		Baseline: y,
	}
}
