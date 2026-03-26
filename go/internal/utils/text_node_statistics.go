// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package utils

import "math"

type TextNodeStatisticsConfig struct {
	FontSizeDominantMin   float64
	FontSizeDominantMax   float64
	FontSizeHeadingMin    float64
	FontSizeHeadingMax    float64
	FontSizeRarityBoost   float64
	FontWeightDominantMin float64
	FontWeightDominantMax float64
	FontWeightHeadingMin  float64
	FontWeightHeadingMax  float64
	FontWeightRarityBoost float64
}

func DefaultTextNodeStatisticsConfig() TextNodeStatisticsConfig {
	return TextNodeStatisticsConfig{
		FontSizeDominantMin:   10.0,
		FontSizeDominantMax:   13.0,
		FontSizeHeadingMin:    10.0,
		FontSizeHeadingMax:    32.0,
		FontSizeRarityBoost:   0.5,
		FontWeightDominantMin: 395.0,
		FontWeightDominantMax: 405.0,
		FontWeightHeadingMin:  400.0,
		FontWeightHeadingMax:  900.0,
		FontWeightRarityBoost: 0.3,
	}
}

type TextNodeStatistics struct {
	fontSizeStatistics   *ModeWeightStatistics
	fontWeightStatistics *ModeWeightStatistics
	config               TextNodeStatisticsConfig
	fontSizeCount        int
	fontSizeSum          float64
	fontSizeSumSquares   float64
	fontWeightCount      int
	fontWeightSum        float64
	fontWeightSumSquares float64
}

func NewTextNodeStatistics() *TextNodeStatistics {
	cfg := DefaultTextNodeStatisticsConfig()
	return &TextNodeStatistics{
		fontSizeStatistics:   NewModeWeightStatistics(cfg.FontSizeHeadingMin, cfg.FontSizeHeadingMax, cfg.FontSizeDominantMin, cfg.FontSizeDominantMax),
		fontWeightStatistics: NewModeWeightStatistics(cfg.FontWeightHeadingMin, cfg.FontWeightHeadingMax, cfg.FontWeightDominantMin, cfg.FontWeightDominantMax),
		config:               cfg,
	}
}

func (s *TextNodeStatistics) Add(fontSize, fontWeight float64) {
	if s == nil {
		return
	}
	s.fontSizeStatistics.AddScore(fontSize)
	s.fontWeightStatistics.AddScore(fontWeight)
	s.fontSizeCount++
	s.fontSizeSum += fontSize
	s.fontSizeSumSquares += fontSize * fontSize
	s.fontWeightCount++
	s.fontWeightSum += fontWeight
	s.fontWeightSumSquares += fontWeight * fontWeight
}

func (s *TextNodeStatistics) FontSizeMode() float64 {
	if s == nil {
		return 0
	}
	return s.fontSizeStatistics.GetMode()
}

func (s *TextNodeStatistics) FontWeightMode() float64 {
	if s == nil {
		return 0
	}
	return s.fontWeightStatistics.GetMode()
}

func (s *TextNodeStatistics) FontSizeMean() float64 {
	if s == nil || s.fontSizeCount == 0 {
		return 0
	}
	return s.fontSizeSum / float64(s.fontSizeCount)
}

func (s *TextNodeStatistics) FontWeightMean() float64 {
	if s == nil || s.fontWeightCount == 0 {
		return 0
	}
	return s.fontWeightSum / float64(s.fontWeightCount)
}

func (s *TextNodeStatistics) FontSizeStdDev() float64 {
	if s == nil || s.fontSizeCount == 0 {
		return 0
	}
	return stddev(s.fontSizeSum, s.fontSizeSumSquares, s.fontSizeCount)
}

func (s *TextNodeStatistics) FontWeightStdDev() float64 {
	if s == nil || s.fontWeightCount == 0 {
		return 0
	}
	return stddev(s.fontWeightSum, s.fontWeightSumSquares, s.fontWeightCount)
}

func (s *TextNodeStatistics) FontSizeRarityBoost(fontSize float64) float64 {
	if s == nil {
		return 0
	}
	return s.fontSizeStatistics.GetBoost(fontSize) * s.config.FontSizeRarityBoost
}

func (s *TextNodeStatistics) FontWeightRarityBoost(fontWeight float64) float64 {
	if s == nil {
		return 0
	}
	return s.fontWeightStatistics.GetBoost(fontWeight) * s.config.FontWeightRarityBoost
}

func stddev(sum, sumSquares float64, count int) float64 {
	mean := sum / float64(count)
	variance := sumSquares/float64(count) - mean*mean
	if variance < 0 {
		variance = 0
	}
	return math.Sqrt(variance)
}
