// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid

const (
	BackendOff         = "off"
	BackendDoclingFast = "docling-fast"
	BackendHancom      = "hancom"
	BackendAzure       = "azure"
	BackendGoogle      = "google"
)

const (
	TriageDecisionJava    = "JAVA"
	TriageDecisionBackend = "BACKEND"
)

const (
	defaultTimeoutMS     = 30000
	defaultMaxConcurrent = 4
	defaultDoclingFast   = "http://localhost:5002"
	defaultHancomURL     = "https://dataloader.cloud.hancom.com/studio-lite/api"
	defaultMode          = "auto"
)

type HybridConfig struct {
	Backend        string
	Mode           string
	URL            string
	TimeoutMS      int
	Fallback       bool
	MaxConcurrency int
}

func DefaultHybridConfig() *HybridConfig {
	return &HybridConfig{
		Backend:        BackendOff,
		Mode:           defaultMode,
		TimeoutMS:      defaultTimeoutMS,
		MaxConcurrency: defaultMaxConcurrent,
	}
}

func (c *HybridConfig) effectiveURL() string {
	if c != nil && c.URL != "" {
		return c.URL
	}
	backend := BackendOff
	if c != nil && c.Backend != "" {
		backend = c.Backend
	}
	switch backend {
	case BackendDoclingFast:
		return defaultDoclingFast
	case BackendHancom:
		return defaultHancomURL
	default:
		return ""
	}
}

func (c *HybridConfig) timeoutMS() int {
	if c != nil && c.TimeoutMS > 0 {
		return c.TimeoutMS
	}
	return defaultTimeoutMS
}
