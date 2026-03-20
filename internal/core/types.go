package core

import "io"

type DocumentID string

type OutputFormat string

const (
	OutputFormatJSON     OutputFormat = "json"
	OutputFormatMarkdown OutputFormat = "markdown"
	OutputFormatHTML     OutputFormat = "html"
	OutputFormatText     OutputFormat = "text"
)

type Stage string

const (
	StageIngestion   Stage = "ingestion"
	StageHeuristics  Stage = "heuristics"
	StageEmission    Stage = "emission"
	StagePostProcess Stage = "post_process"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Source struct {
	Path     string            `json:"path,omitempty"`
	Name     string            `json:"name,omitempty"`
	MIMEType string            `json:"mime_type,omitempty"`
	Reader   io.Reader         `json:"-"`
	Size     int64             `json:"size,omitempty"`
	Meta     map[string]string `json:"meta,omitempty"`
}

type Issue struct {
	Severity  Severity          `json:"severity"`
	Code      string            `json:"code,omitempty"`
	Message   string            `json:"message"`
	Stage     Stage             `json:"stage,omitempty"`
	Page      *PageNumber       `json:"page,omitempty"`
	ElementID *ElementID        `json:"element_id,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
	Data      map[string]any    `json:"data,omitempty"`
}
