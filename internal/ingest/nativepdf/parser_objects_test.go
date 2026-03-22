package nativepdf

import "testing"

func TestParsePageTreeMetadataUsesCatalogAndInheritedBoxes(t *testing.T) {
	pdf := `%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Count 2 /Kids [3 0 R 4 0 R] /MediaBox [0 0 612 792] >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R >>
endobj
4 0 obj
<< /Type /Page /Parent 2 0 R /CropBox [10 20 210 320] /Rotate 90 >>
endobj
`

	pages, ok := parsePageTreeMetadata([]byte(pdf))
	if !ok {
		t.Fatal("parsePageTreeMetadata() = false, want true")
	}
	if got, want := len(pages), 2; got != want {
		t.Fatalf("len(pages) = %d, want %d", got, want)
	}
	if got, want := pages[0].Size.Width, 612.0; got != want {
		t.Fatalf("pages[0].Size.Width = %v, want %v", got, want)
	}
	if got, want := pages[0].Size.Height, 792.0; got != want {
		t.Fatalf("pages[0].Size.Height = %v, want %v", got, want)
	}
	if got, want := pages[1].Bounds.Left, 10.0; got != want {
		t.Fatalf("pages[1].Bounds.Left = %v, want %v", got, want)
	}
	if got, want := pages[1].Bounds.Bottom, 20.0; got != want {
		t.Fatalf("pages[1].Bounds.Bottom = %v, want %v", got, want)
	}
	if got, want := pages[1].Size.Width, 200.0; got != want {
		t.Fatalf("pages[1].Size.Width = %v, want %v", got, want)
	}
	if got, want := pages[1].Size.Height, 300.0; got != want {
		t.Fatalf("pages[1].Size.Height = %v, want %v", got, want)
	}
	if got, want := pages[1].Rotation, 90; got != want {
		t.Fatalf("pages[1].Rotation = %d, want %d", got, want)
	}
}

func TestDocumentParserPromotesPhaseWhenPageTreeParses(t *testing.T) {
	pdf := `%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Count 1 /Kids [3 0 R] /MediaBox [0 0 400 500] >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R >>
endobj
`

	parser := NewDocumentParser()
	result, err := parser.Parse("page-tree.pdf", []byte(pdf), OpenOptions{}, nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got, want := result.Phase, ParsePhaseContainer; got != want {
		t.Fatalf("result.Phase = %q, want %q", got, want)
	}
	if got, want := result.Pages[0].Metadata.Size.Width, 400.0; got != want {
		t.Fatalf("page width = %v, want %v", got, want)
	}
	if got, want := result.Pages[0].Metadata.Size.Height, 500.0; got != want {
		t.Fatalf("page height = %v, want %v", got, want)
	}
}
