// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid_test

import (
	"testing"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/hybrid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHybridClientReturnsHancomStub(t *testing.T) {
	client, err := hybrid.NewHybridClient(&hybrid.HybridConfig{Backend: hybrid.BackendHancom})
	require.NoError(t, err)
	require.NotNil(t, client)

	_, err = client.Convert(&hybrid.ConvertRequest{})
	assert.ErrorContains(t, err, "hancom backend not yet implemented")
}
