// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package entities

type SemanticImage struct {
	BaseObject
	Alt          string
	Data         []byte
	ExternalPath string
	Width        float64
	Height       float64
}

func (i *SemanticImage) GetObjectType() ObjectType { return ObjectTypeImage }
