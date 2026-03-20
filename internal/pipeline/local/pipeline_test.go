package local

import (
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/fixture"
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
