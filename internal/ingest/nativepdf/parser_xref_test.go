package nativepdf

import (
	"fmt"
	"strings"
	"testing"
)

func TestFindStartXrefAndParseXrefInfo(t *testing.T) {
	pdf, xrefOffset := buildXrefPDF()

	gotOffset, ok := findStartXref([]byte(pdf))
	if !ok {
		t.Fatal("findStartXref() = false, want true")
	}
	if gotOffset != xrefOffset {
		t.Fatalf("findStartXref() = %d, want %d", gotOffset, xrefOffset)
	}

	info, ok := parseXrefInfo([]byte(pdf))
	if !ok {
		t.Fatal("parseXrefInfo() = false, want true")
	}
	if info.StartXref != xrefOffset {
		t.Fatalf("info.StartXref = %d, want %d", info.StartXref, xrefOffset)
	}
	if got, want := len(info.Entries), 4; got != want {
		t.Fatalf("len(info.Entries) = %d, want %d", got, want)
	}
	root, ok := xrefInfoRoot(info)
	if !ok {
		t.Fatal("xrefInfoRoot() = false, want true")
	}
	if got, want := root.ObjectNumber, 1; got != want {
		t.Fatalf("root.ObjectNumber = %d, want %d", got, want)
	}
	if got, want := root.Generation, 0; got != want {
		t.Fatalf("root.Generation = %d, want %d", got, want)
	}
}

func TestParsePageTreeMetadataUsesXrefEntries(t *testing.T) {
	pdf, _ := buildXrefPDF()

	pages, ok := parsePageTreeMetadata([]byte(pdf))
	if !ok {
		t.Fatal("parsePageTreeMetadata() = false, want true")
	}
	if got, want := len(pages), 1; got != want {
		t.Fatalf("len(pages) = %d, want %d", got, want)
	}
	if got, want := pages[0].Size.Width, 400.0; got != want {
		t.Fatalf("pages[0].Size.Width = %v, want %v", got, want)
	}
	if got, want := pages[0].Size.Height, 500.0; got != want {
		t.Fatalf("pages[0].Size.Height = %v, want %v", got, want)
	}
}

func TestDocumentParserPromotesPhaseWithXrefTable(t *testing.T) {
	pdf, _ := buildXrefPDF()

	parser := NewDocumentParser()
	result, err := parser.Parse("xref.pdf", []byte(pdf), OpenOptions{}, nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got, want := result.Phase, ParsePhaseContainer; got != want {
		t.Fatalf("result.Phase = %q, want %q", got, want)
	}
	if got, want := result.Metadata.PageCount, 1; got != want {
		t.Fatalf("result.Metadata.PageCount = %d, want %d", got, want)
	}
}

func buildXrefPDF() (string, int) {
	var b strings.Builder
	write := func(s string) {
		_, _ = b.WriteString(s)
	}
	write("%PDF-1.7\n")

	obj1Offset := b.Len()
	write("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	obj2Offset := b.Len()
	write("2 0 obj\n<< /Type /Pages /Count 1 /Kids [3 0 R] /MediaBox [0 0 400 500] >>\nendobj\n")
	obj3Offset := b.Len()
	write("3 0 obj\n<< /Type /Page /Parent 2 0 R >>\nendobj\n")

	xrefOffset := b.Len()
	write("xref\n")
	write("0 4\n")
	write(fmt.Sprintf("%010d %05d f \n", 0, 65535))
	write(fmt.Sprintf("%010d %05d n \n", obj1Offset, 0))
	write(fmt.Sprintf("%010d %05d n \n", obj2Offset, 0))
	write(fmt.Sprintf("%010d %05d n \n", obj3Offset, 0))
	write("trailer\n")
	write("<< /Size 4 /Root 1 0 R >>\n")
	write("startxref\n")
	write(fmt.Sprintf("%d\n", xrefOffset))
	write("%%EOF\n")

	return b.String(), xrefOffset
}
