package pdftext

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
	pdf "github.com/ledongthuc/pdf"
)

// Ingestor extracts raw text artifacts from digital PDFs using ledongthuc/pdf.
type Ingestor struct{}

// New returns a PDF text ingestor.
func New() *Ingestor {
	return &Ingestor{}
}

// Name identifies the ingestor.
func (i *Ingestor) Name() string {
	return "pdftext"
}

// Ingest reads a PDF and converts page text into raw artifacts.
func (i *Ingestor) Ingest(ctx *core.ProcessingContext, source core.Source) (*core.Document, error) {
	_ = ctx

	path, cleanup, err := resolveSourcePath(source)
	if err != nil {
		return nil, err
	}
	if cleanup != nil {
		defer cleanup()
	}

	file, reader, err := pdf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open pdf: %w", err)
	}
	defer file.Close()

	document := model.NewDocument(model.DocumentMetadata{
		FileName:  resolveFileName(source, path),
		PageCount: reader.NumPage(),
	})

	sequence := 0
	totalPages := reader.NumPage()
	for pageNumber := 1; pageNumber <= totalPages; pageNumber++ {
		page := reader.Page(pageNumber)
		pageModel := &model.Page{
			Metadata: model.PageMetadata{
				Index:  model.PageIndex(pageNumber - 1),
				Number: model.PageNumber(pageNumber),
			},
			Artifacts: make([]*model.RawArtifact, 0),
			Kids:      make([]model.ContentElement, 0),
		}

		if page.V.IsNull() || page.V.Key("Contents").Kind() == pdf.Null {
			document.AddPage(pageModel)
			continue
		}

		content := page.Content()
		for _, text := range content.Text {
			if text.S == "" {
				continue
			}

			artifact := &model.RawArtifact{
				ID:         document.NewArtifactID(),
				Kind:       model.ArtifactKindText,
				PageIndex:  model.PageIndex(pageNumber - 1),
				PageNumber: model.PageNumber(pageNumber),
				Bounds: model.Box{
					Left:   text.X,
					Bottom: text.Y,
					Right:  text.X + text.W,
					Top:    text.Y + text.FontSize,
				}.Normalize(),
				Text:     text.S,
				Sequence: sequence,
				Style: model.TextProperties{
					Font:     text.Font,
					FontSize: text.FontSize,
					Content:  text.S,
				},
			}
			sequence++
			pageModel.Artifacts = append(pageModel.Artifacts, artifact)
		}

		document.AddPage(pageModel)
	}

	return document, nil
}

func resolveSourcePath(source core.Source) (string, func(), error) {
	if source.Path != "" {
		return source.Path, nil, nil
	}
	if source.Reader == nil {
		return "", nil, fmt.Errorf("pdftext ingest requires source path or reader")
	}

	file, err := os.CreateTemp("", "opendataloader-pdf-*.pdf")
	if err != nil {
		return "", nil, err
	}
	if _, err := io.Copy(file, source.Reader); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return "", nil, err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(file.Name())
		return "", nil, err
	}

	cleanup := func() {
		_ = os.Remove(file.Name())
	}
	return file.Name(), cleanup, nil
}

func resolveFileName(source core.Source, path string) string {
	if source.Name != "" {
		return source.Name
	}
	return filepath.Base(path)
}
