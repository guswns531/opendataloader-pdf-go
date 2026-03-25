// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package entities

// BoundingBox - PDF 좌표계 (좌하단 원점, Y축 위 방향)
type BoundingBox struct {
	X, Y, Width, Height float64
	Page                int
}

type ObjectType string

const (
	ObjectTypeParagraph    ObjectType = "paragraph"
	ObjectTypeHeading      ObjectType = "heading"
	ObjectTypeTable        ObjectType = "table"
	ObjectTypeList         ObjectType = "list"
	ObjectTypeImage        ObjectType = "image"
	ObjectTypeFormula      ObjectType = "formula"
	ObjectTypeCaption      ObjectType = "caption"
	ObjectTypeHeaderFooter ObjectType = "header_footer"
	ObjectTypeTextLine     ObjectType = "text_line"
	ObjectTypeTextChunk    ObjectType = "text_chunk"
)

type IObject interface {
	GetID() string
	GetBBox() BoundingBox
	GetObjectType() ObjectType
}

type BaseObject struct {
	ID   string
	BBox BoundingBox
}

func (b *BaseObject) GetID() string        { return b.ID }
func (b *BaseObject) GetBBox() BoundingBox { return b.BBox }
