// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid

import "fmt"

func NewHybridClient(config *HybridConfig) (HybridClient, error) {
	if config == nil {
		config = DefaultHybridConfig()
	}

	switch config.Backend {
	case "", BackendOff:
		return nil, nil
	case BackendDocling, BackendDoclingFast:
		return NewDoclingFastServerClient(config), nil
	case BackendHancom:
		return NewHancomClient(config), nil
	default:
		return nil, fmt.Errorf("unknown hybrid backend: %s", config.Backend)
	}
}
