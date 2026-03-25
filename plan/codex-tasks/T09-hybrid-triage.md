# T09: Hybrid 모드 포팅 (HTTP 클라이언트, Triage, Schema 변환)

## 선행 태스크
T03, T04 완료

## 참조 Java 파일
- `hybrid/HybridConfig.java` (200줄)
- `hybrid/HybridClient.java` (인터페이스)
- `hybrid/HybridClientFactory.java`
- `hybrid/DoclingFastServerClient.java` (308줄)
- `hybrid/DoclingSchemaTransformer.java`
- `hybrid/HancomClient.java`
- `hybrid/HancomSchemaTransformer.java`
- `hybrid/TriageProcessor.java`
- `hybrid/TriageLogger.java`
- `processors/HybridDocumentProcessor.java`

## 지시사항

### 1. internal/hybrid/hybrid_config.go

plan/06-phase4-hybrid.md의 Task 4-1 참조.
Java `HybridConfig.java` 포팅. 백엔드 상수 + HybridConfig struct + DefaultHybridConfig().

### 2. internal/hybrid/hybrid_client.go

plan/06-phase4-hybrid.md의 Task 4-2 참조.
`HybridClient` 인터페이스 + `ConvertRequest` / `ConvertResponse` 구조체.

### 3. internal/hybrid/docling_fast_server_client.go

Java `DoclingFastServerClient.java` (308줄) 포팅.
plan/06-phase4-hybrid.md의 Task 4-3 참조.

**OkHttp → net/http 변환**:
- `OkHttpClient` → `http.Client` (Timeout 설정)
- `MultipartBody.Builder` → `multipart.NewWriter(&buf)`
- `Response.body().string()` → `io.ReadAll(resp.Body)`
- OkHttp retry → 단순 재시도 없음 (첫 번째 오류에서 fallback)

메서드 구현:
- `Convert(req *ConvertRequest) (*ConvertResponse, error)` - multipart POST
- `HealthCheck() error` - GET /health
- `Close() error` - http.Client는 Close 불필요, no-op

### 4. internal/hybrid/docling_schema_transformer.go

Java `DoclingSchemaTransformer.java` 포팅.
plan/06-phase4-hybrid.md의 Task 4-6 참조.

Docling API JSON 응답 → `[]*entities.Page` 변환.
Docling 응답 JSON 스키마는 Java 소스에서 `DoclingSchemaTransformer.java` 파싱 로직 확인 후 Go struct 정의.

### 5. internal/hybrid/hancom_client.go (스텁)

Java `HancomClient.java` 포팅 (스텁으로만 구현):
```go
// HancomClient is a stub for the Hancom proprietary backend.
// Full implementation requires Hancom API credentials.
type HancomClient struct {
    config *HybridConfig
}

func (c *HancomClient) Convert(req *ConvertRequest) (*ConvertResponse, error) {
    return nil, fmt.Errorf("hancom backend not yet implemented")
}
```

### 6. internal/hybrid/hybrid_client_factory.go

Java `HybridClientFactory.java` 포팅:
```go
func NewHybridClient(config *HybridConfig) (HybridClient, error) {
    switch config.Backend {
    case BackendDoclingFast:
        return NewDoclingFastServerClient(config), nil
    case BackendHancom:
        return NewHancomClient(config), nil
    case BackendOff:
        return nil, nil
    default:
        return nil, fmt.Errorf("unknown hybrid backend: %s", config.Backend)
    }
}
```

### 7. internal/hybrid/triage_processor.go

Java `TriageProcessor.java` 완전 포팅.
plan/06-phase4-hybrid.md의 Task 4-4 참조.

**상수** (Java 소스에서 정확한 값 확인):
```go
const (
    LineRatioThreshold        = 0.3
    AlignedLineGroupsMin      = 5
    GridGapMultiplier         = 3.0
    MinLineCountForTable      = 8
    MinGridLines              = 3
    MinRowSeparatorPatterns   = 5
    MinAlignedShortLines      = 2
    MinConsecutivePatterns    = 2
    LargeImageRatio           = 0.11
    ImageAspectRatioThreshold = 1.75
)
```

`TriageSignals` struct + `TriageResult` struct + `Triage(page *entities.Page) *TriageResult`.

### 8. internal/hybrid/triage_logger.go

Java `TriageLogger.java` 포팅.
트리아지 결과를 구조화된 로그로 출력 (Go `log/slog` 사용).

### 9. internal/processors/hybrid_document_processor.go

Java `HybridDocumentProcessor.java` 완전 포팅.
plan/06-phase4-hybrid.md의 Task 4-5 참조.

6단계 파이프라인 구현:
- Phase 0: HealthCheck (Fallback 처리 포함)
- Phase 1: ContentFilter 적용
- Phase 2: Triage (auto/full 분기)
- Phase 3: Java/Backend 페이지 분리
- Phase 4: sync.WaitGroup으로 병렬 처리
- Phase 5: 페이지 순서 유지 병합
- Phase 6: 크로스 페이지 후처리 (목록 연속성)

### 10. 테스트

`go/tests/unit/hybrid/triage_processor_test.go`:
- 단순 텍스트 페이지 → JAVA 결정 테스트
- 많은 LineArtChunk 포함 페이지 → BACKEND 결정 테스트
- 신뢰도 점수 0.0~1.0 범위 확인

## 완료 기준
- `go build ./internal/hybrid/...` `./internal/processors/hybrid_document_processor.go` 성공
- `go test ./tests/unit/hybrid/...` 통과
- 실제 Docling 서버 없이도 HealthCheck 실패 → Fallback 경로 동작 확인
