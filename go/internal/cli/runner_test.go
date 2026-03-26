package cli

import (
	"os"
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
