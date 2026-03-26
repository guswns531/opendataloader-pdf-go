// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/processors"
)

func TestContentFilterRemovesHiddenTinyOffPageAndHiddenOCGByDefault(t *testing.T) {
	chunks := []*entities.TextChunk{
		{Text: "visible", FontStyle: entities.FontStyle{FontSize: 12}, BaseObject: entities.BaseObject{BBox: entities.BoundingBox{X: 10, Y: 10, Width: 20, Height: 10}}},
		{Text: "hidden", IsHidden: true, FontStyle: entities.FontStyle{FontSize: 12}, BaseObject: entities.BaseObject{BBox: entities.BoundingBox{X: 10, Y: 10, Width: 20, Height: 10}}},
		{Text: "tiny", FontStyle: entities.FontStyle{FontSize: 0.8}, BaseObject: entities.BaseObject{BBox: entities.BoundingBox{X: 10, Y: 10, Width: 20, Height: 0.8}}},
		{Text: "offpage", FontStyle: entities.FontStyle{FontSize: 12}, BaseObject: entities.BaseObject{BBox: entities.BoundingBox{X: 110, Y: 10, Width: 20, Height: 10}}},
		{Text: "hidden-ocg", IsHiddenOCG: true, FontStyle: entities.FontStyle{FontSize: 12}, BaseObject: entities.BaseObject{BBox: entities.BoundingBox{X: 10, Y: 10, Width: 20, Height: 10}}},
	}

	filtered := processors.FilterContent(chunks, nil, 100, 100)

	assert.Len(t, filtered, 1)
	assert.Equal(t, "visible", filtered[0].Text)
}

func TestContentFilterHonorsDisabledFlagsFromAPIConfig(t *testing.T) {
	chunks := []*entities.TextChunk{
		{Text: "hidden", IsHidden: true, FontStyle: entities.FontStyle{FontSize: 12}, BaseObject: entities.BaseObject{BBox: entities.BoundingBox{X: 10, Y: 10, Width: 20, Height: 10}}},
		{Text: "tiny", FontStyle: entities.FontStyle{FontSize: 0.8}, BaseObject: entities.BaseObject{BBox: entities.BoundingBox{X: 10, Y: 10, Width: 20, Height: 0.8}}},
	}

	config := api.FilterConfigFromStrings([]string{"hidden-text,tiny"})
	filtered := processors.FilterContent(chunks, config, 100, 100)

	assert.Len(t, filtered, 2)
	assert.Equal(t, "hidden", filtered[0].Text)
	assert.Equal(t, "tiny", filtered[1].Text)
}
