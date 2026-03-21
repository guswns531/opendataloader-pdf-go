package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestRunWithoutInputsPrintsUsageAndSucceeds(t *testing.T) {
	if got := run(nil); got != 0 {
		t.Fatalf("run() = %d, want 0", got)
	}
}

func TestRunExportOptionsWritesJSONAndSucceeds(t *testing.T) {
	output := captureStdout(t, func() {
		if got := run([]string{"--export-options"}); got != 0 {
			t.Fatalf("run() = %d, want 0", got)
		}
	})
	if !strings.Contains(output, `"options"`) {
		t.Fatalf("stdout = %q, want exported options json", output)
	}
}

func TestRunFixtureWritesRequestedOutputs(t *testing.T) {
	dir := t.TempDir()
	fixturePath := filepath.Join(dir, "fixture_input.json")
	outputDir := filepath.Join(dir, "out")
	const fixtureJSON = `{
	  "metadata": {"file_name": "fixture_input.json", "page_count": 1},
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

	if got := run([]string{
		"--fixture",
		"--quiet",
		"--output-dir", outputDir,
		"--format", "json,markdown",
		fixturePath,
	}); got != 0 {
		t.Fatalf("run() = %d, want 0", got)
	}

	for _, name := range []string{"fixture_input.json", "fixture_input.md"} {
		path := filepath.Join(outputDir, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected output file %s: %v", path, err)
		}
	}
}

func TestRunNativeRawFixtureWritesJSONOutput(t *testing.T) {
	dir := t.TempDir()
	fixturePath := filepath.Join(dir, "fixture.raw.json")
	outputDir := filepath.Join(dir, "out")
	const fixtureJSON = `{
	  "metadata": {"file_name": "fixture.raw.json", "page_count": 1},
	  "pages": [
	    {
	      "metadata": {"number": 1, "index": 0, "width": 612, "height": 792},
	      "artifacts": [
	        {
	          "kind": "text",
	          "page_index": 0,
	          "page_number": 1,
	          "sequence": 0,
	          "bounds": {"left": 0, "bottom": 90, "right": 40, "top": 100},
	          "text": "Hello native fixture",
	          "style": {"font": "Times", "font_size": 12, "content": "Hello native fixture"}
	        }
	      ]
	    }
	  ]
	}`
	if err := os.WriteFile(fixturePath, []byte(fixtureJSON), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if got := run([]string{
		"--quiet",
		"--output-dir", outputDir,
		fixturePath,
	}); got != 0 {
		t.Fatalf("run() = %d, want 0", got)
	}

	outputPath := filepath.Join(outputDir, "fixture.raw.json")
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected output file %s: %v", outputPath, err)
	}
}

func TestRunLegacyMarkdownNoJSONWritesOnlyMarkdown(t *testing.T) {
	dir := t.TempDir()
	fixturePath := filepath.Join(dir, "legacy_fixture.json")
	outputDir := filepath.Join(dir, "out")
	const fixtureJSON = `{
	  "metadata": {"file_name": "legacy_fixture.json", "page_count": 1},
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
	          "text": "Legacy markdown"
	        }
	      ]
	    }
	  ]
	}`
	if err := os.WriteFile(fixturePath, []byte(fixtureJSON), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if got := run([]string{
		"--quiet",
		"--output-dir", outputDir,
		"--markdown",
		"--no-json",
		fixturePath,
	}); got != 0 {
		t.Fatalf("run() = %d, want 0", got)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "legacy_fixture.md")); err != nil {
		t.Fatalf("expected markdown output: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "legacy_fixture.json")); !os.IsNotExist(err) {
		t.Fatalf("expected json output to be absent, got err=%v", err)
	}
}

func TestRunRejectsUnsupportedLegacyPDFOutput(t *testing.T) {
	dir := t.TempDir()
	fixturePath := filepath.Join(dir, "unsupported_fixture.json")
	const fixtureJSON = `{
	  "metadata": {"file_name": "unsupported_fixture.json", "page_count": 1},
	  "pages": []
	}`
	if err := os.WriteFile(fixturePath, []byte(fixtureJSON), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if got := run([]string{"--quiet", "--pdf", fixturePath}); got != 1 {
		t.Fatalf("run() = %d, want 1", got)
	}
}

func TestRunPDFWritesJSONOutput(t *testing.T) {
	outputDir := t.TempDir()
	inputPath := filepath.Clean("../../samples/pdf/lorem.pdf")

	if got := run([]string{
		"--quiet",
		"--output-dir", outputDir,
		inputPath,
	}); got != 0 {
		t.Fatalf("run() = %d, want 0", got)
	}

	outputPath := filepath.Join(outputDir, "lorem.json")
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected output file %s: %v", outputPath, err)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	os.Stdout = writer
	defer func() { os.Stdout = original }()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	return string(data)
}

func TestRunDirectoryDiscoversSupportedInputs(t *testing.T) {
	root := t.TempDir()
	outputDir := filepath.Join(root, "out")
	if err := os.MkdirAll(filepath.Join(root, "nested"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	const fixtureJSON = `{
	  "metadata": {"file_name": "dir_fixture.json", "page_count": 1},
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
	fixturePath := filepath.Join(root, "nested", "dir_fixture.json")
	if err := os.WriteFile(fixturePath, []byte(fixtureJSON), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "ignore.txt"), []byte("ignored"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if got := run([]string{
		"--quiet",
		"--output-dir", outputDir,
		root,
	}); got != 0 {
		t.Fatalf("run() = %d, want 0", got)
	}

	outputPath := filepath.Join(outputDir, "dir_fixture.json")
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected discovered output file %s: %v", outputPath, err)
	}
}

func TestRunAppliesPagesFilterAndWritesTextAndHTML(t *testing.T) {
	dir := t.TempDir()
	fixturePath := filepath.Join(dir, "paged_fixture.json")
	outputDir := filepath.Join(dir, "out")
	const fixtureJSON = `{
	  "metadata": {"file_name": "paged_fixture.json", "page_count": 2},
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
	          "text": "Page One"
	        }
	      ]
	    },
	    {
	      "metadata": {"number": 2, "index": 1},
	      "artifacts": [
	        {
	          "kind": "text",
	          "page_index": 1,
	          "page_number": 2,
	          "sequence": 0,
	          "bounds": {"left": 0, "bottom": 90, "right": 40, "top": 100},
	          "text": "Page Two"
	        }
	      ]
	    }
	  ]
	}`
	if err := os.WriteFile(fixturePath, []byte(fixtureJSON), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if got := run([]string{
		"--quiet",
		"--output-dir", outputDir,
		"--format", "text,html",
		"--pages", "2",
		fixturePath,
	}); got != 0 {
		t.Fatalf("run() = %d, want 0", got)
	}

	textPath := filepath.Join(outputDir, "paged_fixture.txt")
	htmlPath := filepath.Join(outputDir, "paged_fixture.html")
	textContent, err := os.ReadFile(textPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", textPath, err)
	}
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", htmlPath, err)
	}

	if strings.Contains(string(textContent), "Page One") {
		t.Fatalf("text output should not contain filtered page content: %s", textContent)
	}
	if !strings.Contains(string(textContent), "Page Two") {
		t.Fatalf("text output should contain selected page content: %s", textContent)
	}
	if strings.Contains(string(htmlContent), "Page One") {
		t.Fatalf("html output should not contain filtered page content: %s", htmlContent)
	}
	if !strings.Contains(string(htmlContent), "Page Two") {
		t.Fatalf("html output should contain selected page content: %s", htmlContent)
	}
}

func TestRunAppliesSanitizeAndTextCleanup(t *testing.T) {
	dir := t.TempDir()
	fixturePath := filepath.Join(dir, "sanitize_fixture.json")
	outputDir := filepath.Join(dir, "out")
	const fixtureJSON = `{
	  "metadata": {"file_name": "sanitize_fixture.json", "page_count": 1},
	  "pages": [
	    {
	      "metadata": {"number": 1, "index": 0},
	      "artifacts": [
	        {
	          "kind": "text",
	          "page_index": 0,
	          "page_number": 1,
	          "sequence": 0,
	          "bounds": {"left": 0, "bottom": 90, "right": 80, "top": 100},
	          "text": "reach me at user@example.com\u0000"
	        }
	      ]
	    }
	  ]
	}`
	if err := os.WriteFile(fixturePath, []byte(fixtureJSON), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if got := run([]string{
		"--quiet",
		"--output-dir", outputDir,
		"--format", "text",
		"--replace-invalid-chars", "_",
		"--sanitize",
		fixturePath,
	}); got != 0 {
		t.Fatalf("run() = %d, want 0", got)
	}

	outputPath := filepath.Join(outputDir, "sanitize_fixture.txt")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", outputPath, err)
	}
	text := string(content)
	if !strings.Contains(text, "[EMAIL]_") {
		t.Fatalf("expected sanitized and cleaned content, got %q", text)
	}
}

func TestRunCanKeepHeaderFooterWhenRequested(t *testing.T) {
	dir := t.TempDir()
	fixturePath := filepath.Join(dir, "headers.json")
	outputDir := filepath.Join(dir, "out")
	const fixtureJSON = `{
	  "metadata": {"file_name": "headers.json", "page_count": 2},
	  "pages": [
	    {
	      "metadata": {"number": 1, "index": 0, "size": {"width": 600, "height": 800}},
	      "kids": [
	        {"type": "paragraph", "page_index": 0, "page_number": 1, "bounds": {"left": 40, "bottom": 770, "right": 180, "top": 790}, "content": "Company Report"},
	        {"type": "paragraph", "page_index": 0, "page_number": 1, "bounds": {"left": 40, "bottom": 700, "right": 260, "top": 720}, "content": "Body page one"}
	      ]
	    },
	    {
	      "metadata": {"number": 2, "index": 1, "size": {"width": 600, "height": 800}},
	      "kids": [
	        {"type": "paragraph", "page_index": 1, "page_number": 2, "bounds": {"left": 40, "bottom": 770, "right": 180, "top": 790}, "content": "Company Report"},
	        {"type": "paragraph", "page_index": 1, "page_number": 2, "bounds": {"left": 40, "bottom": 700, "right": 260, "top": 720}, "content": "Body page two"}
	      ]
	    }
	  ]
	}`
	if err := os.WriteFile(fixturePath, []byte(fixtureJSON), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if got := run([]string{
		"--quiet",
		"--output-dir", outputDir,
		"--include-header-footer",
		"--format", "text",
		fixturePath,
	}); got != 0 {
		t.Fatalf("run() = %d, want 0", got)
	}

	outputPath := filepath.Join(outputDir, "headers.txt")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", outputPath, err)
	}
	if !strings.Contains(string(content), "Company Report") {
		t.Fatalf("expected header/footer content to remain when requested: %s", content)
	}
}
