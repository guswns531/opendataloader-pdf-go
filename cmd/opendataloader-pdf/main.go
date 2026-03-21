package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guswns531/opendataloader-pdf-go/internal/cli/discovery"
	clioptions "github.com/guswns531/opendataloader-pdf-go/internal/cli/options"
	"github.com/guswns531/opendataloader-pdf-go/internal/cli/pagerange"
	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/emit/htmlout"
	"github.com/guswns531/opendataloader-pdf-go/internal/emit/schemajson"
	"github.com/guswns531/opendataloader-pdf-go/internal/emit/semanticmd"
	"github.com/guswns531/opendataloader-pdf-go/internal/emit/textout"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/fixture"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/nativepdf"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/pdftext"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
	"github.com/guswns531/opendataloader-pdf-go/internal/pipeline/local"
	"github.com/guswns531/opendataloader-pdf-go/internal/pipeline/pagefilter"
)

const version = "0.0.0-dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	opts, err := clioptions.Parse(args)
	if err != nil {
		return 2
	}
	if opts.Version {
		fmt.Println(version)
		return 0
	}
	if len(opts.Inputs) == 0 {
		printUsage()
		return 2
	}

	inputs, err := discovery.Discover(opts.Inputs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "input discovery failed: %v\n", err)
		return 1
	}
	if len(inputs) == 0 {
		fmt.Fprintln(os.Stderr, "no supported input files found")
		return 1
	}

	pages, err := pagerange.ParseOptional(opts.Pages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "page parsing failed: %v\n", err)
		return 2
	}

	hadFailure := false
	for _, input := range inputs {
		if err := processInput(input, opts, pages); err != nil {
			hadFailure = true
			fmt.Fprintf(os.Stderr, "%s: %v\n", input, err)
		}
	}
	if hadFailure {
		return 1
	}
	return 0
}

func pipelineForInput(path string, useFixture bool) (*local.Pipeline, error) {
	switch {
	case strings.HasSuffix(strings.ToLower(path), ".raw.json"):
		return local.New(nativepdf.NewIngestor(nativepdf.NewFixtureLoader())), nil
	case useFixture, strings.EqualFold(filepath.Ext(path), ".json"):
		return local.New(fixture.New()), nil
	case strings.EqualFold(filepath.Ext(path), ".pdf"):
		return local.New(defaultPDFIngestor()), nil
	default:
		return nil, fmt.Errorf("unsupported input type for %q", path)
	}
}

func defaultPDFIngestor() core.Ingestor {
	// Temporary bridge until the nativepdf package grows a real parser backend.
	return pdftext.New()
}

func processInput(input string, opts clioptions.Options, pages []int) error {
	options := opts.ProcessingOptions(input)
	if len(pages) > 0 {
		if options.Extras == nil {
			options.Extras = make(map[string]any)
		}
		options.Extras["pages"] = append([]int(nil), pages...)
	}

	document := model.NewDocument(model.DocumentMetadata{
		FileName:  options.DocumentName,
		PageCount: 0,
	})
	ctx := core.NewProcessingContext(document, options)
	pipeline, err := pipelineForInput(input, opts.Fixture)
	if err != nil {
		return err
	}
	document, err = pipeline.Run(ctx, core.Source{
		Path: input,
		Name: filepath.Base(input),
	}, nil)
	if err != nil {
		return fmt.Errorf("pipeline failed: %w", err)
	}
	if len(pages) > 0 {
		if err := pagefilter.Apply(document, toPageNumbers(pages)); err != nil {
			return fmt.Errorf("page filtering failed: %w", err)
		}
	}
	outputPaths, err := writeOutputs(document, ctx, options)
	if err != nil {
		return fmt.Errorf("writing outputs failed: %w", err)
	}
	if !opts.Quiet {
		fmt.Fprintf(os.Stderr, "opendataloader-pdf (pure go skeleton) version %s\n", version)
		fmt.Fprintf(os.Stderr, "input=%s pages=%d artifacts=%d nodes=%d stage=%s\n",
			input, len(document.Pages), countArtifacts(document), len(document.Kids), ctx.Stage)
		if len(outputPaths) > 0 {
			fmt.Fprintf(os.Stderr, "wrote=%s\n", strings.Join(outputPaths, ","))
		}
	}
	return nil
}

func printUsage() {
	fs, _ := clioptions.NewFlagSet("opendataloader-pdf")
	fs.SetOutput(os.Stderr)
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
			emitters = append(emitters, schemajson.New())
		case core.OutputFormatMarkdown:
			emitters = append(emitters, semanticmd.New())
		case core.OutputFormatHTML:
			emitters = append(emitters, htmlout.New())
		case core.OutputFormatText:
			emitters = append(emitters, textout.New())
		default:
			return nil, fmt.Errorf("unsupported format %q", format)
		}
	}
	if len(emitters) == 0 {
		emitters = append(emitters, schemajson.New())
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

func toPageNumbers(pages []int) []model.PageNumber {
	out := make([]model.PageNumber, 0, len(pages))
	for _, page := range pages {
		out = append(out, model.PageNumber(page))
	}
	return out
}
