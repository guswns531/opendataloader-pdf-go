# T03: Config, FilterConfig, CLI 옵션 포팅

## 선행 태스크
T01, T02 완료

## 참조 Java 파일
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/api/Config.java` (859줄)
- `java/opendataloader-pdf-cli/src/main/java/org/opendataloader/pdf/cli/CLIOptions.java`
- `java/opendataloader-pdf-cli/src/main/java/org/opendataloader/pdf/cli/CLIMain.java`

## 지시사항

### 1. internal/api/config.go

Java `Config.java` 전체를 Go로 변환. 주요 내용:

**상수** (Java static final → Go const):
```
ReadingOrderOff/XYCut
HybridModeAuto/Full
TableMethodDefault/Cluster
ImageFormatPNG/JPEG
ImageOutputOff/Embedded/External
FormatJSON/Text/HTML/PDF/Markdown/MarkdownWithHTML/MarkdownWithImages
ContentSafetyAll/HiddenText/OffPage/Tiny/HiddenOCG
```

**Config struct**: plan/03-phase1-infrastructure.md의 Task 1-3 참조.
Java getter/setter → Go public 필드로 변환.
`DefaultConfig()` 함수 구현 (기본값 포함).

### 2. internal/api/filter_config.go

plan/03-phase1-infrastructure.md의 Task 1-4 참조.
`FilterConfig` struct + `FilterConfigFromStrings([]string)` 함수.

### 3. internal/containers/processor_context.go

plan/03-phase1-infrastructure.md의 Task 1-5 참조.
Java ThreadLocal → Go ProcessorContext struct.
`NewProcessorContext()` + `NextID() string` (atomic int64 기반).

### 4. internal/cli/options.go

plan/03-phase1-infrastructure.md의 Task 1-6 참조.
- `CLIOptions` struct (23개 옵션 + ExportOptions bool)
- `AddFlags(cmd *cobra.Command, opts *CLIOptions)` 함수
- `(o *CLIOptions) ToConfig() *api.Config` 메서드

### 5. internal/cli/export_options.go

`--export-options` 플래그 처리. plan/07-phase5-integration.md의 Task 5-2 참조.

`OptionDefinition` struct:
```go
type OptionDefinition struct {
    Name        string      `json:"name"`
    ShortName   string      `json:"shortName,omitempty"`
    Description string      `json:"description"`
    Type        string      `json:"type"`   // "string", "boolean", "integer", "[]string"
    Default     interface{} `json:"default,omitempty"`
    Choices     []string    `json:"choices,omitempty"`
    Multiple    bool        `json:"multiple,omitempty"`
}
```

23개 옵션의 OptionDefinition 목록 정의. Java CLIOptions.java의 OPTION_DEFINITIONS 패턴과 동일 구조.
`ExportOptionsJSON()` 함수: JSON 출력 후 os.Exit(0).

### 6. cmd/opendataloader-pdf/main.go

plan/03-phase1-infrastructure.md의 Task 1-7 참조.
cobra 루트 커맨드 + `--export-options` 분기 + 파일/디렉토리 처리 + 종료 코드(0/1/2).

## 완료 기준
- `go build ./cmd/opendataloader-pdf/` 성공
- `./bin/opendataloader-pdf --help` 실행 시 23개 옵션 모두 표시
- `./bin/opendataloader-pdf --export-options` 실행 시 유효한 JSON 출력
- `go test ./internal/api/... ./internal/cli/...` 통과
