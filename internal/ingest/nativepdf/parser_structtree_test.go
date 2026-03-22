package nativepdf

import (
	"context"
	"strings"
	"testing"
)

func TestDocumentParserExtractsPageStructTree(t *testing.T) {
	pdf := `%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R /StructTreeRoot 5 0 R >>
endobj
2 0 obj
<< /Type /Pages /Count 1 /Kids [3 0 R] /MediaBox [0 0 400 500] >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R >>
endobj
5 0 obj
<< /Type /StructTreeRoot /K [6 0 R] >>
endobj
6 0 obj
<< /Type /StructElem /S /P /Pg 3 0 R /K [7 0 R 0] >>
endobj
7 0 obj
<< /Type /StructElem /S /Span /K 1 >>
endobj
`

	result, err := NewDocumentParser().Parse("tagged.pdf", []byte(pdf), OpenOptions{}, nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got, want := result.Phase, ParsePhaseStructured; got != want {
		t.Fatalf("result.Phase = %q, want %q", got, want)
	}
	tree := result.Pages[0].StructTree
	if tree == nil {
		t.Fatal("result.Pages[0].StructTree = nil, want tree")
	}
	if got, want := tree.Type, "StructTreeRoot"; got != want {
		t.Fatalf("tree.Type = %q, want %q", got, want)
	}
	if got := len(tree.Kids); got != 1 {
		t.Fatalf("len(tree.Kids) = %d, want 1", got)
	}
	if got, want := tree.Kids[0].Type, "P"; got != want {
		t.Fatalf("tree.Kids[0].Type = %q, want %q", got, want)
	}
	if tree.Kids[0].PageIndex == nil || *tree.Kids[0].PageIndex != 0 {
		t.Fatalf("tree.Kids[0].PageIndex = %#v, want page 0", tree.Kids[0].PageIndex)
	}
	if got := len(tree.Kids[0].Kids); got != 2 {
		t.Fatalf("len(tree.Kids[0].Kids) = %d, want 2", got)
	}
	if got, want := tree.Kids[0].Kids[0].Type, "Span"; got != want {
		t.Fatalf("tree.Kids[0].Kids[0].Type = %q, want %q", got, want)
	}
	if got, want := tree.Kids[0].Kids[0].Kids[0].MarkedContentIDs[0], 1; got != want {
		t.Fatalf("nested MarkedContentIDs[0] = %d, want %d", got, want)
	}
	if got, want := tree.Kids[0].Kids[1].MarkedContentIDs[0], 0; got != want {
		t.Fatalf("direct MarkedContentIDs[0] = %d, want %d", got, want)
	}
}

func TestDocumentParserBoundsStructTreeCycles(t *testing.T) {
	pdf := `%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R /StructTreeRoot 5 0 R >>
endobj
2 0 obj
<< /Type /Pages /Count 1 /Kids [3 0 R] /MediaBox [0 0 400 500] >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R >>
endobj
5 0 obj
<< /Type /StructTreeRoot /K [6 0 R] >>
endobj
6 0 obj
<< /Type /StructElem /S /P /Pg 3 0 R /K [6 0 R 0] >>
endobj
`

	result, err := NewDocumentParser().Parse("cyclic-tagged.pdf", []byte(pdf), OpenOptions{}, nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	tree := result.Pages[0].StructTree
	if tree == nil {
		t.Fatal("result.Pages[0].StructTree = nil, want bounded tree")
	}
	if got := len(tree.Kids); got != 1 {
		t.Fatalf("len(tree.Kids) = %d, want 1", got)
	}
	if got := len(tree.Kids[0].Kids); got != 1 {
		t.Fatalf("len(tree.Kids[0].Kids) = %d, want 1 after cycle pruning", got)
	}
	if got, want := tree.Kids[0].Kids[0].MarkedContentIDs[0], 0; got != want {
		t.Fatalf("MarkedContentIDs[0] = %d, want %d", got, want)
	}
}

func TestSkeletonLoaderExposesParsedStructTree(t *testing.T) {
	pdf := `%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R /StructTreeRoot 5 0 R >>
endobj
2 0 obj
<< /Type /Pages /Count 1 /Kids [3 0 R] /MediaBox [0 0 400 500] >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R >>
endobj
5 0 obj
<< /Type /StructTreeRoot /K [6 0 R] >>
endobj
6 0 obj
<< /Type /StructElem /S /P /Pg 3 0 R /K 0 >>
endobj
`

	handle, err := NewSkeletonLoader().OpenReader(context.Background(), "tagged.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	page, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}
	tree, err := page.StructTree(context.Background())
	if err != nil {
		t.Fatalf("page.StructTree() error = %v", err)
	}
	if tree == nil || len(tree.Kids) != 1 {
		t.Fatalf("page.StructTree() = %#v, want one parsed child", tree)
	}
	if got, want := tree.Kids[0].Kids[0].MarkedContentIDs[0], 0; got != want {
		t.Fatalf("MarkedContentIDs[0] = %d, want %d", got, want)
	}
}

func TestDocumentParserLinksStructTreeNodesToMarkedContentArtifacts(t *testing.T) {
	pdf := `%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R /StructTreeRoot 6 0 R >>
endobj
2 0 obj
<< /Type /Pages /Count 1 /Kids [3 0 R] >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 300] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>
endobj
4 0 obj
<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
endobj
5 0 obj
<< /Length 100 >>
stream
BT /F1 12 Tf
/P << /MCID 0 >> BDC
10 20 Td (Hello) Tj
EMC
/P << /MCID 1 >> BDC
0 -14 Td (World) Tj
EMC
ET
endstream
endobj
6 0 obj
<< /Type /StructTreeRoot /K [7 0 R] >>
endobj
7 0 obj
<< /Type /StructElem /S /P /Pg 3 0 R /K [8 0 R 9 0 R] >>
endobj
8 0 obj
<< /Type /MCR /Pg 3 0 R /MCID 0 >>
endobj
9 0 obj
<< /Type /MCR /Pg 3 0 R /MCID 1 >>
endobj
`

	result, err := NewDocumentParser().Parse("linked.pdf", []byte(pdf), OpenOptions{}, nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got, want := len(result.Pages), 1; got != want {
		t.Fatalf("len(result.Pages) = %d, want %d", got, want)
	}
	if got, want := len(result.Pages[0].Artifacts), 2; got != want {
		t.Fatalf("len(result.Pages[0].Artifacts) = %d, want %d", got, want)
	}
	first := result.Pages[0].Artifacts[0]
	second := result.Pages[0].Artifacts[1]
	if first.MarkedContentID == nil || *first.MarkedContentID != 0 {
		t.Fatalf("first.MarkedContentID = %#v, want 0", first.MarkedContentID)
	}
	if second.MarkedContentID == nil || *second.MarkedContentID != 1 {
		t.Fatalf("second.MarkedContentID = %#v, want 1", second.MarkedContentID)
	}

	tree := result.Pages[0].StructTree
	if tree == nil || len(tree.Kids) != 1 {
		t.Fatalf("result.Pages[0].StructTree = %#v, want one child", tree)
	}
	if got, want := tree.ArtifactIDs, []int64{1, 2}; len(got) != len(want) || int64(got[0]) != want[0] || int64(got[1]) != want[1] {
		t.Fatalf("tree.ArtifactIDs = %#v, want [1 2]", tree.ArtifactIDs)
	}
	if got, want := tree.Kids[0].ArtifactIDs, []int64{1, 2}; len(got) != len(want) || int64(got[0]) != want[0] || int64(got[1]) != want[1] {
		t.Fatalf("tree.Kids[0].ArtifactIDs = %#v, want [1 2]", tree.Kids[0].ArtifactIDs)
	}
	if got, want := tree.Kids[0].Kids[0].ArtifactIDs[0], first.ID; got != want {
		t.Fatalf("first MCR ArtifactIDs[0] = %d, want %d", got, want)
	}
	if got, want := tree.Kids[0].Kids[1].ArtifactIDs[0], second.ID; got != want {
		t.Fatalf("second MCR ArtifactIDs[0] = %d, want %d", got, want)
	}
}
