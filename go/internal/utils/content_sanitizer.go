/*
 * Copyright 2025-2026 Hancom Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"regexp"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

type SanitizationRule struct {
	Name        string
	Pattern     *regexp.Regexp
	Replacement string
}

var DefaultRules = []SanitizationRule{
	{Name: "email", Pattern: regexp.MustCompile(`(?i)\b[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}\b`), Replacement: "[EMAIL]"},
	{Name: "phone", Pattern: regexp.MustCompile(`\b(?:\+?\d{1,3}[-.\s]?)?(?:\(?\d{2,4}\)?[-.\s]?)?\d{3,4}[-.\s]?\d{4}\b`), Replacement: "[PHONE]"},
	{Name: "ip", Pattern: regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`), Replacement: "[IP]"},
	{Name: "credit_card", Pattern: regexp.MustCompile(`\b(?:\d[ -]*?){13,19}\b`), Replacement: "[CREDIT_CARD]"},
	{Name: "url", Pattern: regexp.MustCompile(`(?i)\b(?:https?://|www\.)\S+\b`), Replacement: "[URL]"},
}

func Sanitize(doc *entities.Document, rules []SanitizationRule) {
	if doc == nil {
		return
	}
	if len(rules) == 0 {
		rules = DefaultRules
	}
	for _, page := range doc.Pages {
		if page == nil {
			continue
		}
		for _, chunk := range page.Chunks {
			if chunk != nil {
				chunk.Text = sanitizeText(chunk.Text, rules)
			}
		}
		for _, element := range page.Elements {
			sanitizeObject(element, rules)
		}
	}
}

func sanitizeObject(obj entities.IObject, rules []SanitizationRule) {
	switch v := obj.(type) {
	case *entities.TextChunk:
		v.Text = sanitizeText(v.Text, rules)
	case *entities.TextLine:
		for _, chunk := range v.Chunks {
			if chunk != nil {
				chunk.Text = sanitizeText(chunk.Text, rules)
			}
		}
	case *entities.SemanticParagraph:
		for _, line := range v.Lines {
			sanitizeObject(line, rules)
		}
	case *entities.SemanticHeading:
		for _, line := range v.Lines {
			sanitizeObject(line, rules)
		}
	case *entities.SemanticTable:
		for _, row := range v.Rows {
			if row == nil {
				continue
			}
			for _, cell := range row.Cells {
				if cell == nil {
					continue
				}
				for _, content := range cell.Content {
					sanitizeObject(content, rules)
				}
			}
		}
	case *entities.PDFList:
		for _, item := range v.Items {
			if item == nil {
				continue
			}
			item.BulletText = sanitizeText(item.BulletText, rules)
			for _, content := range item.Content {
				sanitizeObject(content, rules)
			}
		}
	case *entities.SemanticCaption:
		v.Text = sanitizeText(v.Text, rules)
	case *entities.SemanticFormula:
		v.LaTeX = sanitizeText(v.LaTeX, rules)
	case *entities.SemanticImage:
		v.Alt = sanitizeText(v.Alt, rules)
	case *entities.SemanticHeaderFooter:
		for _, line := range v.Lines {
			sanitizeObject(line, rules)
		}
	}
}

func sanitizeText(input string, rules []SanitizationRule) string {
	if input == "" || len(rules) == 0 {
		return input
	}

	out := input
	for {
		ruleIndex := -1
		start := -1
		end := -1
		for i, rule := range rules {
			if rule.Pattern == nil {
				continue
			}
			loc := rule.Pattern.FindStringIndex(out)
			if loc == nil {
				continue
			}
			if start == -1 || loc[0] < start || (loc[0] == start && i < ruleIndex) {
				ruleIndex = i
				start = loc[0]
				end = loc[1]
			}
		}
		if ruleIndex == -1 {
			return out
		}
		out = out[:start] + rules[ruleIndex].Replacement + out[end:]
	}
}
