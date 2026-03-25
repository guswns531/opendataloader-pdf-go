// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package entities

type ListItem struct {
	BaseObject
	Content    []IObject
	BulletText string
	IsOrdered  bool
	Level      int
}

type PDFList struct {
	BaseObject
	Items     []*ListItem
	IsOrdered bool
}

func (l *PDFList) GetObjectType() ObjectType  { return ObjectTypeList }
func (i *ListItem) GetObjectType() ObjectType { return ObjectTypeListItem }
