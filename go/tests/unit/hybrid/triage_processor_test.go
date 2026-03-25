// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid_test

import (
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/hybrid"
	"github.com/stretchr/testify/assert"
)

func TestTriageTextOnlyPageRoutesToJava(t *testing.T) {
	processor := &hybrid.TriageProcessor{}
	page := &entities.Page{
		PageMetadata: entities.PageMetadata{Number: 0, Width: 595, Height: 842},
		Chunks: []*entities.TextChunk{
			textChunk("a", 10, 700, 40, 12, 700, "Hello"),
			textChunk("b", 55, 700, 60, 12, 700, "world"),
			textChunk("c", 10, 680, 80, 12, 680, "simple paragraph"),
		},
	}

	result := processor.Triage(page)

	assert.NotNil(t, result)
	assert.Equal(t, hybrid.TriageDecisionJava, result.Decision)
	assert.GreaterOrEqual(t, result.Confidence, 0.0)
	assert.LessOrEqual(t, result.Confidence, 1.0)
}

func TestTriageLineArtHeavyPageRoutesToBackend(t *testing.T) {
	processor := &hybrid.TriageProcessor{}
	page := &entities.Page{
		PageMetadata: entities.PageMetadata{Number: 0, Width: 595, Height: 842},
		Chunks: []*entities.TextChunk{
			textChunk("a", 10, 760, 40, 12, 760, "A"),
			textChunk("b", 60, 760, 40, 12, 760, "B"),
		},
		LineArts: []*entities.LineArtChunk{
			lineArt("l1", 10, 730, 200, 1, true, false),
			lineArt("l2", 10, 700, 200, 1, true, false),
			lineArt("l3", 10, 670, 200, 1, true, false),
			lineArt("l4", 10, 640, 200, 1, true, false),
			lineArt("l5", 10, 610, 200, 1, true, false),
			lineArt("l6", 10, 580, 200, 1, true, false),
			lineArt("l7", 10, 550, 200, 1, true, false),
			lineArt("l8", 10, 520, 200, 1, true, false),
		},
	}

	result := processor.Triage(page)

	assert.NotNil(t, result)
	assert.Equal(t, hybrid.TriageDecisionBackend, result.Decision)
	assert.GreaterOrEqual(t, result.Confidence, 0.0)
	assert.LessOrEqual(t, result.Confidence, 1.0)
}

func textChunk(id string, x, y, w, h, baseline float64, text string) *entities.TextChunk {
	return &entities.TextChunk{
		BaseObject: entities.BaseObject{
			ID: id,
			BBox: entities.BoundingBox{
				X:      x,
				Y:      y,
				Width:  w,
				Height: h,
				Page:   0,
			},
		},
		Text:     text,
		Baseline: baseline,
	}
}

func lineArt(id string, x, y, w, h float64, horizontal, vertical bool) *entities.LineArtChunk {
	return &entities.LineArtChunk{
		BaseObject: entities.BaseObject{
			ID: id,
			BBox: entities.BoundingBox{
				X:      x,
				Y:      y,
				Width:  w,
				Height: h,
				Page:   0,
			},
		},
		IsHorizontal: horizontal,
		IsVertical:   vertical,
	}
}
