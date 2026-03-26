package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestExportOptionsJSONMatchesExpectedSchema(t *testing.T) {
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()
	os.Stdout = w

	exportErr := ExportOptionsJSON()

	_ = w.Close()
	os.Stdout = originalStdout
	if exportErr != nil {
		t.Fatalf("ExportOptionsJSON returned error: %v", exportErr)
	}

	payload, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read export payload: %v", err)
	}
	if len(bytes.TrimSpace(payload)) == 0 {
		t.Fatal("expected non-empty JSON output")
	}

	var got struct {
		Options []OptionDefinition `json:"options"`
	}
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, string(payload))
	}
	if len(got.Options) != 25 {
		t.Fatalf("expected 25 exported options, got %d", len(got.Options))
	}
	if got.Options[0].Name != "output-dir" {
		t.Fatalf("expected first option output-dir, got %q", got.Options[0].Name)
	}
	if got.Options[2].Name != "format" || got.Options[2].Type != "[]string" || !got.Options[2].Multiple {
		t.Fatalf("expected format option with []string type and multiple=true, got %+v", got.Options[2])
	}
	if len(got.Options[2].Choices) != 7 {
		t.Fatalf("expected 7 format choices, got %+v", got.Options[2].Choices)
	}
	if got.Options[4].Name != "content-safety-off" || got.Options[4].Type != "[]string" || !got.Options[4].Multiple {
		t.Fatalf("expected content-safety-off option with []string type and multiple=true, got %+v", got.Options[4])
	}
	last := got.Options[len(got.Options)-1]
	if last.Name != "hybrid-fallback" || last.Type != "boolean" {
		t.Fatalf("unexpected last option: %+v", last)
	}
	if got.Options[23].Name != "hybrid-timeout" || got.Options[23].Type != "integer" {
		t.Fatalf("expected hybrid-timeout integer option, got %+v", got.Options[23])
	}
}
