package hybrid

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTriageLoggerCreatesJavaCompatibleJSON(t *testing.T) {
	logger := &TriageLogger{}
	results := map[int]*TriageResult{
		1: {
			Decision:   TriageDecisionJava,
			Confidence: 0.9,
			Signals: TriageSignals{
				LineChunkCount:       2,
				TextChunkCount:       45,
				LineToTextRatio:      0.04,
				AlignedLineGroups:    0,
				HasTableBorder:       false,
				HasSuspiciousPattern: false,
			},
		},
		0: {
			Decision:   TriageDecisionBackend,
			Confidence: 0.95,
			Signals: TriageSignals{
				LineChunkCount:       6,
				TextChunkCount:       22,
				LineToTextRatio:      0.21,
				AlignedLineGroups:    5,
				HasTableBorder:       true,
				HasSuspiciousPattern: true,
			},
		},
	}

	payload, err := logger.CreateTriageJSON("sample.pdf", BackendDocling, results)
	if err != nil {
		t.Fatalf("CreateTriageJSON returned error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if decoded["document"] != "sample.pdf" {
		t.Fatalf("expected document sample.pdf, got %v", decoded["document"])
	}
	if decoded["hybrid"] != BackendDocling {
		t.Fatalf("expected hybrid %q, got %v", BackendDocling, decoded["hybrid"])
	}

	triage, ok := decoded["triage"].([]any)
	if !ok || len(triage) != 2 {
		t.Fatalf("expected 2 triage entries, got %#v", decoded["triage"])
	}

	first := triage[0].(map[string]any)
	if first["page"].(float64) != 1 || first["decision"] != TriageDecisionBackend {
		t.Fatalf("expected first entry to be page 1 backend, got %#v", first)
	}

	second := triage[1].(map[string]any)
	if second["page"].(float64) != 2 || second["decision"] != TriageDecisionJava {
		t.Fatalf("expected second entry to be page 2 java, got %#v", second)
	}

	summary := decoded["summary"].(map[string]any)
	if summary["totalPages"].(float64) != 2 || summary["javaPages"].(float64) != 1 || summary["backendPages"].(float64) != 1 {
		t.Fatalf("unexpected summary %#v", summary)
	}
}

func TestTriageLoggerWritesFile(t *testing.T) {
	logger := &TriageLogger{}
	outputDir := t.TempDir()

	err := logger.LogToFile(outputDir, "sample.pdf", BackendDocling, map[int]*TriageResult{
		0: {Decision: TriageDecisionBackend, Confidence: 1.0},
	})
	if err != nil {
		t.Fatalf("LogToFile returned error: %v", err)
	}

	outputPath := filepath.Join(outputDir, DefaultTriageFilename)
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected triage file at %q: %v", outputPath, err)
	}
}
