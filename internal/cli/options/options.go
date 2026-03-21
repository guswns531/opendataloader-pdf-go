package options

import (
	"flag"
	"io"
	"path/filepath"
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
)

const (
	defaultFormat              = "json"
	defaultReplaceInvalidChars = " "
	defaultTableMethod         = "default"
	defaultReadingOrder        = "xycut"
	defaultImageOutput         = "external"
	defaultImageFormat         = "png"
	defaultHybrid              = "off"
	defaultHybridMode          = "auto"
	defaultHybridTimeout       = "30000"
)

// Options captures the CLI flags we can parse today, plus parity placeholders
// that are recognized but not yet consumed by the Go skeleton.
type Options struct {
	OutputDir string
	Format    string
	Quiet     bool
	Fixture   bool
	Version   bool

	Password              string
	ContentSafetyOff      string
	Sanitize              bool
	KeepLineBreaks        bool
	ReplaceInvalidChars   string
	UseStructTree         bool
	TableMethod           string
	ReadingOrder          string
	MarkdownPageSeparator string
	TextPageSeparator     string
	HTMLPageSeparator     string
	ImageOutput           string
	ImageFormat           string
	ImageDir              string
	Pages                 string
	IncludeHeaderFooter   bool
	DetectStrikethrough   bool
	Hybrid                string
	HybridMode            string
	HybridURL             string
	HybridTimeout         string
	HybridFallback        bool

	Inputs []string
}

// NewFlagSet returns a flag set with the current skeleton flags and the parity
// placeholders already registered.
func NewFlagSet(name string) (*flag.FlagSet, *Options) {
	opts := defaultOptions()

	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.StringVar(&opts.OutputDir, "output-dir", opts.OutputDir, "Directory where output files are written. Default: input file directory")
	fs.StringVar(&opts.OutputDir, "o", opts.OutputDir, "Directory where output files are written. Default: input file directory")
	fs.StringVar(&opts.Password, "password", opts.Password, "Password for encrypted PDF files")
	fs.StringVar(&opts.Password, "p", opts.Password, "Password for encrypted PDF files")
	fs.StringVar(&opts.Format, "format", opts.Format, "Output formats (comma-separated). Values: json, text, html, pdf, markdown, markdown-with-html, markdown-with-images. Default: json")
	fs.StringVar(&opts.Format, "f", opts.Format, "Output formats (comma-separated). Values: json, text, html, pdf, markdown, markdown-with-html, markdown-with-images. Default: json")
	fs.BoolVar(&opts.Quiet, "quiet", opts.Quiet, "Suppress console logging output")
	fs.BoolVar(&opts.Quiet, "q", opts.Quiet, "Suppress console logging output")
	fs.BoolVar(&opts.Fixture, "fixture", opts.Fixture, "Treat input JSON as a document fixture.")
	fs.BoolVar(&opts.Version, "version", opts.Version, "Print version and exit.")

	fs.StringVar(&opts.ContentSafetyOff, "content-safety-off", opts.ContentSafetyOff, "Disable content safety filters. Values: all, hidden-text, off-page, tiny, hidden-ocg")
	fs.BoolVar(&opts.Sanitize, "sanitize", opts.Sanitize, "Enable sensitive data sanitization. Replaces emails, phone numbers, IPs, credit cards, and URLs with placeholders")
	fs.BoolVar(&opts.KeepLineBreaks, "keep-line-breaks", opts.KeepLineBreaks, "Preserve original line breaks in extracted text")
	fs.StringVar(&opts.ReplaceInvalidChars, "replace-invalid-chars", opts.ReplaceInvalidChars, "Replacement character for invalid/unrecognized characters. Default: space")
	fs.BoolVar(&opts.UseStructTree, "use-struct-tree", opts.UseStructTree, "Use PDF structure tree (tagged PDF) for reading order and semantic structure")
	fs.StringVar(&opts.TableMethod, "table-method", opts.TableMethod, "Table detection method. Values: default (border-based), cluster (border + cluster). Default: default")
	fs.StringVar(&opts.ReadingOrder, "reading-order", opts.ReadingOrder, "Reading order algorithm. Values: off, xycut. Default: xycut")
	fs.StringVar(&opts.MarkdownPageSeparator, "markdown-page-separator", opts.MarkdownPageSeparator, "Separator between pages in Markdown output. Use %page-number% for page numbers. Default: none")
	fs.StringVar(&opts.TextPageSeparator, "text-page-separator", opts.TextPageSeparator, "Separator between pages in text output. Use %page-number% for page numbers. Default: none")
	fs.StringVar(&opts.HTMLPageSeparator, "html-page-separator", opts.HTMLPageSeparator, "Separator between pages in HTML output. Use %page-number% for page numbers. Default: none")
	fs.StringVar(&opts.ImageOutput, "image-output", opts.ImageOutput, "Image output mode. Values: off (no images), embedded (Base64 data URIs), external (file references). Default: external")
	fs.StringVar(&opts.ImageFormat, "image-format", opts.ImageFormat, "Output format for extracted images. Values: png, jpeg. Default: png")
	fs.StringVar(&opts.ImageDir, "image-dir", opts.ImageDir, "Directory for extracted images")
	fs.StringVar(&opts.Pages, "pages", opts.Pages, "Pages to extract (e.g., \"1,3,5-7\"). Default: all pages")
	fs.BoolVar(&opts.IncludeHeaderFooter, "include-header-footer", opts.IncludeHeaderFooter, "Include page headers and footers in output")
	fs.BoolVar(&opts.DetectStrikethrough, "detect-strikethrough", opts.DetectStrikethrough, "Detect strikethrough text and wrap with ~~ in Markdown output (experimental)")
	fs.StringVar(&opts.Hybrid, "hybrid", opts.Hybrid, "Hybrid backend for AI processing. Values: off (default), docling-fast")
	fs.StringVar(&opts.HybridMode, "hybrid-mode", opts.HybridMode, "Hybrid triage mode. Values: auto (default, dynamic triage), full (skip triage, all pages to backend)")
	fs.StringVar(&opts.HybridURL, "hybrid-url", opts.HybridURL, "Hybrid backend server URL (overrides default)")
	fs.StringVar(&opts.HybridTimeout, "hybrid-timeout", opts.HybridTimeout, "Hybrid backend request timeout in milliseconds. Default: 30000")
	fs.BoolVar(&opts.HybridFallback, "hybrid-fallback", opts.HybridFallback, "Opt in to Java fallback on hybrid backend error (default: disabled)")

	return fs, &opts
}

// Parse parses args into an Options value and preserves positional inputs.
func Parse(args []string) (Options, error) {
	fs, opts := NewFlagSet("opendataloader-pdf")
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	opts.Inputs = append([]string(nil), fs.Args()...)
	return *opts, nil
}

// ParseFormats normalizes a comma-separated format list into supported output formats.
func ParseFormats(value string) []core.OutputFormat {
	if strings.TrimSpace(value) == "" {
		return []core.OutputFormat{core.OutputFormatJSON}
	}

	parts := strings.Split(value, ",")
	formats := make([]core.OutputFormat, 0, len(parts))
	for _, part := range parts {
		switch normalized := strings.TrimSpace(part); normalized {
		case "":
		case string(core.OutputFormatJSON):
			formats = append(formats, core.OutputFormatJSON)
		case string(core.OutputFormatMarkdown):
			formats = append(formats, core.OutputFormatMarkdown)
		case string(core.OutputFormatHTML):
			formats = append(formats, core.OutputFormatHTML)
		case string(core.OutputFormatText):
			formats = append(formats, core.OutputFormatText)
		}
	}
	if len(formats) == 0 {
		return []core.OutputFormat{core.OutputFormatJSON}
	}
	return formats
}

// ProcessingOptions builds the core processing options used by the pipeline.
func (o Options) ProcessingOptions(inputPath string) core.ProcessingOptions {
	documentName := ""
	if inputPath != "" {
		documentName = filepath.Base(inputPath)
	}

	extras := map[string]any{
		"sanitize":              o.Sanitize,
		"replace_invalid":       o.ReplaceInvalidChars,
		"content_safety_off":    o.ContentSafetyOff,
		"include_header_footer": o.IncludeHeaderFooter,
	}

	return core.ProcessingOptions{
		InputPath:        inputPath,
		OutputPath:       o.OutputDir,
		DocumentName:     documentName,
		RequestedFormats: ParseFormats(o.Format),
		Extras:           extras,
	}
}

func defaultOptions() Options {
	return Options{
		Format:              defaultFormat,
		ReplaceInvalidChars: defaultReplaceInvalidChars,
		TableMethod:         defaultTableMethod,
		ReadingOrder:        defaultReadingOrder,
		ImageOutput:         defaultImageOutput,
		ImageFormat:         defaultImageFormat,
		Hybrid:              defaultHybrid,
		HybridMode:          defaultHybridMode,
		HybridTimeout:       defaultHybridTimeout,
	}
}
