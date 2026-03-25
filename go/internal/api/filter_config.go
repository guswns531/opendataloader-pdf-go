// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package api

import "strings"

type FilterConfig struct {
	DisableHiddenText bool
	DisableOffPage    bool
	DisableTiny       bool
	DisableHiddenOCG  bool
}

func FilterConfigFromStrings(flags []string) *FilterConfig {
	cfg := &FilterConfig{}
	for _, raw := range flags {
		for _, part := range strings.Split(raw, ",") {
			flag := strings.ToLower(strings.TrimSpace(part))
			switch flag {
			case "", "none":
				continue
			case "all":
				cfg.DisableHiddenText = true
				cfg.DisableOffPage = true
				cfg.DisableTiny = true
				cfg.DisableHiddenOCG = true
			case "hidden-text":
				cfg.DisableHiddenText = true
			case "off-page":
				cfg.DisableOffPage = true
			case "tiny":
				cfg.DisableTiny = true
			case "hidden-ocg":
				cfg.DisableHiddenOCG = true
			}
		}
	}
	return cfg
}

func (f *FilterConfig) IsAll() bool {
	if f == nil {
		return false
	}
	return f.DisableHiddenText && f.DisableOffPage && f.DisableTiny && f.DisableHiddenOCG
}
