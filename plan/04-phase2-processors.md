# Phase 2: 핵심 프로세서 포팅

선행 조건: Phase 1 완료 (go.mod, Config, CLI, 엔티티 기본 구조)
예상 소요: 3~5주

처리 파이프라인 순서:
```
PDF 로딩
  → ContentFilterProcessor (숨겨진 텍스트, 페이지 외부 요소 제거)
  → TextProcessor (원시 텍스트 청크 추출)
  → TextLineProcessor (텍스트 청크 → 텍스트 라인)
  → TableBorderProcessor (테두리 테이블 감지)
  → SpecialTableProcessor (경계 없는 테이블)
  → ClusterTableProcessor (클러스터링 기반 테이블)
  → ParagraphProcessor (텍스트 라인 → 문단)
  → ListProcessor (목록 감지 및 구조화)
  → HeadingProcessor (제목 감지)
  → LevelProcessor (제목 레벨 할당)
  → CaptionProcessor (이미지/테이블 캡션)
  → HeaderFooterProcessor (헤더/푸터 필터링)
  → XYCutPlusPlusSorter (읽기 순서)
```

---

## Task 2-1: TextChunk / TextLine 엔티티

**Java 소스**: `entities/` 내 TextChunk, TextLine 관련 클래스

**Go 대상**: `internal/entities/text_chunk.go`, `text_line.go`

```go
// internal/entities/text_chunk.go
package entities

type FontStyle struct {
    FontName   string
    FontSize   float64
    Bold       bool
    Italic     bool
    Color      [3]float64 // RGB
}

type TextChunk struct {
    BaseObject
    Text        string
    FontStyle   FontStyle
    Baseline    float64
    CharSpacing float64
    IsHidden    bool
    IsOffPage   bool
    IsTiny      bool
}

// internal/entities/text_line.go
type TextLine struct {
    BaseObject
    Chunks       []*TextChunk
    LineArtBullet *LineArtChunk  // 라인아트 불릿 연결
    Baseline     float64
    IsFirstLine  bool
    IsLastLine   bool
}
```

---

## Task 2-2: ContentFilterProcessor 포팅

**Java 소스**: `processors/ContentFilterProcessor.java`

**Go 대상**: `internal/processors/content_filter_processor.go`

**필터 로직**:
- `hidden-text`: `TextChunk.IsHidden == true` 제거
- `off-page`: BBox가 페이지 경계 밖인 요소 제거
- `tiny`: 폰트 크기 < 임계값인 요소 제거
- `hidden-ocg`: Optional Content Group 숨김 레이어 요소 제거

```go
func FilterContent(chunks []*entities.TextChunk, config *api.FilterConfig) []*entities.TextChunk {
    var result []*entities.TextChunk
    for _, chunk := range chunks {
        if config.DisableHiddenText && chunk.IsHidden { continue }
        if config.DisableOffPage && chunk.IsOffPage { continue }
        if config.DisableTiny && chunk.FontStyle.FontSize < TinyThreshold { continue }
        result = append(result, chunk)
    }
    return result
}
```

---

## Task 2-3: TextLineProcessor 포팅

**Java 소스**: `processors/TextLineProcessor.java` (127줄)

**Go 대상**: `internal/processors/text_line_processor.go`

**핵심 로직**:
1. 텍스트 청크를 baseline Y 좌표 기준으로 그룹화
2. 같은 라인의 청크를 X 좌표 순으로 정렬 후 병합
3. 인접 청크 사이 간격 계산 (공백 삽입 여부 판단)
4. LineArt 불릿(•, -, ※ 등)과 텍스트 라인 연결

```go
type TextLineProcessor struct {
    config *api.Config
}

func (p *TextLineProcessor) Process(
    chunks []*entities.TextChunk,
    lineArtChunks []*entities.LineArtChunk,
    ctx *containers.ProcessorContext,
) []*entities.TextLine {
    // 1. Y baseline 기준 청크 그룹화 (허용 오차: 0.5pt)
    // 2. X 좌표 정렬
    // 3. 청크 간 간격 → 공백 또는 탭 삽입
    // 4. LineArt 불릿 연결 (x 좌표 인접 여부로 판단)
}
```

---

## Task 2-4: HeadingProcessor 포팅

**Java 소스**: `processors/HeadingProcessor.java` (210줄)

**Go 대상**: `internal/processors/heading_processor.go`

**핵심 로직**:
1. 텍스트 라인의 폰트 크기 + 굵기 + 위치 분석
2. 폰트 스타일 그룹 생성 (크기별 클러스터링)
3. 제목 후보 필터링 (단독 라인, 짧은 텍스트, 굵은 폰트)
4. `HeadingProcessor` → `LevelProcessor` 로 제목 레벨 할당

```go
type HeadingCandidate struct {
    Line      *entities.TextLine
    FontSize  float64
    IsBold    bool
    Score     float64
}

func (p *HeadingProcessor) Process(
    lines []*entities.TextLine,
    ctx *containers.ProcessorContext,
) []*entities.Heading {
    // 1. 모든 라인의 폰트 통계 수집 (ModeWeightStatistics)
    // 2. 제목 후보 스코어링
    // 3. 제목 후보 → SemanticHeading 변환
}
```

---

## Task 2-5: TableBorderProcessor 포팅 (재귀 깊이 제한)

**Java 소스**: `processors/TableBorderProcessor.java` (252줄)

**Go 대상**: `internal/processors/table_border_processor.go`

**핵심 패턴** (Java ThreadLocal depth → Go 함수 매개변수):
```go
const maxTableDepth = 10

type TableBorderProcessor struct{}

func (p *TableBorderProcessor) Process(
    elements []entities.IObject,
    lineArts []*entities.LineArtChunk,
    ctx *containers.ProcessorContext,
) []entities.IObject {
    return p.processNode(elements, lineArts, ctx, 0)
}

func (p *TableBorderProcessor) processNode(
    elements []entities.IObject,
    lineArts []*entities.LineArtChunk,
    ctx *containers.ProcessorContext,
    depth int,
) []entities.IObject {
    if depth >= maxTableDepth {
        return elements
    }
    // 테두리 라인아트 감지
    // 테이블 경계 내 요소 그룹화
    // 중첩 테이블 재귀 처리 (depth+1)
    // 텍스트 청크가 테이블 경계에 걸친 경우 분할
}
```

---

## Task 2-6: ParagraphProcessor 포팅

**Java 소스**: `processors/ParagraphProcessor.java`

**Go 대상**: `internal/processors/paragraph_processor.go`

**핵심 로직**:
- 정렬 감지 (양쪽/왼쪽/오른쪽/가운데)
- 수평 겹침 검사
- 폰트 크기 일관성 체크
- 병합 확률 임계값: **75%**
- `SemanticParagraph` 생성

```go
const mergeProbabilityThreshold = 0.75

func (p *ParagraphProcessor) canMerge(line1, line2 *entities.TextLine) bool {
    // 정렬 일치 여부
    // 폰트 크기 유사도 (±10%)
    // 수평 겹침 비율
    // 동일 페이지 여부
    // → 확률 계산 후 0.75 이상이면 병합
}
```

---

## Task 2-7: ListProcessor 포팅

**Java 소스**: `processors/ListProcessor.java`

**Go 대상**: `internal/processors/list_processor.go`

**핵심 로직**:
- 텍스트 레이블 목록 감지 (1., 2., a., b., •, -, ※)
- 불릿 스타일 인식 (ordered/unordered)
- 수평 정렬 검증
- 베이스라인 차이 임계값: **1.2x**
- 다중 페이지 목록 연속 처리
- 한국어 첨부 패턴 처리: `붙\s*임`

```go
var (
    orderedBulletPattern   = regexp.MustCompile(`^(\d+\.|\([a-zA-Z]\)|[a-zA-Z]\.)`)
    unorderedBulletPattern = regexp.MustCompile(`^[•\-※◦▪▸►]`)
    koreanAttachPattern    = regexp.MustCompile(`^붙\s*임`)
)

const baselineDiffThreshold = 1.2
```

---

## Task 2-8: XYCutPlusPlusSorter 포팅 (읽기 순서)

**Java 소스**: `processors/readingorder/XYCutPlusPlusSorter.java` (677줄)

**Go 대상**: `internal/processors/readingorder/xycut_plus_plus_sorter.go`

**알고리즘**: arXiv:2504.10258 기반 XY-Cut++ 알고리즘

**주요 컴포넌트**:
```go
// 크로스 레이아웃 감지
type CrossLayoutDetector struct{}

// 밀도 기반 축 선택
type DensityAxisSelector struct{}

// 좁은 이상값 필터링 (narrow outlier filtering)
// 최근 커밋(8e3f74a)에서 코드 리뷰 수정됨
type NarrowOutlierFilter struct {
    MinWidthRatio float64 // 기본: 페이지 너비의 5%
}

type XYCutPlusPlusSorter struct {
    CrossLayoutDetector  *CrossLayoutDetector
    DensityAxisSelector  *DensityAxisSelector
    NarrowOutlierFilter  *NarrowOutlierFilter
}

func (s *XYCutPlusPlusSorter) Sort(
    elements []entities.IObject,
    pageWidth, pageHeight float64,
) []entities.IObject {
    // Phase 1: 좁은 이상값 필터링
    // Phase 2: 크로스 레이아웃 감지
    // Phase 3: 밀도 기반 X/Y 축 선택
    // Phase 4: 재귀적 분할 및 정렬
    // Phase 5: 이상값 재삽입
}
```

---

## Task 2-9: 나머지 프로세서 포팅

각 Java → Go 변환:

| Java | Go | 핵심 포인트 |
|------|-----|-----------|
| `SpecialTableProcessor.java` | `special_table_processor.go` | 경계 없는 테이블 (텍스트 정렬 기반) |
| `ClusterTableProcessor.java` | `cluster_table_processor.go` | DBSCAN 스타일 클러스터링 |
| `CaptionProcessor.java` | `caption_processor.go` | Figure/Table 캡션 감지 (정규식) |
| `HeaderFooterProcessor.java` | `header_footer_processor.go` | 페이지 상/하단 20% 영역 필터 |
| `HiddenTextProcessor.java` | `hidden_text_processor.go` | OCR 레이어 숨겨진 텍스트 |
| `StrikethroughProcessor.java` | `strikethrough_processor.go` | 취소선 감지 (experimental) |
| `LevelProcessor.java` | `level_processor.go` | 제목 레벨 1-6 할당 |
| `TaggedDocumentProcessor.java` | `tagged_document_processor.go` | Tagged PDF 구조 트리 처리 |

---

## Task 2-10: DocumentProcessor 메인 파이프라인 포팅

**Java 소스**: `processors/DocumentProcessor.java` (454줄)

**Go 대상**: `internal/processors/document_processor.go`

**파이프라인 흐름**:
```go
func (p *DocumentProcessor) Process(pdfPath string, config *api.Config) (*Document, error) {
    ctx := containers.NewProcessorContext()
    ctx.UseStructTree = config.UseStructTree

    // 1. PDF 로딩 (pdfcpu)
    doc, err := loadPDF(pdfPath, config.Password)

    // 2. 페이지 필터링 (--pages 옵션)
    pages := filterPages(doc.Pages, config.Pages)

    for _, page := range pages {
        // 3. ContentFilter
        chunks := contentFilter.Filter(page.Chunks, filterConfig)

        // 4. TextLine 처리
        lines := textLineProc.Process(chunks, page.LineArts, ctx)

        // 5. 테이블 감지 (method별 분기)
        var tableElements []entities.IObject
        switch config.TableMethod {
        case api.TableMethodCluster:
            tableElements = clusterTableProc.Process(lines, ctx)
        default:
            tableElements = tableBorderProc.Process(lines, page.LineArts, ctx)
            tableElements = specialTableProc.Process(tableElements, ctx)
        }

        // 6. 문단, 목록, 제목
        paragraphs := paragraphProc.Process(lines, ctx)
        lists := listProc.Process(paragraphs, ctx)
        headings := headingProc.Process(lists, ctx)
        leveled := levelProc.Process(headings, ctx)
        captioned := captionProc.Process(leveled, ctx)

        // 7. 헤더/푸터 (옵션)
        final := headerFooterProc.Process(captioned, config.IncludeHeaderFooter, ctx)

        // 8. 읽기 순서 정렬
        if config.ReadingOrder == api.ReadingOrderXYCut {
            final = xycutSorter.Sort(final, page.Width, page.Height)
        }

        page.Elements = final
    }

    // 9. ContentSanitizer (--sanitize)
    if config.Sanitize {
        sanitizer.Sanitize(doc, ctx)
    }

    return doc, nil
}
```

---

## Phase 2 완료 기준

- [ ] 샘플 PDF 1개를 Go로 처리 → Markdown 출력 생성
- [ ] 벤치마크 샘플 5개 Go vs Java 결과 비교 (구조 일치 확인)
- [ ] `go test ./internal/processors/...` 전체 통과
- [ ] XYCutPlusPlusSorter: 기존 Java 테스트 케이스 동일 결과
