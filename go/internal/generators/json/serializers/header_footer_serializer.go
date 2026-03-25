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

func SerializeHeaderFooter(h *entities.SemanticHeaderFooter) map[string]interface{} {
	if h == nil {
		return nil
	}

	objectType := "footer"
	if h.IsHeader {
		objectType = "header"
	}
	out := essentialInfo(h, objectType)
	out[jsonKids] = serializeLineChildren(h.Lines)
	return out
}

func serializeLineChildren(lines []*entities.TextLine) []interface{} {
	out := make([]interface{}, 0, len(lines))
	for _, line := range lines {
		if line == nil {
			continue
		}
		out = append(out, SerializeTextLine(line))
	}
	return out
}
