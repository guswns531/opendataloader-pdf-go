// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
)

const DefaultTriageFilename = "triage.json"

type TriageLogger struct{}

type triageLog struct {
	Document string           `json:"document"`
	Hybrid   string           `json:"hybrid"`
	Triage   []triageLogEntry `json:"triage"`
	Summary  triageSummary    `json:"summary"`
}

type triageLogEntry struct {
	Page       int           `json:"page"`
	Decision   string        `json:"decision"`
	Confidence float64       `json:"confidence"`
	Signals    TriageSignals `json:"signals"`
}

type triageSummary struct {
	TotalPages   int `json:"totalPages"`
	JavaPages    int `json:"javaPages"`
	BackendPages int `json:"backendPages"`
}

func (l *TriageLogger) LogToFile(outputDir, documentName, hybridBackend string, triageResults map[int]*TriageResult) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	payload, err := l.CreateTriageJSON(documentName, hybridBackend, triageResults)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(outputDir, DefaultTriageFilename)
	if err := os.WriteFile(outputPath, payload, 0o644); err != nil {
		return err
	}

	slog.Info("triage log written", "path", outputPath)
	return nil
}

func (l *TriageLogger) CreateTriageJSON(documentName, hybridBackend string, triageResults map[int]*TriageResult) ([]byte, error) {
	if l == nil {
		l = &TriageLogger{}
	}

	root := triageLog{
		Document: documentName,
		Hybrid:   hybridBackend,
		Triage:   make([]triageLogEntry, 0, len(triageResults)),
	}

	pageNums := make([]int, 0, len(triageResults))
	for pageNum := range triageResults {
		pageNums = append(pageNums, pageNum)
	}
	sort.Ints(pageNums)

	for _, pageNum := range pageNums {
		result := triageResults[pageNum]
		if result == nil {
			slog.Warn("triage result missing", "page", pageNum+1)
			continue
		}

		root.Triage = append(root.Triage, triageLogEntry{
			Page:       pageNum + 1,
			Decision:   result.Decision,
			Confidence: result.Confidence,
			Signals:    result.Signals,
		})

		switch result.Decision {
		case TriageDecisionBackend:
			root.Summary.BackendPages++
		default:
			root.Summary.JavaPages++
		}

		slog.Info(
			"triage decision",
			"page", pageNum+1,
			"decision", result.Decision,
			"confidence", result.Confidence,
			"signals", result.Signals,
		)
	}

	root.Summary.TotalPages = len(root.Triage)

	return json.MarshalIndent(root, "", "  ")
}
