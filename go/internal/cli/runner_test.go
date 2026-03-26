package cli

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
)

func TestProcessPathReportsMissingFile(t *testing.T) {
	err := processPath("/definitely/missing/file.pdf", api.DefaultConfig())
	if err == nil || !strings.Contains(err.Error(), "file not found: /definitely/missing/file.pdf") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessPathRejectsNonPDFFile(t *testing.T) {
	t.Helper()
	path := t.TempDir() + "/note.txt"
	if err := os.WriteFile(path, []byte("not a pdf"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	err := processPath(path, api.DefaultConfig())
	if err == nil || !strings.Contains(err.Error(), "not a PDF file: "+path) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToConfigAcceptsLegacyFormatFlags(t *testing.T) {
	opts := &CLIOptions{
		MarkdownReport:     true,
		HTMLReport:         true,
		MarkdownWithImages: true,
		NoJSONReport:       true,
	}

	cfg, err := opts.toConfig()
	if err != nil {
		t.Fatalf("toConfig returned error: %v", err)
	}

	if slices.Contains(cfg.Formats, api.FormatJSON) {
		t.Fatalf("expected json format to be removed, got %#v", cfg.Formats)
	}
	for _, expected := range []string{api.FormatMarkdown, api.FormatHTML, api.FormatMarkdownWithImages} {
		if !slices.Contains(cfg.Formats, expected) {
			t.Fatalf("expected format %q in %#v", expected, cfg.Formats)
		}
	}
}

func TestToConfigRejectsUnsupportedFormat(t *testing.T) {
	opts := &CLIOptions{Format: []string{"json", "docx"}}
	_, err := opts.toConfig()
	if err == nil || !strings.Contains(err.Error(), `unsupported format "docx"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToConfigRejectsUnsupportedContentSafetyFlag(t *testing.T) {
	opts := &CLIOptions{ContentSafetyOff: []string{"all", "bogus"}}
	_, err := opts.toConfig()
	if err == nil || !strings.Contains(err.Error(), `unsupported value "bogus"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunReturnsAggregateFailure(t *testing.T) {
	err := Run(&CLIOptions{}, []string{"/definitely/missing/file.pdf"})
	if err == nil || !strings.Contains(err.Error(), "one or more files failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
