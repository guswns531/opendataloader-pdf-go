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

var captionPattern = regexp.MustCompile(`^(?:그림|Fig|Figure|표|Table|도표)\s*[\d\-\.]+`)

type CaptionProcessor struct{}

func (p *CaptionProcessor) Process(elements []entities.IObject, ctx *containers.ProcessorContext) []entities.IObject {
	out := make([]entities.IObject, 0, len(elements))
	for idx, element := range elements {
		text := objectText(element)
		if text == "" || !captionPattern.MatchString(strings.TrimSpace(text)) {
			out = append(out, element)
			continue
		}
		refType, ok := captionReferenceType(elements, idx, strings.TrimSpace(text))
		if !ok {
			out = append(out, element)
			continue
		}
		out = append(out, &entities.SemanticCaption{
			BaseObject: entities.BaseObject{ID: nextID(ctx), BBox: element.GetBBox()},
			Text:       strings.TrimSpace(text),
			RefType:    refType,
		})
	}
	return out
}

func captionReferenceType(elements []entities.IObject, idx int, text string) (string, bool) {
	if strings.HasPrefix(text, "표") || strings.HasPrefix(strings.ToLower(text), "table") {
		return "table", len(elements) == 1 || hasCaptionTarget(elements, idx, entities.ObjectTypeTable)
	}
	return "figure", len(elements) == 1 || hasCaptionTarget(elements, idx, entities.ObjectTypeImage) || hasCaptionTarget(elements, idx, entities.ObjectTypeTable)
}

func hasCaptionTarget(elements []entities.IObject, idx int, objectType entities.ObjectType) bool {
	return matchesCaptionTarget(neighborObject(elements, idx, -1), objectType) || matchesCaptionTarget(neighborObject(elements, idx, 1), objectType)
}

func neighborObject(elements []entities.IObject, idx, step int) entities.IObject {
	for pos := idx + step; pos >= 0 && pos < len(elements); pos += step {
		if elements[pos] != nil {
			return elements[pos]
		}
	}
	return nil
}

func matchesCaptionTarget(obj entities.IObject, objectType entities.ObjectType) bool {
	if obj == nil {
		return false
	}
	if obj.GetObjectType() == objectType {
		return true
	}
	return objectType == entities.ObjectTypeImage && obj.GetObjectType() == entities.ObjectTypeTable
}
