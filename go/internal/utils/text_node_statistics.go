// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package utils

import "sort"

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

type modeWeightStatistics struct {
	scoreMin float64
	scoreMax float64
	modeMin  float64
	modeMax  float64
	countMap map[float64]int
}

func newModeWeightStatistics(scoreMin, scoreMax, modeMin, modeMax float64) *modeWeightStatistics {
	return &modeWeightStatistics{
		scoreMin: scoreMin,
		scoreMax: scoreMax,
		modeMin:  modeMin,
		modeMax:  modeMax,
		countMap: make(map[float64]int),
	}
}

func (m *modeWeightStatistics) addScore(score float64) {
	m.countMap[score]++
}

func (m *modeWeightStatistics) getBoost(score float64) float64 {
	type entry struct {
		score float64
		count int
	}
	var entries []entry
	for s, c := range m.countMap {
		entries = append(entries, entry{score: s, count: c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count == entries[j].count {
			return entries[i].score < entries[j].score
		}
		return entries[i].count > entries[j].count
	})
	mode := 0.0
	for _, item := range entries {
		if item.score >= m.modeMin && item.score <= m.modeMax {
			mode = item.score
			break
		}
	}
	var higher []float64
	for _, item := range entries {
		if item.score > mode && item.score >= m.scoreMin && item.score <= m.scoreMax {
			higher = append(higher, item.score)
		}
	}
	sort.Float64s(higher)
	if len(higher) == 0 {
		return 0
	}
	for i, value := range higher {
		if value == score {
			return float64(i+1) / float64(len(higher))
		}
	}
	return 0
}

type TextNodeStatistics struct {
	fontSizeStatistics   *modeWeightStatistics
	fontWeightStatistics *modeWeightStatistics
	config               TextNodeStatisticsConfig
}

func NewTextNodeStatistics() *TextNodeStatistics {
	cfg := DefaultTextNodeStatisticsConfig()
	return &TextNodeStatistics{
		fontSizeStatistics:   newModeWeightStatistics(cfg.FontSizeHeadingMin, cfg.FontSizeHeadingMax, cfg.FontSizeDominantMin, cfg.FontSizeDominantMax),
		fontWeightStatistics: newModeWeightStatistics(cfg.FontWeightHeadingMin, cfg.FontWeightHeadingMax, cfg.FontWeightDominantMin, cfg.FontWeightDominantMax),
		config:               cfg,
	}
}

func (s *TextNodeStatistics) Add(fontSize, fontWeight float64) {
	if s == nil {
		return
	}
	s.fontSizeStatistics.addScore(fontSize)
	s.fontWeightStatistics.addScore(fontWeight)
}

func (s *TextNodeStatistics) FontSizeRarityBoost(fontSize float64) float64 {
	if s == nil {
		return 0
	}
	return s.fontSizeStatistics.getBoost(fontSize) * s.config.FontSizeRarityBoost
}

func (s *TextNodeStatistics) FontWeightRarityBoost(fontWeight float64) float64 {
	if s == nil {
		return 0
	}
	return s.fontWeightStatistics.getBoost(fontWeight) * s.config.FontWeightRarityBoost
}
