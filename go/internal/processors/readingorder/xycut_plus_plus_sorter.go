// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package readingorder

import (
	"math"
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
	maxSegmentDepth         = 128
	columnMinObjectCount    = 4
	columnMinRegionRatio    = 0.20
	columnMaxRegionRatio    = 0.80
	columnHistogramBuckets  = 100
	sidebarMaxWidthRatio    = 0.45
	sidebarEdgeSlack        = 20.0
	sidebarMinMainCount     = 3
	stackedEdgeColumnCount  = 3
	stackedEdgeGapSlack     = 12.0
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
	sortedMain := s.sortWithColumnAwareness(remaining, preferHorizontalFirst, entities.BoundingBox{
		X:      0,
		Width:  pageWidth,
		Height: pageHeight,
	})
	return mergeCrossLayoutElements(sortedMain, crossLayout)
}

type cutInfo struct {
	position float64
	gap      float64
}

func (s XYCutPlusPlusSorter) recursiveSegment(objects []entities.IObject, preferHorizontalFirst bool, regionX, pageWidth, pageHeight float64) []entities.IObject {
	return s.recursiveSegmentWithDepth(objects, preferHorizontalFirst, regionX, pageWidth, pageHeight, 0)
}

func (s XYCutPlusPlusSorter) recursiveSegmentWithDepth(objects []entities.IObject, preferHorizontalFirst bool, regionX, pageWidth, pageHeight float64, depth int) []entities.IObject {
	if len(objects) <= 1 {
		return cloneObjects(objects)
	}
	if depth >= maxSegmentDepth {
		return sortByYThenX(objects)
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
		result = append(result, s.recursiveSegmentWithDepth(group, preferHorizontalFirst, nextRegionX, nextPageWidth, pageHeight, depth+1)...)
	}
	return result
}

func (s XYCutPlusPlusSorter) sortWithColumnAwareness(objects []entities.IObject, preferHorizontalFirst bool, region entities.BoundingBox) []entities.IObject {
	if len(objects) <= 1 {
		return cloneObjects(objects)
	}

	marginalFloating := identifyMarginalFloatingElements(objects, region)
	if len(marginalFloating) > 0 && len(marginalFloating) < len(objects) {
		floatingSet := make(map[string]struct{}, len(marginalFloating))
		remaining := make([]entities.IObject, 0, len(objects)-len(marginalFloating))
		for _, obj := range marginalFloating {
			floatingSet[obj.GetID()] = struct{}{}
		}
		for _, obj := range objects {
			if _, ok := floatingSet[obj.GetID()]; !ok {
				remaining = append(remaining, obj)
			}
		}
		if len(remaining) >= sidebarMinMainCount {
			mainRegion := calculateBoundingRegion(remaining)
			if mainRegion.Width <= 0 {
				mainRegion = region
			}
			sortedMain := s.sortWithColumnAwareness(remaining, preferHorizontalFirst, mainRegion)
			edgeColumns, deferredFloating := partitionStackedEdgeFloatingColumns(marginalFloating, region)
			if len(edgeColumns) > 0 {
				sortedMain = mergeDeferredEdgeColumns(sortedMain, edgeColumns)
			}
			return mergeDeferredFloatingElements(sortedMain, deferredFloating)
		}
	}

	if gapX := detectColumnSplit(objects, region.X, region.Width); gapX >= 0 {
		left, right, neutral := splitByVerticalCutWithNeutral(objects, gapX)
		if len(left) > 0 && len(right) > 0 {
			leftRegion := calculateBoundingRegion(left)
			if leftRegion.Width <= 0 {
				leftRegion = region
			}
			rightRegion := calculateBoundingRegion(right)
			if rightRegion.Width <= 0 {
				rightRegion = region
			}

			if len(neutral) == 0 {
				if sidebar, sidebarRegion, main, mainRegion, ok := detectMarginalSidebarSplit(left, leftRegion, right, rightRegion, region); ok {
					sortedMain := s.sortWithColumnAwareness(main, preferHorizontalFirst, mainRegion)
					sortedSidebar := s.sortWithColumnAwareness(sidebar, preferHorizontalFirst, sidebarRegion)
					return mergeDeferredFloatingElements(sortedMain, sortedSidebar)
				}
			}

			sortedLeft := s.sortWithColumnAwareness(left, preferHorizontalFirst, leftRegion)
			sortedRight := s.sortWithColumnAwareness(right, preferHorizontalFirst, rightRegion)
			result := append(cloneObjects(sortedLeft), sortedRight...)
			if (len(neutral) > 0 || columnsAreImbalanced(left, right)) &&
				detectColumnSplit(left, leftRegion.X, leftRegion.Width) < 0 &&
				detectColumnSplit(right, rightRegion.X, rightRegion.Width) < 0 {
				result = mergeBalancedColumns(sortedLeft, sortedRight)
			}
			if len(neutral) == 0 {
				return result
			}
			return mergeDeferredFloatingElements(result, neutral)
		}
	}

	return s.recursiveSegment(objects, preferHorizontalFirst, region.X, region.Width, region.Height)
}

func detectColumnSplit(objects []entities.IObject, regionX, regionWidth float64) float64 {
	split := detectColumnSplitInfo(objects, regionX, regionWidth)

	midpoint := regionX + regionWidth/2
	filtered := make([]entities.IObject, 0, len(objects))
	for _, obj := range objects {
		b := obj.GetBBox()
		leftSpan := midpoint - b.X
		rightSpan := (b.X + b.Width) - midpoint
		if leftSpan >= minGapThreshold && rightSpan >= minGapThreshold {
			continue
		}
		filtered = append(filtered, obj)
	}
	if len(filtered) >= columnMinObjectCount && len(filtered) < len(objects) {
		filteredSplit := detectColumnSplitInfo(filtered, regionX, regionWidth)
		left, right, neutral := splitByVerticalCutWithNeutral(objects, filteredSplit.position)
		if filteredSplit.gap > split.gap &&
			len(left) > 0 &&
			len(right) > 0 &&
			len(neutral) > 0 &&
			len(neutral) < smallestCount(len(left), len(right)) {
			split = filteredSplit
		}
	}

	if split.gap < minGapThreshold {
		return -1
	}
	return split.position
}

func detectColumnSplitInfo(objects []entities.IObject, regionX, regionWidth float64) cutInfo {
	if regionWidth <= 0 || len(objects) < columnMinObjectCount {
		return cutInfo{position: -1}
	}

	minX := regionX + regionWidth*columnMinRegionRatio
	maxX := regionX + regionWidth*columnMaxRegionRatio
	if maxX <= minX {
		return cutInfo{position: -1}
	}

	bucketWidth := (maxX - minX) / float64(columnHistogramBuckets)
	if bucketWidth <= 0 {
		return cutInfo{position: -1}
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
		// Treat spans as half-open so elements ending on a bucket boundary
		// do not erase a real gutter in the next bucket.
		endIdx := int(math.Ceil((endX-minX)/bucketWidth)) - 1
		if endIdx < 0 {
			endIdx = 0
		}
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
		if currentStart > 0 && currentLen > bestGapLen {
			bestGapStart = currentStart
			bestGapLen = currentLen
		}
		currentStart, currentLen = -1, 0
	}
	if bestGapLen <= 0 {
		return cutInfo{position: -1}
	}

	actualGapWidth := float64(bestGapLen) * bucketWidth
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
		return cutInfo{position: -1}
	}

	return cutInfo{position: gapX, gap: actualGapWidth}
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

func splitByVerticalCutWithNeutral(objects []entities.IObject, cutX float64) ([]entities.IObject, []entities.IObject, []entities.IObject) {
	var left, right, neutral []entities.IObject
	for _, obj := range objects {
		b := obj.GetBBox()
		leftSpan := cutX - b.X
		rightSpan := (b.X + b.Width) - cutX
		if leftSpan >= minGapThreshold && rightSpan >= minGapThreshold {
			neutral = append(neutral, obj)
			continue
		}

		centerX := b.X + b.Width/2
		if centerX < cutX {
			left = append(left, obj)
		} else {
			right = append(right, obj)
		}
	}
	return left, right, neutral
}

func mergeBalancedColumns(left, right []entities.IObject) []entities.IObject {
	if len(left) == 0 {
		return cloneObjects(right)
	}
	if len(right) == 0 {
		return cloneObjects(left)
	}

	leftTop, leftBottom := verticalSpan(left)
	rightTop, rightBottom := verticalSpan(right)
	overlapTop := min(leftTop, rightTop)
	overlapBottom := max(leftBottom, rightBottom)
	if overlapTop <= overlapBottom {
		result := cloneObjects(left)
		return append(result, right...)
	}

	leftLead, leftCore, leftTrail := partitionByVerticalBand(left, overlapTop, overlapBottom)
	rightLead, rightCore, rightTrail := partitionByVerticalBand(right, overlapTop, overlapBottom)

	result := make([]entities.IObject, 0, len(left)+len(right))
	if len(leftLead)+len(rightLead) > 0 {
		result = append(result, sortByYThenX(append(cloneObjects(leftLead), rightLead...))...)
	}
	result = append(result, leftCore...)
	result = append(result, rightCore...)
	if len(leftTrail)+len(rightTrail) > 0 {
		result = append(result, sortByYThenX(append(cloneObjects(leftTrail), rightTrail...))...)
	}
	return result
}

func verticalSpan(objects []entities.IObject) (top float64, bottom float64) {
	first := objects[0].GetBBox()
	top = first.Y + first.Height
	bottom = first.Y
	for _, obj := range objects[1:] {
		b := obj.GetBBox()
		top = max(top, b.Y+b.Height)
		bottom = min(bottom, b.Y)
	}
	return top, bottom
}

func columnsAreImbalanced(left, right []entities.IObject) bool {
	leftTop, leftBottom := verticalSpan(left)
	rightTop, rightBottom := verticalSpan(right)
	return math.Abs(leftTop-rightTop) >= minGapThreshold || math.Abs(leftBottom-rightBottom) >= minGapThreshold
}

func smallestCount(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func identifyMarginalFloatingElements(objects []entities.IObject, region entities.BoundingBox) []entities.IObject {
	if len(objects) < sidebarMinMainCount+1 {
		return nil
	}

	maxWidth := 0.0
	for _, obj := range objects {
		if width := obj.GetBBox().Width; width > maxWidth {
			maxWidth = width
		}
	}
	if maxWidth <= 0 {
		return nil
	}

	threshold := maxWidth * sidebarMaxWidthRatio
	floating := make([]entities.IObject, 0, len(objects))
	for _, obj := range objects {
		b := obj.GetBBox()
		if b.Width > threshold || !regionTouchesHorizontalEdge(b, region, sidebarEdgeSlack) {
			continue
		}
		if hasMinimumVerticalOverlaps(obj, objects, minOverlapCount) {
			floating = append(floating, obj)
		}
	}
	return floating
}

func detectMarginalSidebarSplit(
	left []entities.IObject,
	leftRegion entities.BoundingBox,
	right []entities.IObject,
	rightRegion entities.BoundingBox,
	parentRegion entities.BoundingBox,
) ([]entities.IObject, entities.BoundingBox, []entities.IObject, entities.BoundingBox, bool) {
	if len(left) == 0 || len(right) == 0 {
		return nil, entities.BoundingBox{}, nil, entities.BoundingBox{}, false
	}

	sidebar, sidebarRegion := left, leftRegion
	main, mainRegion := right, rightRegion
	if len(left) > len(right) {
		sidebar, sidebarRegion = right, rightRegion
		main, mainRegion = left, leftRegion
	}

	if len(main) < sidebarMinMainCount || len(sidebar) >= len(main) {
		return nil, entities.BoundingBox{}, nil, entities.BoundingBox{}, false
	}
	if sidebarRegion.Width <= 0 || mainRegion.Width <= 0 {
		return nil, entities.BoundingBox{}, nil, entities.BoundingBox{}, false
	}
	if sidebarRegion.Width > mainRegion.Width*sidebarMaxWidthRatio {
		return nil, entities.BoundingBox{}, nil, entities.BoundingBox{}, false
	}
	if !regionTouchesHorizontalEdge(sidebarRegion, parentRegion, sidebarEdgeSlack) {
		return nil, entities.BoundingBox{}, nil, entities.BoundingBox{}, false
	}
	if detectColumnSplit(sidebar, sidebarRegion.X, sidebarRegion.Width) >= 0 {
		return nil, entities.BoundingBox{}, nil, entities.BoundingBox{}, false
	}

	return sidebar, sidebarRegion, main, mainRegion, true
}

func regionTouchesHorizontalEdge(region, parent entities.BoundingBox, slack float64) bool {
	parentRight := parent.X + parent.Width
	regionRight := region.X + region.Width
	return region.X-parent.X <= slack || parentRight-regionRight <= slack
}

func hasMinimumVerticalOverlaps(element entities.IObject, objects []entities.IObject, minCount int) bool {
	overlapCount := 0
	for _, other := range objects {
		if other == element {
			continue
		}
		if calculateVerticalOverlapRatio(element.GetBBox(), other.GetBBox()) >= overlapThreshold {
			overlapCount++
			if overlapCount >= minCount {
				return true
			}
		}
	}
	return false
}

func calculateVerticalOverlapRatio(a, b entities.BoundingBox) float64 {
	bottom := max(a.Y, b.Y)
	top := min(a.Y+a.Height, b.Y+b.Height)
	overlap := top - bottom
	if overlap <= 0 {
		return 0
	}
	smaller := min(a.Height, b.Height)
	if smaller <= 0 {
		return 0
	}
	return overlap / smaller
}

func partitionByVerticalBand(objects []entities.IObject, overlapTop, overlapBottom float64) ([]entities.IObject, []entities.IObject, []entities.IObject) {
	lead := make([]entities.IObject, 0, len(objects))
	core := make([]entities.IObject, 0, len(objects))
	trail := make([]entities.IObject, 0, len(objects))
	for _, obj := range objects {
		b := obj.GetBBox()
		centerY := b.Y + b.Height/2
		switch {
		case centerY > overlapTop:
			lead = append(lead, obj)
		case centerY < overlapBottom:
			trail = append(trail, obj)
		default:
			core = append(core, obj)
		}
	}
	return lead, core, trail
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

func mergeDeferredFloatingElements(sortedMain, floating []entities.IObject) []entities.IObject {
	if len(floating) == 0 {
		return sortedMain
	}
	if len(sortedMain) == 0 {
		return sortByYThenX(floating)
	}

	sortedFloating := sortByYThenX(floating)
	result := make([]entities.IObject, 0, len(sortedMain)+len(sortedFloating))
	mainIndex, floatingIndex := 0, 0
	for floatingIndex < len(sortedFloating) {
		floatingBottom := sortedFloating[floatingIndex].GetBBox().Y
		for mainIndex < len(sortedMain) {
			mainTop := sortedMain[mainIndex].GetBBox().Y + sortedMain[mainIndex].GetBBox().Height
			if mainTop < floatingBottom {
				break
			}
			result = append(result, sortedMain[mainIndex])
			mainIndex++
		}
		result = append(result, sortedFloating[floatingIndex])
		floatingIndex++
	}
	if mainIndex < len(sortedMain) {
		result = append(result, sortedMain[mainIndex:]...)
	}
	return result
}

func partitionStackedEdgeFloatingColumns(floating []entities.IObject, region entities.BoundingBox) ([][]entities.IObject, []entities.IObject) {
	if len(floating) < stackedEdgeColumnCount {
		return nil, floating
	}

	type edgeGroup struct {
		side    string
		members []entities.IObject
	}

	var groups []edgeGroup
	for _, obj := range floating {
		side := floatingEdgeSide(obj.GetBBox(), region)
		if side == "" {
			continue
		}

		placed := false
		for idx := range groups {
			if groups[idx].side != side || !sameStackedEdgeColumn(groups[idx].members[0].GetBBox(), obj.GetBBox(), side) {
				continue
			}
			groups[idx].members = append(groups[idx].members, obj)
			placed = true
			break
		}
		if !placed {
			groups = append(groups, edgeGroup{side: side, members: []entities.IObject{obj}})
		}
	}

	stackedSet := make(map[string]struct{})
	columns := make([][]entities.IObject, 0, len(groups))
	for _, group := range groups {
		if !isStackedEdgeColumn(group.members) {
			continue
		}
		column := sortByYThenX(group.members)
		for _, obj := range column {
			stackedSet[obj.GetID()] = struct{}{}
		}
		columns = append(columns, column)
	}

	if len(columns) == 0 {
		return nil, floating
	}

	regular := make([]entities.IObject, 0, len(floating)-len(stackedSet))
	for _, obj := range floating {
		if _, ok := stackedSet[obj.GetID()]; !ok {
			regular = append(regular, obj)
		}
	}

	sort.Slice(columns, func(i, j int) bool {
		return calculateBoundingRegion(columns[i]).Y+calculateBoundingRegion(columns[i]).Height >
			calculateBoundingRegion(columns[j]).Y+calculateBoundingRegion(columns[j]).Height
	})

	return columns, regular
}

func floatingEdgeSide(b, region entities.BoundingBox) string {
	if regionTouchesHorizontalEdge(b, region, sidebarEdgeSlack) {
		if b.X-region.X <= sidebarEdgeSlack {
			return "left"
		}
		return "right"
	}
	return ""
}

func sameStackedEdgeColumn(a, b entities.BoundingBox, side string) bool {
	if math.Abs(a.Width-b.Width) > minGapThreshold {
		return false
	}
	if math.Abs(a.X-b.X) > minGapThreshold {
		return false
	}

	if side == "right" {
		aRight := a.X + a.Width
		bRight := b.X + b.Width
		return math.Abs(aRight-bRight) <= minGapThreshold
	}

	return true
}

func isStackedEdgeColumn(objects []entities.IObject) bool {
	if len(objects) < stackedEdgeColumnCount {
		return false
	}

	sorted := sortByYThenX(objects)
	for idx := 1; idx < len(sorted); idx++ {
		prev := sorted[idx-1].GetBBox()
		curr := sorted[idx].GetBBox()
		gap := prev.Y - (curr.Y + curr.Height)
		if gap < 0 || gap > stackedEdgeGapSlack {
			return false
		}
	}
	return true
}

func mergeDeferredEdgeColumns(sortedMain []entities.IObject, edgeColumns [][]entities.IObject) []entities.IObject {
	if len(edgeColumns) == 0 {
		return sortedMain
	}

	result := cloneObjects(sortedMain)
	for _, column := range edgeColumns {
		insertAt := findDeferredEdgeColumnInsertionIndex(result, column)
		merged := make([]entities.IObject, 0, len(result)+len(column))
		merged = append(merged, result[:insertAt]...)
		merged = append(merged, column...)
		merged = append(merged, result[insertAt:]...)
		result = merged
	}
	return result
}

func findDeferredEdgeColumnInsertionIndex(sortedMain, column []entities.IObject) int {
	if len(sortedMain) == 0 {
		return 0
	}

	columnRegion := calculateBoundingRegion(column)
	columnBottom := columnRegion.Y
	for idx, obj := range sortedMain {
		b := obj.GetBBox()
		if calculateHorizontalOverlapRatio(b, columnRegion) >= overlapThreshold && b.Y+b.Height/2 <= columnBottom {
			return idx
		}
	}
	return len(sortedMain)
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
