package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
)

func TestParseFormatsFallsBackToJSON(t *testing.T) {
	formats := parseFormats("unknown")
	if len(formats) != 1 || formats[0] != core.OutputFormatJSON {
		t.Fatalf("formats = %v, want [json]", formats)
	}
}

func TestParseFormatsFiltersKnownValues(t *testing.T) {
	formats := parseFormats("markdown,json,unknown,html")
	want := []core.OutputFormat{
		core.OutputFormatMarkdown,
		core.OutputFormatJSON,
		core.OutputFormatHTML,
	}

	if len(formats) != len(want) {
		t.Fatalf("len(formats) = %d, want %d (%v)", len(formats), len(want), formats)
	}
	for i := range want {
		if formats[i] != want[i] {
			t.Fatalf("formats[%d] = %q, want %q", i, formats[i], want[i])
		}
	}
}

func TestRunFixtureMode(t *testing.T) {
	dir := t.TempDir()
	fixturePath := filepath.Join(dir, "fixture.json")
	const fixtureJSON = `{
	  "metadata": {"file_name": "fixture.json", "page_count": 1},
	  "pages": [
	    {
	      "metadata": {"number": 1, "index": 0},
	      "artifacts": [
	        {
	          "kind": "text",
	          "page_index": 0,
	          "page_number": 1,
	          "sequence": 0,
	          "bounds": {"left": 0, "bottom": 90, "right": 40, "top": 100},
	          "text": "Hello fixture"
	        }
	      ]
	    }
	  ]
	}`
	if err := os.WriteFile(fixturePath, []byte(fixtureJSON), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if got := run([]string{"--fixture", "--quiet", fixturePath}); got != 0 {
		t.Fatalf("run() = %d, want 0", got)
	}
}
