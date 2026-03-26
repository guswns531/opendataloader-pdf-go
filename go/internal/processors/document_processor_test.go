package processors

import "testing"

func TestParsePageRangeExpandsAndSorts(t *testing.T) {
	got, err := parsePageRange("3,1,5-7,6", 10)
	if err != nil {
		t.Fatalf("parsePageRange returned error: %v", err)
	}
	want := []int{1, 3, 5, 6, 7}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestParsePageRangeRejectsOutOfBoundsPages(t *testing.T) {
	_, err := parsePageRange("1,11", 10)
	if err == nil {
		t.Fatal("expected out-of-bounds error")
	}
}

func TestParsePageRangeRejectsOutOfBoundsRanges(t *testing.T) {
	_, err := parsePageRange("2-12", 10)
	if err == nil {
		t.Fatal("expected out-of-bounds range error")
	}
}
