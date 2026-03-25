// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package readingorder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/processors/readingorder"
)

func TestXYCutPlusPlusSorterSortsTwoColumnLayout(t *testing.T) {
	sorter := readingorder.XYCutPlusPlusSorter{}
	elements := []entities.IObject{
		&entities.TextLine{BaseObject: entities.BaseObject{ID: "title", BBox: entities.BoundingBox{X: 50, Y: 730, Width: 500, Height: 30, Page: 0}}},
		&entities.TextLine{BaseObject: entities.BaseObject{ID: "l1", BBox: entities.BoundingBox{X: 50, Y: 640, Width: 220, Height: 20, Page: 0}}},
		&entities.TextLine{BaseObject: entities.BaseObject{ID: "l2", BBox: entities.BoundingBox{X: 50, Y: 590, Width: 220, Height: 20, Page: 0}}},
		&entities.TextLine{BaseObject: entities.BaseObject{ID: "r1", BBox: entities.BoundingBox{X: 320, Y: 640, Width: 220, Height: 20, Page: 0}}},
		&entities.TextLine{BaseObject: entities.BaseObject{ID: "r2", BBox: entities.BoundingBox{X: 320, Y: 590, Width: 220, Height: 20, Page: 0}}},
	}

	sorted := sorter.Sort(elements, 600, 800)
	var ids []string
	for _, element := range sorted {
		ids = append(ids, element.GetID())
	}

	assert.Equal(t, []string{"title", "l1", "l2", "r1", "r2"}, ids)
}

func TestXYCutPlusPlusSorterRecursivelySortsMultipleColumnsLeftToRight(t *testing.T) {
	sorter := readingorder.XYCutPlusPlusSorter{}
	elements := []entities.IObject{
		&entities.TextLine{BaseObject: entities.BaseObject{ID: "c1l1", BBox: entities.BoundingBox{X: 40, Y: 640, Width: 140, Height: 20, Page: 0}}},
		&entities.TextLine{BaseObject: entities.BaseObject{ID: "c1l2", BBox: entities.BoundingBox{X: 40, Y: 600, Width: 140, Height: 20, Page: 0}}},
		&entities.TextLine{BaseObject: entities.BaseObject{ID: "c2l1", BBox: entities.BoundingBox{X: 250, Y: 640, Width: 140, Height: 20, Page: 0}}},
		&entities.TextLine{BaseObject: entities.BaseObject{ID: "c2l2", BBox: entities.BoundingBox{X: 250, Y: 600, Width: 140, Height: 20, Page: 0}}},
		&entities.TextLine{BaseObject: entities.BaseObject{ID: "c3l1", BBox: entities.BoundingBox{X: 460, Y: 640, Width: 140, Height: 20, Page: 0}}},
		&entities.TextLine{BaseObject: entities.BaseObject{ID: "c3l2", BBox: entities.BoundingBox{X: 460, Y: 600, Width: 140, Height: 20, Page: 0}}},
	}

	sorted := sorter.Sort(elements, 640, 800)
	var ids []string
	for _, element := range sorted {
		ids = append(ids, element.GetID())
	}

	assert.Equal(t, []string{"c1l1", "c1l2", "c2l1", "c2l2", "c3l1", "c3l2"}, ids)
}
