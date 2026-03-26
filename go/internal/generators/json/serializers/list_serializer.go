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
	if previousID, ok := previousListID(list); ok {
		out[jsonPreviousListID] = previousID
	}
	if nextID, ok := nextListID(list); ok {
		out[jsonNextListID] = nextID
	}

	items := make([]interface{}, 0, len(list.Items))
	for _, item := range list.Items {
		if item == nil {
			continue
		}
		items = append(items, SerializeListItem(item))
	}
	out[jsonListItems] = items
	return out
}

func previousListID(list *entities.PDFList) (int, bool) {
	if len(list.Items) == 0 || list.Items[0] == nil {
		return 0, false
	}
	return parseNumericID(list.Items[0].GetID())
}

func nextListID(list *entities.PDFList) (int, bool) {
	lastIndex := len(list.Items) - 1
	if lastIndex < 0 || list.Items[lastIndex] == nil {
		return 0, false
	}
	return parseNumericID(list.Items[lastIndex].GetID())
}
