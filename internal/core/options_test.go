package core

import "testing"

func TestProcessingOptionsResolveDefaultsToJSON(t *testing.T) {
	cfg := (ProcessingOptions{}).Resolve()

	if len(cfg.RequestedFormats) != 1 || cfg.RequestedFormats[0] != OutputFormatJSON {
		t.Fatalf("requested formats = %v, want [json]", cfg.RequestedFormats)
	}
}

func TestProcessingOptionsResolveClonesMutableState(t *testing.T) {
	options := ProcessingOptions{
		RequestedFormats:  []OutputFormat{OutputFormatMarkdown},
		EnabledHeuristics: []string{"xycut"},
		Metadata:          map[string]string{"author": "hancom"},
		Extras:            map[string]any{"debug": true},
	}

	cfg := options.Resolve()

	options.RequestedFormats[0] = OutputFormatJSON
	options.EnabledHeuristics[0] = "other"
	options.Metadata["author"] = "changed"
	options.Extras["debug"] = false

	if cfg.RequestedFormats[0] != OutputFormatMarkdown {
		t.Fatalf("requested format mutated: %v", cfg.RequestedFormats)
	}
	if cfg.EnabledHeuristics[0] != "xycut" {
		t.Fatalf("enabled heuristics mutated: %v", cfg.EnabledHeuristics)
	}
	if cfg.Metadata["author"] != "hancom" {
		t.Fatalf("metadata mutated: %v", cfg.Metadata)
	}
	if cfg.Extras["debug"] != true {
		t.Fatalf("extras mutated: %v", cfg.Extras)
	}
}
