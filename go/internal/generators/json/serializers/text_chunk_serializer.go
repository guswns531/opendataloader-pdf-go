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

import "github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"

func SerializeTextChunk(c *entities.TextChunk) map[string]interface{} {
	if c == nil {
		return nil
	}

	out := essentialInfo(c, "text chunk")
	for key, value := range textInfoFromChunkAndContent(c, c.Text, isHiddenChunk(c)) {
		out[key] = value
	}
	return out
}

func SerializeTextLine(line *entities.TextLine) map[string]interface{} {
	if line == nil {
		return nil
	}

	out := essentialInfo(line, "text chunk")
	for key, value := range textInfoFromLines([]*entities.TextLine{line}) {
		out[key] = value
	}
	return out
}
