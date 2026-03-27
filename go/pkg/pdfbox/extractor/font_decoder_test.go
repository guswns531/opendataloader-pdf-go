package extractor

import "testing"

func TestMacRomanTableCoversAllBytes(t *testing.T) {
	table := macRomanTable()

	if len(table) != 256 {
		t.Fatalf("expected 256 entries, got %d", len(table))
	}
	if table[0] != '\x00' {
		t.Fatalf("expected entry 0 to remain NUL, got %q", table[0])
	}
	if table[255] == 0 {
		t.Fatal("expected final MacRoman entry to be populated")
	}
}
