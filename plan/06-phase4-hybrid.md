# Phase 4: Hybrid 모드 + 고급 기능 포팅

선행 조건: Phase 2, 3 완료
예상 소요: 2주

---

## Task 4-1: HybridConfig 포팅

**Java 소스**: `hybrid/HybridConfig.java` (200줄)

**Go 대상**: `internal/hybrid/hybrid_config.go`

```go
package hybrid

const (
    BackendOff         = "off"
    BackendDoclingFast = "docling-fast"
    BackendHancom      = "hancom"
    BackendAzure       = "azure"
    BackendGoogle      = "google"

    TriageDecisionJava    = "JAVA"
    TriageDecisionBackend = "BACKEND"
)

type HybridConfig struct {
    Backend     string
    Mode        string // auto, full
    URL         string
    TimeoutMS   int
    Fallback    bool
    MaxConcurrency int
}

func DefaultHybridConfig() *HybridConfig {
    return &HybridConfig{
        Backend:        BackendOff,
        Mode:           "auto",
        TimeoutMS:      30000,
        MaxConcurrency: 4,
        Fallback:       false,
    }
}
```

---

## Task 4-2: HybridClient 인터페이스 포팅

**Java 소스**: `hybrid/HybridClient.java` (인터페이스)

**Go 대상**: `internal/hybrid/hybrid_client.go`

```go
type ConvertRequest struct {
    PDFPath  string
    PageNums []int
    Config   *HybridConfig
}

type ConvertResponse struct {
    Pages   []*entities.Page
    Error   error
}

// HybridClient is the interface for external AI backends.
type HybridClient interface {
    Convert(req *ConvertRequest) (*ConvertResponse, error)
    HealthCheck() error
    Close() error
}
```

---

## Task 4-3: DoclingFastServerClient 포팅

**Java 소스**: `hybrid/DoclingFastServerClient.java` (308줄, OkHttp → net/http)

**Go 대상**: `internal/hybrid/docling_fast_server_client.go`

**OkHttp → net/http 변환 패턴**:
```go
type DoclingFastServerClient struct {
    config     *HybridConfig
    httpClient *http.Client
    baseURL    string
}

func NewDoclingFastServerClient(config *HybridConfig) *DoclingFastServerClient {
    return &DoclingFastServerClient{
        config: config,
        httpClient: &http.Client{
            Timeout: time.Duration(config.TimeoutMS) * time.Millisecond,
        },
        baseURL: config.URL,
    }
}

// Java: OkHttp MultipartBody → Go: multipart.Writer
func (c *DoclingFastServerClient) Convert(req *ConvertRequest) (*ConvertResponse, error) {
    var buf bytes.Buffer
    w := multipart.NewWriter(&buf)

    // PDF 파일 첨부
    fw, err := w.CreateFormFile("pdf_file", filepath.Base(req.PDFPath))
    if err != nil {
        return nil, err
    }
    f, err := os.Open(req.PDFPath)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    if _, err := io.Copy(fw, f); err != nil {
        return nil, err
    }

    // 페이지 범위 첨부
    if len(req.PageNums) > 0 {
        pageStr := formatPageNums(req.PageNums)
        w.WriteField("pages", pageStr)
    }
    w.Close()

    httpReq, err := http.NewRequest("POST", c.baseURL+"/convert", &buf)
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", w.FormDataContentType())

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        if c.config.Fallback {
            return &ConvertResponse{Error: err}, nil // fallback signal
        }
        return nil, err
    }
    defer resp.Body.Close()

    return c.parseResponse(resp)
}

func (c *DoclingFastServerClient) HealthCheck() error {
    resp, err := c.httpClient.Get(c.baseURL + "/health")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("health check failed: %d", resp.StatusCode)
    }
    return nil
}
```

---

## Task 4-4: TriageProcessor 포팅

**Java 소스**: `hybrid/TriageProcessor.java`

**Go 대상**: `internal/hybrid/triage_processor.go`

**상수** (Java에서 추출):
```go
const (
    LineRatioThreshold       = 0.3
    AlignedLineGroupsMin     = 5      // 증가됨 (이전: 3)
    GridGapMultiplier        = 3.0    // text height의 배수
    MinLineCountForTable     = 8
    MinGridLines             = 3
    MinRowSeparatorPatterns  = 5
    MinAlignedShortLines     = 2
    MinConsecutivePatterns   = 2
    LargeImageRatio          = 0.11   // 페이지 면적의 11%
    ImageAspectRatioThreshold = 1.75
)

type TriageSignals struct {
    LineChunkCount      int
    TextChunkCount      int
    BaselineGroups      int
    HasTableBorder      bool
    HorizontalLines     int
    VerticalLines       int
    GridPatterns        int
    AlignedShortLines   int
    TextPatternCount    int
    LargeImageCount     int
    RowSeparatorCount   int
}

type TriageResult struct {
    Decision   string  // "JAVA" or "BACKEND"
    Confidence float64 // 0.0-1.0
    Signals    TriageSignals
}

func (p *TriageProcessor) Triage(page *entities.Page) *TriageResult {
    signals := p.extractSignals(page)
    decision, confidence := p.decide(signals)
    return &TriageResult{
        Decision:   decision,
        Confidence: confidence,
        Signals:    signals,
    }
}
```

---

## Task 4-5: HybridDocumentProcessor 포팅

**Java 소스**: `processors/HybridDocumentProcessor.java`

**Go 대상**: `internal/processors/hybrid_document_processor.go`

**6단계 파이프라인**:
```go
func (p *HybridDocumentProcessor) Process(
    doc *entities.Document,
    config *api.Config,
    ctx *containers.ProcessorContext,
) (*entities.Document, error) {

    // Phase 0: 백엔드 가용성 체크
    client := hybrid.GetClient(config)
    if err := client.HealthCheck(); err != nil {
        if !config.HybridFallback {
            return nil, fmt.Errorf("backend unavailable: %w", err)
        }
        // 전체 fallback to Java
        return p.javaProcessor.Process(doc, config, ctx)
    }

    // Phase 1: 모든 페이지 ContentFilter
    for _, page := range doc.Pages {
        page.Chunks = contentFilter.Filter(page.Chunks, filterConfig)
    }

    // Phase 2: 페이지 트리아지 (auto vs full)
    decisions := make(map[int]string) // pageNum → JAVA/BACKEND
    if config.HybridMode == api.HybridModeFull {
        for _, page := range doc.Pages {
            decisions[page.Number] = hybrid.TriageDecisionBackend
        }
    } else {
        triageProc := hybrid.NewTriageProcessor()
        for _, page := range doc.Pages {
            result := triageProc.Triage(page)
            decisions[page.Number] = result.Decision
        }
    }

    // Phase 3: Java/Backend 페이지 분리
    var javaPages, backendPages []*entities.Page
    for _, page := range doc.Pages {
        if decisions[page.Number] == hybrid.TriageDecisionBackend {
            backendPages = append(backendPages, page)
        } else {
            javaPages = append(javaPages, page)
        }
    }

    // Phase 4: 병렬 처리
    var wg sync.WaitGroup
    var javaResults, backendResults []*entities.Page
    var javaErr, backendErr error

    wg.Add(2)
    go func() {
        defer wg.Done()
        javaResults, javaErr = p.processJavaPages(javaPages, config, ctx)
    }()
    go func() {
        defer wg.Done()
        backendResults, backendErr = p.processBackendPages(backendPages, client, config)
        if backendErr != nil && config.HybridFallback {
            // Fallback: backend 실패 페이지를 Java로 재처리
            backendResults, backendErr = p.processJavaPages(backendPages, config, ctx)
        }
    }()
    wg.Wait()

    // Phase 5: 페이지 순서 유지하며 병합
    doc.Pages = mergePageResults(doc.Pages, javaResults, backendResults)

    // Phase 6: 크로스 페이지 후처리 (목록 연속성 등)
    p.crossPagePostProcess(doc, ctx)

    return doc, nil
}
```

---

## Task 4-6: DoclingSchemaTransformer 포팅

**Java 소스**: `hybrid/DoclingSchemaTransformer.java`

**Go 대상**: `internal/hybrid/docling_schema_transformer.go`

Docling API 응답 JSON → 내부 `entities.Page` 구조체 변환:
```go
type DoclingResponse struct {
    Pages []DoclingPage `json:"pages"`
}

type DoclingPage struct {
    PageNum  int              `json:"page_no"`
    Elements []DoclingElement `json:"body"`
}

func TransformDoclingResponse(resp *DoclingResponse) ([]*entities.Page, error) {
    pages := make([]*entities.Page, len(resp.Pages))
    for i, dp := range resp.Pages {
        pages[i] = transformDoclingPage(&dp)
    }
    return pages, nil
}
```

---

## Task 4-7: ContentSanitizer 포팅

**Java 소스**: `utils/ContentSanitizer.java`

**Go 대상**: `internal/utils/content_sanitizer.go`

```go
import "regexp"

type SanitizationRule struct {
    Name        string
    Pattern     *regexp.Regexp
    Replacement string
}

var DefaultRules = []SanitizationRule{
    {Name: "email",      Pattern: regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`), Replacement: "[EMAIL]"},
    {Name: "phone",      Pattern: regexp.MustCompile(`(\+?[\d\s\-().]{10,})`), Replacement: "[PHONE]"},
    {Name: "ip",         Pattern: regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`), Replacement: "[IP]"},
    {Name: "credit_card",Pattern: regexp.MustCompile(`\b\d{4}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`), Replacement: "[CARD]"},
    {Name: "url",        Pattern: regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`), Replacement: "[URL]"},
}

// Sanitize applies all rules to all text content in the document.
// 겹치는 매칭은 첫 번째 규칙 우선 (Java와 동일)
func Sanitize(doc *entities.Document, rules []SanitizationRule) {
    for _, page := range doc.Pages {
        for _, el := range page.Elements {
            sanitizeElement(el, rules)
        }
    }
}
```

---

## Task 4-8: PDFWriter (Tagged PDF) 포팅

**Java 소스**: `pdf/PDFWriter.java`

**Go 대상**: `internal/pdf/pdf_writer.go`

```go
// pdfcpu 기반 Tagged PDF 생성
// PDF/UA 접근성 태그 추가
// Structure Tree 생성 (Document → Article → Section → P/H1-H6/Table/List)
type PDFWriter struct {
    config *api.Config
}

func (w *PDFWriter) Write(doc *entities.Document, outputPath string) error {
    // pdfcpu API로 Tagged PDF 생성
    // veraPDF(MPL-2.0) 검증 로직 적용
}
```

---

## Task 4-9: StrikethroughProcessor 포팅 (experimental)

**Java 소스**: `processors/StrikethroughProcessor.java`

**Go 대상**: `internal/processors/strikethrough_processor.go`

```go
// --detect-strikethrough 옵션 활성화 시 실행
// LineArt 중 텍스트 중간을 가로지르는 수평선 감지
// baseline과 lineArt Y 좌표 비교 (허용 오차: 폰트 크기의 20-60%)
func (p *StrikethroughProcessor) Process(
    lines []*entities.TextLine,
    lineArts []*entities.LineArtChunk,
) []*entities.TextLine {
    for _, line := range lines {
        for _, la := range lineArts {
            if isStrikethrough(line, la) {
                line.HasStrikethrough = true
            }
        }
    }
    return lines
}
```

---

## Phase 4 완료 기준

- [ ] `--hybrid docling-fast --hybrid-mode full` 옵션으로 실제 Docling 서버 연결 테스트
- [ ] `--hybrid-fallback` 옵션: 백엔드 다운 시 Java 경로로 자동 전환
- [ ] `--sanitize` 옵션: 이메일, 전화번호 등 마스킹 확인
- [ ] `--detect-strikethrough` 옵션 동작 확인
- [ ] Phase 4 완료 후 벤치마크 Triage Recall ≥ 0.95 확인
