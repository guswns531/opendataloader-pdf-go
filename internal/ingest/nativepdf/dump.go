package nativepdf

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

const rawDumpEnv = "OPENDATALOADER_GO_RAW_DUMP"

func writeRawDumpIfRequested(document *model.Document) error {
	if document == nil {
		return nil
	}
	path := os.Getenv(rawDumpEnv)
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	payload := dumpDocument{
		Metadata: dumpDocumentMetadata{
			FileName:  document.Metadata.FileName,
			PageCount: document.Metadata.PageCount,
		},
		Pages: make([]dumpPage, 0, len(document.Pages)),
	}
	for _, page := range document.Pages {
		if page == nil {
			continue
		}
		item := dumpPage{
			Metadata: dumpPageMetadata{
				Index:    page.Metadata.Index,
				Number:   page.Metadata.Number,
				Width:    page.Metadata.Size.Width,
				Height:   page.Metadata.Size.Height,
				Rotation: page.Metadata.Rotation,
				Bounds:   page.Metadata.Bounds,
			},
			Artifacts: make([]dumpArtifact, 0, len(page.Artifacts)),
		}
		for _, artifact := range page.Artifacts {
			if artifact == nil {
				continue
			}
			item.Artifacts = append(item.Artifacts, dumpArtifact{
				Kind:       artifact.Kind,
				PageIndex:  artifact.PageIndex,
				PageNumber: artifact.PageNumber,
				Sequence:   artifact.Sequence,
				Bounds:     artifact.Bounds,
				Boxes:      append(model.MultiBox(nil), artifact.Boxes...),
				Text:       artifact.Text,
				Format:     artifact.Format,
				Style: dumpTextStyle{
					Font:       artifact.Style.Font,
					FontSize:   artifact.Style.FontSize,
					TextColor:  artifact.Style.TextColor,
					Content:    artifact.Style.Content,
					HiddenText: artifact.Style.HiddenText,
					Bold:       artifact.Style.Bold,
					Italic:     artifact.Style.Italic,
					Underline:  artifact.Style.Underline,
				},
			})
		}
		payload.Pages = append(payload.Pages, item)
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

type dumpDocument struct {
	Metadata dumpDocumentMetadata `json:"metadata"`
	Pages    []dumpPage           `json:"pages"`
}

type dumpDocumentMetadata struct {
	FileName  string `json:"file_name"`
	PageCount int    `json:"page_count"`
}

type dumpPage struct {
	Metadata  dumpPageMetadata `json:"metadata"`
	Artifacts []dumpArtifact   `json:"artifacts"`
}

type dumpPageMetadata struct {
	Index    model.PageIndex  `json:"index"`
	Number   model.PageNumber `json:"number"`
	Width    float64          `json:"width,omitempty"`
	Height   float64          `json:"height,omitempty"`
	Rotation int              `json:"rotation,omitempty"`
	Bounds   model.Box        `json:"bounds,omitempty"`
}

type dumpArtifact struct {
	Kind       model.ArtifactKind `json:"kind"`
	PageIndex  model.PageIndex    `json:"page_index"`
	PageNumber model.PageNumber   `json:"page_number"`
	Sequence   int                `json:"sequence"`
	Bounds     model.Box          `json:"bounds,omitempty"`
	Boxes      model.MultiBox     `json:"boxes,omitempty"`
	Text       string             `json:"text,omitempty"`
	Format     model.ImageFormat  `json:"format,omitempty"`
	Style      dumpTextStyle      `json:"style,omitempty"`
}

type dumpTextStyle struct {
	Font       string  `json:"font,omitempty"`
	FontSize   float64 `json:"font_size,omitempty"`
	TextColor  string  `json:"text_color,omitempty"`
	Content    string  `json:"content,omitempty"`
	HiddenText bool    `json:"hidden_text,omitempty"`
	Bold       bool    `json:"bold,omitempty"`
	Italic     bool    `json:"italic,omitempty"`
	Underline  bool    `json:"underline,omitempty"`
}
