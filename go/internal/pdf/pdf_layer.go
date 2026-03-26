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

package pdf

type PDFLayer string

const (
	PDFLayerContent             PDFLayer = "content"
	PDFLayerTableCells          PDFLayer = "table cells"
	PDFLayerListItems           PDFLayer = "list items"
	PDFLayerTableContent        PDFLayer = "table content"
	PDFLayerListContent         PDFLayer = "list content"
	PDFLayerTextBlockContent    PDFLayer = "text blocks content"
	PDFLayerHeaderFooterContent PDFLayer = "header and footer content"
)
