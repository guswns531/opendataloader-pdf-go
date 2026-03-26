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

func TestParagraphProcessorMergesConsecutiveLines(t *testing.T) {
	processor := &processors.ParagraphProcessor{}
	ctx := containers.NewProcessorContext()
	lines := []*entities.TextLine{
		newTextLine("First line", 50, 120, 120, 12, 12, 0),
		newTextLine("Second line", 50, 104, 108, 12, 12, 0),
	}

	paragraphs := processor.Process(lines, ctx)

	assert.Len(t, paragraphs, 1)
	assert.Len(t, paragraphs[0].Lines, 2)
	assert.Equal(t, entities.AlignLeft, paragraphs[0].Alignment)
}

func TestParagraphProcessorDoesNotMergeDifferentFontSizes(t *testing.T) {
	processor := &processors.ParagraphProcessor{}
	ctx := containers.NewProcessorContext()
	lines := []*entities.TextLine{
		newTextLine("Heading-ish", 50, 120, 120, 12, 16, 0),
		newTextLine("Body text", 50, 104, 108, 12, 12, 0),
	}

	paragraphs := processor.Process(lines, ctx)

	assert.Len(t, paragraphs, 2)
	assert.Len(t, paragraphs[0].Lines, 1)
	assert.Len(t, paragraphs[1].Lines, 1)
}

func newTextLine(text string, x, y, width, height, fontSize float64, page int) *entities.TextLine {
	chunk := &entities.TextChunk{
		BaseObject: entities.BaseObject{
			BBox: entities.BoundingBox{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
				Page:   page,
			},
		},
		Text:      text,
		Baseline:  y,
		FontStyle: entities.FontStyle{FontSize: fontSize},
	}
	return &entities.TextLine{
		BaseObject: entities.BaseObject{
			BBox: chunk.BBox,
		},
		Chunks:   []*entities.TextChunk{chunk},
		Baseline: y,
	}
}
