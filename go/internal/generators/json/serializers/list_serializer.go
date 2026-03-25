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

func SerializeList(list *entities.PDFList) map[string]interface{} {
	if list == nil {
		return nil
	}

	out := essentialInfo(list, "list")
	out[jsonNumberingStyle] = numberingStyle(list)
	out[jsonNumberOfListItems] = len(list.Items)

	items := make([]interface{}, 0, len(list.Items))
	for idx, item := range list.Items {
		if item == nil {
			continue
		}
		if idx > 0 && list.Items[idx-1] != nil && list.Items[idx-1].ID != "" {
			item = cloneListItem(item)
			item.Level = max(item.Level, 1)
		}
		items = append(items, SerializeListItem(item))
	}
	out[jsonListItems] = items
	return out
}

func cloneListItem(item *entities.ListItem) *entities.ListItem {
	cloned := *item
	return &cloned
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
