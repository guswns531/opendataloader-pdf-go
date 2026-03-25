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
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

func SerializeListItem(item *entities.ListItem) map[string]interface{} {
	if item == nil {
		return nil
	}

	out := essentialInfo(item, "list item")
	content := strings.TrimSpace(item.BulletText + " " + contentText(item.Content))
	for key, value := range textInfoFromObjects(item.Content, content) {
		out[key] = value
	}
	out[jsonKids] = serializeElements(item.Content)
	if item.Level > 0 {
		out[jsonLevel] = item.Level
	}
	return out
}

func contentText(content []entities.IObject) string {
	parts := make([]string, 0, len(content))
	for _, object := range content {
		switch typed := object.(type) {
		case *entities.TextChunk:
			parts = append(parts, typed.Text)
		case *entities.TextLine:
			parts = append(parts, typed.GetText())
		case *entities.SemanticParagraph:
			parts = append(parts, serializeLines(typed.Lines))
		}
	}
	return strings.Join(parts, "\n")
}
