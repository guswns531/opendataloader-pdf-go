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

package json_gen

import (
	gojson "github.com/goccy/go-json"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/json/serializers"
)

type JsonWriter struct {
	ImageOutput string
}

func (w *JsonWriter) Write(doc *entities.Document) ([]byte, error) {
	if doc == nil {
		return gojson.Marshal(map[string]interface{}{
			Kids: []interface{}{},
		})
	}

	payload := serializeMetadata(doc.Metadata)
	payload["pages"] = w.serializePages(doc.Pages)

	kids := make([]interface{}, 0)
	for _, page := range doc.Pages {
		if page == nil {
			continue
		}
		kids = append(kids, w.serializeElements(page.Elements)...)
	}
	payload[Kids] = kids

	return gojson.Marshal(payload)
}

func (w *JsonWriter) serializeElements(elements []entities.IObject) []interface{} {
	out := make([]interface{}, 0, len(elements))
	for _, element := range elements {
		if element == nil {
			continue
		}
		switch typed := element.(type) {
		case *entities.SemanticTable:
			out = append(out, serializers.SerializeTable(typed))
		case *entities.SemanticHeading:
			out = append(out, serializers.SerializeHeading(typed))
		case *entities.SemanticParagraph:
			out = append(out, serializers.SerializeParagraph(typed))
		case *entities.PDFList:
			out = append(out, serializers.SerializeList(typed))
		case *entities.SemanticImage:
			out = append(out, serializers.SerializeImage(typed, w.ImageOutput))
		case *entities.SemanticFormula:
			out = append(out, serializers.SerializeFormula(typed))
		case *entities.SemanticCaption:
			out = append(out, serializers.SerializeCaption(typed))
		case *entities.TextChunk:
			out = append(out, serializers.SerializeTextChunk(typed))
		case *entities.SemanticHeaderFooter:
			out = append(out, serializers.SerializeHeaderFooter(typed))
		}
	}
	return out
}

func (w *JsonWriter) serializePages(pages []*entities.Page) []interface{} {
	out := make([]interface{}, 0, len(pages))
	for _, page := range pages {
		if page == nil {
			continue
		}
		out = append(out, map[string]interface{}{
			PageNumber: page.Number + 1,
			"width":    SerializeDouble(page.Width),
			"height":   SerializeDouble(page.Height),
			Kids:       w.serializeElements(page.Elements),
		})
	}
	return out
}

func serializeMetadata(metadata entities.DocumentMetadata) map[string]interface{} {
	out := map[string]interface{}{
		NumberOfPages: metadata.PageCount,
	}
	if metadata.Author != "" {
		out[Author] = metadata.Author
	}
	if metadata.Title != "" {
		out[Title] = metadata.Title
	}
	if metadata.Creator != "" {
		out["creator"] = metadata.Creator
	}
	if metadata.Producer != "" {
		out["producer"] = metadata.Producer
	}
	return out
}
