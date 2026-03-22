package nativepdf

import (
	"strings"
	"testing"
)

func TestSkeletonLoaderUsesPageReferencedContentStreams(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Count 2 /Kids [3 0 R 4 0 R] >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Contents 6 0 R >> endobj
4 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Contents 7 0 R >> endobj
5 0 obj << /Length 17 >> stream
BT (Noise) Tj ET
endstream
endobj
6 0 obj << /Length 19 >> stream
BT (PageOne) Tj ET
endstream
endobj
7 0 obj << /Length 19 >> stream
BT (PageTwo) Tj ET
endstream
endobj
`

	loader := NewSkeletonLoader()
	handle, err := loader.OpenReader(nil, "sample.pdf", strings.NewReader(pdf), OpenOptions{})
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer handle.Close()

	if got, want := handle.PageCount(), 2; got != want {
		t.Fatalf("handle.PageCount() = %d, want %d", got, want)
	}

	plans, ok := parsePagePlans([]byte(pdf))
	if !ok {
		t.Fatal("parsePagePlans() = false, want true")
	}
	if got, want := len(plans), 2; got != want {
		t.Fatalf("len(plans) = %d, want %d", got, want)
	}
	if got, want := len(plans[0].ContentRefs), 1; got != want {
		t.Fatalf("len(plans[0].ContentRefs) = %d, want %d", got, want)
	}
	if got, want := plans[0].ContentRefs[0].ObjectNumber, 6; got != want {
		t.Fatalf("plans[0].ContentRefs[0].ObjectNumber = %d, want %d", got, want)
	}

	graph, err := scanIndirectObjects([]byte(pdf))
	if err != nil {
		t.Fatalf("scanIndirectObjects() error = %v", err)
	}
	if got, want := len(graph.objects), 7; got != want {
		t.Fatalf("len(graph.objects) = %d, want %d", got, want)
	}
	if object := graph.objects[plans[0].ContentRefs[0]]; object == nil {
		t.Fatal("graph.objects[plans[0].ContentRefs[0]] = nil")
	} else if got := string(object.Raw); !strings.Contains(got, "stream") {
		t.Fatalf("object.Raw missing stream marker: %q", got)
	}
	streams := graph.decodedStreamsForRefs(plans[0].ContentRefs)
	if got, want := len(streams), 1; got != want {
		t.Fatalf("len(streams) = %d, want %d", got, want)
	}
	if got, want := string(streams[0]), "BT (PageOne) Tj ET"; got != want {
		t.Fatalf("stream[0] = %q, want %q", got, want)
	}
	fragments := extractTextFragmentsFromStreams(streams, nil)
	if got, want := len(fragments), 1; got != want {
		t.Fatalf("len(fragments) = %d, want %d", got, want)
	}

	plannedArtifacts := shellArtifactsFromPagePlans([]byte(pdf), plans)
	if plannedArtifacts == nil {
		t.Fatal("shellArtifactsFromPagePlans() = nil, want page-specific artifacts")
	}
	if got, want := len(plannedArtifacts[0]), 1; got != want {
		t.Fatalf("len(plannedArtifacts[0]) = %d, want %d", got, want)
	}
	if got, want := plannedArtifacts[0][0].Text, "PageOne"; got != want {
		t.Fatalf("plannedArtifacts[0][0].Text = %q, want %q", got, want)
	}

	page1, err := handle.Page(0)
	if err != nil {
		t.Fatalf("handle.Page(0) error = %v", err)
	}
	artifacts1, err := page1.Artifacts(nil, ArtifactOptions{IncludeText: true})
	if err != nil {
		t.Fatalf("page1.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts1), 1; got != want {
		t.Fatalf("len(page1 artifacts) = %d, want %d", got, want)
	}
	if got, want := artifacts1[0].Text, "PageOne"; got != want {
		t.Fatalf("page1 artifact text = %q, want %q", got, want)
	}

	page2, err := handle.Page(1)
	if err != nil {
		t.Fatalf("handle.Page(1) error = %v", err)
	}
	artifacts2, err := page2.Artifacts(nil, ArtifactOptions{IncludeText: true})
	if err != nil {
		t.Fatalf("page2.Artifacts() error = %v", err)
	}
	if got, want := len(artifacts2), 1; got != want {
		t.Fatalf("len(page2 artifacts) = %d, want %d", got, want)
	}
	if got, want := artifacts2[0].Text, "PageTwo"; got != want {
		t.Fatalf("page2 artifact text = %q, want %q", got, want)
	}
}

func TestToUnicodeBfrangeAndVariableWidthDecoding(t *testing.T) {
	const pdf = `%PDF-1.7
1 0 obj << /Length 57 >> stream
2 beginbfrange
<00> <02> <0041>
endbfrange
endstream
`

	glyphMap := extractToUnicodeMap([]byte(pdf))
	if got, want := glyphMap["00"], "A"; got != want {
		t.Fatalf("glyphMap[00] = %q, want %q", got, want)
	}
	if got, want := glyphMap["01"], "B"; got != want {
		t.Fatalf("glyphMap[01] = %q, want %q", got, want)
	}
	if got, want := glyphMap["02"], "C"; got != want {
		t.Fatalf("glyphMap[02] = %q, want %q", got, want)
	}

	if got, want := decodeHexGlyphString([]byte("000102"), glyphMap), "ABC"; got != want {
		t.Fatalf("decodeHexGlyphString() = %q, want %q", got, want)
	}
}
