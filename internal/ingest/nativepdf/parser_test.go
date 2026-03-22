package nativepdf

import (
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestDocumentParserBuildsShellParseResult(t *testing.T) {
	parser := NewDocumentParser()

	result, err := parser.Parse("sample.pdf", []byte(onePageShellPDF), OpenOptions{}, nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.Phase != ParsePhaseShell {
		t.Fatalf("result.Phase = %q, want %q", result.Phase, ParsePhaseShell)
	}
	if result.Metadata.FileName != "sample.pdf" {
		t.Fatalf("result.Metadata.FileName = %q, want sample.pdf", result.Metadata.FileName)
	}
	if result.Metadata.PageCount != 1 {
		t.Fatalf("result.Metadata.PageCount = %d, want 1", result.Metadata.PageCount)
	}
	if len(result.Pages) != 1 {
		t.Fatalf("len(result.Pages) = %d, want 1", len(result.Pages))
	}
	if len(result.Pages[0].Artifacts) != 2 {
		t.Fatalf("len(result.Pages[0].Artifacts) = %d, want 2", len(result.Pages[0].Artifacts))
	}
	if got, want := result.Pages[0].Artifacts[0].Text, "Hello"; got != want {
		t.Fatalf("artifact text = %q, want %q", got, want)
	}
}

func TestDocumentParserRespectsRequestedPages(t *testing.T) {
	parser := NewDocumentParser()

	result, err := parser.Parse("sample.pdf", []byte(twoPageShellPDF), OpenOptions{
		Pages: []int{2},
	}, nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.Metadata.PageCount != 1 {
		t.Fatalf("result.Metadata.PageCount = %d, want 1", result.Metadata.PageCount)
	}
	if len(result.Pages) != 1 {
		t.Fatalf("len(result.Pages) = %d, want 1", len(result.Pages))
	}
	if got, want := result.Pages[0].Metadata.Number, model.PageNumber(2); got != want {
		t.Fatalf("page number = %d, want %d", got, want)
	}
}

func TestSkeletonLoaderOpenReaderFiltersRequestedPages(t *testing.T) {
	loader := NewSkeletonLoader()

	handle, err := loader.OpenReader(nil, "sample.pdf", strings.NewReader(twoPageShellPDF), OpenOptions{
		Pages: []int{2},
	})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	if got, want := handle.PageCount(), 1; got != want {
		t.Fatalf("handle.PageCount() = %d, want %d", got, want)
	}
	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}
	if got, want := page.Metadata().Number, model.PageNumber(2); got != want {
		t.Fatalf("page.Metadata().Number = %d, want %d", got, want)
	}
}
