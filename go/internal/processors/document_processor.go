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
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/generators"
	json_gen "github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/json"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/processors/readingorder"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/utils"
	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/extractor"
	pdfbox_loader "github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/loader"
)

type DocumentProcessor struct{}

func NewDocumentProcessor() *DocumentProcessor {
	return &DocumentProcessor{}
}

func init() {
	api.RegisterDocumentProcessor(func() api.DocumentFileProcessor {
		return NewDocumentProcessor()
	})
}

func (p *DocumentProcessor) Process(pdfPath string, config *api.Config) (*entities.Document, error) {
	if config == nil {
		config = api.DefaultConfig()
	}

	ctx := containers.NewProcessorContext()
	ctx.UseStructTree = config.UseStructTree
	ctx.ImageDir = resolveImageDir(pdfPath, config)
	ctx.EmbedImages = config.ImageOutput == api.ImageOutputEmbedded
	ctx.ImageFormat = config.ImageFormat

	doc, err := p.loadDocument(pdfPath, config, ctx)
	if err != nil {
		return nil, err
	}

	var processed *entities.Document
	if config.Hybrid != "" && config.Hybrid != api.HybridOff {
		processed, err = NewHybridDocumentProcessor(p).Process(doc, pdfPath, config, ctx)
	} else {
		processed, err = p.processJavaDocument(doc, config, ctx)
	}
	if err != nil {
		return nil, err
	}

	if err := p.writeOutputs(pdfPath, processed, config); err != nil {
		return nil, err
	}

	return processed, nil
}

func (p *DocumentProcessor) processJavaDocument(doc *entities.Document, config *api.Config, ctx *containers.ProcessorContext) (*entities.Document, error) {
	var headings []*entities.SemanticHeading
	apiFilterConfig := api.FilterConfigFromStrings(config.ContentSafetyOff)

	for _, page := range doc.Pages {
		if page == nil {
			continue
		}

		page.Chunks = FilterContent(page.Chunks, apiFilterConfig, page.Width, page.Height)

		rawElements := make([]entities.IObject, 0, len(page.Elements)+len(page.Chunks))
		rawElements = append(rawElements, page.Elements...)
		for _, chunk := range page.Chunks {
			rawElements = append(rawElements, chunk)
		}

		pageElements := rawElements
		if config.TableMethod == api.TableMethodCluster {
			pageElements = (&ClusterTableProcessor{}).Process(pageElements, ctx)
		}
		pageElements = (&TableBorderProcessor{}).Process(pageElements, page.LineArts, ctx)

		tablesAndOther, textChunks := splitTextChunks(pageElements)
		lines := (&TextLineProcessor{}).Process(textChunks, page.LineArts, ctx)
		if config.DetectStrikethrough {
			lines = (&StrikethroughProcessor{}).Process(lines, page.LineArts)
		}

		pageElements = append(tablesAndOther, textLinesToObjects(lines)...)
		pageElements = sortObjects(pageElements)

		pageElements = (&SpecialTableProcessor{}).Process(pageElements, ctx)

		pageElements = (&HeaderFooterProcessor{}).Process(pageElements, page.Height, config.IncludeHeaderFooter, ctx)
		pageElements = (&ListProcessor{}).Process(pageElements, ctx)
		var pageHeadings []*entities.SemanticHeading
		pageElements, pageHeadings = transformTextRuns(pageElements, ctx)
		headings = append(headings, pageHeadings...)
		pageElements, pageHeadings = (&HeadingProcessor{}).PromoteListHeadings(pageElements, ctx)
		headings = append(headings, pageHeadings...)
		pageElements = (&CaptionProcessor{}).Process(pageElements, ctx)
		page.Elements = sortObjects(pageElements)
	}

	MergeListsAcrossPages(doc.Pages)
	(&LevelProcessor{}).Process(headings)
	sortDocumentContents(doc, config)

	if config.Sanitize {
		utils.Sanitize(doc, utils.DefaultRules)
	}
	return doc, nil
}

func (p *DocumentProcessor) loadDocument(pdfPath string, config *api.Config, ctx *containers.ProcessorContext) (*entities.Document, error) {
	doc, err := pdfbox_loader.Open(pdfPath, config.Password)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	pageCount := doc.PageCount()
	selectedPages, explicitSelection, err := parsePageRange(config.Pages, pageCount)
	if err != nil {
		return nil, err
	}
	if !explicitSelection {
		selectedPages = make([]int, 0, pageCount)
		for pageNum := 1; pageNum <= pageCount; pageNum++ {
			selectedPages = append(selectedPages, pageNum)
		}
	}

	out := &entities.Document{
		Metadata: entities.DocumentMetadata{
			Title:     strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath)),
			PageCount: pageCount,
		},
		Pages: make([]*entities.Page, 0, len(selectedPages)),
	}

	for _, pageNum := range selectedPages {
		pageIdx := pageNum - 1

		page, err := doc.GetPage(pageIdx)
		if err != nil {
			return nil, err
		}

		textChunks, err := extractor.ExtractTextChunks(doc, pageIdx)
		if err != nil {
			return nil, err
		}
		var images []*extractor.ExtractedImage
		if config.ImageOutput != api.ImageOutputOff {
			images, err = extractor.ExtractImages(doc, pageIdx, ctx.ImageDir)
			if err != nil {
				return nil, err
			}
		}
		lineArts, err := extractor.ExtractLineArts(doc, pageIdx)
		if err != nil {
			return nil, err
		}

		semanticImages := make([]entities.IObject, 0, len(images))
		for _, image := range images {
			semanticImages = append(semanticImages, toSemanticImage(image, ctx))
		}

		out.Pages = append(out.Pages, &entities.Page{
			PageMetadata: entities.PageMetadata{
				Number: page.Number,
				Width:  page.Width,
				Height: page.Height,
			},
			Chunks:   toTextChunks(textChunks, ctx),
			LineArts: toLineArtChunks(lineArts, ctx),
			Elements: semanticImages,
		})
	}

	return out, nil
}

func toTextChunks(extracted []*extractor.ExtractedText, ctx *containers.ProcessorContext) []*entities.TextChunk {
	out := make([]*entities.TextChunk, 0, len(extracted))
	for _, chunk := range extracted {
		if chunk == nil {
			continue
		}
		out = append(out, &entities.TextChunk{
			BaseObject: entities.BaseObject{
				ID: nextObjectID(ctx),
				BBox: entities.BoundingBox{
					X:      chunk.X,
					Y:      chunk.Y,
					Width:  chunk.Width,
					Height: chunk.Height,
					Page:   chunk.Page,
				},
			},
			Text: chunk.Text,
			FontStyle: entities.FontStyle{
				FontName:   chunk.FontName,
				FontSize:   chunk.FontSize,
				FontWeight: fontWeight(chunk.Bold),
				Bold:       chunk.Bold,
				Italic:     chunk.Italic,
				Color:      chunk.Color,
			},
			Baseline: chunk.Baseline,
		})
	}
	return out
}

func toLineArtChunks(extracted []*extractor.ExtractedLineArt, ctx *containers.ProcessorContext) []*entities.LineArtChunk {
	out := make([]*entities.LineArtChunk, 0, len(extracted))
	for _, line := range extracted {
		if line == nil {
			continue
		}
		out = append(out, &entities.LineArtChunk{
			BaseObject: entities.BaseObject{
				ID: nextObjectID(ctx),
				BBox: entities.BoundingBox{
					X:      line.X,
					Y:      line.Y,
					Width:  line.Width,
					Height: line.Height,
					Page:   line.Page,
				},
			},
			IsHorizontal: line.IsHorizontal,
			IsVertical:   line.IsVertical,
			LineWidth:    line.LineWidth,
		})
	}
	return out
}

func toSemanticImage(image *extractor.ExtractedImage, ctx *containers.ProcessorContext) *entities.SemanticImage {
	if image == nil {
		return nil
	}
	return &entities.SemanticImage{
		BaseObject: entities.BaseObject{
			ID: nextObjectID(ctx),
			BBox: entities.BoundingBox{
				X:      image.X,
				Y:      image.Y,
				Width:  image.Width,
				Height: image.Height,
				Page:   image.Page,
			},
		},
		Data:         image.Data,
		ExternalPath: image.ExternalPath,
		Width:        image.Width,
		Height:       image.Height,
	}
}

func fontWeight(bold bool) float64 {
	if bold {
		return 700
	}
	return 400
}

func parsePageRange(spec string, totalPages int) ([]int, bool, error) {
	pages, err := api.ParsePageRanges(spec)
	if err != nil {
		return nil, false, err
	}
	if len(pages) == 0 {
		return nil, false, nil
	}
	validPages := make([]int, 0, len(pages))
	for _, page := range pages {
		if page >= 1 && page <= totalPages {
			validPages = append(validPages, page)
		}
	}
	return validPages, true, nil
}

func splitTextChunks(elements []entities.IObject) ([]entities.IObject, []*entities.TextChunk) {
	others := make([]entities.IObject, 0, len(elements))
	chunks := make([]*entities.TextChunk, 0, len(elements))
	for _, element := range elements {
		chunk, ok := element.(*entities.TextChunk)
		if ok {
			chunks = append(chunks, chunk)
			continue
		}
		others = append(others, element)
	}
	return others, chunks
}

func textLinesToObjects(lines []*entities.TextLine) []entities.IObject {
	out := make([]entities.IObject, 0, len(lines))
	for _, line := range lines {
		out = append(out, line)
	}
	return out
}

func transformTextRuns(elements []entities.IObject, ctx *containers.ProcessorContext) ([]entities.IObject, []*entities.SemanticHeading) {
	out := make([]entities.IObject, 0, len(elements))
	headings := make([]*entities.SemanticHeading, 0)

	flush := func(lines []*entities.TextLine) {
		if len(lines) == 0 {
			return
		}
		runObjects, runHeadings := convertTextLines(lines, ctx)
		out = append(out, runObjects...)
		headings = append(headings, runHeadings...)
	}

	var run []*entities.TextLine
	for _, element := range elements {
		line, ok := element.(*entities.TextLine)
		if !ok {
			flush(run)
			run = nil
			out = append(out, element)
			continue
		}
		run = append(run, line)
	}
	flush(run)

	return sortObjects(out), headings
}

func convertTextLines(lines []*entities.TextLine, ctx *containers.ProcessorContext) ([]entities.IObject, []*entities.SemanticHeading) {
	if len(lines) == 0 {
		return nil, nil
	}

	headingCandidates := (&HeadingProcessor{}).Process(lines, ctx)
	headingByLine := make(map[*entities.TextLine]*entities.SemanticHeading, len(headingCandidates))
	for _, heading := range headingCandidates {
		if heading == nil || len(heading.Lines) == 0 {
			continue
		}
		headingByLine[heading.Lines[0]] = heading
	}

	out := make([]entities.IObject, 0, len(lines))
	headings := make([]*entities.SemanticHeading, 0, len(headingCandidates))
	var paragraphRun []*entities.TextLine

	flushParagraphs := func() {
		if len(paragraphRun) == 0 {
			return
		}
		paragraphs := (&ParagraphProcessor{}).Process(paragraphRun, ctx)
		for _, paragraph := range paragraphs {
			out = append(out, paragraph)
		}
		paragraphRun = nil
	}

	for _, line := range lines {
		if heading, ok := headingByLine[line]; ok {
			flushParagraphs()
			out = append(out, heading)
			headings = append(headings, heading)
			continue
		}
		paragraphRun = append(paragraphRun, line)
	}
	flushParagraphs()

	return out, headings
}

func sortObjects(elements []entities.IObject) []entities.IObject {
	out := append([]entities.IObject(nil), elements...)
	sort.SliceStable(out, func(i, j int) bool {
		left := out[i].GetBBox()
		right := out[j].GetBBox()
		if left.Page != right.Page {
			return left.Page < right.Page
		}
		if left.Y != right.Y {
			return left.Y > right.Y
		}
		return left.X < right.X
	})
	return out
}

func sortDocumentContents(doc *entities.Document, config *api.Config) {
	if doc == nil || config == nil || config.ReadingOrder != api.ReadingOrderXYCut {
		return
	}
	sorter := readingorder.XYCutPlusPlusSorter{}
	for _, page := range doc.Pages {
		if page == nil {
			continue
		}
		page.Elements = sorter.Sort(page.Elements, page.Width, page.Height)
	}
}

func resolveImageDir(pdfPath string, config *api.Config) string {
	if config == nil || config.ImageOutput == api.ImageOutputOff {
		return ""
	}
	if trimmed := strings.TrimSpace(config.ImageDir); trimmed != "" {
		return trimmed
	}
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	return filepath.Join(outputDirForConfig(pdfPath, config), baseName+"_images")
}

func outputDirForConfig(pdfPath string, config *api.Config) string {
	if config != nil {
		if trimmed := strings.TrimSpace(config.OutputDir); trimmed != "" {
			return trimmed
		}
	}
	return filepath.Dir(pdfPath)
}

func (p *DocumentProcessor) writeOutputs(pdfPath string, doc *entities.Document, config *api.Config) error {
	outputDir := outputDirForConfig(pdfPath, config)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	baseName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	for _, format := range config.Formats {
		if err := writeOutputForFormat(pdfPath, doc, config, outputDir, baseName, format); err != nil {
			return err
		}
	}
	return nil
}

func writeOutputForFormat(pdfPath string, doc *entities.Document, config *api.Config, outputDir, baseName, format string) error {
	switch format {
	case api.FormatJSON:
		payload, err := (&json_gen.JsonWriter{ImageOutput: config.ImageOutput}).Write(doc)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(outputDir, baseName+".json"), payload, 0o644)
	case api.FormatPDF:
		data, err := os.ReadFile(pdfPath)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(outputDir, baseName+".pdf"), data, 0o644)
	default:
		generator := generators.GetGenerator(format, config)
		if generator == nil {
			return fmt.Errorf("unsupported output format %q", format)
		}
		content, err := generator.Generate(doc)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(outputDir, baseName+"."+outputExtension(format)), []byte(content), 0o644)
	}
}

func outputExtension(format string) string {
	switch format {
	case api.FormatText:
		return "txt"
	case api.FormatHTML:
		return "html"
	case api.FormatMarkdown, api.FormatMarkdownWithHTML, api.FormatMarkdownWithImages:
		return "md"
	case api.FormatJSON:
		return "json"
	case api.FormatPDF:
		return "pdf"
	default:
		return format
	}
}
