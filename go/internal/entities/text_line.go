// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package entities

import "strings"

type TextLine struct {
	BaseObject
	Chunks           []*TextChunk
	LineArtBullet    *LineArtChunk
	Baseline         float64
	IsFirstLine      bool
	IsLastLine       bool
	HasStrikethrough bool
}

func (l *TextLine) GetObjectType() ObjectType { return ObjectTypeTextLine }

func (l *TextLine) GetText() string {
	var parts []string
	for _, c := range l.Chunks {
		parts = append(parts, c.Text)
	}

	return strings.Join(parts, "")
}
