// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0

package wcag

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
