// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package entities

const (
	AlignLeft    = "left"
	AlignRight   = "right"
	AlignCenter  = "center"
	AlignJustify = "justify"
)

type SemanticParagraph struct {
	BaseObject
	Lines     []*TextLine
	Alignment string
}

func (p *SemanticParagraph) GetObjectType() ObjectType { return ObjectTypeParagraph }
