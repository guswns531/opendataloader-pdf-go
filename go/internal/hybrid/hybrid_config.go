// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid

const (
	BackendOff         = "off"
	BackendDocling     = "docling"
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
	DefaultTimeoutMS         = 30000
	DefaultMaxConcurrent     = 4
	DefaultDoclingURL        = "http://localhost:5001"
	DefaultDoclingFastURL    = "http://localhost:5002"
	DefaultHancomURL         = "https://dataloader.cloud.hancom.com/studio-lite/api"
	DefaultHybridMode        = "auto"
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
		Mode:           DefaultHybridMode,
		TimeoutMS:      DefaultTimeoutMS,
		MaxConcurrency: DefaultMaxConcurrent,
	}
}

func DefaultURLForBackend(backend string) string {
	switch backend {
	case BackendDocling, BackendDoclingFast:
		return DefaultDoclingFastURL
	case BackendHancom:
		return DefaultHancomURL
	default:
		return ""
	}
}

func (c *HybridConfig) EffectiveURL() string {
	if c != nil && c.URL != "" {
		return c.URL
	}
	backend := BackendOff
	if c != nil && c.Backend != "" {
		backend = c.Backend
	}
	return DefaultURLForBackend(backend)
}

func (c *HybridConfig) Timeout() int {
	if c != nil && c.TimeoutMS > 0 {
		return c.TimeoutMS
	}
	return DefaultTimeoutMS
}

func (c *HybridConfig) IsFullMode() bool {
	return c != nil && c.Mode == "full"
}
