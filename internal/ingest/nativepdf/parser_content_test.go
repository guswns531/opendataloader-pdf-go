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

func TestContentTokenizerReadsMarkedContentPropertyDict(t *testing.T) {
	tokenizer := newContentTokenizer([]byte("/P << /MCID 7 >> BDC"))

	tag, ok := tokenizer.next()
	if !ok {
		t.Fatal("first token missing")
	}
	if got, want := tag.text, "P"; got != want {
		t.Fatalf("tag = %q, want %q", got, want)
	}

	properties, ok := tokenizer.next()
	if !ok {
		t.Fatal("property dict token missing")
	}
	if got := markedContentIDFromToken(properties); got == nil || *got != 7 {
		t.Fatalf("markedContentIDFromToken() = %#v, want 7", got)
	}

	op, ok := tokenizer.next()
	if !ok {
		t.Fatal("operator token missing")
	}
	if got, want := op.text, "BDC"; got != want {
		t.Fatalf("operator = %q, want %q", got, want)
	}
}
