// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package hybrid

import "log"

func LogTriageResult(pageNum int, result *TriageResult) {
	if result == nil {
		log.Printf("triage page=%d result=nil", pageNum)
		return
	}

	log.Printf(
		"triage page=%d decision=%s confidence=%.2f signals=%+v",
		pageNum,
		result.Decision,
		result.Confidence,
		result.Signals,
	)
}
