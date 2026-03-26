// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package utils

import "sort"

type scoreWeightEntry struct {
	Score  float64
	Weight float64
}

type ModeWeightStatistics struct {
	scoreMin float64
	scoreMax float64
	modeMin  float64
	modeMax  float64
	countMap map[float64]float64
	sorted   []scoreWeightEntry
	higher   []float64
	ready    bool
}

func NewModeWeightStatistics(scoreMin, scoreMax, modeMin, modeMax float64) *ModeWeightStatistics {
	return &ModeWeightStatistics{
		scoreMin: scoreMin,
		scoreMax: scoreMax,
		modeMin:  modeMin,
		modeMax:  modeMax,
		countMap: make(map[float64]float64),
	}
}

func (s *ModeWeightStatistics) AddScore(score float64) {
	s.Add(score, 1)
}

func (s *ModeWeightStatistics) Add(value, weight float64) {
	if s.countMap == nil {
		s.countMap = make(map[float64]float64)
	}
	s.countMap[value] += weight
	s.sorted = nil
	s.higher = nil
	s.ready = false
}

func (s *ModeWeightStatistics) SortByFrequency() {
	if s == nil {
		return
	}
	s.sorted = s.sorted[:0]
	for score, weight := range s.countMap {
		s.sorted = append(s.sorted, scoreWeightEntry{Score: score, Weight: weight})
	}
	sort.Slice(s.sorted, func(i, j int) bool {
		if s.sorted[i].Weight == s.sorted[j].Weight {
			return s.sorted[i].Score < s.sorted[j].Score
		}
		return s.sorted[i].Weight > s.sorted[j].Weight
	})
}

func (s *ModeWeightStatistics) GetMode() float64 {
	if s == nil {
		return 0
	}
	if len(s.sorted) == 0 {
		s.SortByFrequency()
	}
	for _, entry := range s.sorted {
		if entry.Score >= s.modeMin && entry.Score <= s.modeMax {
			return entry.Score
		}
	}
	return 0
}

func (s *ModeWeightStatistics) Mode() float64 {
	return s.GetMode()
}

func (s *ModeWeightStatistics) GetBoost(score float64) float64 {
	if s == nil {
		return 0
	}
	s.initHigherScores()
	if len(s.higher) == 0 {
		return 0
	}
	for idx, candidate := range s.higher {
		if candidate == score {
			return float64(idx+1) / float64(len(s.higher))
		}
	}
	return 0
}

func (s *ModeWeightStatistics) initHigherScores() {
	if s == nil || s.ready {
		return
	}
	if len(s.sorted) == 0 {
		s.SortByFrequency()
	}
	mode := s.GetMode()
	s.higher = s.higher[:0]
	for _, entry := range s.sorted {
		if entry.Score > mode && entry.Score >= s.scoreMin && entry.Score <= s.scoreMax {
			s.higher = append(s.higher, entry.Score)
		}
	}
	sort.Float64s(s.higher)
	s.ready = true
}
