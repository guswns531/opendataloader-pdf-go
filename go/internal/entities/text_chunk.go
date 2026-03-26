// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package entities

type FontStyle struct {
	FontName   string
	FontSize   float64
	FontWeight float64
	Bold       bool
	Italic     bool
	Color      [3]float64
}

type TextChunk struct {
	BaseObject
	Text            string
	FontStyle       FontStyle
	Baseline        float64
	CharSpacing     float64
	IsHidden        bool
	IsHiddenOCG     bool
	IsOffPage       bool
	IsTiny          bool
	IsStrikethrough bool
}

func (c *TextChunk) GetObjectType() ObjectType { return ObjectTypeTextChunk }

type LineArtChunk struct {
	BaseObject
	IsHorizontal bool
	IsVertical   bool
	LineWidth    float64
}

func (c *LineArtChunk) GetObjectType() ObjectType { return ObjectTypeLineArt }
