// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0

package model

type ValidationError struct {
	RuleID      string
	Description string
	Location    string
	Object      string
}

type ValidationResult struct {
	Flavour     PDFAFlavour
	IsCompliant bool
	TotalRules  int
	PassedRules int
	FailedRules int
	Errors      []ValidationError
}
