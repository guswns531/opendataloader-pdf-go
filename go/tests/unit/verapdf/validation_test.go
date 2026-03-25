// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0

package verapdf_test

import (
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/verapdf/model"
	"github.com/stretchr/testify/assert"
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
