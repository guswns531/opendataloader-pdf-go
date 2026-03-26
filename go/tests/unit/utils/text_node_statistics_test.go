// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/utils"
)

func TestModeWeightStatisticsUsesWeightedModeAndBoost(t *testing.T) {
	stats := utils.NewModeWeightStatistics(10, 32, 10, 13)
	stats.Add(12, 3)
	stats.Add(18, 1)
	stats.Add(24, 2)

	assert.Equal(t, 12.0, stats.GetMode())
	assert.Equal(t, 0.5, stats.GetBoost(18))
	assert.Equal(t, 1.0, stats.GetBoost(24))
	assert.Equal(t, 0.0, stats.GetBoost(12))
}

func TestTextNodeStatisticsComputesModeMeanAndStdDev(t *testing.T) {
	stats := utils.NewTextNodeStatistics()
	stats.Add(12, 400)
	stats.Add(12, 400)
	stats.Add(24, 700)

	assert.Equal(t, 12.0, stats.FontSizeMode())
	assert.Equal(t, 400.0, stats.FontWeightMode())
	assert.InDelta(t, 16.0, stats.FontSizeMean(), 0.001)
	assert.InDelta(t, 500.0, stats.FontWeightMean(), 0.001)
	assert.Greater(t, stats.FontSizeStdDev(), 0.0)
	assert.Greater(t, stats.FontWeightStdDev(), 0.0)
	assert.Equal(t, 0.0, stats.FontSizeRarityBoost(12))
	assert.Greater(t, stats.FontSizeRarityBoost(24), 0.0)
}
