package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
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
		if !*quiet {
			fmt.Fprintf(os.Stderr, "opendataloader-pdf (pure go skeleton) version %s\n", version)
			fmt.Fprintf(os.Stderr, "pages=%d artifacts=%d nodes=%d stage=%s\n",
				len(document.Pages), countArtifacts(document), len(document.Kids), ctx.Stage)
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
