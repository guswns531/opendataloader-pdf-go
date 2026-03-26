// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid

import "fmt"

// HancomClient is a stub for the Hancom proprietary backend.
// Full implementation requires Hancom API credentials.
type HancomClient struct {
	config *HybridConfig
}

func NewHancomClient(config *HybridConfig) *HancomClient {
	if config == nil {
		config = DefaultHybridConfig()
	}
	return &HancomClient{config: config}
}

func (c *HancomClient) Convert(req *ConvertRequest) (*ConvertResponse, error) {
	return nil, fmt.Errorf("hancom backend not yet implemented")
}

func (c *HancomClient) HealthCheck() error {
	return fmt.Errorf("hancom backend not yet implemented")
}

func (c *HancomClient) Close() error {
	return nil
}
