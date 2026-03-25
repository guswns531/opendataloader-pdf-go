# Phase 0: Go 프로젝트 구조

## 최종 디렉토리 레이아웃

```
go/                                         # Java의 java/ 와 동일 레벨
├── cmd/
│   └── opendataloader-pdf/
│       └── main.go                         # CLI 진입점 (CLIMain.java)
├── internal/
│   ├── api/
│   │   ├── config.go                       # Config.java (859줄 → Go struct)
│   │   ├── filter_config.go                # FilterConfig.java
│   │   └── opendataloader_pdf.go           # OpenDataLoaderPDF.java
│   ├── cli/
│   │   ├── options.go                      # CLIOptions.java (cobra 플래그 정의)
│   │   └── main.go                         # CLIMain.java 로직
│   ├── containers/
│   │   └── static_layout.go               # StaticLayoutContainers.java (ThreadLocal → context)
│   ├── entities/
│   │   ├── semantic_formula.go             # SemanticFormula.java
│   │   ├── semantic_picture.go             # SemanticPicture.java
│   │   ├── semantic_paragraph.go           # SemanticParagraph
│   │   ├── semantic_table.go               # Table 관련
│   │   ├── semantic_list.go                # PDFList, ListItem
│   │   ├── semantic_heading.go             # Heading
│   │   ├── semantic_caption.go             # Caption
│   │   ├── text_chunk.go                   # TextChunk
│   │   ├── text_line.go                    # TextLine
│   │   └── base_object.go                  # IObject 인터페이스 + BaseObject
│   ├── processors/
│   │   ├── document_processor.go           # DocumentProcessor.java (454줄)
│   │   ├── text_processor.go               # TextProcessor.java
│   │   ├── text_line_processor.go          # TextLineProcessor.java (127줄)
│   │   ├── paragraph_processor.go          # ParagraphProcessor.java
│   │   ├── heading_processor.go            # HeadingProcessor.java (210줄)
│   │   ├── list_processor.go               # ListProcessor.java
│   │   ├── level_processor.go              # LevelProcessor.java
│   │   ├── table_border_processor.go       # TableBorderProcessor.java (252줄)
│   │   ├── abstract_table_processor.go     # AbstractTableProcessor.java
│   │   ├── special_table_processor.go      # SpecialTableProcessor.java
│   │   ├── cluster_table_processor.go      # ClusterTableProcessor.java
│   │   ├── caption_processor.go            # CaptionProcessor.java
│   │   ├── content_filter_processor.go     # ContentFilterProcessor.java
│   │   ├── hidden_text_processor.go        # HiddenTextProcessor.java
│   │   ├── header_footer_processor.go      # HeaderFooterProcessor.java
│   │   ├── tagged_document_processor.go    # TaggedDocumentProcessor.java
│   │   ├── strikethrough_processor.go      # StrikethroughProcessor.java
│   │   ├── hybrid_document_processor.go    # HybridDocumentProcessor.java
│   │   └── readingorder/
│   │       └── xycut_plus_plus_sorter.go   # XYCutPlusPlusSorter.java (677줄)
│   ├── generators/
│   │   ├── markdown/
│   │   │   ├── markdown_generator.go       # MarkdownGenerator.java (359줄)
│   │   │   ├── markdown_generator_factory.go
│   │   │   ├── markdown_html_generator.go  # MarkdownHTMLGenerator.java
│   │   │   └── markdown_syntax.go          # MarkdownSyntax.java
│   │   ├── html/
│   │   │   ├── html_generator.go           # HtmlGenerator.java
│   │   │   ├── html_generator_factory.go
│   │   │   └── html_syntax.go              # HtmlSyntax.java
│   │   ├── json/
│   │   │   ├── json_writer.go              # JsonWriter.java (89줄)
│   │   │   ├── json_name.go                # JsonName.java (필드명 상수)
│   │   │   ├── object_mapper.go            # ObjectMapperHolder.java
│   │   │   └── serializers/
│   │   │       ├── caption_serializer.go
│   │   │       ├── formula_serializer.go
│   │   │       ├── heading_serializer.go
│   │   │       ├── image_serializer.go
│   │   │       ├── list_item_serializer.go
│   │   │       ├── list_serializer.go
│   │   │       ├── paragraph_serializer.go
│   │   │       ├── picture_serializer.go
│   │   │       ├── table_serializer.go     # (TableSerializer.java)
│   │   │       ├── table_row_serializer.go
│   │   │       ├── table_cell_serializer.go
│   │   │       ├── text_chunk_serializer.go
│   │   │       ├── text_line_serializer.go
│   │   │       ├── line_chunk_serializer.go
│   │   │       ├── semantic_text_node_serializer.go
│   │   │       ├── header_footer_serializer.go
│   │   │       └── double_serializer.go
│   │   └── text/
│   │       └── text_generator.go           # TextGenerator.java
│   ├── hybrid/
│   │   ├── hybrid_config.go                # HybridConfig.java (200줄)
│   │   ├── hybrid_client.go                # HybridClient.java (인터페이스)
│   │   ├── hybrid_client_factory.go        # HybridClientFactory.java
│   │   ├── hybrid_schema_transformer.go    # HybridSchemaTransformer.java
│   │   ├── docling_fast_server_client.go   # DoclingFastServerClient.java (308줄)
│   │   ├── docling_schema_transformer.go   # DoclingSchemaTransformer.java
│   │   ├── hancom_client.go                # HancomClient.java
│   │   ├── hancom_schema_transformer.go
│   │   ├── triage_processor.go             # TriageProcessor.java
│   │   └── triage_logger.go                # TriageLogger.java
│   ├── pdf/
│   │   ├── pdf_layer.go                    # PDFLayer.java
│   │   └── pdf_writer.go                   # PDFWriter.java
│   └── utils/
│       ├── base64_image_utils.go           # Base64ImageUtils.java
│       ├── bulleted_paragraph_utils.go     # BulletedParagraphUtils.java
│       ├── content_sanitizer.go            # ContentSanitizer.java
│       ├── images_utils.go                 # ImagesUtils.java
│       ├── mode_weight_statistics.go       # ModeWeightStatistics.java
│       ├── text_node_statistics.go         # TextNodeStatistics.java
│       ├── text_node_statistics_config.go  # TextNodeStatisticsConfig.java
│       ├── sanitization_rule.go            # SanitizationRule.java
│       └── levels/
│           ├── level_info.go               # LevelInfo.java
│           ├── list_level_info.go
│           ├── table_level_info.go
│           ├── text_bullet_paragraph_level_info.go
│           └── line_art_bullet_paragraph_level_info.go
├── pkg/
│   └── verapdf/                            # MPL-2.0 포팅
│       ├── LICENSE_MPL2                    # MPL-2.0 전문 복사
│       ├── model/                          # veraPDF validation-model
│       │   ├── pdfa/
│       │   ├── wcag/
│       │   └── results/
│       ├── pdfbox/                         # veraPDF pdfbox-validation
│       │   ├── features/
│       │   └── checker/
│       └── doc.go                          # 패키지 문서 + 라이센스 안내
├── tests/
│   ├── unit/
│   │   ├── processors/
│   │   ├── generators/
│   │   └── utils/
│   └── integration/
│       └── end_to_end_test.go
├── go.mod
├── go.sum
└── Makefile                                # build, test, bench 타겟
```

---

## Java 패턴 → Go 변환 규칙

### 1. ThreadLocal → 컨텍스트 매개변수

```java
// Java
ThreadLocal<State> state = new ThreadLocal<>();

// Go
type ProcessorContext struct {
    CurrentID    int
    Headings     []entities.Heading
    ImageIndex   int
    UseStructTree bool
}
```

### 2. Java 인터페이스 → Go 인터페이스

```java
// Java
public interface IObject {
    String getID();
    BoundingBox getBBox();
}

// Go
type IObject interface {
    GetID() string
    GetBBox() BoundingBox
}
```

### 3. Jackson Serializer → MarshalJSON

```java
// Java
@JsonSerialize(using = TableSerializer.class)
public class SemanticTable { ... }

// Go
func (t *SemanticTable) MarshalJSON() ([]byte, error) {
    // TableSerializer 로직 직접 구현
}
```

### 4. static 메서드 → 패키지 함수

```java
// Java
public static List<Page> processFile(Config config) { ... }

// Go
func ProcessFile(config *api.Config) ([]Page, error) { ... }
```

### 5. 재귀 깊이 제한 (TableBorderProcessor)

```java
// Java: ThreadLocal<Integer> depth
// Go: 함수 매개변수로 depth 전달
func processNode(node Node, depth int) {
    if depth >= 10 { return }
    processNode(child, depth+1)
}
```

---

## Makefile 타겟

```makefile
.PHONY: build test bench sync clean

build:
	go build -o bin/opendataloader-pdf ./cmd/opendataloader-pdf/

test:
	go test ./...

bench:
	cd .. && python tests/benchmark/run.py

sync: build
	./bin/opendataloader-pdf --export-options > options.json
	npm run sync

clean:
	rm -rf bin/
```
