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
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/utils"
)

func TestLevelProcessorAssignsDescendingHeadingLevels(t *testing.T) {
	headings := []*entities.SemanticHeading{
		{FontSize: 24},
		{FontSize: 18},
		{FontSize: 18},
		{FontSize: 14},
	}

	result := (&processors.LevelProcessor{}).Process(headings)

	require.Len(t, result, 4)
	assert.Equal(t, 1, result[0].Level)
	assert.Equal(t, 2, result[1].Level)
	assert.Equal(t, 2, result[2].Level)
	assert.Equal(t, 3, result[3].Level)
}

func TestCaptionProcessorCreatesSemanticCaption(t *testing.T) {
	ctx := containers.NewProcessorContext()
	line := plainLine("Figure 1. Example caption", 50, 700, 180)

	result := (&processors.CaptionProcessor{}).Process([]entities.IObject{line}, ctx)

	require.Len(t, result, 1)
	caption, ok := result[0].(*entities.SemanticCaption)
	require.True(t, ok)
	assert.Equal(t, "figure", caption.RefType)
	assert.Equal(t, "Figure 1. Example caption", caption.Text)
}

func TestHeaderFooterProcessorDropsHeaderFooterByDefault(t *testing.T) {
	ctx := containers.NewProcessorContext()
	elements := []entities.IObject{
		plainLine("header", 50, 950, 100),
		plainLine("body", 50, 500, 100),
		plainLine("footer", 50, 20, 100),
	}

	result := (&processors.HeaderFooterProcessor{}).Process(elements, 1000, false, ctx)

	require.Len(t, result, 1)
	assert.Equal(t, "body", result[0].(*entities.TextLine).GetText())
}

func TestHeaderFooterProcessorWrapsHeaderFooterWhenIncluded(t *testing.T) {
	ctx := containers.NewProcessorContext()
	elements := []entities.IObject{
		plainLine("header", 50, 950, 100),
		plainLine("body", 50, 500, 100),
		plainLine("footer", 50, 20, 100),
	}

	result := (&processors.HeaderFooterProcessor{}).Process(elements, 1000, true, ctx)

	require.Len(t, result, 3)
	header, ok := result[0].(*entities.SemanticHeaderFooter)
	require.True(t, ok)
	assert.True(t, header.IsHeader)
	assert.Equal(t, "header", header.Lines[0].GetText())
	footer, ok := result[2].(*entities.SemanticHeaderFooter)
	require.True(t, ok)
	assert.False(t, footer.IsHeader)
	assert.Equal(t, "footer", footer.Lines[0].GetText())
}

func TestBulletedParagraphUtilsDetectsBulletsAndIndent(t *testing.T) {
	assert.True(t, utils.IsOrderedBullet("1. item"))
	assert.True(t, utils.IsUnorderedBullet("• item"))
	assert.False(t, utils.IsUnorderedBullet("1. item"))

	line := plainLine("• indented item", 120, 600, 120)
	assert.Equal(t, 1, utils.GetIndentLevel(line, 1000))
}
