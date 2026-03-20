package core

type ProcessingOptions struct {
	InputPath          string            `json:"input_path,omitempty"`
	OutputPath         string            `json:"output_path,omitempty"`
	DocumentName       string            `json:"document_name,omitempty"`
	RequestedFormats   []OutputFormat    `json:"requested_formats,omitempty"`
	EnabledHeuristics  []string          `json:"enabled_heuristics,omitempty"`
	DisabledHeuristics []string          `json:"disabled_heuristics,omitempty"`
	EmitGeometry       bool              `json:"emit_geometry,omitempty"`
	EmitDiagnostics    bool              `json:"emit_diagnostics,omitempty"`
	PreservePageBreaks bool              `json:"preserve_page_breaks,omitempty"`
	Strict             bool              `json:"strict,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	Extras             map[string]any    `json:"extras,omitempty"`
}

type ProcessingConfig struct {
	InputPath          string            `json:"input_path,omitempty"`
	OutputPath         string            `json:"output_path,omitempty"`
	DocumentName       string            `json:"document_name,omitempty"`
	RequestedFormats   []OutputFormat    `json:"requested_formats,omitempty"`
	EnabledHeuristics  []string          `json:"enabled_heuristics,omitempty"`
	EmitGeometry       bool              `json:"emit_geometry,omitempty"`
	EmitDiagnostics    bool              `json:"emit_diagnostics,omitempty"`
	PreservePageBreaks bool              `json:"preserve_page_breaks,omitempty"`
	Strict             bool              `json:"strict,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	Extras             map[string]any    `json:"extras,omitempty"`
}

func (o ProcessingOptions) Resolve() ProcessingConfig {
	cfg := ProcessingConfig{
		InputPath:          o.InputPath,
		OutputPath:         o.OutputPath,
		DocumentName:       o.DocumentName,
		RequestedFormats:   append([]OutputFormat(nil), o.RequestedFormats...),
		EnabledHeuristics:  append([]string(nil), o.EnabledHeuristics...),
		EmitGeometry:       o.EmitGeometry,
		EmitDiagnostics:    o.EmitDiagnostics,
		PreservePageBreaks: o.PreservePageBreaks,
		Strict:             o.Strict,
		Metadata:           cloneStringMap(o.Metadata),
		Extras:             cloneAnyMap(o.Extras),
	}
	if len(cfg.RequestedFormats) == 0 {
		cfg.RequestedFormats = []OutputFormat{OutputFormatJSON}
	}
	return cfg
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneAnyMap(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
