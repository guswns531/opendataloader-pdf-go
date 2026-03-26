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
	"fmt"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

func SerializeTextChunk(c *entities.TextChunk) map[string]interface{} {
	if c == nil {
		return nil
	}

	out := essentialInfo(c, "text chunk")
	out[jsonContent] = c.Text
	if c.FontStyle.FontName != "" {
		out[jsonFontType] = c.FontStyle.FontName
	}
	if size := serializeDouble(c.FontStyle.FontSize); size != nil {
		out[jsonFontSize] = size
	}
	out[jsonTextColor] = fmt.Sprintf("%v", c.FontStyle.Color)
	if c.IsHidden || c.IsHiddenOCG || c.IsOffPage || c.IsTiny {
		out[jsonHiddenText] = true
	}
	return out
}

func SerializeTextLine(line *entities.TextLine) map[string]interface{} {
	if line == nil {
		return nil
	}

	out := essentialInfo(line, "text chunk")
	out[jsonContent] = line.GetText()
	if chunk := firstChunkFromLines([]*entities.TextLine{line}); chunk != nil {
		if chunk.FontStyle.FontName != "" {
			out[jsonFontType] = chunk.FontStyle.FontName
		}
		if size := serializeDouble(chunk.FontStyle.FontSize); size != nil {
			out[jsonFontSize] = size
		}
		out[jsonTextColor] = fmt.Sprintf("%v", chunk.FontStyle.Color)
	}
	return out
}

func SerializeTextChunkContentElement(c *entities.TextChunk) map[string]interface{} {
	if c == nil {
		return nil
	}

	out := essentialInfo(c, "paragraph")
	for key, value := range textInfoFromChunkAndContent(c, c.Text, c.IsHidden || c.IsHiddenOCG || c.IsOffPage || c.IsTiny) {
		out[key] = value
	}
	return out
}

func SerializeTextLineContentElement(line *entities.TextLine) map[string]interface{} {
	if line == nil {
		return nil
	}

	out := essentialInfo(line, "paragraph")
	for key, value := range textInfoFromLines([]*entities.TextLine{line}) {
		out[key] = value
	}
	return out
}
