// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0

package verapdf_test

import (
	"path/filepath"
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/verapdf"
	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/verapdf/model"
	verapdfpdfbox "github.com/opendataloader-project/opendataloader-pdf-go/pkg/verapdf/pdfbox"
	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/verapdf/wcag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPDFAFlavourString(t *testing.T) {
	assert.Equal(t, "NONE", model.PDFAFlavourNone.String())
	assert.Equal(t, "PDF/A-1B", model.PDFA_1_B.String())
	assert.Equal(t, "PDF/A-2A", model.PDFA_2_A.String())
	assert.Equal(t, "PDF/A-4F", model.PDFA_4_F.String())
}

func TestParsePDFAFlavour(t *testing.T) {
	assert.Equal(t, model.PDFAFlavourNone, model.ParsePDFAFlavour(""))
	assert.Equal(t, model.PDFA_1_B, model.ParsePDFAFlavour("pdf/a-1b"))
	assert.Equal(t, model.PDFA_2_A, model.ParsePDFAFlavour("PDFA-2A"))
	assert.Equal(t, model.PDFA_3_U, model.ParsePDFAFlavour("3u"))
	assert.Equal(t, model.PDFA_4_E, model.ParsePDFAFlavour("pdf_a_4e"))
	assert.Equal(t, model.PDFAFlavourNone, model.ParsePDFAFlavour("unknown"))
}

func TestValidationResultCreation(t *testing.T) {
	result := model.ValidationResult{
		Flavour:     model.PDFA_2_B,
		IsCompliant: false,
		TotalRules:  2,
		PassedRules: 1,
		FailedRules: 1,
		Errors: []model.ValidationError{
			{
				RuleID:      "PDFA_METADATA_PRESENT",
				Description: "XMP metadata is required for PDF/A compliance",
				Location:    "Catalog",
				Object:      "document",
			},
		},
	}

	assert.Equal(t, model.PDFA_2_B, result.Flavour)
	assert.False(t, result.IsCompliant)
	assert.Equal(t, 2, result.TotalRules)
	assert.Equal(t, 1, result.PassedRules)
	assert.Equal(t, 1, result.FailedRules)
	if assert.Len(t, result.Errors, 1) {
		assert.Equal(t, "PDFA_METADATA_PRESENT", result.Errors[0].RuleID)
		assert.Equal(t, "Catalog", result.Errors[0].Location)
	}
}

func TestPDFACheckerRejectsMissingRequirements(t *testing.T) {
	checker := verapdfpdfbox.PDFAChecker{Flavour: model.PDFA_2_A}

	result := checker.Check(&verapdfpdfbox.PDFFeatures{
		HasMetadata:      false,
		HasStructureTree: false,
		HasDocumentTitle: false,
		HasLanguage:      false,
		FontsEmbedded:    false,
		ColorSpaces:      []string{"DeviceRGB"},
		HasOutputIntent:  false,
	})

	assert.False(t, result.IsCompliant)
	assert.Equal(t, 6, result.TotalRules)
	assert.Equal(t, 6, result.FailedRules)
	assert.Len(t, result.Errors, 6)
	assert.Equal(t, "PDFA_METADATA_PRESENT", result.Errors[0].RuleID)
	assert.Equal(t, "Catalog.Metadata", result.Errors[0].Location)
}

func TestPDFACheckerAcceptsMinimalTaggedDocument(t *testing.T) {
	checker := verapdfpdfbox.PDFAChecker{Flavour: model.PDFA_3_A}

	result := checker.Check(&verapdfpdfbox.PDFFeatures{
		HasMetadata:      true,
		HasStructureTree: true,
		HasDocumentTitle: true,
		HasLanguage:      true,
		FontsEmbedded:    true,
		ColorSpaces:      []string{"ICCBased"},
	})

	assert.True(t, result.IsCompliant)
	assert.Equal(t, 6, result.TotalRules)
	assert.Equal(t, 6, result.PassedRules)
	assert.Empty(t, result.Errors)
}

func TestWCAGEvaluateDetectsViolations(t *testing.T) {
	result := wcag.Evaluate(&verapdfpdfbox.PDFFeatures{
		Images:           []verapdfpdfbox.ImageFeature{{Width: 100, Height: 100}},
		HasStructureTree: false,
		HasDocumentTitle: false,
		HasLanguage:      false,
	})

	assert.False(t, result.IsAccessible)
	assert.Len(t, result.Violations, 4)
	assert.Equal(t, wcag.WCAG_1_1_1, result.Violations[0].Criterion)
}

func TestPDFACheckerRejectsNilFeatures(t *testing.T) {
	checker := verapdfpdfbox.PDFAChecker{Flavour: model.PDFA_1_B}

	result := checker.Check(nil)

	assert.False(t, result.IsCompliant)
	assert.Equal(t, 1, result.TotalRules)
	assert.Equal(t, 1, result.FailedRules)
	if assert.Len(t, result.Errors, 1) {
		assert.Equal(t, "GENERAL_FEATURES_PRESENT", result.Errors[0].RuleID)
		assert.Equal(t, "input", result.Errors[0].Location)
	}
}

func TestPDFACheckerSkipsValidationWhenFlavourIsNone(t *testing.T) {
	checker := verapdfpdfbox.PDFAChecker{Flavour: model.PDFAFlavourNone}

	result := checker.Check(nil)

	assert.True(t, result.IsCompliant)
	assert.Equal(t, 0, result.TotalRules)
	assert.Equal(t, 0, result.PassedRules)
	assert.Equal(t, 0, result.FailedRules)
	assert.Empty(t, result.Errors)
}

func TestFeatureExtractorExtractsRealSample(t *testing.T) {
	samplePath := filepath.Join("..", "..", "..", "..", "samples", "pdf", "1901.03003.pdf")

	features, err := (&verapdfpdfbox.FeatureExtractor{}).Extract(samplePath)

	require.NoError(t, err)
	require.NotNil(t, features)
	assert.NotNil(t, features.ColorSpaces)
	assert.NotNil(t, features.Annotations)
	assert.NotNil(t, features.Images)
	assert.Greater(t, len(features.Annotations), 0)
}

func TestValidateFacadeRunsEndToEnd(t *testing.T) {
	samplePath := filepath.Join("..", "..", "..", "..", "samples", "pdf", "1901.03003.pdf")

	report, err := verapdf.Validate(samplePath, model.PDFA_1_B)

	require.NoError(t, err)
	require.NotNil(t, report)
	require.NotNil(t, report.Features)
	require.NotNil(t, report.PDFA)
	assert.Equal(t, model.PDFA_1_B, report.PDFA.Flavour)
}
