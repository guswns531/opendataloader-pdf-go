package local

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/fixture"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/nativepdf"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestPipelineBuildsParagraphsFromArtifactsAndSorts(t *testing.T) {
	const fixtureJSON = `{
	  "metadata": {"file_name": "fixture.json", "page_count": 1},
	  "pages": [
	    {
	      "metadata": {"number": 1, "index": 0},
	      "artifacts": [
		        {
		          "kind": "text",
		          "page_index": 0,
		          "page_number": 1,
		          "sequence": 1,
		          "bounds": {"left": 0, "bottom": 70, "right": 40, "top": 80},
		          "text": "Second"
		        },
	        {
	          "kind": "text",
	          "page_index": 0,
	          "page_number": 1,
	          "sequence": 0,
	          "bounds": {"left": 0, "bottom": 110, "right": 40, "top": 120},
	          "text": "First"
	        }
	      ]
	    }
	  ]
	}`

	pipeline := New(fixture.New())
	ctx := core.NewProcessingContext(nil, core.ProcessingOptions{})

	document, err := pipeline.Run(ctx, core.Source{
		Name:   "fixture.json",
		Reader: strings.NewReader(fixtureJSON),
	}, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if ctx.Document != document {
		t.Fatal("processing context document was not updated")
	}
	if len(document.Pages) != 1 {
		t.Fatalf("len(document.Pages) = %d, want 1", len(document.Pages))
	}
	if len(document.Kids) != 2 {
		t.Fatalf("len(document.Kids) = %d, want 2", len(document.Kids))
	}

	first, ok := document.Kids[0].(*model.Paragraph)
	if !ok {
		t.Fatalf("first kid type = %T, want *model.Paragraph", document.Kids[0])
	}
	if first.Content != "First" {
		t.Fatalf("first paragraph content = %q, want %q", first.Content, "First")
	}
}

func TestPipelinePromotesHeadingsAndBuildsLists(t *testing.T) {
	const fixtureJSON = `{
	  "metadata": {"file_name": "fixture.json", "page_count": 1},
	  "pages": [
	    {
	      "metadata": {"number": 1, "index": 0},
	      "kids": [
	        {
	          "type": "paragraph",
	          "page_index": 0,
	          "page_number": 1,
	          "bounds": {"left": 20, "bottom": 700, "right": 260, "top": 724},
	          "font_size": 20,
	          "bold": true,
	          "content": "INTRODUCTION"
	        },
	        {
	          "type": "paragraph",
	          "page_index": 0,
	          "page_number": 1,
	          "bounds": {"left": 40, "bottom": 640, "right": 260, "top": 652},
	          "content": "- first item"
	        },
	        {
	          "type": "paragraph",
	          "page_index": 0,
	          "page_number": 1,
	          "bounds": {"left": 40, "bottom": 620, "right": 260, "top": 632},
	          "content": "- second item"
	        }
	      ]
	    }
	  ]
	}`

	pipeline := New(fixture.New())
	ctx := core.NewProcessingContext(nil, core.ProcessingOptions{})

	document, err := pipeline.Run(ctx, core.Source{
		Name:   "fixture.json",
		Reader: strings.NewReader(fixtureJSON),
	}, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(document.Kids) != 2 {
		t.Fatalf("len(document.Kids) = %d, want 2", len(document.Kids))
	}

	if _, ok := document.Kids[0].(*model.Heading); !ok {
		t.Fatalf("first node type = %T, want *model.Heading", document.Kids[0])
	}
	listNode, ok := document.Kids[1].(*model.List)
	if !ok {
		t.Fatalf("second node type = %T, want *model.List", document.Kids[1])
	}
	if listNode.NumberOfItems != 2 {
		t.Fatalf("list item count = %d, want 2", listNode.NumberOfItems)
	}
}

func TestPipelineRemovesRepeatedHeadersAcrossPages(t *testing.T) {
	const fixtureJSON = `{
	  "metadata": {"file_name": "fixture.json", "page_count": 2},
	  "pages": [
	    {
	      "metadata": {"number": 1, "index": 0, "size": {"width": 600, "height": 800}},
	      "kids": [
	        {
	          "type": "paragraph",
	          "page_index": 0,
	          "page_number": 1,
	          "bounds": {"left": 40, "bottom": 770, "right": 180, "top": 790},
	          "content": "Company Report"
	        },
	        {
	          "type": "paragraph",
	          "page_index": 0,
	          "page_number": 1,
	          "bounds": {"left": 40, "bottom": 700, "right": 260, "top": 720},
	          "content": "Body page one"
	        }
	      ]
	    },
	    {
	      "metadata": {"number": 2, "index": 1, "size": {"width": 600, "height": 800}},
	      "kids": [
	        {
	          "type": "paragraph",
	          "page_index": 1,
	          "page_number": 2,
	          "bounds": {"left": 40, "bottom": 770, "right": 180, "top": 790},
	          "content": "Company Report"
	        },
	        {
	          "type": "paragraph",
	          "page_index": 1,
	          "page_number": 2,
	          "bounds": {"left": 40, "bottom": 700, "right": 260, "top": 720},
	          "content": "Body page two"
	        }
	      ]
	    }
	  ]
	}`

	pipeline := New(fixture.New())
	ctx := core.NewProcessingContext(nil, core.ProcessingOptions{})

	document, err := pipeline.Run(ctx, core.Source{
		Name:   "fixture.json",
		Reader: strings.NewReader(fixtureJSON),
	}, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(document.Kids) != 2 {
		t.Fatalf("len(document.Kids) = %d, want 2", len(document.Kids))
	}
	for i, element := range document.Kids {
		para, ok := element.(*model.Paragraph)
		if !ok {
			t.Fatalf("document.Kids[%d] type = %T, want *model.Paragraph", i, element)
		}
		if strings.Contains(para.Content, "Company Report") {
			t.Fatalf("header text should have been removed, got %q", para.Content)
		}
	}
}

func TestPipelineBuildsParagraphsFromNativeRawFixture(t *testing.T) {
	pipeline := New(nativepdf.NewIngestor(nativepdf.NewFixtureLoader()))
	ctx := core.NewProcessingContext(nil, core.ProcessingOptions{})

	document, err := pipeline.Run(ctx, core.Source{
		Path: filepath.Clean("../../../testdata/fixtures/native/basic.raw.json"),
		Name: "basic.raw.json",
	}, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(document.Kids) != 2 {
		t.Fatalf("len(document.Kids) = %d, want 2", len(document.Kids))
	}

	first, ok := document.Kids[0].(*model.Paragraph)
	if !ok {
		t.Fatalf("document.Kids[0] type = %T, want *model.Paragraph", document.Kids[0])
	}
	second, ok := document.Kids[1].(*model.Paragraph)
	if !ok {
		t.Fatalf("document.Kids[1] type = %T, want *model.Paragraph", document.Kids[1])
	}

	if first.Content != "Pure Go native fixture paragraph one." {
		t.Fatalf("first paragraph content = %q", first.Content)
	}
	if second.Content != "Second paragraph placed lower on the page." {
		t.Fatalf("second paragraph content = %q", second.Content)
	}
}

func TestPipelineKeepsNativeTwoColumnLinesAsSeparateParagraphs(t *testing.T) {
	pipeline := New(nativepdf.NewIngestor(nativepdf.NewFixtureLoader()))
	ctx := core.NewProcessingContext(nil, core.ProcessingOptions{})

	document, err := pipeline.Run(ctx, core.Source{
		Path: filepath.Clean("../../../testdata/fixtures/native/two_column_reading_order.raw.json"),
		Name: "two_column_reading_order.raw.json",
	}, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(document.Kids) != 4 {
		t.Fatalf("len(document.Kids) = %d, want 4", len(document.Kids))
	}

	got := make([]string, 0, len(document.Kids))
	for _, element := range document.Kids {
		para, ok := element.(*model.Paragraph)
		if !ok {
			t.Fatalf("element type = %T, want *model.Paragraph", element)
		}
		got = append(got, para.Content)
	}

	want := []string{
		"Left column first line.",
		"Left column second line.",
		"Right column first line.",
		"Right column second line.",
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("paragraph contents = %v, want %v", got, want)
	}
}
