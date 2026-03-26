// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package serializers

import "github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"

func SerializeLineChunk(line *entities.LineArtChunk) map[string]interface{} {
	if line == nil {
		return nil
	}
	return essentialInfo(line, "line")
}
