// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid

import "log/slog"

func LogTriageResult(pageNum int, result *TriageResult) {
	if result == nil {
		slog.Warn("triage result missing", "page", pageNum)
		return
	}

	slog.Info(
		"triage decision",
		"page", pageNum,
		"decision", result.Decision,
		"confidence", result.Confidence,
		"signals", result.Signals,
	)
}
