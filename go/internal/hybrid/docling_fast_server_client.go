// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const (
	convertEndpoint      = "/v1/convert/file"
	healthEndpoint       = "/health"
	healthCheckTimeoutMS = 3000
	defaultPDFName       = "document.pdf"
)

type DoclingFastServerClient struct {
	config     *HybridConfig
	httpClient *http.Client
	baseURL    string
}

func NewDoclingFastServerClient(config *HybridConfig) *DoclingFastServerClient {
	return NewDoclingFastServerClientWithHTTPClient(config, nil)
}

func NewDoclingFastServerClientWithHTTPClient(config *HybridConfig, httpClient *http.Client) *DoclingFastServerClient {
	if config == nil {
		config = DefaultHybridConfig()
	}
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: time.Duration(config.Timeout()) * time.Millisecond,
		}
	}
	return &DoclingFastServerClient{
		config:  config,
		baseURL: strings.TrimRight(config.EffectiveURL(), "/"),
		httpClient: httpClient,
	}
}

func (c *DoclingFastServerClient) Convert(request *ConvertRequest) (*ConvertResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("convert request is nil")
	}
	if request.PDFPath == "" {
		return nil, fmt.Errorf("convert request missing PDFPath")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileWriter, err := writer.CreateFormFile("files", defaultPDFName)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(request.PDFPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := io.Copy(fileWriter, file); err != nil {
		return nil, err
	}

	if len(request.PageNums) > 0 {
		pageNums := append([]int(nil), request.PageNums...)
		sort.Ints(pageNums)
		pageRange := fmt.Sprintf("%d-%d", pageNums[0], pageNums[len(pageNums)-1])
		if err := writer.WriteField("page_ranges", pageRange); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+convertEndpoint, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("docling fast server request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	parsed, err := parseDoclingResponse(respBody)
	if err != nil {
		return &ConvertResponse{Error: err}, err
	}

	pages := parsed.Pages
	if len(request.PageNums) > 0 {
		requested := make(map[int]struct{}, len(request.PageNums))
		for _, pageNum := range request.PageNums {
			requested[pageNum] = struct{}{}
		}
		filtered := make([]*entities.Page, 0, len(pages))
		for _, page := range pages {
			if page == nil {
				continue
			}
			if _, ok := requested[page.Number+1]; ok {
				filtered = append(filtered, page)
			}
		}
		pages = filtered
	}

	return &ConvertResponse{
		Pages:          pages,
		FailedPageNums: parsed.FailedPages,
	}, nil
}

func (c *DoclingFastServerClient) HealthCheck() error {
	client := &http.Client{
		Timeout:   time.Duration(healthCheckTimeoutMS) * time.Millisecond,
		Transport: c.httpClient.Transport,
	}
	resp, err := client.Get(c.baseURL + healthEndpoint)
	if err != nil {
		return fmt.Errorf(
			"hybrid server is not available at %s\nPlease start the server with: opendataloader-pdf-hybrid\nOr run without --hybrid flag for Java-only processing: %w",
			c.baseURL, err,
		)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf(
			"hybrid server at %s returned HTTP %d during health check.\nThe server is reachable but may be starting up or unhealthy.",
			c.baseURL, resp.StatusCode,
		)
	}
	return nil
}

func (c *DoclingFastServerClient) Close() error {
	return nil
}

type parsedDoclingResponse struct {
	Pages       []*entities.Page
	FailedPages []int
}

func parseDoclingResponse(data []byte) (*parsedDoclingResponse, error) {
	var root doclingRoot
	if err := json.Unmarshal(data, &root); err == nil && root.Document != nil && len(root.Document.JSONContent) > 0 {
		if root.Status == "failure" {
			return nil, fmt.Errorf("docling processing failed: %s", string(root.Errors))
		}
		pages, err := transformDoclingDocument(root.Document.JSONContent)
		if err != nil {
			return nil, err
		}
		return &parsedDoclingResponse{
			Pages:       pages,
			FailedPages: append([]int(nil), root.FailedPages...),
		}, nil
	}

	pages, err := transformDoclingDocument(data)
	if err != nil {
		return nil, err
	}
	return &parsedDoclingResponse{Pages: pages}, nil
}
