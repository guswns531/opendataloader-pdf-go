/*
 * Copyright 2025-2026 Hancom Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package processors

import (
	"fmt"
	"sort"
	"sync"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/hybrid"
)

type javaPipeline interface {
	processJavaDocument(doc *entities.Document, config *api.Config, ctx *containers.ProcessorContext) (*entities.Document, error)
}

type HybridDocumentProcessor struct {
	javaProcessor javaPipeline
}

func NewHybridDocumentProcessor(javaProcessor javaPipeline) *HybridDocumentProcessor {
	return &HybridDocumentProcessor{javaProcessor: javaProcessor}
}

func (p *HybridDocumentProcessor) Process(doc *entities.Document, pdfPath string, config *api.Config, ctx *containers.ProcessorContext) (*entities.Document, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}
	if config == nil || config.Hybrid == "" || config.Hybrid == api.HybridOff {
		return p.javaProcessor.processJavaDocument(doc, config, ctx)
	}

	applyHybridContentFilter(doc, config)

	hybridConfig := &hybrid.HybridConfig{
		Backend:        config.Hybrid,
		Mode:           config.HybridMode,
		URL:            config.HybridURL,
		TimeoutMS:      config.HybridTimeout,
		Fallback:       config.HybridFallback,
		MaxConcurrency: hybrid.DefaultHybridConfig().MaxConcurrency,
	}
	client, err := hybrid.NewHybridClient(hybridConfig)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return p.javaProcessor.processJavaDocument(doc, config, ctx)
	}
	defer client.Close()

	if err := client.HealthCheck(); err != nil {
		if !config.HybridFallback {
			return nil, fmt.Errorf("backend unavailable: %w", err)
		}
		return p.javaProcessor.processJavaDocument(doc, config, ctx)
	}

	javaPageMap := map[int]*entities.Page{}
	backendPageNums := make([]int, 0)
	for _, page := range doc.Pages {
		if page == nil {
			continue
		}
		decision := hybrid.TriageDecisionJava
		if config.HybridMode == api.HybridModeFull {
			decision = hybrid.TriageDecisionBackend
		} else {
			result := (&hybrid.TriageProcessor{}).Triage(page)
			hybrid.LogTriageResult(page.Number+1, result)
			decision = result.Decision
		}
		if decision == hybrid.TriageDecisionBackend {
			backendPageNums = append(backendPageNums, page.Number+1)
		} else {
			javaPageMap[page.Number] = page
		}
	}

	if len(backendPageNums) == 0 {
		return p.javaProcessor.processJavaDocument(doc, config, ctx)
	}

	resultPages := make(map[int]*entities.Page, len(doc.Pages))
	for pageNum, page := range javaPageMap {
		resultPages[pageNum] = page
	}

	var (
		wg          sync.WaitGroup
		javaErr     error
		backendErr  error
		javaDoc     *entities.Document
		backendResp *hybrid.ConvertResponse
	)

	if len(javaPageMap) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			javaSubset := cloneDocumentWithPages(doc, pagesFromMap(javaPageMap))
			javaDoc, javaErr = p.javaProcessor.processJavaDocument(javaSubset, config, containers.NewProcessorContext())
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		backendResp, backendErr = client.Convert(&hybrid.ConvertRequest{
			PDFPath:  pdfPath,
			PageNums: backendPageNums,
			Config:   hybridConfig,
		})
	}()

	wg.Wait()

	if javaErr != nil {
		return nil, javaErr
	}
	if backendErr != nil {
		if !config.HybridFallback {
			return nil, backendErr
		}
		return p.javaProcessor.processJavaDocument(doc, config, ctx)
	}
	if backendResp == nil || backendResp.Error != nil {
		if !config.HybridFallback {
			if backendResp != nil && backendResp.Error != nil {
				return nil, backendResp.Error
			}
			return nil, fmt.Errorf("hybrid backend returned no pages")
		}
		return p.javaProcessor.processJavaDocument(doc, config, ctx)
	}
	if len(backendResp.FailedPageNums) > 0 {
		failedPages := make(map[int]*entities.Page, len(backendResp.FailedPageNums))
		for _, pageNum := range backendResp.FailedPageNums {
			pageIndex := pageNum - 1
			if pageIndex >= 0 && pageIndex < len(doc.Pages) {
				failedPages[pageIndex] = doc.Pages[pageIndex]
			}
		}
		if len(failedPages) > 0 {
			if !config.HybridFallback {
				return nil, fmt.Errorf("hybrid backend returned partial_success for pages %v", backendResp.FailedPageNums)
			}
			fallbackDoc, err := p.javaProcessor.processJavaDocument(
				cloneDocumentWithPages(doc, pagesFromMap(failedPages)),
				config,
				containers.NewProcessorContext(),
			)
			if err != nil {
				return nil, err
			}
			if fallbackDoc != nil {
				for _, page := range fallbackDoc.Pages {
					if page != nil {
						resultPages[page.Number] = page
					}
				}
			}
		}
	}

	if javaDoc != nil {
		for _, page := range javaDoc.Pages {
			if page != nil {
				resultPages[page.Number] = page
			}
		}
	}
	for _, page := range backendResp.Pages {
		if page != nil {
			resultPages[page.Number] = page
		}
	}

	merged := &entities.Document{
		Metadata: doc.Metadata,
		Pages:    make([]*entities.Page, 0, len(doc.Pages)),
	}
	for _, original := range doc.Pages {
		if original == nil {
			continue
		}
		page := resultPages[original.Number]
		if page == nil {
			page = original
		}
		if page.Width == 0 {
			page.Width = original.Width
		}
		if page.Height == 0 {
			page.Height = original.Height
		}
		merged.Pages = append(merged.Pages, page)
	}
	MergeListsAcrossPages(merged.Pages)
	return merged, nil
}

func applyHybridContentFilter(doc *entities.Document, config *api.Config) {
	if doc == nil || config == nil {
		return
	}
	filterConfig := api.FilterConfigFromStrings(config.ContentSafetyOff)
	for _, page := range doc.Pages {
		if page == nil {
			continue
		}
		page.Chunks = FilterContent(page.Chunks, filterConfig, page.Width, page.Height)
	}
}

func cloneDocumentWithPages(doc *entities.Document, pages []*entities.Page) *entities.Document {
	if doc == nil {
		return nil
	}
	cloned := &entities.Document{
		Metadata: doc.Metadata,
		Pages:    make([]*entities.Page, 0, len(pages)),
	}
	cloned.Pages = append(cloned.Pages, pages...)
	return cloned
}

func pagesFromMap(pageMap map[int]*entities.Page) []*entities.Page {
	pageNums := make([]int, 0, len(pageMap))
	for pageNum := range pageMap {
		pageNums = append(pageNums, pageNum)
	}
	sort.Ints(pageNums)

	pages := make([]*entities.Page, 0, len(pageMap))
	for _, pageNum := range pageNums {
		pages = append(pages, pageMap[pageNum])
	}
	return pages
}
