# T07: JSON 직렬화 레이어 포팅 (19개 Serializer)

## 선행 태스크
T02 완료 (엔티티 구조체 필요)

## 참조 Java 파일
- `json/JsonName.java`
- `json/JsonWriter.java` (89줄)
- `json/ObjectMapperHolder.java`
- `json/serializers/*.java` (16+ 파일)

## 지시사항

### 1. internal/generators/json/json_name.go

Java `JsonName.java`의 모든 public static final String 상수를 Go const 블록으로 변환.
Java 소스를 직접 읽어 필드명 정확하게 매핑.

### 2. internal/generators/json/serializers/ (17개 파일)

각 Java Serializer를 Go 함수로 변환. 패턴:

```go
// SerializeXxx(x *entities.SemanticXxx) map[string]interface{}
// Java StdSerializer → Go 직렬화 함수
```

**포팅 파일 목록**:

| 파일 | Java 소스 | 특이사항 |
|------|-----------|---------|
| `table_serializer.go` | `TableSerializer.java` | LineArtChunk 필터링, prev/next ID 처리 |
| `table_row_serializer.go` | `TableRowSerializer.java` | |
| `table_cell_serializer.go` | `TableCellSerializer.java` | rowspan/colspan |
| `paragraph_serializer.go` | `ParagraphSerializer.java` | |
| `heading_serializer.go` | `HeadingSerializer.java` | level 1-6 |
| `list_serializer.go` | `ListSerializer.java` | |
| `list_item_serializer.go` | `ListItemSerializer.java` | |
| `image_serializer.go` | `ImageSerializer.java` | embedded vs external 분기 |
| `picture_serializer.go` | `PictureSerializer.java` | |
| `caption_serializer.go` | `CaptionSerializer.java` | |
| `formula_serializer.go` | `FormulaSerializer.java` | LaTeX 문자열 |
| `text_chunk_serializer.go` | `TextChunkSerializer.java` | |
| `text_line_serializer.go` | `TextLineSerializer.java` | |
| `line_chunk_serializer.go` | `LineChunkSerializer.java` | |
| `semantic_text_node_serializer.go` | `SemanticTextNodeSerializer.java` | |
| `header_footer_serializer.go` | `HeaderFooterSerializer.java` | |
| `double_serializer.go` | `DoubleSerializer.java` | 소수점 6자리 정밀도, NaN/Inf 처리 |

**TableSerializer 포팅 예시** (Java 소스 핵심 부분):
```go
func SerializeTable(t *entities.SemanticTable) map[string]interface{} {
    result := map[string]interface{}{
        json_name.FieldType: "table",
        json_name.FieldBBox: SerializeBBox(t.GetBBox()),
        json_name.FieldRows: len(t.Rows),
        json_name.FieldCols: maxCols(t.Rows),
    }
    if t.PrevTableID != "" {
        result[json_name.FieldPrev] = t.PrevTableID
    }
    if t.NextTableID != "" {
        result[json_name.FieldNext] = t.NextTableID
    }
    rows := make([]interface{}, 0, len(t.Rows))
    for _, row := range t.Rows {
        cells := make([]interface{}, 0)
        for _, cell := range row.Cells {
            cells = append(cells, SerializeCell(cell))
        }
        rows = append(rows, map[string]interface{}{
            json_name.FieldCells: cells,
        })
    }
    result["rows_data"] = rows
    return result
}
```

### 3. internal/generators/json/json_writer.go

Java `JsonWriter.java` (89줄) 포팅.
plan/05-phase3-generators.md의 Task 3-3 참조.

`JsonWriter` struct + `Write(doc *entities.Document) ([]byte, error)` 메서드.
`serializeElements([]entities.IObject)` 함수에서 타입 스위치로 각 직렬화 함수 호출.
`github.com/goccy/go-json` 사용 (성능).

### 4. 공통 헬퍼

`internal/generators/json/helpers.go`:
```go
func SerializeBBox(bbox entities.BoundingBox) map[string]interface{} {
    return map[string]interface{}{
        "x": bbox.X, "y": bbox.Y,
        "width": bbox.Width, "height": bbox.Height,
        "page": bbox.Page,
    }
}

func SerializeDouble(v float64) interface{} {
    // NaN → null, Inf → null, 정상 → 소수점 6자리
    if math.IsNaN(v) || math.IsInf(v, 0) {
        return nil
    }
    return math.Round(v*1e6) / 1e6
}
```

### 5. 테스트

`go/tests/unit/generators/json_writer_test.go`:
- `SemanticTable` → JSON 직렬화 테스트
- `SemanticHeading` → JSON 직렬화 테스트
- BBox 직렬화 정밀도 테스트
- `schema.json` 유효성 검증 (루트의 schema.json 사용)

## 완료 기준
- `go build ./internal/generators/json/...` 성공
- `go test ./tests/unit/generators/...` 통과
- 샘플 PDF JSON 출력이 `schema.json` (Draft 7) 유효성 검증 통과
