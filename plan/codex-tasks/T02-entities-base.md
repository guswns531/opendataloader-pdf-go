# T02: 핵심 엔티티 구조체 정의

## 선행 태스크
T01 완료

## 참조 Java 파일
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/entities/`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/api/Config.java`

## 지시사항

### 1. internal/entities/base_object.go

Apache-2.0 헤더 포함. 아래 타입 정의:
- `BoundingBox` struct (X, Y, Width, Height float64, Page int)
- `ObjectType` string 타입 + 상수 (paragraph, heading, table, list, image, formula, caption, header_footer, text_line, text_chunk)
- `IObject` 인터페이스 (GetID() string, GetBBox() BoundingBox, GetObjectType() ObjectType)
- `BaseObject` struct + 메서드 구현

### 2. internal/entities/text_chunk.go

- `FontStyle` struct (FontName string, FontSize float64, Bold/Italic bool, Color [3]float64)
- `TextChunk` struct (BaseObject 임베드, Text string, FontStyle, Baseline float64, IsHidden/IsOffPage/IsTiny bool, IsStrikethrough bool)
- `LineArtChunk` struct (BaseObject 임베드, IsHorizontal/IsVertical bool, LineWidth float64)

### 3. internal/entities/text_line.go

- `TextLine` struct (BaseObject 임베드, Chunks []*TextChunk, LineArtBullet *LineArtChunk, Baseline float64, IsFirstLine/IsLastLine bool)
- `GetText() string` 메서드 (모든 청크의 Text 연결)

### 4. internal/entities/semantic_paragraph.go

- `SemanticParagraph` struct (BaseObject 임베드, Lines []*TextLine, Alignment string)
- Alignment 상수: AlignLeft, AlignRight, AlignCenter, AlignJustify

### 5. internal/entities/semantic_heading.go

- `SemanticHeading` struct (BaseObject 임베드, Lines []*TextLine, Level int, FontSize float64)

### 6. internal/entities/semantic_table.go

- `TableCell` struct (Content []IObject, Rowspan/Colspan int, BBox BoundingBox)
- `TableRow` struct (Cells []*TableCell, BBox BoundingBox)
- `SemanticTable` struct (BaseObject 임베드, Rows []*TableRow, PrevTableID/NextTableID string)

### 7. internal/entities/semantic_list.go

- `ListItem` struct (BaseObject 임베드, Content []IObject, BulletText string, IsOrdered bool, Level int)
- `PDFList` struct (BaseObject 임베드, Items []*ListItem, IsOrdered bool)

### 8. internal/entities/semantic_image.go

- `SemanticImage` struct (BaseObject 임베드, Alt string, Data []byte, ExternalPath string, Width/Height float64)

### 9. internal/entities/semantic_formula.go

Java `SemanticFormula.java` 참조:
- `SemanticFormula` struct (BaseObject 임베드, LaTeX string)

### 10. internal/entities/semantic_caption.go

- `SemanticCaption` struct (BaseObject 임베드, Text string, RefType string "figure"/"table")

### 11. internal/entities/document.go

- `PageMetadata` struct (Number int, Width/Height float64)
- `Page` struct (PageMetadata, Chunks []*TextChunk, LineArts []*LineArtChunk, Elements []IObject)
- `DocumentMetadata` struct (Title/Author/Creator/Producer string, PageCount int)
- `Document` struct (Metadata DocumentMetadata, Pages []*Page)

## 완료 기준
- `go build ./internal/entities/...` 성공
- 모든 파일에 Apache-2.0 라이센스 헤더 포함
- `go vet ./internal/entities/...` 경고 없음
