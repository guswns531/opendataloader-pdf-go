package pagerange

import (
	"fmt"
	"strconv"
	"strings"
)

const expectedFormat = "1,3,5-7"

// Parse converts a page selection string like "1,3,5-7" into 1-based page
// numbers. Ranges are inclusive and input order is preserved.
func Parse(spec string) ([]int, error) {
	return parse(spec, false)
}

// ParseOptional behaves like Parse, but treats blank input as no page selection.
func ParseOptional(spec string) ([]int, error) {
	return parse(spec, true)
}

func parse(spec string, allowEmpty bool) ([]int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		if allowEmpty {
			return nil, nil
		}
		return nil, fmt.Errorf("page range cannot be empty")
	}

	parts := strings.Split(spec, ",")
	pages := make([]int, 0, len(parts))

	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			return nil, invalidFormat(spec)
		}

		if startStr, endStr, ok := strings.Cut(token, "-"); ok {
			if strings.Contains(endStr, "-") || startStr == "" || endStr == "" {
				return nil, invalidFormat(spec)
			}

			start, err := parsePositivePage(startStr, spec)
			if err != nil {
				return nil, err
			}
			end, err := parsePositivePage(endStr, spec)
			if err != nil {
				return nil, err
			}
			if start > end {
				return nil, fmt.Errorf("invalid page range %q: start page cannot be greater than end page", token)
			}

			for page := start; page <= end; page++ {
				pages = append(pages, page)
			}
			continue
		}

		page, err := parsePositivePage(token, spec)
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}

	return pages, nil
}

func parsePositivePage(token, fullInput string) (int, error) {
	page, err := strconv.Atoi(strings.TrimSpace(token))
	if err != nil {
		return 0, invalidFormat(fullInput)
	}
	if page < 1 {
		return 0, fmt.Errorf("page numbers must be positive: %q", fullInput)
	}
	return page, nil
}

func invalidFormat(input string) error {
	return fmt.Errorf("invalid page range format: %q. expected format: %s", input, expectedFormat)
}
