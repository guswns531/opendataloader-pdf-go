// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package levels

import "math"

const xGapMultiplier = 0.3

type LevelInfo struct {
	Level    int
	FontSize float64
	IsBold   bool
	Count    int

	Left    float64
	Right   float64
	MaxXGap float64
}

func AreSameLevelInfos(left, right *LevelInfo) bool {
	if left == nil || right == nil {
		return false
	}
	if left.Right < right.Left || right.Right < left.Left {
		return true
	}
	maxGap := math.Max(left.MaxXGap, right.MaxXGap)
	return closeNumber(left.Left, right.Left, maxGap) || closeNumber(left.Right, right.Right, maxGap)
}

func closeNumber(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps
}
