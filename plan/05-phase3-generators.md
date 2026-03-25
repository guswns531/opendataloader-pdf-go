# Phase 3: 출력 생성기 포팅

선행 조건: Phase 2 완료 (엔티티 구조체 모두 정의됨)
예상 소요: 2~3주

---

## Task 3-1: JsonName 상수 포팅

**Java 소스**: `json/JsonName.java`

**Go 대상**: `internal/generators/json/json_name.go`

```go
// Java의 public static final String 상수들 → Go const
package json

const (
    FieldPage      = "page"
    FieldElements  = "elements"
    FieldType      = "type"
    FieldBBox      = "bbox"
    FieldText      = "text"
    FieldRows      = "rows"
    FieldCells     = "cells"
    FieldRowspan   = "rowspan"
    FieldColspan   = "colspan"
    FieldLevel     = "level"
    FieldPrev      = "prev"
    FieldNext      = "next"
    FieldSrc       = "src"
    FieldAlt       = "alt"
    FieldWidth     = "width"
    FieldHeight    = "height"
    // ... 전체 JsonName.java에서 추출
)
```

---

## Task 3-2: JSON Serializer 19개 포팅 (Jackson → Go MarshalJSON)

**Java 소스**: `json/serializers/*.java` (16+ 파일)

**Go 대상**: `internal/generators/json/serializers/`

**변환 패턴** (공통):
```go
// Java: StdSerializer<SemanticTable>
// Go: MarshalJSON() 메서드 또는 별도 직렬화 함수

// TableSerializer.java → table_serializer.go
func SerializeTable(t *entities.SemanticTable) (map[string]interface{}, error) {
    result := map[string]interface{}{
        json_name.FieldType:  "table",
        json_name.FieldBBox:  serializeBBox(t.BBox),
        json_name.FieldRows:  len(t.Rows),
    }
    if t.PrevTableID != "" {
        result[json_name.FieldPrev] = t.PrevTableID
    }
    if t.NextTableID != "" {
        result[json_name.FieldNext] = t.NextTableID
    }
    // rows 직렬화 (LineArtChunk 필터링)
    rows := make([]interface{}, 0)
    for _, row := range t.Rows {
        cells := make([]interface{}, 0)
        for _, cell := range row.Cells {
            // LineArtChunk는 건너뜀
            cells = append(cells, SerializeCell(cell))
        }
        rows = append(rows, map[string]interface{}{
            json_name.FieldCells: cells,
        })
    }
    result[json_name.FieldRows] = rows
    return result, nil
}
```

**직렬화 파일 목록** (Java → Go):

| Java 파일 | Go 파일 | 특이사항 |
|-----------|---------|---------|
| `TableSerializer.java` | `table_serializer.go` | LineArtChunk 필터링, prev/next ID |
| `TableRowSerializer.java` | `table_row_serializer.go` | |
| `TableCellSerializer.java` | `table_cell_serializer.go` | rowspan/colspan |
| `ParagraphSerializer.java` | `paragraph_serializer.go` | |
| `HeadingSerializer.java` | `heading_serializer.go` | level 1-6 |
| `ListSerializer.java` | `list_serializer.go` | |
| `ListItemSerializer.java` | `list_item_serializer.go` | |
| `ImageSerializer.java` | `image_serializer.go` | embedded vs external |
| `PictureSerializer.java` | `picture_serializer.go` | |
| `CaptionSerializer.java` | `caption_serializer.go` | |
| `FormulaSerializer.java` | `formula_serializer.go` | LaTeX 문자열 |
| `TextChunkSerializer.java` | `text_chunk_serializer.go` | |
| `TextLineSerializer.java` | `text_line_serializer.go` | |
| `LineChunkSerializer.java` | `line_chunk_serializer.go` | |
| `SemanticTextNodeSerializer.java` | `semantic_text_node_serializer.go` | |
| `HeaderFooterSerializer.java` | `header_footer_serializer.go` | |
| `DoubleSerializer.java` | `double_serializer.go` | 소수점 정밀도 처리 |

---

## Task 3-3: JsonWriter 포팅

**Java 소스**: `json/JsonWriter.java` (89줄)

**Go 대상**: `internal/generators/json/json_writer.go`

```go
// internal/generators/json/json_writer.go
package json

import (
    gojson "github.com/goccy/go-json"
    "github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

type JsonWriter struct{}

// Write serializes the document to JSON bytes.
// Java: writeDocument(Document doc, Writer writer)
func (w *JsonWriter) Write(doc *entities.Document) ([]byte, error) {
    root := map[string]interface{}{
        "metadata": serializeMetadata(doc.Metadata),
        "pages":    serializePages(doc.Pages),
    }
    return gojson.Marshal(root)
}

func serializePages(pages []*entities.Page) []interface{} {
    result := make([]interface{}, len(pages))
    for i, page := range pages {
        result[i] = map[string]interface{}{
            FieldPage:     page.Number,
            FieldElements: serializeElements(page.Elements),
        }
    }
    return result
}

func serializeElements(elements []entities.IObject) []interface{} {
    result := make([]interface{}, 0, len(elements))
    for _, el := range elements {
        switch v := el.(type) {
        case *entities.SemanticTable:
            serialized, _ := SerializeTable(v)
            result = append(result, serialized)
        case *entities.SemanticHeading:
            serialized, _ := SerializeHeading(v)
            result = append(result, serialized)
        // ... 모든 타입 처리
        }
    }
    return result
}
```

---

## Task 3-4: MarkdownSyntax 포팅

**Java 소스**: `markdown/MarkdownSyntax.java`

**Go 대상**: `internal/generators/markdown/markdown_syntax.go`

```go
// MarkdownSyntax는 Markdown 포맷 규칙을 정의
package markdown

const (
    HeadingPrefix  = "#"   // H1: #, H2: ##, ...
    BoldWrapper    = "**"
    ItalicWrapper  = "_"
    TableSeparator = "|"
    CodeBlock      = "```"
    HorizontalRule = "---"
    ListUnordered  = "- "
    ListOrdered    = "%d. "  // fmt.Sprintf 사용
    ImageSyntax    = "![%s](%s)"  // alt, src
    FormulaSyntax  = "$%s$"       // LaTeX inline
)
```

---

## Task 3-5: MarkdownGenerator 포팅

**Java 소스**: `markdown/MarkdownGenerator.java` (359줄)

**Go 대상**: `internal/generators/markdown/markdown_generator.go`

**주요 처리**:
```go
type MarkdownGenerator struct {
    config   *api.Config
    withHTML bool   // markdown-with-html variant
    withImages bool  // markdown-with-images variant
}

func (g *MarkdownGenerator) Generate(doc *entities.Document) (string, error) {
    var sb strings.Builder

    for i, page := range doc.Pages {
        if i > 0 && g.config.MarkdownPageSeparator != "" {
            sb.WriteString(g.config.MarkdownPageSeparator + "\n")
        }
        for _, el := range page.Elements {
            switch v := el.(type) {
            case *entities.SemanticHeading:
                g.writeHeading(&sb, v)
            case *entities.SemanticParagraph:
                g.writeParagraph(&sb, v)
            case *entities.SemanticTable:
                g.writeTable(&sb, v)
            case *entities.PDFList:
                g.writeList(&sb, v)
            case *entities.SemanticImage:
                g.writeImage(&sb, v)
            case *entities.SemanticFormula:
                g.writeFormula(&sb, v)
            case *entities.SemanticCaption:
                g.writeCaption(&sb, v)
            }
        }
    }
    return sb.String(), nil
}

// 이미지 처리: embedded vs external
func (g *MarkdownGenerator) writeImage(sb *strings.Builder, img *entities.SemanticImage) {
    switch g.config.ImageOutput {
    case api.ImageOutputEmbedded:
        // Base64 인코딩하여 data URL 생성
        dataURL := utils.EncodeImageBase64(img.Data, g.config.ImageFormat)
        fmt.Fprintf(sb, "![%s](%s)\n\n", img.Alt, dataURL)
    case api.ImageOutputExternal:
        fmt.Fprintf(sb, "![%s](%s)\n\n", img.Alt, img.ExternalPath)
    case api.ImageOutputOff:
        // 이미지 생략
    }
}
```

---

## Task 3-6: HtmlGenerator 포팅

**Java 소스**: `html/HtmlGenerator.java`, `HtmlSyntax.java`

**Go 대상**: `internal/generators/html/html_generator.go`

```go
type HtmlGenerator struct {
    config *api.Config
}

func (g *HtmlGenerator) Generate(doc *entities.Document) (string, error) {
    var sb strings.Builder
    sb.WriteString("<!DOCTYPE html>\n<html>\n<body>\n")
    for i, page := range doc.Pages {
        if i > 0 && g.config.HTMLPageSeparator != "" {
            sb.WriteString(g.config.HTMLPageSeparator)
        }
        for _, el := range page.Elements {
            g.writeElement(&sb, el)
        }
    }
    sb.WriteString("</body>\n</html>")
    return sb.String(), nil
}
```

---

## Task 3-7: TextGenerator 포팅

**Java 소스**: `text/TextGenerator.java`

**Go 대상**: `internal/generators/text/text_generator.go`

순수 텍스트 추출 (서식 제거):
```go
func (g *TextGenerator) Generate(doc *entities.Document) (string, error) {
    var sb strings.Builder
    for i, page := range doc.Pages {
        if i > 0 && g.config.TextPageSeparator != "" {
            sb.WriteString(g.config.TextPageSeparator + "\n")
        }
        for _, el := range page.Elements {
            text := extractText(el)
            if text != "" {
                sb.WriteString(text + "\n\n")
            }
        }
    }
    return sb.String(), nil
}
```

---

## Task 3-8: GeneratorFactory 및 DocumentProcessor 출력 통합

**Java 소스**: `MarkdownGeneratorFactory.java`, `HtmlGeneratorFactory.java`

**Go 대상**: `internal/generators/` 내 각 factory

```go
// 출력 형식별 생성기 선택
func GetGenerator(format string, config *api.Config) Generator {
    switch format {
    case api.FormatMarkdown:
        return markdown.NewMarkdownGenerator(config, false, false)
    case api.FormatMarkdownWithHTML:
        return markdown.NewMarkdownGenerator(config, true, false)
    case api.FormatMarkdownWithImages:
        return markdown.NewMarkdownGenerator(config, false, true)
    case api.FormatHTML:
        return html.NewHtmlGenerator(config)
    case api.FormatJSON:
        return json_gen.NewJsonWriter(config)
    case api.FormatText:
        return text.NewTextGenerator(config)
    }
    return nil
}
```

---

## Phase 3 완료 기준

- [ ] 샘플 PDF → JSON 출력: `schema.json` 유효성 검증 통과
- [ ] 샘플 PDF → Markdown 출력: Java 결과와 diff 최소화
- [ ] 샘플 PDF → HTML 출력 생성
- [ ] 이미지 embedded/external 모드 모두 동작
- [ ] `go test ./internal/generators/...` 전체 통과
