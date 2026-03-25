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
	"strconv"
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

	doc, err := p.loadDocument(pdfPath, config)
	if err != nil {
		return nil, err
	}

	ctx := containers.NewProcessorContext()
	ctx.UseStructTree = config.UseStructTree
	ctx.ImageDir = config.ImageDir
	ctx.EmbedImages = config.ImageOutput == api.ImageOutputEmbedded
	ctx.ImageFormat = config.ImageFormat

	var headings []*entities.SemanticHeading
	apiFilterConfig := api.FilterConfigFromStrings(config.ContentSafetyOff)
	filterConfig := &FilterConfig{
		DisableHiddenText: apiFilterConfig.DisableHiddenText,
		DisableOffPage:    apiFilterConfig.DisableOffPage,
		DisableTiny:       apiFilterConfig.DisableTiny,
		DisableHiddenOCG:  apiFilterConfig.DisableHiddenOCG,
	}

	for _, page := range doc.Pages {
		if page == nil {
			continue
		}

		page.Chunks = FilterContent(page.Chunks, filterConfig, page.Width, page.Height)

		rawElements := make([]entities.IObject, 0, len(page.Elements)+len(page.Chunks))
		rawElements = append(rawElements, page.Elements...)
		for _, chunk := range page.Chunks {
			rawElements = append(rawElements, chunk)
		}

		var pageElements []entities.IObject
		if config.TableMethod == api.TableMethodCluster {
			pageElements = (&ClusterTableProcessor{}).Process(rawElements, ctx)
		} else {
			pageElements = (&TableBorderProcessor{}).Process(rawElements, page.LineArts, ctx)
		}

		tablesAndOther, textChunks := splitTextChunks(pageElements)
		lines := (&TextLineProcessor{}).Process(textChunks, page.LineArts, ctx)
		if config.DetectStrikethrough {
			lines = (&StrikethroughProcessor{}).Process(lines, page.LineArts)
		}

		pageElements = append(tablesAndOther, textLinesToObjects(lines)...)
		pageElements = sortObjects(pageElements)

		if config.TableMethod != api.TableMethodCluster {
			pageElements = (&SpecialTableProcessor{}).Process(pageElements, ctx)
		}

		pageElements = (&HeaderFooterProcessor{}).Process(pageElements, page.Height, config.IncludeHeaderFooter, ctx)
		pageElements = (&ListProcessor{}).Process(pageElements, ctx)
		var pageHeadings []*entities.SemanticHeading
		pageElements, pageHeadings = transformTextRuns(pageElements, ctx)
		headings = append(headings, pageHeadings...)
		pageElements = (&CaptionProcessor{}).Process(pageElements, ctx)

		if config.ReadingOrder == api.ReadingOrderXYCut {
			pageElements = readingorder.XYCutPlusPlusSorter{}.Sort(pageElements, page.Width, page.Height)
		}

		page.Elements = sortObjects(pageElements)
	}

	(&LevelProcessor{}).Process(headings)

	if config.Sanitize {
		utils.Sanitize(doc, utils.DefaultRules)
	}

	if err := p.writeOutputs(pdfPath, doc, config); err != nil {
		return nil, err
	}

	return doc, nil
}

func (p *DocumentProcessor) loadDocument(pdfPath string, config *api.Config) (*entities.Document, error) {
	doc, err := pdfbox_loader.Open(pdfPath, config.Password)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	pageCount := doc.PageCount()
	selectedPages, err := parsePageRange(config.Pages, pageCount)
	if err != nil {
		return nil, err
	}
	if len(selectedPages) == 0 {
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
		images, err := extractor.ExtractImages(doc, pageIdx, config.ImageDir)
		if err != nil {
			return nil, err
		}
		lineArts, err := extractor.ExtractLineArts(doc, pageIdx)
		if err != nil {
			return nil, err
		}

		semanticImages := make([]entities.IObject, 0, len(images))
		for _, image := range images {
			semanticImages = append(semanticImages, toSemanticImage(image))
		}

		out.Pages = append(out.Pages, &entities.Page{
			PageMetadata: entities.PageMetadata{
				Number: pageIdx,
				Width:  page.Width,
				Height: page.Height,
			},
			Chunks:   toTextChunks(textChunks),
			LineArts: toLineArtChunks(lineArts),
			Elements: semanticImages,
		})
	}

	return out, nil
}

func toTextChunks(extracted []*extractor.ExtractedText) []*entities.TextChunk {
	out := make([]*entities.TextChunk, 0, len(extracted))
	for _, chunk := range extracted {
		if chunk == nil {
			continue
		}
		out = append(out, &entities.TextChunk{
			BaseObject: entities.BaseObject{
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

func toLineArtChunks(extracted []*extractor.ExtractedLineArt) []*entities.LineArtChunk {
	out := make([]*entities.LineArtChunk, 0, len(extracted))
	for _, line := range extracted {
		if line == nil {
			continue
		}
		out = append(out, &entities.LineArtChunk{
			BaseObject: entities.BaseObject{
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

func toSemanticImage(image *extractor.ExtractedImage) *entities.SemanticImage {
	if image == nil {
		return nil
	}
	return &entities.SemanticImage{
		BaseObject: entities.BaseObject{
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

func parsePageRange(spec string, totalPages int) ([]int, error) {
	if strings.TrimSpace(spec) == "" {
		return nil, nil
	}

	seen := map[int]struct{}{}
	var pages []int
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			if len(bounds) != 2 {
				return nil, fmt.Errorf("invalid page range %q", part)
			}
			start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid page %q", part)
			}
			end, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid page %q", part)
			}
			if start > end {
				start, end = end, start
			}
			for page := start; page <= end; page++ {
				if page < 1 || page > totalPages {
					continue
				}
				if _, ok := seen[page]; ok {
					continue
				}
				seen[page] = struct{}{}
				pages = append(pages, page)
			}
			continue
		}

		page, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid page %q", part)
		}
		if page < 1 || page > totalPages {
			continue
		}
		if _, ok := seen[page]; ok {
			continue
		}
		seen[page] = struct{}{}
		pages = append(pages, page)
	}

	sort.Ints(pages)
	return pages, nil
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

func (p *DocumentProcessor) writeOutputs(pdfPath string, doc *entities.Document, config *api.Config) error {
	outputDir := config.OutputDir
	if outputDir == "" {
		outputDir = filepath.Dir(pdfPath)
	}
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
