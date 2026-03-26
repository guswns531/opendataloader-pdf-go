// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors

import (
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const tinyTextThreshold = 1.0

type ContentFilterProcessor struct{}

func FilterContent(chunks []*entities.TextChunk, config *api.FilterConfig, pageWidth, pageHeight float64) []*entities.TextChunk {
	filtered := make([]*entities.TextChunk, 0, len(chunks))
	for _, chunk := range chunks {
		if chunk == nil {
			continue
		}

		chunk.IsTiny = isTinyText(chunk)
		chunk.IsOffPage = isOffPage(chunk.BBox, pageWidth, pageHeight)

		if shouldFilterHiddenText(config) && chunk.IsHidden {
			continue
		}
		if shouldFilterHiddenOCG(config) && chunk.IsHiddenOCG {
			continue
		}
		if shouldFilterTinyText(config) && chunk.IsTiny {
			continue
		}
		if shouldFilterOffPage(config) && chunk.IsOffPage {
			continue
		}

		filtered = append(filtered, chunk)
	}

	return filtered
}

func shouldFilterHiddenText(config *api.FilterConfig) bool {
	return config == nil || !config.DisableHiddenText
}

func shouldFilterHiddenOCG(config *api.FilterConfig) bool {
	return config == nil || !config.DisableHiddenOCG
}

func shouldFilterTinyText(config *api.FilterConfig) bool {
	return config == nil || !config.DisableTiny
}

func shouldFilterOffPage(config *api.FilterConfig) bool {
	return config == nil || !config.DisableOffPage
}

func isTinyText(chunk *entities.TextChunk) bool {
	return chunk != nil && chunk.FontStyle.FontSize < tinyTextThreshold
}

func isOffPage(bbox entities.BoundingBox, pageWidth, pageHeight float64) bool {
	if pageWidth <= 0 || pageHeight <= 0 {
		return false
	}
	return bbox.X+bbox.Width <= 0 ||
		bbox.Y+bbox.Height <= 0 ||
		bbox.X >= pageWidth ||
		bbox.Y >= pageHeight
}
