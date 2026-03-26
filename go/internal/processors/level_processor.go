// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package processors

import (
	"sort"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

type HeadingLevelProcessor struct{}
type LevelProcessor struct{}

type headingStyleKey struct {
	FontSize   float64
	IsBold     bool
	IsItalic   bool
	FontFamily string
}

func (p *LevelProcessor) Process(headings []*entities.SemanticHeading) []*entities.SemanticHeading {
	if len(headings) == 0 {
		return headings
	}

	styleGroups := map[headingStyleKey][]*entities.SemanticHeading{}
	for _, heading := range headings {
		if heading == nil {
			continue
		}
		key := headingStyleKey{
			FontSize:   heading.FontSize,
			IsBold:     heading.IsBold,
			IsItalic:   heading.IsItalic,
			FontFamily: heading.FontFamily,
		}
		styleGroups[key] = append(styleGroups[key], heading)
	}

	ordered := make([]headingStyleKey, 0, len(styleGroups))
	for key := range styleGroups {
		ordered = append(ordered, key)
	}
	sort.Slice(ordered, func(i, j int) bool {
		return compareHeadingStyle(ordered[i], ordered[j]) < 0
	})

	for i, key := range ordered {
		level := i + 1
		if level > 6 {
			level = 6
		}
		for _, heading := range styleGroups[key] {
			heading.Level = level
		}
	}
	return headings
}

func compareHeadingStyle(left, right headingStyleKey) int {
	switch {
	case left.FontSize > right.FontSize:
		return -1
	case left.FontSize < right.FontSize:
		return 1
	case left.IsBold && !right.IsBold:
		return -1
	case !left.IsBold && right.IsBold:
		return 1
	case left.IsItalic && !right.IsItalic:
		return -1
	case !left.IsItalic && right.IsItalic:
		return 1
	case left.FontFamily < right.FontFamily:
		return -1
	case left.FontFamily > right.FontFamily:
		return 1
	default:
		return 0
	}
}
