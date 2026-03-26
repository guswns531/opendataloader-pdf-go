// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generators_test

import (
	"math"
	"os"
	"testing"

	gojson "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
	json_gen "github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/json"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/generators/json/serializers"
)

func TestSerializeTableSkipsLineArtChunksInCells(t *testing.T) {
	table := &entities.SemanticTable{
		BaseObject: entities.BaseObject{
			ID: "table-1",
			BBox: entities.BoundingBox{
				X: 10, Y: 20, Width: 100, Height: 50, Page: 0,
			},
		},
		PrevTableID: "1",
		NextTableID: "2",
		Rows: []*entities.TableRow{
			{
				Cells: []*entities.TableCell{
					{
						BBox:    entities.BoundingBox{X: 10, Y: 45, Width: 50, Height: 25, Page: 0},
						Rowspan: 1,
						Colspan: 1,
						Content: []entities.IObject{
							&entities.TextChunk{
								BaseObject: entities.BaseObject{
									ID:   "chunk-1",
									BBox: entities.BoundingBox{X: 12, Y: 50, Width: 20, Height: 8, Page: 0},
								},
								Text: "A1",
							},
							&entities.LineArtChunk{
								BaseObject: entities.BaseObject{
									BBox: entities.BoundingBox{X: 10, Y: 45, Width: 50, Height: 1, Page: 0},
								},
							},
						},
					},
				},
			},
		},
	}

	got := serializers.SerializeTable(table)

	assert.Equal(t, "table", got["type"])
	assert.Equal(t, 1, got["number of rows"])
	assert.Equal(t, 1, got["number of columns"])
	assert.EqualValues(t, 1, got["previous table id"])
	assert.EqualValues(t, 2, got["next table id"])

	rows := got["rows"].([]interface{})
	cells := rows[0].(map[string]interface{})["cells"].([]interface{})
	kids := cells[0].(map[string]interface{})["kids"].([]interface{})
	assert.Len(t, kids, 1)
	assert.Equal(t, "A1", kids[0].(map[string]interface{})["content"])
	assert.Len(t, cells[0].(map[string]interface{})["bounding box"].([]interface{}), 4)
}

func TestJsonWriterWritesHeadingLevelField(t *testing.T) {
	doc := &entities.Document{
		Metadata: entities.DocumentMetadata{
			Title:     "Doc",
			Author:    "Author",
			PageCount: 1,
		},
		Pages: []*entities.Page{
			{
				PageMetadata: entities.PageMetadata{Number: 0, Width: 595, Height: 842},
				Elements: []entities.IObject{
					&entities.SemanticHeading{
						BaseObject: entities.BaseObject{
							ID:   "h-1",
							BBox: entities.BoundingBox{X: 10, Y: 700, Width: 120, Height: 20, Page: 0},
						},
						Level: 3,
						Lines: []*entities.TextLine{
							{
								BaseObject: entities.BaseObject{
									BBox: entities.BoundingBox{X: 10, Y: 700, Width: 120, Height: 20, Page: 0},
								},
								Chunks: []*entities.TextChunk{
									{
										BaseObject: entities.BaseObject{
											BBox: entities.BoundingBox{X: 10, Y: 700, Width: 120, Height: 20, Page: 0},
										},
										Text: "Section Title",
										FontStyle: entities.FontStyle{
											FontName: "Helvetica-Bold",
											FontSize: 18,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	writer := &json_gen.JsonWriter{}
	data, err := writer.Write(doc)

	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = gojson.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	kids := decoded["kids"].([]interface{})
	assert.Len(t, kids, 1)
	assert.Equal(t, "heading", kids[0].(map[string]interface{})["type"])
	assert.EqualValues(t, 3, kids[0].(map[string]interface{})["heading level"])
	assert.Len(t, kids[0].(map[string]interface{})["bounding box"].([]interface{}), 4)
}

func TestJsonWriterOutputMatchesSchema(t *testing.T) {
	doc := &entities.Document{
		Metadata: entities.DocumentMetadata{
			Title:     "sample",
			Author:    "tester",
			PageCount: 1,
		},
		Pages: []*entities.Page{
			{
				PageMetadata: entities.PageMetadata{Number: 0, Width: 595, Height: 842},
				Elements: []entities.IObject{
					&entities.SemanticParagraph{
						BaseObject: entities.BaseObject{
							BBox: entities.BoundingBox{X: 10, Y: 10, Width: 100, Height: 20, Page: 0},
						},
						Lines: []*entities.TextLine{
							{
								BaseObject: entities.BaseObject{BBox: entities.BoundingBox{X: 10, Y: 10, Width: 100, Height: 20, Page: 0}},
								Chunks: []*entities.TextChunk{
									{
										BaseObject: entities.BaseObject{BBox: entities.BoundingBox{X: 10, Y: 10, Width: 100, Height: 20, Page: 0}},
										Text:       "hello",
										FontStyle: entities.FontStyle{
											FontName: "Helvetica",
											FontSize: 12,
											Color:    [3]float64{0, 0, 0},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	writer := &json_gen.JsonWriter{}
	data, err := writer.Write(doc)
	require.NoError(t, err)

	schemaBytes, err := os.ReadFile("../../../../schema.json")
	require.NoError(t, err)

	result, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schemaBytes),
		gojsonschema.NewBytesLoader(data),
	)
	require.NoError(t, err)
	assert.Truef(t, result.Valid(), "schema errors: %v", result.Errors())
}

func TestSerializeBBoxArrayPrecisionAndInvalidValues(t *testing.T) {
	bbox := entities.BoundingBox{
		X: 1.1234567, Y: 2.7654321, Width: math.Inf(1), Height: math.NaN(), Page: 0,
	}

	got := json_gen.SerializeBBoxArray(bbox)

	require.Len(t, got, 4)
	assert.Equal(t, 1.123457, got[0])
	assert.Equal(t, 2.765432, got[1])
	assert.Nil(t, got[2])
	assert.Nil(t, got[3])
}

func TestTextSerializersMatchJavaTypeNames(t *testing.T) {
	chunk := &entities.TextChunk{
		BaseObject: entities.BaseObject{
			ID:   "1",
			BBox: entities.BoundingBox{X: 10, Y: 20, Width: 30, Height: 8, Page: 0},
		},
		Text: "chunk text",
		FontStyle: entities.FontStyle{
			FontName: "Helvetica",
			FontSize: 12,
		},
	}
	line := &entities.TextLine{
		BaseObject: entities.BaseObject{
			ID:   "2",
			BBox: entities.BoundingBox{X: 10, Y: 40, Width: 30, Height: 8, Page: 0},
		},
		Chunks: []*entities.TextChunk{chunk},
	}
	lineArt := &entities.LineArtChunk{
		BaseObject: entities.BaseObject{
			ID:   "3",
			BBox: entities.BoundingBox{X: 10, Y: 50, Width: 30, Height: 1, Page: 0},
		},
	}

	chunkJSON := serializers.SerializeTextChunk(chunk)
	lineJSON := serializers.SerializeTextLine(line)
	lineArtJSON := serializers.SerializeLineChunk(lineArt)

	assert.Equal(t, "text chunk", chunkJSON["type"])
	assert.Equal(t, "text chunk", lineJSON["type"])
	assert.Equal(t, "line", lineArtJSON["type"])
	assert.Equal(t, "chunk text", chunkJSON["content"])
	assert.Equal(t, "chunk text", lineJSON["content"])
}

func TestJsonWriterSerializesTopLevelTextLine(t *testing.T) {
	doc := &entities.Document{
		Metadata: entities.DocumentMetadata{PageCount: 1},
		Pages: []*entities.Page{
			{
				PageMetadata: entities.PageMetadata{Number: 0, Width: 595, Height: 842},
				Elements: []entities.IObject{
					&entities.TextLine{
						BaseObject: entities.BaseObject{
							ID:   "9",
							BBox: entities.BoundingBox{X: 10, Y: 700, Width: 120, Height: 20, Page: 0},
						},
						Chunks: []*entities.TextChunk{
							{
								BaseObject: entities.BaseObject{
									BBox: entities.BoundingBox{X: 10, Y: 700, Width: 120, Height: 20, Page: 0},
								},
								Text: "Standalone line",
								FontStyle: entities.FontStyle{
									FontName: "Helvetica",
									FontSize: 12,
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := (&json_gen.JsonWriter{}).Write(doc)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = gojson.Unmarshal(data, &decoded)
	require.NoError(t, err)

	kids := decoded["kids"].([]interface{})
	require.Len(t, kids, 1)
	assert.Equal(t, "text chunk", kids[0].(map[string]interface{})["type"])
	assert.Equal(t, "Standalone line", kids[0].(map[string]interface{})["content"])
}
