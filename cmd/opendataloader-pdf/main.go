package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/emit/jsonout"
	"github.com/guswns531/opendataloader-pdf-go/internal/emit/markdown"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/fixture"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
	"github.com/guswns531/opendataloader-pdf-go/internal/pipeline/local"
)

const version = "0.0.0-dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("opendataloader-pdf", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	outputDir := fs.String("output-dir", "", "Directory where output files are written.")
	format := fs.String("format", "json", "Output formats (comma-separated).")
	useFixture := fs.Bool("fixture", false, "Treat input JSON as a document fixture.")
	quiet := fs.Bool("quiet", false, "Suppress non-error output.")
	showVersion := fs.Bool("version", false, "Print version and exit.")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *showVersion {
		fmt.Println(version)
		return 0
	}

	inputs := fs.Args()
	if len(inputs) == 0 {
		printUsage(fs)
		return 2
	}

	options := core.ProcessingOptions{
		InputPath:        inputs[0],
		OutputPath:       *outputDir,
		DocumentName:     filepath.Base(inputs[0]),
		RequestedFormats: parseFormats(*format),
	}

	document := model.NewDocument(model.DocumentMetadata{
		FileName:  options.DocumentName,
		PageCount: 0,
	})
	ctx := core.NewProcessingContext(document, options)
	if *useFixture || strings.EqualFold(filepath.Ext(inputs[0]), ".json") {
		pipeline := local.New(fixture.New())
		document, err := pipeline.Run(ctx, core.Source{
			Path: inputs[0],
			Name: filepath.Base(inputs[0]),
		}, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fixture pipeline failed: %v\n", err)
			return 1
		}
		outputPaths, err := writeOutputs(document, ctx, options)
		if err != nil {
			fmt.Fprintf(os.Stderr, "writing outputs failed: %v\n", err)
			return 1
		}
		if !*quiet {
			fmt.Fprintf(os.Stderr, "opendataloader-pdf (pure go skeleton) version %s\n", version)
			fmt.Fprintf(os.Stderr, "pages=%d artifacts=%d nodes=%d stage=%s\n",
				len(document.Pages), countArtifacts(document), len(document.Kids), ctx.Stage)
			if len(outputPaths) > 0 {
				fmt.Fprintf(os.Stderr, "wrote=%s\n", strings.Join(outputPaths, ","))
			}
		}
		return 0
	}

	fmt.Fprintln(os.Stderr, "pure go core skeleton: PDF ingestion pipeline not implemented yet")
	return 1
}

func parseFormats(value string) []core.OutputFormat {
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

func printUsage(fs *flag.FlagSet) {
	fmt.Fprintln(os.Stderr, "Usage: opendataloader-pdf [options] <INPUT FILE OR FOLDER>...")
	fs.PrintDefaults()
}

func countArtifacts(document *model.Document) int {
	if document == nil {
		return 0
	}
	total := len(document.Artifacts)
	for _, page := range document.Pages {
		if page != nil {
			total += len(page.Artifacts)
		}
	}
	return total
}

func writeOutputs(document *model.Document, ctx *core.ProcessingContext, options core.ProcessingOptions) ([]string, error) {
	emitters, err := emittersForFormats(options.RequestedFormats)
	if err != nil {
		return nil, err
	}

	outputDir := options.OutputPath
	if outputDir == "" {
		outputDir = filepath.Dir(options.InputPath)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, err
	}

	baseName := strings.TrimSuffix(options.DocumentName, filepath.Ext(options.DocumentName))
	if baseName == "" {
		baseName = "output"
	}

	written := make([]string, 0, len(emitters))
	for _, emitter := range emitters {
		path := filepath.Join(outputDir, baseName+extensionForFormat(emitter.Format()))
		file, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		if emitErr := emitter.Emit(ctx, document, file); emitErr != nil {
			_ = file.Close()
			return nil, emitErr
		}
		if closeErr := file.Close(); closeErr != nil {
			return nil, closeErr
		}
		written = append(written, path)
	}

	return written, nil
}

func emittersForFormats(formats []core.OutputFormat) ([]core.Emitter, error) {
	emitters := make([]core.Emitter, 0, len(formats))
	for _, format := range formats {
		switch format {
		case core.OutputFormatJSON:
			emitters = append(emitters, jsonout.New())
		case core.OutputFormatMarkdown:
			emitters = append(emitters, markdown.New())
		case core.OutputFormatHTML, core.OutputFormatText:
			return nil, fmt.Errorf("format %q is not implemented in the pure go skeleton", format)
		default:
			return nil, fmt.Errorf("unsupported format %q", format)
		}
	}
	if len(emitters) == 0 {
		emitters = append(emitters, jsonout.New())
	}
	return emitters, nil
}

func extensionForFormat(format core.OutputFormat) string {
	switch format {
	case core.OutputFormatMarkdown:
		return ".md"
	case core.OutputFormatHTML:
		return ".html"
	case core.OutputFormatText:
		return ".txt"
	default:
		return ".json"
	}
}
