package processors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/hybrid"
)

func TestHybridDocumentProcessorWritesTriageLog(t *testing.T) {
	outputDir := t.TempDir()
	pdfPath := filepath.Join(outputDir, "sample.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	cfg := api.DefaultConfig()
	cfg.Hybrid = api.HybridDoclingFast
	cfg.OutputDir = outputDir

	triageResults := map[int]*hybrid.TriageResult{
		0: {
			Decision:   hybrid.TriageDecisionBackend,
			Confidence: 1.0,
		},
	}

	logTriageToFile(pdfPath, cfg, triageResults)

	triagePath := filepath.Join(outputDir, "triage.json")
	data, err := os.ReadFile(triagePath)
	if err != nil {
		t.Fatalf("expected triage.json at %q: %v", triagePath, err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if decoded["document"] != "sample.pdf" {
		t.Fatalf("expected sample.pdf document, got %v", decoded["document"])
	}
}
