// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid_test

import (
	"bytes"
	"io"
	"net/http"
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
	pdfPath := filepath.Join(t.TempDir(), "dummy.pdf")
	require.NoError(t, os.WriteFile(pdfPath, []byte("%PDF-1.4\n%%EOF"), 0o644))

	client := hybrid.NewDoclingFastServerClientWithHTTPClient(&hybrid.HybridConfig{
		Backend: hybrid.BackendDoclingFast,
		URL:     "http://docling.test",
	}, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/health":
			return jsonResponse(http.StatusOK, "ok"), nil
		case "/v1/convert/file":
			return jsonResponse(http.StatusOK, `{
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
			}`), nil
		default:
			return jsonResponse(http.StatusNotFound, ""), nil
		}
	})})

	require.NoError(t, client.HealthCheck())
	resp, err := client.Convert(&hybrid.ConvertRequest{
		PDFPath:  pdfPath,
		PageNums: []int{1},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Pages, 1)
	require.NotNil(t, resp.Pages[0])
	assert.Equal(t, 0, resp.Pages[0].Number)
	heading, ok := resp.Pages[0].Elements[0].(*entities.SemanticHeading)
	assert.True(t, ok)
	assert.Equal(t, "Backend heading", heading.Lines[0].GetText())
}

func TestDoclingFastServerClientReturnsFailedPagesOnPartialSuccess(t *testing.T) {
	pdfPath := filepath.Join(t.TempDir(), "dummy.pdf")
	require.NoError(t, os.WriteFile(pdfPath, []byte("%PDF-1.4\n%%EOF"), 0o644))

	client := hybrid.NewDoclingFastServerClientWithHTTPClient(&hybrid.HybridConfig{
		Backend: hybrid.BackendDoclingFast,
		URL:     "http://docling.test",
	}, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/health":
			return jsonResponse(http.StatusOK, "ok"), nil
		case "/v1/convert/file":
			return jsonResponse(http.StatusOK, `{
			  "document": {
			    "json_content": {
			      "pages": {
			        "2": {"width": 600, "height": 800},
			        "3": {"width": 600, "height": 800}
			      },
			      "texts": [{
			        "label": "text",
			        "text": "page two",
			        "prov": [{"page_no": 2, "bbox": {"l": 10, "t": 20, "r": 110, "b": 60, "coord_origin": "TOPLEFT"}}]
			      }],
			      "tables": [],
			      "pictures": []
			    }
			  },
			  "status": "partial_success",
			  "failed_pages": [3]
			}`), nil
		default:
			return jsonResponse(http.StatusNotFound, ""), nil
		}
	})})

	resp, err := client.Convert(&hybrid.ConvertRequest{
		PDFPath:  pdfPath,
		PageNums: []int{2, 3},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, []int{3}, resp.FailedPageNums)
	require.Len(t, resp.Pages, 2)
	assert.Equal(t, 1, resp.Pages[0].Number)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}
