package api

import "testing"

func TestDefaultConfigUsesJSONOutputByDefault(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}
	if len(cfg.Formats) != 1 || cfg.Formats[0] != FormatJSON {
		t.Fatalf("expected default format %q, got %#v", FormatJSON, cfg.Formats)
	}
	if cfg.Hybrid != HybridOff {
		t.Fatalf("expected default hybrid %q, got %q", HybridOff, cfg.Hybrid)
	}
	if cfg.ReadingOrder != ReadingOrderXYCut {
		t.Fatalf("expected default reading order %q, got %q", ReadingOrderXYCut, cfg.ReadingOrder)
	}
}
