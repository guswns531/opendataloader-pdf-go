package local

import (
	"strings"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/fixture"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
)

func TestPipelineReconstructsHeadingAndListFromSemanticFixture(t *testing.T) {
	const fixtureJSON = `{
	  "metadata": {"file_name": "semantic_phase4.json", "page_count": 1},
	  "pages": [
	    {
	      "metadata": {"number": 1, "index": 0},
	      "kids": [
	        {
	          "type": "paragraph",
	          "page_index": 0,
	          "page_number": 1,
	          "bounds": {"left": 40, "bottom": 760, "right": 300, "top": 790},
	          "font_size": 22,
	          "bold": true,
	          "content": "Document Title"
	        },
	        {
	          "type": "paragraph",
	          "page_index": 0,
	          "page_number": 1,
	          "bounds": {"left": 48, "bottom": 700, "right": 240, "top": 718},
	          "content": "• First item"
	        },
	        {
	          "type": "paragraph",
	          "page_index": 0,
	          "page_number": 1,
	          "bounds": {"left": 66, "bottom": 678, "right": 360, "top": 696},
	          "content": "Continuation detail"
	        },
	        {
	          "type": "paragraph",
	          "page_index": 0,
	          "page_number": 1,
	          "bounds": {"left": 48, "bottom": 652, "right": 240, "top": 670},
	          "content": "• Second item"
	        },
	        {
	          "type": "paragraph",
	          "page_index": 0,
	          "page_number": 1,
	          "bounds": {"left": 40, "bottom": 520, "right": 340, "top": 540},
	          "content": "Closing paragraph"
	        }
	      ]
	    }
	  ]
	}`

	pipeline := New(fixture.New())
	ctx := core.NewProcessingContext(nil, core.ProcessingOptions{})

	document, err := pipeline.Run(ctx, core.Source{
		Name:   "semantic_phase4.json",
		Reader: strings.NewReader(fixtureJSON),
	}, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(document.Kids) != 3 {
		t.Fatalf("len(document.Kids) = %d, want 3", len(document.Kids))
	}

	if _, ok := document.Kids[0].(*model.Heading); !ok {
		t.Fatalf("document.Kids[0] type = %T, want *model.Heading", document.Kids[0])
	}
	listNode, ok := document.Kids[1].(*model.List)
	if !ok {
		t.Fatalf("document.Kids[1] type = %T, want *model.List", document.Kids[1])
	}
	if listNode.NumberOfItems != 2 {
		t.Fatalf("list item count = %d, want 2", listNode.NumberOfItems)
	}
	if len(listNode.ListItems[0].Kids) != 1 {
		t.Fatalf("first list item kids = %d, want 1", len(listNode.ListItems[0].Kids))
	}

	closing, ok := document.Kids[2].(*model.Paragraph)
	if !ok {
		t.Fatalf("document.Kids[2] type = %T, want *model.Paragraph", document.Kids[2])
	}
	if closing.Content != "Closing paragraph" {
		t.Fatalf("closing content = %q, want %q", closing.Content, "Closing paragraph")
	}
}
