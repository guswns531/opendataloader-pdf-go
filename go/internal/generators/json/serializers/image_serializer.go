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
	"path/filepath"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

func SerializeImage(img *entities.SemanticImage, imageOutput string) map[string]interface{} {
	if img == nil {
		return nil
	}

	out := essentialInfo(img, "image")

	switch imageOutput {
	case api.ImageOutputEmbedded:
		imageFormat := imageFormatFromPath(img.ExternalPath)
		if dataURL := dataURL(img.Data, imageFormat); dataURL != "" {
			out[jsonData] = dataURL
			out[jsonImageFormat] = imageFormat
		}
	case api.ImageOutputExternal:
		if img.ExternalPath != "" {
			out[jsonSource] = img.ExternalPath
		}
	}

	return out
}

func imageFormatFromPath(path string) string {
	ext := filepath.Ext(path)
	if len(ext) > 1 {
		format := ext[1:]
		if format == "jpg" {
			return "jpeg"
		}
		return format
	}
	return api.ImageFormatPNG
}
