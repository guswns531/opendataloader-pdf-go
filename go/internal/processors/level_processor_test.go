package processors

import (
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

func TestLevelProcessorPreservesSizeBasedOrdering(t *testing.T) {
	processor := &LevelProcessor{}
	headings := []*entities.SemanticHeading{
		{FontSize: 20, FontFamily: "Helvetica"},
		{FontSize: 16, FontFamily: "Helvetica"},
		{FontSize: 14, FontFamily: "Helvetica"},
	}

	processor.Process(headings)

	if headings[0].Level != 1 || headings[1].Level != 2 || headings[2].Level != 3 {
		t.Fatalf("unexpected levels: got %d, %d, %d", headings[0].Level, headings[1].Level, headings[2].Level)
	}
}

func TestLevelProcessorSeparatesSameSizeByWeightAndItalic(t *testing.T) {
	processor := &LevelProcessor{}
	headings := []*entities.SemanticHeading{
		{FontSize: 18, IsBold: true, FontFamily: "Helvetica"},
		{FontSize: 18, IsItalic: true, FontFamily: "Helvetica"},
		{FontSize: 18, FontFamily: "Helvetica"},
	}

	processor.Process(headings)

	if headings[0].Level != 1 {
		t.Fatalf("expected bold heading to rank first, got level %d", headings[0].Level)
	}
	if headings[1].Level != 2 {
		t.Fatalf("expected italic heading to rank second, got level %d", headings[1].Level)
	}
	if headings[2].Level != 3 {
		t.Fatalf("expected regular heading to rank third, got level %d", headings[2].Level)
	}
}

func TestLevelProcessorSeparatesSameSizeByFontFamily(t *testing.T) {
	processor := &LevelProcessor{}
	headings := []*entities.SemanticHeading{
		{FontSize: 18, IsBold: true, FontFamily: "Helvetica"},
		{FontSize: 18, IsBold: true, FontFamily: "Times New Roman"},
	}

	processor.Process(headings)

	if headings[0].Level == headings[1].Level {
		t.Fatalf("expected distinct font families to map to distinct levels, got %d and %d", headings[0].Level, headings[1].Level)
	}
}

func TestNormalizeFontFamily(t *testing.T) {
	tests := map[string]string{
		"ABCDEE+Helvetica-Bold":    "Helvetica",
		"TimesNewRomanPS-ItalicMT": "TimesNewRomanPS ItalicMT",
		"NotoSansCJKkr-Regular":    "NotoSansCJKkr",
		"SourceSerifPro Semibold":  "SourceSerifPro",
		"MinionPro":                "MinionPro",
	}

	for input, want := range tests {
		if got := normalizeFontFamily(input); got != want {
			t.Fatalf("normalizeFontFamily(%q) = %q, want %q", input, got, want)
		}
	}
}
