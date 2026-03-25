// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0

package pdfbox

import (
	"fmt"
	"os"
)

type AnnotationFeature struct {
	Type    string
	SubType string
	BBox    [4]float64
}

type ImageFeature struct {
	Width            int
	Height           int
	ColorSpace       string
	BitsPerComponent int
}

type PDFFeatures struct {
	HasStructureTree bool
	HasMetadata      bool
	HasDocumentTitle bool
	HasLanguage      bool
	FontsEmbedded    bool
	ColorSpaces      []string
	Annotations      []AnnotationFeature
	Images           []ImageFeature
}

type FeatureExtractor struct{}

func (e *FeatureExtractor) Extract(pdfPath string) (*PDFFeatures, error) {
	if pdfPath == "" {
		return nil, fmt.Errorf("pdf path is empty")
	}

	info, err := os.Stat(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("stat pdf: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("pdf path points to a directory: %s", pdfPath)
	}

	// Phase 5 will replace this with pdfcpu-backed extraction.
	return &PDFFeatures{
		ColorSpaces: []string{},
		Annotations: []AnnotationFeature{},
		Images:      []ImageFeature{},
	}, nil
}
