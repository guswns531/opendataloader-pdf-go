// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package processors

import (
	"regexp"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

var captionPattern = regexp.MustCompile(`^(?:그림|Fig(?:ure)?|표|Table|도표)\s*[:.]?\s*[A-Za-z0-9가-힣\-\.]+`)

type CaptionProcessor struct{}

func (p *CaptionProcessor) Process(elements []entities.IObject, ctx *containers.ProcessorContext) []entities.IObject {
	out := make([]entities.IObject, 0, len(elements))
	for _, element := range elements {
		text := objectText(element)
		if text == "" || !captionPattern.MatchString(strings.TrimSpace(text)) {
			out = append(out, element)
			continue
		}
		refType := "figure"
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(text)), "표") || strings.HasPrefix(strings.ToLower(strings.TrimSpace(text)), "table") {
			refType = "table"
		}
		out = append(out, &entities.SemanticCaption{
			BaseObject: entities.BaseObject{ID: nextID(ctx), BBox: element.GetBBox()},
			Text:       strings.TrimSpace(text),
			RefType:    refType,
		})
	}
	return out
}
