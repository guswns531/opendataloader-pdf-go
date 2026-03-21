package control

import (
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
)

const (
	defaultReadingOrderMode  = "xycut"
	defaultReplacementChar   = " "
	defaultTableMethod       = "default"
	extraReadingOrderKey     = "reading_order"
	extraTableMethodKey      = "table_method"
	extraIncludeHeaderFooter = "include_header_footer"
	extraReplaceInvalidChars = "replace_invalid"
	extraContentSafetyOff    = "content_safety_off"
	extraSanitizeKey         = "sanitize"
	extraKeepLineBreaksKey   = "keep_line_breaks"
	extraUseStructTreeKey    = "use_struct_tree"
	extraDetectStrikethrough = "detect_strikethrough"
)

// Decisions captures the runtime switches derived from processing options.
//
// The struct centralizes the small set of heuristics that can be resolved from
// core.ProcessingOptions today without touching downstream pipeline code.
type Decisions struct {
	ReadingOrderEnabled    bool
	ReadingOrderMode       string
	HeaderFooterIncluded   bool
	TableMethod            string
	TableHeuristicsEnabled bool
	SanitizeEnabled        bool
	ReplaceInvalidChars    string
	ContentSafetyOff       string
	LayoutFilteringEnabled bool
	KeepLineBreaks         bool
	UseStructTree          bool
	DetectStrikethrough    bool
	PreservePageBreaks     bool
	EmitGeometry           bool
	EmitDiagnostics        bool
	Strict                 bool
}

// Resolve normalizes ProcessingOptions into a decision set.
func Resolve(options core.ProcessingOptions) Decisions {
	return Decisions{
		ReadingOrderEnabled:    ReadingOrderEnabled(options),
		ReadingOrderMode:       ReadingOrderMode(options),
		HeaderFooterIncluded:   HeaderFooterIncluded(options),
		TableMethod:            TableMethod(options),
		TableHeuristicsEnabled: TableHeuristicsEnabled(options),
		SanitizeEnabled:        SanitizeEnabled(options),
		ReplaceInvalidChars:    ReplaceInvalidChars(options),
		ContentSafetyOff:       ContentSafetyOff(options),
		LayoutFilteringEnabled: LayoutFilteringEnabled(options),
		KeepLineBreaks:         KeepLineBreaks(options),
		UseStructTree:          UseStructTree(options),
		DetectStrikethrough:    DetectStrikethrough(options),
		PreservePageBreaks:     options.PreservePageBreaks,
		EmitGeometry:           options.EmitGeometry,
		EmitDiagnostics:        options.EmitDiagnostics,
		Strict:                 options.Strict,
	}
}

// ReadingOrderEnabled reports whether reading-order sorting should run.
func ReadingOrderEnabled(options core.ProcessingOptions) bool {
	return ReadingOrderMode(options) != "off"
}

// ReadingOrderMode returns the normalized reading-order mode.
func ReadingOrderMode(options core.ProcessingOptions) string {
	if mode, ok := extraString(options.Extras, extraReadingOrderKey, "reading-order", "reading_order_mode", "reading-order-mode"); ok {
		return normalizeMode(mode, defaultReadingOrderMode, "off", defaultReadingOrderMode)
	}
	if hasHeuristic(options.DisabledHeuristics, "xycut", "reading-order", "reading_order") {
		return "off"
	}
	if hasHeuristic(options.EnabledHeuristics, "xycut", "reading-order", "reading_order") {
		return defaultReadingOrderMode
	}
	return defaultReadingOrderMode
}

// HeaderFooterIncluded reports whether repeated headers and footers should be retained.
func HeaderFooterIncluded(options core.ProcessingOptions) bool {
	if value, ok := extraBool(options.Extras, extraIncludeHeaderFooter, "include-header-footer"); ok {
		return value
	}
	if hasHeuristic(options.DisabledHeuristics, "headerfooter", "header-footer", "include-header-footer") {
		return false
	}
	if hasHeuristic(options.EnabledHeuristics, "headerfooter", "header-footer", "include-header-footer") {
		return true
	}
	return false
}

// TableMethod returns the normalized table heuristic mode.
func TableMethod(options core.ProcessingOptions) string {
	if method, ok := extraString(options.Extras, extraTableMethodKey, "table-method"); ok {
		return normalizeMode(method, defaultTableMethod, "off", defaultTableMethod, "cluster")
	}
	if hasHeuristic(options.DisabledHeuristics, "table", "table-heuristic", "cluster") {
		return "off"
	}
	if hasHeuristic(options.EnabledHeuristics, "cluster") {
		return "cluster"
	}
	if hasHeuristic(options.EnabledHeuristics, "table", "default") {
		return defaultTableMethod
	}
	return defaultTableMethod
}

// TableHeuristicsEnabled reports whether table detection should run.
func TableHeuristicsEnabled(options core.ProcessingOptions) bool {
	return TableMethod(options) != "off"
}

// SanitizeEnabled reports whether post-processing sanitization should run.
func SanitizeEnabled(options core.ProcessingOptions) bool {
	if value, ok := extraBool(options.Extras, extraSanitizeKey); ok {
		return value
	}
	if hasHeuristic(options.DisabledHeuristics, "sanitize") {
		return false
	}
	if hasHeuristic(options.EnabledHeuristics, "sanitize") {
		return true
	}
	return false
}

// ReplaceInvalidChars returns the configured replacement character.
func ReplaceInvalidChars(options core.ProcessingOptions) string {
	if value, ok := extraString(options.Extras, extraReplaceInvalidChars, "replace-invalid-chars"); ok {
		return value
	}
	return defaultReplacementChar
}

// ContentSafetyOff returns the raw content-safety selector string.
func ContentSafetyOff(options core.ProcessingOptions) string {
	if value, ok := extraString(options.Extras, extraContentSafetyOff, "content-safety-off"); ok {
		return value
	}
	return ""
}

// LayoutFilteringEnabled mirrors the current content-safety layout gating.
func LayoutFilteringEnabled(options core.ProcessingOptions) bool {
	spec := strings.TrimSpace(ContentSafetyOff(options))
	if spec == "" {
		return true
	}
	for _, token := range strings.Split(spec, ",") {
		switch strings.TrimSpace(strings.ToLower(token)) {
		case "all", "off-page", "tiny":
			return false
		}
	}
	return true
}

// KeepLineBreaks reports whether line breaks should be preserved when present in extras.
func KeepLineBreaks(options core.ProcessingOptions) bool {
	if value, ok := extraBool(options.Extras, extraKeepLineBreaksKey, "keep-line-breaks"); ok {
		return value
	}
	return false
}

// UseStructTree reports whether structure-tree-based processing is enabled in extras.
func UseStructTree(options core.ProcessingOptions) bool {
	if value, ok := extraBool(options.Extras, extraUseStructTreeKey, "use-struct-tree"); ok {
		return value
	}
	return false
}

// DetectStrikethrough reports whether markdown strikethrough detection is enabled in extras.
func DetectStrikethrough(options core.ProcessingOptions) bool {
	if value, ok := extraBool(options.Extras, extraDetectStrikethrough, "detect-strikethrough"); ok {
		return value
	}
	return false
}

func extraBool(extras map[string]any, keys ...string) (bool, bool) {
	for _, key := range keys {
		value, ok := lookupExtra(extras, key)
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case bool:
			return typed, true
		case string:
			switch strings.ToLower(strings.TrimSpace(typed)) {
			case "1", "t", "true", "yes", "y", "on":
				return true, true
			case "0", "f", "false", "no", "n", "off":
				return false, true
			}
		}
	}
	return false, false
}

func extraString(extras map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		value, ok := lookupExtra(extras, key)
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) == "" {
				continue
			}
			return typed, true
		case core.OutputFormat:
			return string(typed), true
		}
	}
	return "", false
}

func lookupExtra(extras map[string]any, key string) (any, bool) {
	if len(extras) == 0 {
		return nil, false
	}
	if value, ok := extras[key]; ok {
		return value, true
	}
	alt := strings.ReplaceAll(key, "-", "_")
	if alt != key {
		if value, ok := extras[alt]; ok {
			return value, true
		}
	}
	alt = strings.ReplaceAll(key, "_", "-")
	if alt != key {
		if value, ok := extras[alt]; ok {
			return value, true
		}
	}
	return nil, false
}

func hasHeuristic(values []string, names ...string) bool {
	for _, value := range values {
		normalizedValue := normalizeToken(value)
		for _, name := range names {
			if normalizedValue == normalizeToken(name) {
				return true
			}
		}
	}
	return false
}

func normalizeMode(value, fallback string, allowed ...string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return fallback
	}
	for _, candidate := range allowed {
		if normalized == candidate {
			return normalized
		}
	}
	return fallback
}

func normalizeToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}
