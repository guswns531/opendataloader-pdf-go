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

	if c.Flavour == model.PDFAFlavourNone {
		result.IsCompliant = true
		return result
	}

	if features == nil {
		addRuleResult(result, "GENERAL_FEATURES_PRESENT", false,
			"PDF features are required for validation", "input", "document")
		finalizeValidationResult(result)
		return result
	}

	addMetadataRule(result, features)
	addEmbeddedFontsRule(result, features)
	addColorSpaceRule(result, features)

	switch c.Flavour {
	case model.PDFA_1_A, model.PDFA_2_A, model.PDFA_3_A:
		addAccessibleStructureRules(result, features)
	case model.PDFA_1_B, model.PDFA_2_B, model.PDFA_2_U, model.PDFA_3_B, model.PDFA_3_U, model.PDFA_4, model.PDFA_4_E, model.PDFA_4_F:
	default:
	}

	finalizeValidationResult(result)
	return result
}

func addMetadataRule(result *model.ValidationResult, features *PDFFeatures) {
	addRuleResult(result, "PDFA_METADATA_PRESENT", features.HasMetadata,
		"XMP metadata is required for PDF/A compliance", "Catalog.Metadata", "document")
}

func addEmbeddedFontsRule(result *model.ValidationResult, features *PDFFeatures) {
	addRuleResult(result, "PDFA_FONTS_EMBEDDED", features.FontsEmbedded,
		"All fonts must be embedded for PDF/A compliance", "Resources.Font", "document")
}

func addColorSpaceRule(result *model.ValidationResult, features *PDFFeatures) {
	requiresOutputIntent := false
	for _, colorSpace := range features.ColorSpaces {
		switch colorSpace {
		case "DeviceRGB", "DeviceCMYK", "DeviceGray":
			requiresOutputIntent = true
		}
	}
	addRuleResult(result, "PDFA_OUTPUT_INTENT_OR_DEVICE_INDEPENDENT_COLOR", !requiresOutputIntent || features.HasOutputIntent,
		"Device color spaces require an output intent for PDF/A compliance", "Catalog.OutputIntents", "document")
}

func addAccessibleStructureRules(result *model.ValidationResult, features *PDFFeatures) {
	addRuleResult(result, "PDFA_A_STRUCTURE_TREE", features.HasStructureTree,
		"Structure tree is required for tagged PDF/A level A compliance", "Catalog.StructTreeRoot", "document")
	addRuleResult(result, "PDFA_A_DOCUMENT_TITLE", features.HasDocumentTitle,
		"Document title is required for tagged PDF/A level A compliance", "Info.Title", "document")
	addRuleResult(result, "PDFA_A_DOCUMENT_LANGUAGE", features.HasLanguage,
		"Document language is required for tagged PDF/A level A compliance", "Catalog.Lang", "document")
}

func addRuleResult(result *model.ValidationResult, ruleID string, passed bool, description string, location string, object string) {
	result.TotalRules++
	if passed {
		result.PassedRules++
		return
	}

	result.FailedRules++
	result.Errors = append(result.Errors, model.ValidationError{
		RuleID:      ruleID,
		Description: description,
		Location:    location,
		Object:      object,
	})
}

func finalizeValidationResult(result *model.ValidationResult) {
	result.IsCompliant = result.FailedRules == 0
}
