// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0

package verapdf

import (
	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/verapdf/model"
	verapdfpdfbox "github.com/opendataloader-project/opendataloader-pdf-go/pkg/verapdf/pdfbox"
	"github.com/opendataloader-project/opendataloader-pdf-go/pkg/verapdf/wcag"
)

type Report struct {
	Features *verapdfpdfbox.PDFFeatures
	PDFA     *model.ValidationResult
	WCAG     wcag.WCAGResult
}

func Validate(pdfPath string, flavour model.PDFAFlavour) (*Report, error) {
	features, err := (&verapdfpdfbox.FeatureExtractor{}).Extract(pdfPath)
	if err != nil {
		return nil, err
	}

	return &Report{
		Features: features,
		PDFA:     (&verapdfpdfbox.PDFAChecker{Flavour: flavour}).Check(features),
		WCAG:     wcag.Evaluate(features),
	}, nil
}

func ValidatePDFA(pdfPath string, flavour model.PDFAFlavour) (*model.ValidationResult, error) {
	report, err := Validate(pdfPath, flavour)
	if err != nil {
		return nil, err
	}
	return report.PDFA, nil
}

func ValidateWCAG(pdfPath string) (wcag.WCAGResult, error) {
	report, err := Validate(pdfPath, model.PDFAFlavourNone)
	if err != nil {
		return wcag.WCAGResult{}, err
	}
	return report.WCAG, nil
}
