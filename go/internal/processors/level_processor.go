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
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/utils/levels"
)

type HeadingLevelProcessor struct{}
type LevelProcessor struct{}

func (p *LevelProcessor) Process(headings []*entities.SemanticHeading) []*entities.SemanticHeading {
	if len(headings) == 0 {
		return headings
	}

	infoBySize := map[float64]*levels.LevelInfo{}
	for _, heading := range headings {
		if heading == nil {
			continue
		}
		info := infoBySize[heading.FontSize]
		if info == nil {
			info = &levels.LevelInfo{FontSize: heading.FontSize, IsBold: heading.IsBold}
			infoBySize[heading.FontSize] = info
		}
		info.Count++
		info.IsBold = info.IsBold || heading.IsBold
	}

	var ordered []*levels.LevelInfo
	for _, info := range infoBySize {
		ordered = append(ordered, info)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].FontSize == ordered[j].FontSize {
			if ordered[i].IsBold == ordered[j].IsBold {
				return ordered[i].Count > ordered[j].Count
			}
			return ordered[i].IsBold && !ordered[j].IsBold
		}
		return ordered[i].FontSize > ordered[j].FontSize
	})

	levelBySize := map[float64]int{}
	for i, info := range ordered {
		level := i + 1
		if level > 6 {
			level = 6
		}
		info.Level = level
		levelBySize[info.FontSize] = level
	}

	for _, heading := range headings {
		if heading != nil {
			heading.Level = levelBySize[heading.FontSize]
		}
	}
	return headings
}
