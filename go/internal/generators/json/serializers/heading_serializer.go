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

func SerializeHeading(h *entities.SemanticHeading) map[string]interface{} {
	if h == nil {
		return nil
	}

	out := essentialInfo(h, "heading")
	level := h.Level
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	out[jsonHeadingLevel] = level
	for key, value := range textInfoFromLines(h.Lines) {
		out[key] = value
	}
	return out
}
