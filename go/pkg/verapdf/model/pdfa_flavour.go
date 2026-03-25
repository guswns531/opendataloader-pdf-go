// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0

package model

import "strings"

type PDFAFlavour int

const (
	PDFAFlavourNone PDFAFlavour = iota
	PDFA_1_A
	PDFA_1_B
	PDFA_2_A
	PDFA_2_B
	PDFA_2_U
	PDFA_3_A
	PDFA_3_B
	PDFA_3_U
	PDFA_4
	PDFA_4_E
	PDFA_4_F
)

func (f PDFAFlavour) String() string {
	switch f {
	case PDFA_1_A:
		return "PDF/A-1A"
	case PDFA_1_B:
		return "PDF/A-1B"
	case PDFA_2_A:
		return "PDF/A-2A"
	case PDFA_2_B:
		return "PDF/A-2B"
	case PDFA_2_U:
		return "PDF/A-2U"
	case PDFA_3_A:
		return "PDF/A-3A"
	case PDFA_3_B:
		return "PDF/A-3B"
	case PDFA_3_U:
		return "PDF/A-3U"
	case PDFA_4:
		return "PDF/A-4"
	case PDFA_4_E:
		return "PDF/A-4E"
	case PDFA_4_F:
		return "PDF/A-4F"
	default:
		return "NONE"
	}
}

func ParsePDFAFlavour(s string) PDFAFlavour {
	normalized := strings.ToUpper(strings.TrimSpace(s))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, "/", "-")

	switch normalized {
	case "", "NONE":
		return PDFAFlavourNone
	case "PDF-A-1A", "PDFA-1A", "1A":
		return PDFA_1_A
	case "PDF-A-1B", "PDFA-1B", "1B":
		return PDFA_1_B
	case "PDF-A-2A", "PDFA-2A", "2A":
		return PDFA_2_A
	case "PDF-A-2B", "PDFA-2B", "2B":
		return PDFA_2_B
	case "PDF-A-2U", "PDFA-2U", "2U":
		return PDFA_2_U
	case "PDF-A-3A", "PDFA-3A", "3A":
		return PDFA_3_A
	case "PDF-A-3B", "PDFA-3B", "3B":
		return PDFA_3_B
	case "PDF-A-3U", "PDFA-3U", "3U":
		return PDFA_3_U
	case "PDF-A-4", "PDFA-4", "4":
		return PDFA_4
	case "PDF-A-4E", "PDFA-4E", "4E":
		return PDFA_4_E
	case "PDF-A-4F", "PDFA-4F", "4F":
		return PDFA_4_F
	default:
		return PDFAFlavourNone
	}
}
