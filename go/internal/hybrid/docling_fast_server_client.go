// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
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
	if config == nil {
		config = DefaultHybridConfig()
	}
	return &DoclingFastServerClient{
		config:  config,
		baseURL: strings.TrimRight(config.effectiveURL(), "/"),
		httpClient: &http.Client{
			Timeout: time.Duration(config.timeoutMS()) * time.Millisecond,
		},
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

	pages, err := TransformDoclingResponse(respBody)
	if err != nil {
		return &ConvertResponse{Error: err}, err
	}
	if len(request.PageNums) == len(pages) {
		for idx, pageNum := range request.PageNums {
			if pages[idx] != nil {
				pages[idx].Number = pageNum - 1
			}
		}
	}

	return &ConvertResponse{Pages: pages}, nil
}

func (c *DoclingFastServerClient) HealthCheck() error {
	client := &http.Client{Timeout: time.Duration(healthCheckTimeoutMS) * time.Millisecond}
	resp, err := client.Get(c.baseURL + healthEndpoint)
	if err != nil {
		return fmt.Errorf("hybrid server is not available at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("hybrid server at %s returned HTTP %d during health check", c.baseURL, resp.StatusCode)
	}
	return nil
}

func (c *DoclingFastServerClient) Close() error {
	return nil
}
