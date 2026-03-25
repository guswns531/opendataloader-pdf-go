// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package utils

type ModeWeightStatistics struct {
	weights map[float64]float64
}

func (s *ModeWeightStatistics) Add(value, weight float64) {
	if s.weights == nil {
		s.weights = make(map[float64]float64)
	}
	s.weights[value] += weight
}

func (s *ModeWeightStatistics) Mode() float64 {
	if len(s.weights) == 0 {
		return 0
	}

	var (
		mode      float64
		maxWeight float64
		init      bool
	)
	for value, weight := range s.weights {
		if !init || weight > maxWeight || (weight == maxWeight && value < mode) {
			mode = value
			maxWeight = weight
			init = true
		}
	}

	return mode
}
