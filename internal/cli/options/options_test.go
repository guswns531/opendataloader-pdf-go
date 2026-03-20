package options

import (
	"reflect"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
)

func TestParseFormatsFallsBackToJSON(t *testing.T) {
	formats := ParseFormats("unknown")
	if len(formats) != 1 || formats[0] != core.OutputFormatJSON {
		t.Fatalf("formats = %v, want [json]", formats)
	}
}

func TestParseFormatsFiltersKnownValues(t *testing.T) {
	formats := ParseFormats("markdown,json,unknown,html")
	want := []core.OutputFormat{
		core.OutputFormatMarkdown,
		core.OutputFormatJSON,
		core.OutputFormatHTML,
	}

	if !reflect.DeepEqual(formats, want) {
		t.Fatalf("formats = %v, want %v", formats, want)
	}
}

func TestParseCapturesCurrentSkeletonFlags(t *testing.T) {
	opts, err := Parse([]string{
		"-o", "out",
		"-f", "markdown,json",
		"-q",
		"--fixture",
		"input.pdf",
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if opts.OutputDir != "out" {
		t.Fatalf("OutputDir = %q, want %q", opts.OutputDir, "out")
	}
	if opts.Format != "markdown,json" {
		t.Fatalf("Format = %q, want %q", opts.Format, "markdown,json")
	}
	if !opts.Quiet {
		t.Fatalf("Quiet = false, want true")
	}
	if !opts.Fixture {
		t.Fatalf("Fixture = false, want true")
	}
	if got, want := opts.Inputs, []string{"input.pdf"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Inputs = %v, want %v", got, want)
	}
}

func TestParseAcceptsParityPlaceholders(t *testing.T) {
	opts, err := Parse([]string{
		"--password", "secret",
		"--content-safety-off", "hidden-text",
		"--sanitize",
		"--keep-line-breaks",
		"--replace-invalid-chars", "?",
		"--use-struct-tree",
		"--table-method", "cluster",
		"--reading-order", "off",
		"--markdown-page-separator", "---",
		"--text-page-separator", "|page|",
		"--html-page-separator", "<hr>",
		"--image-output", "embedded",
		"--image-format", "jpeg",
		"--image-dir", "images",
		"--pages", "1,3-5",
		"--include-header-footer",
		"--detect-strikethrough",
		"--hybrid", "docling-fast",
		"--hybrid-mode", "full",
		"--hybrid-url", "http://localhost:8080",
		"--hybrid-timeout", "12345",
		"--hybrid-fallback",
		"input.json",
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if opts.Password != "secret" {
		t.Fatalf("Password = %q, want %q", opts.Password, "secret")
	}
	if opts.ContentSafetyOff != "hidden-text" {
		t.Fatalf("ContentSafetyOff = %q, want %q", opts.ContentSafetyOff, "hidden-text")
	}
	if !opts.Sanitize || !opts.KeepLineBreaks || !opts.UseStructTree || !opts.IncludeHeaderFooter || !opts.DetectStrikethrough || !opts.HybridFallback {
		t.Fatalf("boolean parity flags were not preserved: %+v", opts)
	}
	if opts.ReplaceInvalidChars != "?" || opts.TableMethod != "cluster" || opts.ReadingOrder != "off" {
		t.Fatalf("string parity flags were not preserved: %+v", opts)
	}
	if opts.MarkdownPageSeparator != "---" || opts.TextPageSeparator != "|page|" || opts.HTMLPageSeparator != "<hr>" {
		t.Fatalf("page separator flags were not preserved: %+v", opts)
	}
	if opts.ImageOutput != "embedded" || opts.ImageFormat != "jpeg" || opts.ImageDir != "images" {
		t.Fatalf("image flags were not preserved: %+v", opts)
	}
	if opts.Pages != "1,3-5" || opts.Hybrid != "docling-fast" || opts.HybridMode != "full" || opts.HybridURL != "http://localhost:8080" || opts.HybridTimeout != "12345" {
		t.Fatalf("hybrid/page flags were not preserved: %+v", opts)
	}
	if got, want := opts.Inputs, []string{"input.json"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Inputs = %v, want %v", got, want)
	}
}

func TestProcessingOptionsUsesNormalizedFormats(t *testing.T) {
	opts := Options{
		OutputDir: "out",
		Format:    "markdown,unknown,json",
	}

	cfg := opts.ProcessingOptions("/tmp/input.pdf")

	if cfg.InputPath != "/tmp/input.pdf" {
		t.Fatalf("InputPath = %q, want %q", cfg.InputPath, "/tmp/input.pdf")
	}
	if cfg.OutputPath != "out" {
		t.Fatalf("OutputPath = %q, want %q", cfg.OutputPath, "out")
	}
	if cfg.DocumentName != "input.pdf" {
		t.Fatalf("DocumentName = %q, want %q", cfg.DocumentName, "input.pdf")
	}
	wantFormats := []core.OutputFormat{core.OutputFormatMarkdown, core.OutputFormatJSON}
	if !reflect.DeepEqual(cfg.RequestedFormats, wantFormats) {
		t.Fatalf("RequestedFormats = %v, want %v", cfg.RequestedFormats, wantFormats)
	}
}
