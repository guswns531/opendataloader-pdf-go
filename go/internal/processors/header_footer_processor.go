// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package processors

import (
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

type HeaderFooterProcessor struct{}

func (p *HeaderFooterProcessor) Process(elements []entities.IObject, pageHeight float64, includeHeaderFooter bool, ctx *containers.ProcessorContext) []entities.IObject {
	if len(elements) == 0 || pageHeight <= 0 {
		return elements
	}

	headerThreshold := pageHeight * 0.9
	footerThreshold := pageHeight * 0.1
	var headerLines []*entities.TextLine
	var footerLines []*entities.TextLine
	body := make([]entities.IObject, 0, len(elements))

	for _, element := range elements {
		b := element.GetBBox()
		top := b.Y + b.Height
		bottom := b.Y
		line, isLine := element.(*entities.TextLine)
		switch {
		case top >= headerThreshold:
			if isLine {
				headerLines = append(headerLines, line)
			}
			if includeHeaderFooter && !isLine {
				body = append(body, element)
			}
		case bottom <= footerThreshold:
			if isLine {
				footerLines = append(footerLines, line)
			}
			if includeHeaderFooter && !isLine {
				body = append(body, element)
			}
		default:
			body = append(body, element)
		}
	}

	if !includeHeaderFooter {
		return body
	}

	out := make([]entities.IObject, 0, len(body)+2)
	if len(headerLines) > 0 {
		out = append(out, &entities.SemanticHeaderFooter{
			BaseObject: entities.BaseObject{ID: nextID(ctx), BBox: linesBBox(headerLines)},
			Lines:      headerLines,
			IsHeader:   true,
		})
	}
	out = append(out, body...)
	if len(footerLines) > 0 {
		out = append(out, &entities.SemanticHeaderFooter{
			BaseObject: entities.BaseObject{ID: nextID(ctx), BBox: linesBBox(footerLines)},
			Lines:      footerLines,
			IsHeader:   false,
		})
	}
	return out
}
