package nativepdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTextShellStringsFromSampleLoremPDF(t *testing.T) {
	data, err := os.ReadFile(filepath.Clean("../../..//samples/pdf/lorem.pdf"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	texts := extractTextShellStrings(data)
	if len(texts) == 0 {
		t.Fatal("extractTextShellStrings() returned no texts for sample lorem.pdf")
	}
}

func TestDocumentParserBuildsArtifactsFromSampleLoremPDF(t *testing.T) {
	data, err := os.ReadFile(filepath.Clean("../../..//samples/pdf/lorem.pdf"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	parser := NewDocumentParser()
	result, err := parser.Parse("lorem.pdf", data, OpenOptions{}, nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(result.Pages) == 0 {
		t.Fatal("len(result.Pages) = 0, want non-zero")
	}
	if len(result.Pages[0].Artifacts) == 0 {
		t.Fatal("len(result.Pages[0].Artifacts) = 0, want non-zero")
	}
}
