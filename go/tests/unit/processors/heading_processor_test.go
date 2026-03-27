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

func TestHeadingProcessorDetectsStyleDistinctHeadingAtBodySize(t *testing.T) {
	processor := &processors.HeadingProcessor{}
	ctx := containers.NewProcessorContext()
	lines := []*entities.TextLine{
		headingLineWithFont("body paragraph text", 50, 700, 220, 11, false, false, "T1_0"),
		headingLineWithFont("more body text", 50, 680, 220, 11, false, false, "T1_0"),
		headingLineWithFont("7 Variants of SJ Observer Models", 50, 150, 180, 11, false, false, "T1_1"),
		headingLineWithFont("In this chapter, I have presented two variants", 80, 132, 320, 11, false, false, "T1_0"),
	}

	headings := processor.Process(lines, ctx)

	require.Len(t, headings, 1)
	assert.Equal(t, "7 Variants of SJ Observer Models", headings[0].Lines[0].GetText())
	assert.Equal(t, "T1 1", headings[0].FontFamily)
}

func TestHeadingProcessorIgnoresShortBodySizedLineWithoutStyleChange(t *testing.T) {
	processor := &processors.HeadingProcessor{}
	ctx := containers.NewProcessorContext()
	lines := []*entities.TextLine{
		headingLineWithFont("body paragraph text", 50, 700, 220, 11, false, false, "T1_0"),
		headingLineWithFont("more body text", 50, 680, 220, 11, false, false, "T1_0"),
		headingLineWithFont("Short note", 50, 150, 120, 11, false, false, "T1_0"),
		headingLineWithFont("continued body copy follows here", 80, 132, 320, 11, false, false, "T1_0"),
	}

	headings := processor.Process(lines, ctx)

	assert.Empty(t, headings)
}

func TestHeadingProcessorSkipsListItemLines(t *testing.T) {
	processor := &processors.HeadingProcessor{}
	ctx := containers.NewProcessorContext()
	line := headingLine("Appendix", 50, 700, 180, 16, true)
	line.InListItem = true

	headings := processor.Process([]*entities.TextLine{line}, ctx)

	assert.Empty(t, headings)
}

func TestHeadingProcessorSkipsTableCellLines(t *testing.T) {
	processor := &processors.HeadingProcessor{}
	ctx := containers.NewProcessorContext()
	line := headingLine("Total", 50, 700, 180, 16, true)
	line.InTableCell = true

	headings := processor.Process([]*entities.TextLine{line}, ctx)

	assert.Empty(t, headings)
}

func headingLine(text string, x, y, width, fontSize float64, bold bool) *entities.TextLine {
	return headingLineWithFont(text, x, y, width, fontSize, bold, false, "")
}

func headingLineWithFont(text string, x, y, width, fontSize float64, bold, italic bool, fontName string) *entities.TextLine {
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
			Italic:     italic,
			FontName:   fontName,
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
