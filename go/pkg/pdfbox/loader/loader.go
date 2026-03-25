// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// This package provides functionality equivalent to Apache PDFBox 3.0.4
// (https://pdfbox.apache.org/), implemented using pdfcpu.

package loader

import "github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/model"

func Open(path, password string) (*model.PDDocument, error) {
	return model.Open(path, password)
}
