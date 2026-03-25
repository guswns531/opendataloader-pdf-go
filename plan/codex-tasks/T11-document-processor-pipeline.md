# T11: DocumentProcessor 메인 파이프라인 통합

## 선행 태스크
T04, T05, T06, T07, T08, T09 완료

## 참조 Java 파일
- `processors/DocumentProcessor.java` (454줄)
- `api/OpenDataLoaderPDF.java` (52줄)

## 지시사항

### 1. internal/processors/document_processor.go

Java `DocumentProcessor.java` (454줄) 완전 포팅.
plan/04-phase2-processors.md의 Task 2-10 참조.

pdfcpu 기반 PDF 로딩 구현:
```go
import pdfcpu_api "github.com/pdfcpu/pdfcpu/pkg/api"
import pdfcpu_model "github.com/pdfcpu/pdfcpu/pkg/model"

func loadPDF(path string, password string) (*entities.Document, error) {
    conf := pdfcpu_model.NewDefaultConfiguration()
    if password != "" {
        conf.UserPW = password
        conf.OwnerPW = password
    }
    // pdfcpu로 페이지 수 확인
    // pdfcpu로 각 페이지 텍스트 청크 추출
    // pdfcpu로 각 페이지 이미지 추출
    // pdfcpu로 각 페이지 선분(LineArt) 추출
}
```

**페이지 필터링** (`--pages` 옵션):
```go
// "1,3,5-7" → [1,3,5,6,7]
func parsePageRange(pagesStr string, totalPages int) ([]int, error) {
    // 쉼표로 분리
    // 범위(-)는 시작-끝 확장
    // 경계 검사 (1 ~ totalPages)
}
```

**전체 처리 파이프라인** (plan/04-phase2-processors.md Task 2-10 참조):
1. PDF 로딩
2. 페이지 필터링
3. ContentFilter
4. TextLine 처리
5. 테이블 감지 (method 분기)
6. 문단, 목록, 제목, 레벨, 캡션
7. 헤더/푸터 필터
8. 읽기 순서 정렬
9. ContentSanitizer (옵션)
10. 출력 생성 (formats별 분기)

Hybrid 모드 분기:
```go
if config.Hybrid != api.HybridModeOff {
    hybridProc := processors.NewHybridDocumentProcessor(...)
    return hybridProc.Process(doc, config, ctx)
}
// 일반 Java 경로
```

### 2. internal/api/opendataloader_pdf.go

Java `OpenDataLoaderPDF.java` (52줄) 포팅:
```go
package api

// ProcessFile processes a single PDF file and returns the results.
func ProcessFile(pdfPath string, config *Config) error {
    proc := processors.NewDocumentProcessor()
    return proc.Process(pdfPath, config)
}

// Shutdown releases any global resources.
func Shutdown() {
    // Go에서는 대부분 불필요하나 HTTP 클라이언트 등 정리
}
```

### 3. internal/cli/runner.go

CLI에서 파일/디렉토리 처리 로직 (CLIMain.java의 파일 처리 부분):
```go
// Run processes the given PDF files/directories.
func Run(opts *CLIOptions, args []string) error {
    config := opts.ToConfig()
    var exitCode int
    for _, arg := range args {
        if err := processPath(arg, config); err != nil {
            // 에러 기록, exitCode = 1, 계속 처리
        }
    }
    if exitCode != 0 {
        return fmt.Errorf("one or more files failed")
    }
    return nil
}

// processPath handles both files and directories recursively.
func processPath(path string, config *api.Config) error {
    info, err := os.Stat(path)
    if os.IsNotExist(err) {
        return fmt.Errorf("file not found: %s", path)
    }
    if info.IsDir() {
        return filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
            if !d.IsDir() && strings.ToLower(filepath.Ext(p)) == ".pdf" {
                return api.ProcessFile(p, config)
            }
            return nil
        })
    }
    if strings.ToLower(filepath.Ext(path)) != ".pdf" {
        return fmt.Errorf("not a PDF file: %s", path)
    }
    return api.ProcessFile(path, config)
}
```

### 4. 통합 테스트

`go/tests/integration/end_to_end_test.go`:
- `samples/` 디렉토리에서 샘플 PDF 찾아 처리
- markdown, json, html 출력 생성 확인
- 출력 파일 존재 여부 확인

```go
func TestEndToEnd(t *testing.T) {
    samplePDFs, _ := filepath.Glob("../../samples/**/*.pdf")
    if len(samplePDFs) == 0 {
        t.Skip("no sample PDFs found")
    }
    for _, pdf := range samplePDFs[:1] { // 1개만 빠른 테스트
        config := api.DefaultConfig()
        config.Formats = []string{"markdown", "json"}
        config.OutputDir = t.TempDir()
        err := api.ProcessFile(pdf, config)
        assert.NoError(t, err)
    }
}
```

## 완료 기준
- `go build ./...` 전체 성공
- `go test ./tests/integration/...` 통과
- `./bin/opendataloader-pdf samples/` 명령어 실행 시 markdown 출력 파일 생성
- `./bin/opendataloader-pdf --format json samples/` 실행 시 schema.json 유효성 검증 통과
