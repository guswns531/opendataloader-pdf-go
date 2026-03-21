package nativepdf

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
)

func TestFixtureLoaderOpenReaderAndFilterArtifacts(t *testing.T) {
	loader := NewFixtureLoader()
	handle, err := loader.OpenReader(context.Background(), "fixture.json", strings.NewReader(sampleRawFixture), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	if handle.PageCount() != 2 {
		t.Fatalf("handle.PageCount() = %d, want 2", handle.PageCount())
	}
	if handle.Metadata().FileName != "fixture.json" {
		t.Fatalf("handle.Metadata().FileName = %q, want fixture.json", handle.Metadata().FileName)
	}

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}
	textArtifacts, err := page.Artifacts(context.Background(), ArtifactOptions{IncludeText: true})
	if err != nil {
		t.Fatalf("page.Artifacts(text) error = %v", err)
	}
	if len(textArtifacts) != 1 {
		t.Fatalf("len(textArtifacts) = %d, want 1", len(textArtifacts))
	}
	if textArtifacts[0].Text != "Hello" {
		t.Fatalf("textArtifacts[0].Text = %q, want Hello", textArtifacts[0].Text)
	}

	lineArtifacts, err := page.Artifacts(context.Background(), ArtifactOptions{IncludeLine: true})
	if err != nil {
		t.Fatalf("page.Artifacts(line) error = %v", err)
	}
	if len(lineArtifacts) != 1 {
		t.Fatalf("len(lineArtifacts) = %d, want 1", len(lineArtifacts))
	}
	if lineArtifacts[0].Kind != "line" {
		t.Fatalf("lineArtifacts[0].Kind = %q, want line", lineArtifacts[0].Kind)
	}
}

func TestFixtureLoaderOpenPathAndSelectPages(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "raw.json")
	if err := os.WriteFile(path, []byte(sampleRawFixture), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	loader := NewFixtureLoader()
	handle, err := loader.OpenPath(context.Background(), path, OpenOptions{Pages: []int{2}})
	if err != nil {
		t.Fatalf("OpenPath() error = %v", err)
	}
	defer handle.Close()

	if handle.PageCount() != 1 {
		t.Fatalf("handle.PageCount() = %d, want 1", handle.PageCount())
	}
	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}
	if page.Metadata().Number != 2 {
		t.Fatalf("page.Metadata().Number = %d, want 2", page.Metadata().Number)
	}
}

func TestFixtureLoaderRejectsNonContiguousSequences(t *testing.T) {
	loader := NewFixtureLoader()
	_, err := loader.OpenReader(context.Background(), "broken.json", strings.NewReader(`{
	  "metadata": {"file_name": "broken.json", "page_count": 1},
	  "pages": [
	    {
	      "metadata": {"number": 1, "index": 0},
	      "artifacts": [
	        {"kind": "text", "page_index": 0, "page_number": 1, "sequence": 1, "text": "oops", "bounds": {"left": 0, "bottom": 0, "right": 10, "top": 10}}
	      ]
	    }
	  ]
	}`), OpenOptions{})
	if err == nil {
		t.Fatal("OpenReader() error = nil, want error")
	}
}

func TestFixtureLoaderComposesWithNativeIngestor(t *testing.T) {
	ingestor := NewIngestor(NewFixtureLoader())
	document, err := ingestor.Ingest(nil, core.Source{
		Name:   "fixture.json",
		Reader: strings.NewReader(sampleRawFixture),
	})
	if err != nil {
		t.Fatalf("Ingest() error = %v", err)
	}

	if document.Metadata.FileName != "fixture.json" {
		t.Fatalf("document.Metadata.FileName = %q, want fixture.json", document.Metadata.FileName)
	}
	if len(document.Pages) != 2 {
		t.Fatalf("len(document.Pages) = %d, want 2", len(document.Pages))
	}
	if len(document.Pages[0].Artifacts) != 2 {
		t.Fatalf("len(document.Pages[0].Artifacts) = %d, want 2", len(document.Pages[0].Artifacts))
	}
	if document.Pages[0].Artifacts[0].ID == 0 {
		t.Fatal("document.Pages[0].Artifacts[0].ID = 0, want non-zero")
	}
}

const sampleRawFixture = `{
  "metadata": {
    "file_name": "fixture.json",
    "page_count": 2,
    "language": "en"
  },
  "pages": [
    {
      "metadata": {
        "number": 1,
        "index": 0,
        "width": 612,
        "height": 792
      },
      "artifacts": [
        {
          "kind": "text",
          "page_index": 0,
          "page_number": 1,
          "sequence": 0,
          "text": "Hello",
          "bounds": {"left": 0, "bottom": 90, "right": 40, "top": 100},
          "style": {"font": "Times", "font_size": 12, "content": "Hello"}
        },
        {
          "kind": "line",
          "page_index": 0,
          "page_number": 1,
          "sequence": 1,
          "bounds": {"left": 0, "bottom": 80, "right": 100, "top": 81}
        }
      ]
    },
    {
      "metadata": {
        "number": 2,
        "index": 1,
        "width": 612,
        "height": 792
      },
      "artifacts": [
        {
          "kind": "text",
          "page_index": 1,
          "page_number": 2,
          "sequence": 0,
          "text": "World",
          "bounds": {"left": 0, "bottom": 90, "right": 40, "top": 100},
          "style": {"font": "Times", "font_size": 12, "content": "World"}
        }
      ]
    }
  ]
}`
