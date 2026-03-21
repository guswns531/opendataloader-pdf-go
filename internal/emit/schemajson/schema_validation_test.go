package schemajson

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/guswns531/opendataloader-pdf-go/internal/core"
	"github.com/guswns531/opendataloader-pdf-go/internal/ingest/nativepdf"
	"github.com/guswns531/opendataloader-pdf-go/internal/model"
	"github.com/guswns531/opendataloader-pdf-go/internal/pipeline/local"
	"github.com/xeipuuv/gojsonschema"
)

func TestEmitterOutputValidatesAgainstPublishedSchema(t *testing.T) {
	doc := model.NewDocument(model.DocumentMetadata{
		FileName:  "sample.pdf",
		PageCount: 1,
	})
	doc.Kids = []model.ContentElement{
		&model.Paragraph{
			TextNode: model.TextNode{
				BaseNode: model.BaseNode{
					ID:         1,
					Type:       model.ElementTypeParagraph,
					PageNumber: 1,
					Bounds:     model.NewBox(10, 20, 200, 40),
				},
				TextProperties: model.TextProperties{
					Font:      "Times-Roman",
					FontSize:  12,
					TextColor: "#000000",
					Content:   "Hello schema",
				},
			},
		},
	}

	payload := emitJSON(t, doc)
	validateAgainstPublishedSchema(t, payload)
}

func TestNativeSkeletonLoremOutputValidatesAgainstPublishedSchema(t *testing.T) {
	pipeline := local.New(nativepdf.NewIngestor(nativepdf.NewSkeletonLoader()))
	ctx := core.NewProcessingContext(nil, core.ProcessingOptions{})

	document, err := pipeline.Run(ctx, core.Source{
		Path: filepath.Clean("../../../samples/pdf/lorem.pdf"),
		Name: "lorem.pdf",
	}, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	payload := emitJSON(t, document)
	validateAgainstPublishedSchema(t, payload)
}

func emitJSON(t *testing.T, document *model.Document) []byte {
	t.Helper()

	var buf bytes.Buffer
	ctx := core.NewProcessingContext(document, core.ProcessingOptions{})
	if err := New().Emit(ctx, document, &buf); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}
	return buf.Bytes()
}

func validateAgainstPublishedSchema(t *testing.T, payload []byte) {
	t.Helper()

	schemaPath := filepath.Clean("../../../schema.json")
	var value any
	if err := json.Unmarshal(payload, &value); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewReferenceLoader("file://"+schemaPath),
		gojsonschema.NewGoLoader(value),
	)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !result.Valid() {
		t.Fatalf("schema validation failed: %v\npayload=%s", result.Errors(), payload)
	}
}
