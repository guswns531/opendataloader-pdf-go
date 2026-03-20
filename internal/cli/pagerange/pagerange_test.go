package pagerange

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr string
	}{
		{
			name:  "comma separated pages and range",
			input: "1,3,5-7",
			want:  []int{1, 3, 5, 6, 7},
		},
		{
			name:  "whitespace around tokens",
			input: " 2 , 4-6 , 8 ",
			want:  []int{2, 4, 5, 6, 8},
		},
		{
			name:  "single page",
			input: "9",
			want:  []int{9},
		},
		{
			name:  "preserves duplicates and order",
			input: "1,2-3,3",
			want:  []int{1, 2, 3, 3},
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: "cannot be empty",
		},
		{
			name:    "invalid token",
			input:   "abc",
			wantErr: "invalid page range format",
		},
		{
			name:    "zero page",
			input:   "0",
			wantErr: "page numbers must be positive",
		},
		{
			name:    "negative page",
			input:   "-1",
			wantErr: "invalid page range format",
		},
		{
			name:    "reversed range",
			input:   "5-3",
			wantErr: "start page cannot be greater than end page",
		},
		{
			name:    "malformed range",
			input:   "1-2-3",
			wantErr: "invalid page range format",
		},
		{
			name:    "empty segment",
			input:   "1,,2",
			wantErr: "invalid page range format",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(tt.input)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Parse(%q) = nil error, want %q", tt.input, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Parse(%q) error = %q, want substring %q", tt.input, err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("Parse(%q) len = %d, want %d (%v)", tt.input, len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("Parse(%q)[%d] = %d, want %d (full=%v)", tt.input, i, got[i], tt.want[i], got)
				}
			}
		})
	}
}

func TestParseOptional(t *testing.T) {
	t.Parallel()

	got, err := ParseOptional("   ")
	if err != nil {
		t.Fatalf("ParseOptional returned error: %v", err)
	}
	if got != nil {
		t.Fatalf("ParseOptional blank input = %v, want nil", got)
	}

	got, err = ParseOptional("1-2")
	if err != nil {
		t.Fatalf("ParseOptional returned error: %v", err)
	}
	want := []int{1, 2}
	if len(got) != len(want) {
		t.Fatalf("ParseOptional len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParseOptional[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}
