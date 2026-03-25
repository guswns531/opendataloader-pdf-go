// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0

package pdfbox

import "github.com/opendataloader-project/opendataloader-pdf-go/pkg/verapdf/model"

type PDFAChecker struct {
	Flavour model.PDFAFlavour
}

func (c *PDFAChecker) Check(features *PDFFeatures) *model.ValidationResult {
	result := &model.ValidationResult{
		Flavour: c.Flavour,
	}

	if features == nil {
		result.Errors = append(result.Errors, model.ValidationError{
			RuleID:      "GENERAL_FEATURES_PRESENT",
			Description: "PDF features are required for validation",
			Object:      "document",
		})
		finalizeValidationResult(result)
		return result
	}

	addMetadataRule(result, features)

	switch c.Flavour {
	case model.PDFA_1_B:
		addRuleResult(result, "PDFA_1_B_FONTS_EMBEDDED", features.FontsEmbedded,
			"All fonts must be embedded for PDF/A-1b", "document")
	case model.PDFA_2_A, model.PDFA_3_A:
		addRuleResult(result, "PDFA_A_STRUCTURE_TREE", features.HasStructureTree,
			"Structure tree is required for tagged PDF/A level A compliance", "document")
	}

	finalizeValidationResult(result)
	return result
}

func addMetadataRule(result *model.ValidationResult, features *PDFFeatures) {
	addRuleResult(result, "PDFA_METADATA_PRESENT", features.HasMetadata,
		"XMP metadata is required for PDF/A compliance", "document")
}

func addRuleResult(result *model.ValidationResult, ruleID string, passed bool, description string, object string) {
	result.TotalRules++
	if passed {
		result.PassedRules++
		return
	}

	result.FailedRules++
	result.Errors = append(result.Errors, model.ValidationError{
		RuleID:      ruleID,
		Description: description,
		Object:      object,
	})
}

func finalizeValidationResult(result *model.ValidationResult) {
	result.IsCompliant = result.FailedRules == 0
}
