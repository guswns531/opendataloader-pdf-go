// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package readingorder

import (
	"sort"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const (
	defaultBeta             = 2.0
	defaultDensityThreshold = 0.9
	overlapThreshold        = 0.1
	minOverlapCount         = 2
	minGapThreshold         = 5.0
	defaultMinWidthRatio    = 0.05
	columnMinObjectCount    = 4
	columnMinRegionRatio    = 0.20
	columnMaxRegionRatio    = 0.80
	columnMinGapRatio       = 0.03
	columnHistogramBuckets  = 100
)

type XYCutPlusPlusSorter struct {
	MinWidthRatio float64
}

func (s XYCutPlusPlusSorter) Sort(elements []entities.IObject, pageWidth, pageHeight float64) []entities.IObject {
	if len(elements) <= 1 {
		return cloneObjects(elements)
	}

	valid := make([]entities.IObject, 0, len(elements))
	for _, element := range elements {
		if element != nil {
			valid = append(valid, element)
		}
	}
	if len(valid) <= 1 {
		return valid
	}

	crossLayout := identifyCrossLayoutElements(valid, defaultBeta)
	remaining := make([]entities.IObject, 0, len(valid))
	crossSet := make(map[string]struct{}, len(crossLayout))
	for _, item := range crossLayout {
		crossSet[item.GetID()] = struct{}{}
	}
	for _, item := range valid {
		if _, ok := crossSet[item.GetID()]; !ok {
			remaining = append(remaining, item)
		}
	}
	if len(remaining) == 0 {
		return sortByYThenX(valid)
	}

	preferHorizontalFirst := computeDensityRatio(remaining) > defaultDensityThreshold
	sortedMain := s.recursiveSegment(remaining, preferHorizontalFirst, 0, pageWidth, pageHeight)
	return mergeCrossLayoutElements(sortedMain, crossLayout)
}

type cutInfo struct {
	position float64
	gap      float64
}

func (s XYCutPlusPlusSorter) recursiveSegment(objects []entities.IObject, preferHorizontalFirst bool, regionX, pageWidth, pageHeight float64) []entities.IObject {
	if len(objects) <= 1 {
		return cloneObjects(objects)
	}

	if gapX := detectColumnSplit(objects, regionX, pageWidth); gapX >= 0 {
		groups := splitByVerticalCut(objects, gapX)
		if len(groups) > 1 {
			result := make([]entities.IObject, 0, len(objects))
			currentX := regionX
			for idx, group := range groups {
				width := pageWidth
				if idx == 0 {
					width = gapX - regionX
				} else if idx == len(groups)-1 {
					width = (regionX + pageWidth) - currentX
				}
				result = append(result, s.recursiveSegment(group, preferHorizontalFirst, currentX, width, pageHeight)...)
				if idx == 0 {
					currentX = gapX
				}
			}
			return result
		}
	}

	horizontalCut := findBestHorizontalCutWithProjection(objects)
	verticalCut := s.findBestVerticalCutWithProjection(objects, pageWidth)

	hasHorizontal := horizontalCut.gap >= minGapThreshold
	hasVertical := verticalCut.gap >= minGapThreshold

	var useHorizontal bool
	switch {
	case hasHorizontal && hasVertical:
		if horizontalCut.gap == verticalCut.gap {
			useHorizontal = preferHorizontalFirst
		} else {
			useHorizontal = horizontalCut.gap > verticalCut.gap
		}
	case hasHorizontal:
		useHorizontal = true
	case hasVertical:
		useHorizontal = false
	default:
		return sortByYThenX(objects)
	}

	var groups [][]entities.IObject
	if useHorizontal {
		groups = splitByHorizontalCut(objects, horizontalCut.position)
	} else {
		groups = splitByVerticalCut(objects, verticalCut.position)
	}
	if len(groups) <= 1 {
		return sortByYThenX(objects)
	}

	result := make([]entities.IObject, 0, len(objects))
	for _, group := range groups {
		nextRegionX := regionX
		nextPageWidth := pageWidth
		if !useHorizontal {
			groupRegion := calculateBoundingRegion(group)
			nextRegionX = groupRegion.X
			nextPageWidth = groupRegion.Width
		}
		result = append(result, s.recursiveSegment(group, preferHorizontalFirst, nextRegionX, nextPageWidth, pageHeight)...)
	}
	return result
}

func detectColumnSplit(objects []entities.IObject, regionX, regionWidth float64) float64 {
	if regionWidth <= 0 || len(objects) < columnMinObjectCount {
		return -1
	}

	minX := regionX + regionWidth*columnMinRegionRatio
	maxX := regionX + regionWidth*columnMaxRegionRatio
	if maxX <= minX {
		return -1
	}

	bucketWidth := (maxX - minX) / float64(columnHistogramBuckets)
	if bucketWidth <= 0 {
		return -1
	}

	occupied := make([]bool, columnHistogramBuckets)
	for _, obj := range objects {
		b := obj.GetBBox()
		startX := max(b.X, minX)
		endX := min(b.X+b.Width, maxX)
		if endX <= startX {
			continue
		}

		startIdx := int((startX - minX) / bucketWidth)
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx := int((endX - minX) / bucketWidth)
		if endIdx >= columnHistogramBuckets {
			endIdx = columnHistogramBuckets - 1
		}
		for i := startIdx; i <= endIdx; i++ {
			occupied[i] = true
		}
	}

	bestGapStart, bestGapLen := -1, 0
	currentStart, currentLen := -1, 0
	for idx, isOccupied := range occupied {
		if !isOccupied {
			if currentStart < 0 {
				currentStart = idx
			}
			currentLen++
			continue
		}
		if currentLen > bestGapLen {
			bestGapStart = currentStart
			bestGapLen = currentLen
		}
		currentStart, currentLen = -1, 0
	}
	if currentLen > bestGapLen {
		bestGapStart = currentStart
		bestGapLen = currentLen
	}
	if bestGapLen <= 0 {
		return -1
	}

	actualGapWidth := float64(bestGapLen) * bucketWidth
	if actualGapWidth < regionWidth*columnMinGapRatio {
		return -1
	}

	gapX := minX + float64(bestGapStart)*bucketWidth + actualGapWidth/2
	leftCount, rightCount := 0, 0
	for _, obj := range objects {
		b := obj.GetBBox()
		centerX := b.X + b.Width/2
		if centerX < gapX {
			leftCount++
		} else {
			rightCount++
		}
	}
	if leftCount == 0 || rightCount == 0 {
		return -1
	}

	return gapX
}

func identifyCrossLayoutElements(objects []entities.IObject, beta float64) []entities.IObject {
	if len(objects) < 3 {
		return nil
	}
	maxWidth := 0.0
	for _, obj := range objects {
		width := obj.GetBBox().Width
		if width > maxWidth {
			maxWidth = width
		}
	}
	threshold := beta * maxWidth
	var cross []entities.IObject
	for _, obj := range objects {
		if obj.GetBBox().Width >= threshold && hasMinimumOverlaps(obj, objects, minOverlapCount) {
			cross = append(cross, obj)
		}
	}
	return cross
}

func hasMinimumOverlaps(element entities.IObject, objects []entities.IObject, minCount int) bool {
	overlapCount := 0
	for _, other := range objects {
		if other == element {
			continue
		}
		if calculateHorizontalOverlapRatio(element.GetBBox(), other.GetBBox()) >= overlapThreshold {
			overlapCount++
			if overlapCount >= minCount {
				return true
			}
		}
	}
	return false
}

func calculateHorizontalOverlapRatio(a, b entities.BoundingBox) float64 {
	left := max(a.X, b.X)
	right := min(a.X+a.Width, b.X+b.Width)
	overlap := right - left
	if overlap <= 0 {
		return 0
	}
	smaller := min(a.Width, b.Width)
	if smaller <= 0 {
		return 0
	}
	return overlap / smaller
}

func computeDensityRatio(objects []entities.IObject) float64 {
	if len(objects) == 0 {
		return 1.0
	}
	region := calculateBoundingRegion(objects)
	area := region.Width * region.Height
	if area <= 0 {
		return 1.0
	}
	total := 0.0
	for _, obj := range objects {
		b := obj.GetBBox()
		total += b.Width * b.Height
	}
	ratio := total / area
	if ratio > 1.0 {
		return 1.0
	}
	return ratio
}

func calculateBoundingRegion(objects []entities.IObject) entities.BoundingBox {
	if len(objects) == 0 {
		return entities.BoundingBox{}
	}
	first := objects[0].GetBBox()
	minX, minY := first.X, first.Y
	maxX, maxY := first.X+first.Width, first.Y+first.Height
	page := first.Page
	for _, obj := range objects[1:] {
		b := obj.GetBBox()
		minX = min(minX, b.X)
		minY = min(minY, b.Y)
		maxX = max(maxX, b.X+b.Width)
		maxY = max(maxY, b.Y+b.Height)
	}
	return entities.BoundingBox{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY, Page: page}
}

func (s XYCutPlusPlusSorter) findBestVerticalCutWithProjection(objects []entities.IObject, pageWidth float64) cutInfo {
	edgeCut := findVerticalCutByEdges(objects)
	if edgeCut.gap >= minGapThreshold {
		return edgeCut
	}

	ratio := s.MinWidthRatio
	if ratio <= 0 {
		ratio = defaultMinWidthRatio
	}
	region := calculateBoundingRegion(objects)
	regionWidth := region.Width
	if regionWidth <= 0 {
		regionWidth = pageWidth
	}
	if regionWidth <= 0 {
		return edgeCut
	}

	narrowThreshold := regionWidth * ratio
	filtered := make([]entities.IObject, 0, len(objects))
	for _, obj := range objects {
		if obj.GetBBox().Width >= narrowThreshold {
			filtered = append(filtered, obj)
		}
	}
	if len(filtered) >= 2 && len(filtered) < len(objects) {
		filteredCut := findVerticalCutByEdges(filtered)
		if filteredCut.gap > edgeCut.gap && filteredCut.gap >= minGapThreshold {
			return filteredCut
		}
	}
	return edgeCut
}

func findVerticalCutByEdges(objects []entities.IObject) cutInfo {
	sorted := cloneObjects(objects)
	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i].GetBBox(), sorted[j].GetBBox()
		if a.X == b.X {
			return a.X+a.Width < b.X+b.Width
		}
		return a.X < b.X
	})

	var prevRight *float64
	largestGap := 0.0
	cutPosition := 0.0
	for _, obj := range sorted {
		b := obj.GetBBox()
		left := b.X
		right := b.X + b.Width
		if prevRight != nil && left > *prevRight {
			gap := left - *prevRight
			if gap > largestGap {
				largestGap = gap
				cutPosition = (*prevRight + left) / 2.0
			}
		}
		if prevRight == nil {
			prevRight = &right
		} else if right > *prevRight {
			*prevRight = right
		}
	}
	return cutInfo{position: cutPosition, gap: largestGap}
}

func findBestHorizontalCutWithProjection(objects []entities.IObject) cutInfo {
	if len(objects) < 2 {
		return cutInfo{}
	}
	sorted := cloneObjects(objects)
	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i].GetBBox(), sorted[j].GetBBox()
		at, bt := a.Y+a.Height, b.Y+b.Height
		if at == bt {
			return a.Y > b.Y
		}
		return at > bt
	})

	var prevBottom *float64
	largestGap := 0.0
	cutPosition := 0.0
	for _, obj := range sorted {
		b := obj.GetBBox()
		top := b.Y + b.Height
		bottom := b.Y
		if prevBottom != nil && *prevBottom > top {
			gap := *prevBottom - top
			if gap > largestGap {
				largestGap = gap
				cutPosition = (*prevBottom + top) / 2.0
			}
		}
		if prevBottom == nil {
			prevBottom = &bottom
		} else if bottom < *prevBottom {
			*prevBottom = bottom
		}
	}
	return cutInfo{position: cutPosition, gap: largestGap}
}

func splitByHorizontalCut(objects []entities.IObject, cutY float64) [][]entities.IObject {
	var above, below []entities.IObject
	for _, obj := range objects {
		b := obj.GetBBox()
		centerY := b.Y + b.Height/2
		if centerY > cutY {
			above = append(above, obj)
		} else {
			below = append(below, obj)
		}
	}
	var groups [][]entities.IObject
	if len(above) > 0 {
		groups = append(groups, above)
	}
	if len(below) > 0 {
		groups = append(groups, below)
	}
	return groups
}

func splitByVerticalCut(objects []entities.IObject, cutX float64) [][]entities.IObject {
	var left, right []entities.IObject
	for _, obj := range objects {
		b := obj.GetBBox()
		centerX := b.X + b.Width/2
		if centerX < cutX {
			left = append(left, obj)
		} else {
			right = append(right, obj)
		}
	}
	var groups [][]entities.IObject
	if len(left) > 0 {
		groups = append(groups, left)
	}
	if len(right) > 0 {
		groups = append(groups, right)
	}
	return groups
}

func mergeCrossLayoutElements(sortedMain, crossLayout []entities.IObject) []entities.IObject {
	if len(crossLayout) == 0 {
		return sortedMain
	}
	if len(sortedMain) == 0 {
		return sortByYThenX(crossLayout)
	}

	sortedCross := sortByYThenX(crossLayout)
	result := make([]entities.IObject, 0, len(sortedMain)+len(sortedCross))
	mainIndex, crossIndex := 0, 0
	for mainIndex < len(sortedMain) || crossIndex < len(sortedCross) {
		switch {
		case crossIndex >= len(sortedCross):
			result = append(result, sortedMain[mainIndex])
			mainIndex++
		case mainIndex >= len(sortedMain):
			result = append(result, sortedCross[crossIndex])
			crossIndex++
		default:
			mainTop := sortedMain[mainIndex].GetBBox().Y + sortedMain[mainIndex].GetBBox().Height
			crossTop := sortedCross[crossIndex].GetBBox().Y + sortedCross[crossIndex].GetBBox().Height
			if crossTop >= mainTop {
				result = append(result, sortedCross[crossIndex])
				crossIndex++
			} else {
				result = append(result, sortedMain[mainIndex])
				mainIndex++
			}
		}
	}
	return result
}

func sortByYThenX(objects []entities.IObject) []entities.IObject {
	sorted := cloneObjects(objects)
	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i].GetBBox(), sorted[j].GetBBox()
		at, bt := a.Y+a.Height, b.Y+b.Height
		if at == bt {
			return a.X < b.X
		}
		return at > bt
	})
	return sorted
}

func cloneObjects(objects []entities.IObject) []entities.IObject {
	if len(objects) == 0 {
		return nil
	}
	out := make([]entities.IObject, len(objects))
	copy(out, objects)
	return out
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
