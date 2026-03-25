// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serializers

import (
	"encoding/base64"
	"fmt"
	"math"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const (
	jsonPageNumber        = "page number"
	jsonLevel             = "level"
	jsonBoundingBox       = "bounding box"
	jsonType              = "type"
	jsonID                = "id"
	jsonContent           = "content"
	jsonHiddenText        = "hidden text"
	jsonFontType          = "font"
	jsonFontSize          = "font size"
	jsonTextColor         = "text color"
	jsonKids              = "kids"
	jsonListItems         = "list items"
	jsonNumberOfListItems = "number of list items"
	jsonPreviousListID    = "previous list id"
	jsonNextListID        = "next list id"
	jsonPreviousTableID   = "previous table id"
	jsonNextTableID       = "next table id"
	jsonColumnNumber      = "column number"
	jsonRowNumber         = "row number"
	jsonColumnSpan        = "column span"
	jsonRowSpan           = "row span"
	jsonNumberOfRows      = "number of rows"
	jsonNumberOfColumns   = "number of columns"
	jsonHeadingLevel      = "heading level"
	jsonRows              = "rows"
	jsonCells             = "cells"
	jsonNumberingStyle    = "numbering style"
	jsonSource            = "source"
	jsonData              = "data"
	jsonImageFormat       = "format"
	jsonDescription       = "description"
)

func essentialInfo(object entities.IObject, objectType string) map[string]interface{} {
	bbox := object.GetBBox()
	out := map[string]interface{}{
		jsonType:       objectType,
		jsonPageNumber: bbox.Page + 1,
		jsonBoundingBox: map[string]interface{}{
			"left":   serializeDouble(bbox.X),
			"bottom": serializeDouble(bbox.Y),
			"right":  serializeDouble(bbox.X + bbox.Width),
			"top":    serializeDouble(bbox.Y + bbox.Height),
		},
	}
	if id := object.GetID(); id != "" {
		out[jsonID] = id
	}
	return out
}

func serializeTextChunks(chunks []*entities.TextChunk) string {
	var parts []string
	for _, chunk := range chunks {
		if chunk == nil {
			continue
		}
		parts = append(parts, chunk.Text)
	}
	return strings.Join(parts, "")
}

func serializeLines(lines []*entities.TextLine) string {
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == nil {
			continue
		}
		parts = append(parts, line.GetText())
	}
	return strings.Join(parts, "\n")
}

func firstChunkFromLines(lines []*entities.TextLine) *entities.TextChunk {
	for _, line := range lines {
		for _, chunk := range line.Chunks {
			if chunk != nil {
				return chunk
			}
		}
	}
	return nil
}

func firstChunkFromObjects(objects []entities.IObject) *entities.TextChunk {
	for _, object := range objects {
		switch typed := object.(type) {
		case *entities.TextChunk:
			return typed
		case *entities.TextLine:
			for _, chunk := range typed.Chunks {
				if chunk != nil {
					return chunk
				}
			}
		case *entities.SemanticParagraph:
			return firstChunkFromLines(typed.Lines)
		case *entities.SemanticHeading:
			return firstChunkFromLines(typed.Lines)
		}
	}
	return nil
}

func textInfoFromLines(lines []*entities.TextLine) map[string]interface{} {
	return textInfoFromChunkAndContent(firstChunkFromLines(lines), serializeLines(lines), false)
}

func textInfoFromObjects(objects []entities.IObject, content string) map[string]interface{} {
	hidden := false
	for _, object := range objects {
		switch typed := object.(type) {
		case *entities.TextChunk:
			if typed.IsHidden || typed.IsHiddenOCG || typed.IsOffPage || typed.IsTiny {
				hidden = true
			}
		case *entities.TextLine:
			for _, chunk := range typed.Chunks {
				if chunk != nil && (chunk.IsHidden || chunk.IsHiddenOCG || chunk.IsOffPage || chunk.IsTiny) {
					hidden = true
					break
				}
			}
		}
		if hidden {
			break
		}
	}
	return textInfoFromChunkAndContent(firstChunkFromObjects(objects), content, hidden)
}

func textInfoFromChunkAndContent(chunk *entities.TextChunk, content string, hidden bool) map[string]interface{} {
	out := map[string]interface{}{
		jsonContent: content,
	}
	if chunk != nil {
		if chunk.FontStyle.FontName != "" {
			out[jsonFontType] = chunk.FontStyle.FontName
		}
		if size := serializeDouble(chunk.FontStyle.FontSize); size != nil {
			out[jsonFontSize] = size
		}
		out[jsonTextColor] = fmt.Sprintf("%v", chunk.FontStyle.Color)
		if hidden || chunk.IsHidden || chunk.IsHiddenOCG || chunk.IsOffPage || chunk.IsTiny {
			out[jsonHiddenText] = true
		}
	}
	return out
}

func serializeElements(elements []entities.IObject) []interface{} {
	out := make([]interface{}, 0, len(elements))
	for _, element := range elements {
		if element == nil {
			continue
		}
		if _, isLineArt := element.(*entities.LineArtChunk); isLineArt {
			continue
		}
		switch typed := element.(type) {
		case *entities.SemanticTable:
			out = append(out, SerializeTable(typed))
		case *entities.SemanticHeading:
			out = append(out, SerializeHeading(typed))
		case *entities.SemanticParagraph:
			out = append(out, SerializeParagraph(typed))
		case *entities.PDFList:
			out = append(out, SerializeList(typed))
		case *entities.ListItem:
			out = append(out, SerializeListItem(typed))
		case *entities.SemanticImage:
			out = append(out, SerializeImage(typed, api.ImageOutputExternal))
		case *entities.SemanticFormula:
			out = append(out, SerializeFormula(typed))
		case *entities.SemanticCaption:
			out = append(out, SerializeCaption(typed))
		case *entities.TextChunk:
			out = append(out, SerializeTextChunk(typed))
		case *entities.SemanticHeaderFooter:
			out = append(out, SerializeHeaderFooter(typed))
		case *entities.TextLine:
			out = append(out, SerializeTextLine(typed))
		}
	}
	return out
}

func serializeDouble(v float64) interface{} {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return math.Round(v*1_000_000) / 1_000_000
}

func numberingStyle(list *entities.PDFList) string {
	if list != nil && list.IsOrdered {
		return "ordered"
	}
	return "unordered"
}

func dataURL(data []byte, format string) string {
	if len(data) == 0 {
		return ""
	}
	if format == "" {
		format = api.ImageFormatPNG
	}
	return "data:image/" + format + ";base64," + base64.StdEncoding.EncodeToString(data)
}
