// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0

package wcag

import "github.com/opendataloader-project/opendataloader-pdf-go/pkg/verapdf/pdfbox"

type WCAGCriterion string

const (
	WCAG_1_1_1 WCAGCriterion = "1.1.1"
	WCAG_1_3_1 WCAGCriterion = "1.3.1"
	WCAG_1_4_3 WCAGCriterion = "1.4.3"
	WCAG_2_4_6 WCAGCriterion = "2.4.6"
	WCAG_3_1_1 WCAGCriterion = "3.1.1"
)

type WCAGViolation struct {
	Criterion   WCAGCriterion
	Description string
	Element     string
}

type WCAGResult struct {
	Violations   []WCAGViolation
	IsAccessible bool
}

func Evaluate(features *pdfbox.PDFFeatures) WCAGResult {
	result := WCAGResult{}
	if features == nil {
		result.Violations = append(result.Violations, WCAGViolation{
			Criterion:   WCAG_1_3_1,
			Description: "Document features are required for accessibility evaluation",
			Element:     "document",
		})
		result.IsAccessible = false
		return result
	}

	if len(features.Images) > 0 && !features.HasStructureTree {
		result.Violations = append(result.Violations, WCAGViolation{
			Criterion:   WCAG_1_1_1,
			Description: "Images require tagged structure to carry accessible alternatives",
			Element:     "image",
		})
	}
	if !features.HasStructureTree {
		result.Violations = append(result.Violations, WCAGViolation{
			Criterion:   WCAG_1_3_1,
			Description: "Document structure tree is required to preserve reading relationships",
			Element:     "document",
		})
	}
	if !features.HasDocumentTitle {
		result.Violations = append(result.Violations, WCAGViolation{
			Criterion:   WCAG_2_4_6,
			Description: "Document title is missing",
			Element:     "document",
		})
	}
	if !features.HasLanguage {
		result.Violations = append(result.Violations, WCAGViolation{
			Criterion:   WCAG_3_1_1,
			Description: "Primary document language is missing",
			Element:     "document",
		})
	}

	result.IsAccessible = len(result.Violations) == 0
	return result
}
