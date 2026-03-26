// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package processors_test

import (
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/processors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHybridDocumentProcessorFallsBackToJavaWhenHealthCheckFails(t *testing.T) {
	javaProcessor := processors.NewDocumentProcessor()
	processor := processors.NewHybridDocumentProcessor(javaProcessor)
	cfg := api.DefaultConfig()
	cfg.Hybrid = api.HybridDoclingFast
	cfg.HybridURL = "http://127.0.0.1:1"
	cfg.HybridFallback = true

	doc := &entities.Document{
		Pages: []*entities.Page{
			{
				PageMetadata: entities.PageMetadata{Number: 0, Width: 595, Height: 842},
				Chunks: []*entities.TextChunk{
					{
						BaseObject: entities.BaseObject{
							ID:   "1",
							BBox: entities.BoundingBox{X: 40, Y: 700, Width: 100, Height: 12, Page: 0},
						},
						Text:      "Fallback works",
						Baseline:  700,
						FontStyle: entities.FontStyle{FontSize: 12},
					},
				},
			},
		},
	}

	result, err := processor.Process(doc, "/tmp/nonexistent.pdf", cfg, containers.NewProcessorContext())
	require.NoError(t, err)
	require.Len(t, result.Pages, 1)
	require.NotEmpty(t, result.Pages[0].Elements)

	_, ok := result.Pages[0].Elements[0].(*entities.SemanticParagraph)
	assert.True(t, ok)
}
