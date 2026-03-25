# Phase 1: 기반 인프라

선행 조건: 없음 (첫 번째 Phase)
예상 소요: 1~2주

---

## Task 1-1: go.mod 초기화 및 프로젝트 스캐폴딩

**목표**: Go 모듈 초기화, 디렉토리 구조 생성, 기본 의존성 설치

```bash
# 작업 위치: tripoli/go/
cd tripoli/go
go mod init github.com/opendataloader-project/opendataloader-pdf-go
go get github.com/pdfcpu/pdfcpu@v0.9.0
go get github.com/spf13/cobra@v1.8.1
go get github.com/spf13/pflag@v1.0.5
go get github.com/stretchr/testify@v1.9.0
go get github.com/goccy/go-json@v0.10.3
```

**생성 파일**: `go.mod`, `go.sum`, 전체 디렉토리 구조 (02-project-structure.md 참조)

---

## Task 1-2: 핵심 인터페이스 및 엔티티 정의

**목표**: `IObject` 인터페이스, `BoundingBox`, `BaseObject` 포팅

**Java 소스**: `entities/` 전체

**Go 대상 파일**: `internal/entities/base_object.go`

```go
// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package entities

// BoundingBox represents a rectangular bounding box in PDF coordinate space.
// PDF 좌표계: 좌하단 원점, Y축 위 방향
type BoundingBox struct {
    X      float64
    Y      float64
    Width  float64
    Height float64
    Page   int
}

// IObject is the base interface for all semantic PDF elements.
type IObject interface {
    GetID() string
    GetBBox() BoundingBox
    GetObjectType() ObjectType
}

type ObjectType string

const (
    ObjectTypeParagraph  ObjectType = "paragraph"
    ObjectTypeHeading    ObjectType = "heading"
    ObjectTypeTable      ObjectType = "table"
    ObjectTypeList       ObjectType = "list"
    ObjectTypeImage      ObjectType = "image"
    ObjectTypeFormula    ObjectType = "formula"
    ObjectTypeCaption    ObjectType = "caption"
    ObjectTypeHeaderFooter ObjectType = "header_footer"
)

// BaseObject provides common fields for all semantic elements.
type BaseObject struct {
    ID   string
    BBox BoundingBox
}

func (b *BaseObject) GetID() string           { return b.ID }
func (b *BaseObject) GetBBox() BoundingBox    { return b.BBox }
```

---

## Task 1-3: Config 구조체 포팅

**목표**: `Config.java` (859줄) 전체를 Go struct + 상수로 변환

**Java 소스**: `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/api/Config.java`

**Go 대상 파일**: `internal/api/config.go`

**주요 상수 목록** (Java static final → Go const):
```go
const (
    ReadingOrderOff   = "off"
    ReadingOrderXYCut = "xycut"

    HybridModeAuto = "auto"
    HybridModeFull = "full"

    TableMethodDefault = "default"
    TableMethodCluster = "cluster"

    ImageFormatPNG  = "png"
    ImageFormatJPEG = "jpeg"

    ImageOutputOff      = "off"
    ImageOutputEmbedded = "embedded"
    ImageOutputExternal = "external"

    FormatJSON              = "json"
    FormatText              = "text"
    FormatHTML              = "html"
    FormatPDF               = "pdf"
    FormatMarkdown          = "markdown"
    FormatMarkdownWithHTML  = "markdown-with-html"
    FormatMarkdownWithImages = "markdown-with-images"
)
```

**Config 구조체** (Java getters/setters → Go public fields):
```go
type Config struct {
    OutputDir              string
    Password               string
    Formats                []string
    Quiet                  bool
    ContentSafetyOff       []string
    Sanitize               bool
    KeepLineBreaks         bool
    ReplaceInvalidChars    string
    UseStructTree          bool
    TableMethod            string
    ReadingOrder           string
    DetectStrikethrough    bool
    MarkdownPageSeparator  string
    TextPageSeparator      string
    HTMLPageSeparator      string
    ImageOutput            string
    ImageFormat            string
    ImageDir               string
    Pages                  string
    IncludeHeaderFooter    bool
    Hybrid                 string
    HybridMode             string
    HybridURL              string
    HybridTimeout          int
    HybridFallback         bool
}

func DefaultConfig() *Config {
    return &Config{
        ReplaceInvalidChars: " ",
        TableMethod:         TableMethodDefault,
        ReadingOrder:        ReadingOrderXYCut,
        ImageOutput:         ImageOutputExternal,
        ImageFormat:         ImageFormatPNG,
        HybridTimeout:       30000,
        HybridMode:          HybridModeAuto,
    }
}
```

---

## Task 1-4: FilterConfig 포팅

**목표**: `FilterConfig.java` → `internal/api/filter_config.go`

```go
type FilterConfig struct {
    DisableHiddenText bool
    DisableOffPage    bool
    DisableTiny       bool
    DisableHiddenOCG  bool
}

func FilterConfigFromStrings(flags []string) *FilterConfig {
    fc := &FilterConfig{}
    for _, f := range flags {
        switch f {
        case "all":
            fc.DisableHiddenText = true
            fc.DisableOffPage = true
            fc.DisableTiny = true
            fc.DisableHiddenOCG = true
        case "hidden-text":
            fc.DisableHiddenText = true
        case "off-page":
            fc.DisableOffPage = true
        case "tiny":
            fc.DisableTiny = true
        case "hidden-ocg":
            fc.DisableHiddenOCG = true
        }
    }
    return fc
}
```

---

## Task 1-5: StaticLayoutContainers 포팅 (ThreadLocal 제거)

**목표**: `StaticLayoutContainers.java` (ThreadLocal) → Go ProcessorContext 구조체

**Java 소스**: `containers/StaticLayoutContainers.java`

**핵심 변환**: Java `ThreadLocal` → Go 함수 매개변수로 전달하는 `ProcessorContext`

```go
// internal/containers/processor_context.go

package containers

import (
    "sync/atomic"
)

// ProcessorContext replaces Java ThreadLocal<> pattern.
// 각 PDF 처리 고루틴이 독립적인 컨텍스트를 가짐.
type ProcessorContext struct {
    idCounter       int64  // atomic
    Headings        []interface{}
    ImageIndex      int
    UseStructTree   bool
    ContrastConsumer func(float64)
    ImageDir        string
    EmbedImages     bool
    ImageFormat     string
}

func NewProcessorContext() *ProcessorContext {
    return &ProcessorContext{}
}

func (c *ProcessorContext) NextID() string {
    id := atomic.AddInt64(&c.idCounter, 1)
    return fmt.Sprintf("%d", id)
}
```

---

## Task 1-6: CLI 옵션 포팅 (cobra)

**목표**: `CLIOptions.java` (23개 옵션) → cobra 플래그 정의

**Java 소스**: `java/opendataloader-pdf-cli/src/main/java/org/opendataloader/pdf/cli/CLIOptions.java`

**Go 대상 파일**: `internal/cli/options.go`

**OPTION_DEFINITIONS 패턴 유지**: `--export-options` 플래그로 `options.json` 재생성 지원

```go
// internal/cli/options.go
package cli

import (
    "github.com/spf13/cobra"
    "github.com/opendataloader-project/opendataloader-pdf-go/internal/api"
)

type CLIOptions struct {
    OutputDir             string
    Password              string
    Format                []string
    Quiet                 bool
    ContentSafetyOff      []string
    Sanitize              bool
    KeepLineBreaks        bool
    ReplaceInvalidChars   string
    UseStructTree         bool
    TableMethod           string
    ReadingOrder          string
    DetectStrikethrough   bool
    MarkdownPageSeparator string
    TextPageSeparator     string
    HTMLPageSeparator     string
    ImageOutput           string
    ImageFormat           string
    ImageDir              string
    Pages                 string
    IncludeHeaderFooter   bool
    Hybrid                string
    HybridMode            string
    HybridURL             string
    HybridTimeout         int
    HybridFallback        bool
    ExportOptions         bool   // --export-options (options.json 재생성)
}

func AddFlags(cmd *cobra.Command, opts *CLIOptions) {
    f := cmd.Flags()
    f.StringVarP(&opts.OutputDir, "output-dir", "o", "", "Output directory")
    f.StringVarP(&opts.Password, "password", "p", "", "PDF encryption password")
    f.StringSliceVarP(&opts.Format, "format", "f", []string{"markdown"}, "Output formats: json,text,html,pdf,markdown,markdown-with-html,markdown-with-images")
    f.BoolVarP(&opts.Quiet, "quiet", "q", false, "Suppress console output")
    f.StringSliceVar(&opts.ContentSafetyOff, "content-safety-off", nil, "Disable content safety filters: all,hidden-text,off-page,tiny,hidden-ocg")
    f.BoolVar(&opts.Sanitize, "sanitize", false, "Sanitize PII (emails, phones, IPs, credit cards, URLs)")
    f.BoolVar(&opts.KeepLineBreaks, "keep-line-breaks", false, "Preserve original line breaks")
    f.StringVar(&opts.ReplaceInvalidChars, "replace-invalid-chars", " ", "Replacement for invalid characters")
    f.BoolVar(&opts.UseStructTree, "use-struct-tree", false, "Use tagged PDF structure tree for reading order")
    f.StringVar(&opts.TableMethod, "table-method", "default", "Table detection method: default,cluster")
    f.StringVar(&opts.ReadingOrder, "reading-order", "xycut", "Reading order algorithm: off,xycut")
    f.BoolVar(&opts.DetectStrikethrough, "detect-strikethrough", false, "Detect strikethrough text (experimental)")
    f.StringVar(&opts.MarkdownPageSeparator, "markdown-page-separator", "", "Page separator for Markdown output")
    f.StringVar(&opts.TextPageSeparator, "text-page-separator", "", "Page separator for Text output")
    f.StringVar(&opts.HTMLPageSeparator, "html-page-separator", "", "Page separator for HTML output")
    f.StringVar(&opts.ImageOutput, "image-output", "external", "Image output mode: off,embedded,external")
    f.StringVar(&opts.ImageFormat, "image-format", "png", "Image format: png,jpeg")
    f.StringVar(&opts.ImageDir, "image-dir", "", "Image extraction directory")
    f.StringVar(&opts.Pages, "pages", "", "Page range (e.g., 1,3,5-7)")
    f.BoolVar(&opts.IncludeHeaderFooter, "include-header-footer", false, "Include headers and footers")
    f.StringVar(&opts.Hybrid, "hybrid", "off", "Hybrid backend: off,docling-fast")
    f.StringVar(&opts.HybridMode, "hybrid-mode", "auto", "Hybrid triage mode: auto,full")
    f.StringVar(&opts.HybridURL, "hybrid-url", "", "Custom hybrid backend URL")
    f.IntVar(&opts.HybridTimeout, "hybrid-timeout", 30000, "Hybrid request timeout in ms")
    f.BoolVar(&opts.HybridFallback, "hybrid-fallback", false, "Fallback to Java on backend error")
    f.BoolVar(&opts.ExportOptions, "export-options", false, "Export CLI options as JSON to stdout")
}

func (o *CLIOptions) ToConfig() *api.Config {
    return &api.Config{
        OutputDir:             o.OutputDir,
        Password:              o.Password,
        Formats:               o.Format,
        Quiet:                 o.Quiet,
        ContentSafetyOff:      o.ContentSafetyOff,
        Sanitize:              o.Sanitize,
        KeepLineBreaks:        o.KeepLineBreaks,
        ReplaceInvalidChars:   o.ReplaceInvalidChars,
        UseStructTree:         o.UseStructTree,
        TableMethod:           o.TableMethod,
        ReadingOrder:          o.ReadingOrder,
        DetectStrikethrough:   o.DetectStrikethrough,
        MarkdownPageSeparator: o.MarkdownPageSeparator,
        TextPageSeparator:     o.TextPageSeparator,
        HTMLPageSeparator:     o.HTMLPageSeparator,
        ImageOutput:           o.ImageOutput,
        ImageFormat:           o.ImageFormat,
        ImageDir:              o.ImageDir,
        Pages:                 o.Pages,
        IncludeHeaderFooter:   o.IncludeHeaderFooter,
        Hybrid:                o.Hybrid,
        HybridMode:            o.HybridMode,
        HybridURL:             o.HybridURL,
        HybridTimeout:         o.HybridTimeout,
        HybridFallback:        o.HybridFallback,
    }
}
```

---

## Task 1-7: main.go (CLIMain) 포팅

**Java 소스**: `CLIMain.java`
- 0 = 성공, 1 = 처리 실패, 2 = 파싱 오류
- `--export-options` → options.json 출력
- 디렉토리 재귀 처리 (`.pdf` 확장자만)
- `OpenDataLoaderPDF.shutdown()` → defer 정리

```go
// cmd/opendataloader-pdf/main.go
func main() {
    opts := &cli.CLIOptions{}
    rootCmd := &cobra.Command{
        Use:   "opendataloader-pdf [flags] <file.pdf> [file2.pdf ...]",
        Short: "PDF parser for AI-ready data extraction",
        RunE: func(cmd *cobra.Command, args []string) error {
            if opts.ExportOptions {
                return cli.ExportOptionsJSON(opts)
            }
            return cli.Run(opts, args)
        },
    }
    cli.AddFlags(rootCmd, opts)
    if err := rootCmd.Execute(); err != nil {
        os.Exit(2)
    }
}
```

---

## Task 1-8: PDF 파일 로딩 (pdfcpu 통합)

**목표**: veraPDF/PDFBox의 PDF 로딩 → pdfcpu 기반으로 대체

**Go 대상 파일**: `internal/processors/document_processor.go` (초기 스텁)

```go
// pdfcpu로 PDF 열기 + 페이지 정보 추출
import "github.com/pdfcpu/pdfcpu/pkg/api"

func loadPDF(path string, password string) (*PDFDocument, error) {
    conf := pdfcpu_model.NewDefaultConfiguration()
    if password != "" {
        conf.UserPW = password
        conf.OwnerPW = password
    }
    // pdfcpu API로 페이지 수, 메타데이터 추출
}
```

---

## Phase 1 완료 기준

- [ ] `go build ./cmd/opendataloader-pdf/` 성공
- [ ] `./bin/opendataloader-pdf --help` 모든 23개 옵션 표시
- [ ] `./bin/opendataloader-pdf --export-options` → 유효한 JSON 출력
- [ ] `npm run sync` 후 Python/Node.js 래퍼 옵션 동기화 확인
- [ ] `go test ./internal/api/...` 통과
