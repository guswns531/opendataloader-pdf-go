// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/hybrid"
)

func TestTransformDoclingResponseBuildsPagesAndElements(t *testing.T) {
	payload := []byte(`{
	  "document": {
	    "json_content": {
	      "pages": {
	        "1": {"width": 600, "height": 800}
	      },
	      "texts": [
	        {
	          "label": "section_header",
	          "text": "Intro",
	          "meta": {"level": 2},
	          "prov": [{"page_no": 1, "bbox": {"l": 10, "t": 20, "r": 110, "b": 60, "coord_origin": "TOPLEFT"}}]
	        }
	      ],
	      "tables": [
	        {
	          "prov": [{"page_no": 1, "bbox": {"l": 100, "t": 100, "r": 300, "b": 220, "coord_origin": "TOPLEFT"}}],
	          "data": {
	            "grid": [[{}, {}], [{}, {}]],
	            "table_cells": [
	              {"start_row_offset_idx": 0, "start_col_offset_idx": 0, "row_span": 1, "col_span": 1, "text": "A1"},
	              {"start_row_offset_idx": 0, "start_col_offset_idx": 1, "row_span": 1, "col_span": 1, "text": "B1"},
	              {"start_row_offset_idx": 1, "start_col_offset_idx": 0, "row_span": 1, "col_span": 1, "text": "A2"},
	              {"start_row_offset_idx": 1, "start_col_offset_idx": 1, "row_span": 1, "col_span": 1, "text": "B2"}
	            ]
	          }
	        }
	      ],
	      "pictures": [
	        {
	          "prov": [{"page_no": 1, "bbox": {"l": 320, "t": 100, "r": 420, "b": 200, "coord_origin": "TOPLEFT"}}],
	          "annotations": [{"kind": "description", "text": "diagram"}]
	        }
	      ]
	    }
	  },
	  "status": "success"
	}`)

	pages, err := hybrid.TransformDoclingResponse(payload)

	require.NoError(t, err)
	require.Len(t, pages, 1)
	require.Len(t, pages[0].Elements, 3)
	_, ok := pages[0].Elements[0].(*entities.SemanticHeading)
	assert.True(t, ok)
	_, ok = pages[0].Elements[1].(*entities.SemanticTable)
	assert.True(t, ok)
	image, ok := pages[0].Elements[2].(*entities.SemanticImage)
	assert.True(t, ok)
	assert.Equal(t, "diagram", image.Alt)
}

func TestDoclingFastServerClientHealthCheckAndConvert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		case "/v1/convert/file":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
			  "document": {
			    "json_content": {
			      "pages": {"1": {"width": 600, "height": 800}},
			      "texts": [{
			        "label": "section_header",
			        "text": "Backend heading",
			        "meta": {"level": 1},
			        "prov": [{"page_no": 1, "bbox": {"l": 10, "t": 20, "r": 110, "b": 60, "coord_origin": "TOPLEFT"}}]
			      }],
			      "tables": [],
			      "pictures": []
			    }
			  },
			  "status": "success"
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	pdfPath := filepath.Join(t.TempDir(), "dummy.pdf")
	require.NoError(t, os.WriteFile(pdfPath, []byte("%PDF-1.4\n%%EOF"), 0o644))

	client := hybrid.NewDoclingFastServerClient(&hybrid.HybridConfig{
		Backend: hybrid.BackendDoclingFast,
		URL:     server.URL,
	})

	require.NoError(t, client.HealthCheck())
	resp, err := client.Convert(&hybrid.ConvertRequest{
		PDFPath:  pdfPath,
		PageNums: []int{2},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Pages, 1)
	require.NotNil(t, resp.Pages[0])
	assert.Equal(t, 1, resp.Pages[0].Number)
	heading, ok := resp.Pages[0].Elements[0].(*entities.SemanticHeading)
	assert.True(t, ok)
	assert.Equal(t, "Backend heading", heading.Lines[0].GetText())
}
