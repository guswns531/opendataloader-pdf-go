package nativepdf

import (
	"strings"
	"testing"
)

func TestExtractContentTextByOperatorsHandlesTextOperators(t *testing.T) {
	stream := []byte("BT /F1 12 Tf (Hello) Tj 0 -14 Td (World) Tj ET")

	got := extractContentTextByOperators(stream, nil)
	want := []string{"Hello", "World"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("extractContentTextByOperators() = %v, want %v", got, want)
	}
}

func TestExtractContentTextByOperatorsHandlesTJSpacingAndHex(t *testing.T) {
	stream := []byte("BT [<0048><0069>-250<0021>] TJ ET")

	got := extractContentTextByOperators(stream, nil)
	want := []string{"Hi !"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("extractContentTextByOperators() = %v, want %v", got, want)
	}
}

func TestExtractContentTextByOperatorsHandlesQuoteOperator(t *testing.T) {
	stream := []byte("BT (Line1) Tj (Line2) ' ET")

	got := extractContentTextByOperators(stream, nil)
	want := []string{"Line1", "Line2"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("extractContentTextByOperators() = %v, want %v", got, want)
	}
}
