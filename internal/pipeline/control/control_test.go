package control

import (
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
)

func TestResolveDefaultsToCurrentRuntimeBaseline(t *testing.T) {
	got := Resolve(core.ProcessingOptions{})

	if !got.ReadingOrderEnabled {
		t.Fatal("reading order should be enabled by default")
	}
	if got.ReadingOrderMode != "xycut" {
		t.Fatalf("ReadingOrderMode = %q, want %q", got.ReadingOrderMode, "xycut")
	}
	if got.HeaderFooterIncluded {
		t.Fatal("header/footer should be excluded by default")
	}
	if got.TableMethod != "default" {
		t.Fatalf("TableMethod = %q, want %q", got.TableMethod, "default")
	}
	if !got.TableHeuristicsEnabled {
		t.Fatal("table heuristics should be enabled by default")
	}
	if got.SanitizeEnabled {
		t.Fatal("sanitize should be disabled by default")
	}
	if got.ReplaceInvalidChars != " " {
		t.Fatalf("ReplaceInvalidChars = %q, want %q", got.ReplaceInvalidChars, " ")
	}
	if got.ContentSafetyOff != "" {
		t.Fatalf("ContentSafetyOff = %q, want empty string", got.ContentSafetyOff)
	}
	if !got.LayoutFilteringEnabled {
		t.Fatal("layout filtering should be enabled by default")
	}
	if got.KeepLineBreaks || got.UseStructTree || got.DetectStrikethrough {
		t.Fatalf("unexpected extras-derived flags: %+v", got)
	}
	if got.PreservePageBreaks || got.EmitGeometry || got.EmitDiagnostics || got.Strict {
		t.Fatalf("unexpected direct flags: %+v", got)
	}
}

func TestResolveReadsExtrasAndDirectFlags(t *testing.T) {
	options := core.ProcessingOptions{
		EmitGeometry:       true,
		EmitDiagnostics:    true,
		PreservePageBreaks: true,
		Strict:             true,
		EnabledHeuristics:  []string{"xycut"},
		DisabledHeuristics: []string{"table"},
		Extras: map[string]any{
			"reading_order":         "off",
			"include_header_footer": true,
			"table_method":          "cluster",
			"sanitize":              true,
			"replace_invalid":       "?",
			"content_safety_off":    "off-page",
			"keep_line_breaks":      true,
			"use_struct_tree":       true,
			"detect_strikethrough":  true,
		},
	}

	got := Resolve(options)

	if got.ReadingOrderEnabled {
		t.Fatal("reading order should be disabled when extras request off")
	}
	if got.ReadingOrderMode != "off" {
		t.Fatalf("ReadingOrderMode = %q, want %q", got.ReadingOrderMode, "off")
	}
	if !got.HeaderFooterIncluded {
		t.Fatal("header/footer should be included when requested")
	}
	if got.TableMethod != "cluster" {
		t.Fatalf("TableMethod = %q, want %q", got.TableMethod, "cluster")
	}
	if !got.TableHeuristicsEnabled {
		t.Fatal("table heuristics should be enabled for cluster mode")
	}
	if !got.SanitizeEnabled {
		t.Fatal("sanitize should be enabled")
	}
	if got.ReplaceInvalidChars != "?" {
		t.Fatalf("ReplaceInvalidChars = %q, want %q", got.ReplaceInvalidChars, "?")
	}
	if got.ContentSafetyOff != "off-page" {
		t.Fatalf("ContentSafetyOff = %q, want %q", got.ContentSafetyOff, "off-page")
	}
	if got.LayoutFilteringEnabled {
		t.Fatal("layout filtering should be disabled for off-page content-safety override")
	}
	if !got.KeepLineBreaks || !got.UseStructTree || !got.DetectStrikethrough {
		t.Fatalf("extras-derived flags were not set: %+v", got)
	}
	if !got.PreservePageBreaks || !got.EmitGeometry || !got.EmitDiagnostics || !got.Strict {
		t.Fatalf("direct flags were not preserved: %+v", got)
	}
}

func TestResolveFallsBackToHeuristicLists(t *testing.T) {
	options := core.ProcessingOptions{
		EnabledHeuristics:  []string{"reading-order", "cluster"},
		DisabledHeuristics: []string{"header-footer"},
	}

	got := Resolve(options)

	if !got.ReadingOrderEnabled {
		t.Fatal("reading order should be enabled from heuristic list")
	}
	if got.ReadingOrderMode != "xycut" {
		t.Fatalf("ReadingOrderMode = %q, want %q", got.ReadingOrderMode, "xycut")
	}
	if got.HeaderFooterIncluded {
		t.Fatal("header/footer should remain excluded when disabled")
	}
	if got.TableMethod != "cluster" {
		t.Fatalf("TableMethod = %q, want %q", got.TableMethod, "cluster")
	}
	if !got.TableHeuristicsEnabled {
		t.Fatal("table heuristics should be enabled from heuristic list")
	}
}
