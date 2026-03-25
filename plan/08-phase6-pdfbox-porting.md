# Phase 6: Apache PDFBox 직접 포팅

선행 조건: Phase 1~5 완료 (go/ 기본 구조, pdfcpu 설치됨)
예상 소요: 2~3주
라이센스: **Apache-2.0** (PDFBox 원본과 동일)

---

## PDFBox가 이 프로젝트에서 하는 일 (분석 결과)

### 직접 사용 (PDFWriter.java만)
| PDFBox 클래스 | 역할 |
|--------------|------|
| `Loader.loadPDF()` | PDF 파일 로딩 |
| `PDDocument` | PDF 문서 객체 |
| `PDAnnotationSquare` | 컬러 사각형 주석 |
| `PDOptionalContentGroup` | OCG 레이어 |
| `PDOptionalContentProperties` | OCG 프로퍼티 관리 |
| `PDColor` / `PDDeviceRGB` | RGB 색상 |
| `PDRectangle` | 사각형 영역 |
| `COSDictionary` / `COSName` | 저수준 COS 객체 |

### 간접 사용 (veraPDF 내부 → DocumentProcessor)
veraPDF의 `org.verapdf.pd.PDDocument`가 실제 PDF 콘텐츠 파싱을 담당:
- 텍스트 청크 추출 (좌표, 폰트, 베이스라인)
- 이미지 추출 (좌표, 크기)
- 선분/선아트 추출 (수평/수직 LineArt)
- 페이지 크기, 메타데이터

---

## 포팅 전략

Go에서는 `pdfcpu`를 기반으로 두 레이어를 구현:

```
pkg/pdfbox/
├── loader/          # PDF 파일 로딩 (pdfcpu 기반)
├── extractor/       # 콘텐츠 추출 (텍스트, 이미지, 선분)
├── annotation/      # 주석 추가 (PDFWriter 기능)
├── ocg/             # Optional Content Groups (레이어)
└── model/           # PDDocument, PDPage, PDRectangle 등 모델
```

**라이센스**: `pkg/pdfbox/` 모든 파일에 Apache-2.0 헤더 + PDFBox 원본 참조 주석:
```go
// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0
//
// This package provides functionality equivalent to Apache PDFBox 3.0.4
// (https://pdfbox.apache.org/) for PDF content extraction and annotation,
// implemented using pdfcpu (https://github.com/pdfcpu/pdfcpu).
```

---

## Task 구성

| 태스크 | 상태 | 내용 |
|--------|------|------|
| T13 | ✅ 완료 | PDF 콘텐츠 추출 (텍스트 청크, 이미지, 선분) — 스텁 → 실구현 |
| T14 | ✅ 완료 | PDFWriter 포팅 (주석, OCG 레이어) |
| T15 | ✅ 완료 | DocumentProcessor 스텁 제거 → T13 실구현으로 교체 |
| T16 | ⏳ 대기 | 텍스트 추출 품질 수정 (TJ kerning, TrimSpace, WinAnsiEncoding) |
| T17 | ⏳ 대기 | 다단 컬럼 읽기 순서 수정 (XYCutPlusPlusSorter 2-column 지원) |

---

## T13: PDF 콘텐츠 추출 실구현

### pkg/pdfbox/model/ (모델 레이어)

pdfcpu 기반 PDF 구조 모델:
```go
// PDDocument: PDF 문서
type PDDocument struct {
    path     string
    password string
    ctx      *model.Context  // pdfcpu context
    Pages    []*PDPage
}

// PDPage: 단일 페이지
type PDPage struct {
    Number  int
    Width   float64
    Height  float64
    Rotate  int
}
```

### pkg/pdfbox/extractor/ (콘텐츠 추출)

pdfcpu의 content stream 파싱으로 추출:

**텍스트 추출** (`text_extractor.go`):
```go
type ExtractedText struct {
    Text      string
    X, Y      float64    // 좌하단 기준 좌표
    Width     float64
    Height    float64
    FontName  string
    FontSize  float64
    Bold      bool
    Italic    bool
    Color     [3]float64 // RGB
    Baseline  float64
    Page      int
}

func ExtractTextChunks(doc *PDDocument, pageNum int) ([]*ExtractedText, error)
```

**이미지 추출** (`image_extractor.go`):
```go
type ExtractedImage struct {
    X, Y     float64
    Width    float64
    Height   float64
    Data     []byte
    Format   string  // "png", "jpeg"
    Page     int
}

func ExtractImages(doc *PDDocument, pageNum int, outputDir string) ([]*ExtractedImage, error)
```

**선분 추출** (`lineArt_extractor.go`):
```go
type ExtractedLineArt struct {
    X, Y         float64
    Width        float64
    Height       float64
    IsHorizontal bool
    IsVertical   bool
    LineWidth    float64
    Page         int
}

func ExtractLineArts(doc *PDDocument, pageNum int) ([]*ExtractedLineArt, error)
```

---

## T14: PDFWriter 포팅

### pkg/pdfbox/annotation/ (주석 레이어)

Java PDFWriter.java 의 핵심 기능:
- 컬러 사각형 주석 추가 (PDAnnotationSquare → pdfcpu Rectangle annotation)
- 내용 텍스트 (tooltip) 설정
- 레이어(OCG) 할당

```go
type SquareAnnotation struct {
    Rect     [4]float64  // [x, y, width, height]
    Color    [3]float64  // RGB 0.0~1.0
    Opacity  float64
    Contents string
    Layer    string  // OCG 레이어명
}

func AddSquareAnnotation(doc *PDDocument, pageNum int, ann *SquareAnnotation) error
```

### pkg/pdfbox/ocg/ (Optional Content Groups)

OCG 레이어 관리:
```go
type OptionalContentGroup struct {
    Name    string
    Visible bool
}

func AddOCG(doc *PDDocument, group *OptionalContentGroup) error
func SetOCGEnabled(doc *PDDocument, name string, enabled bool) error
```

### internal/pdf/pdf_writer.go 완성

T11에서 스텁으로 남긴 PDFWriter를 pkg/pdfbox 기반으로 완전 구현:
- PDFLayer 열거형 (content, table cells, list items, ...)
- 각 시맨틱 요소별 색상 매핑 (헤딩=파랑, 리스트=초록, ...)
- updatePDF() 메서드

---

## T15: DocumentProcessor 스텁 교체

T11의 스텁 텍스트 추출을 T13 실구현으로 교체:

```go
// Before (stub):
func (p *DocumentProcessor) loadPDFContent(path string) (*entities.Document, error) {
    // TODO: real extraction
    return &entities.Document{...}, nil
}

// After (real implementation using pkg/pdfbox):
func (p *DocumentProcessor) loadPDFContent(path string, config *api.Config) (*entities.Document, error) {
    doc, err := pdfbox_loader.Open(path, config.Password)
    defer doc.Close()

    document := &entities.Document{}
    for pageNum := 0; pageNum < doc.PageCount(); pageNum++ {
        page := &entities.Page{PageMetadata: entities.PageMetadata{
            Number: pageNum + 1,
            Width:  doc.Pages[pageNum].Width,
            Height: doc.Pages[pageNum].Height,
        }}

        // 텍스트 청크 추출
        texts, _ := extractor.ExtractTextChunks(doc, pageNum)
        for _, t := range texts {
            page.Chunks = append(page.Chunks, textToChunk(t))
        }

        // 이미지 추출
        images, _ := extractor.ExtractImages(doc, pageNum, config.ImageDir)
        // ...

        // 선분 추출
        lineArts, _ := extractor.ExtractLineArts(doc, pageNum)
        // ...

        document.Pages = append(document.Pages, page)
    }
    return document, nil
}
```

---

## THIRD_PARTY_LICENSES.md 추가 항목

```markdown
## PDFBox Equivalent Implementation

The pkg/pdfbox/ package provides functionality equivalent to Apache PDFBox 3.0.4
and is implemented using pdfcpu. The implementation is licensed under Apache-2.0.

| Original Library | Version | License | URL |
|-----------------|---------|---------|-----|
| Apache PDFBox | 3.0.4 | Apache-2.0 | https://pdfbox.apache.org/ |
```

---

---

## T16: 텍스트 추출 품질 수정

벤치마크 실행 후 발견된 3가지 버그 수정. 수정 파일: `pkg/pdfbox/extractor/`.

### 버그 1: TJ 배열 kerning 오프셋 무시 → 단어 사이 공백 손실

`decodeTJText` 함수가 TJ 배열의 숫자 항목을 완전히 무시. 숫자가 -100 이하이면 공백 삽입:

```go
const tjSpaceThreshold = -100.0
// item.kind == "number" && v < tjSpaceThreshold → sb.WriteByte(' ')
```

### 버그 2: appendText에서 strings.TrimSpace → 의도적 공백 제거

`text_extractor.go`의 `appendText`에서 `strings.TrimSpace(text)` 제거. 빈 문자열 체크만 유지.

### 버그 3: font encoding 미처리 → 8비트 문자열 깨짐

`normalizePDFString`에 UTF-16 BOM 처리 외에 Windows-1252 (WinAnsiEncoding) 변환 추가:
- 0x80~0x9F: `win1252Extras` 맵으로 유니코드 변환
- 0xA0~0xFF: Latin-1과 동일 (rune 직변환)
- 0x00~0x7F: ASCII 그대로

기대 결과: `"AMulti-ObjectRectiedAttentionNetwork"` → `"A Multi-Object Rectified Attention Network"`

---

## T17: 다단 컬럼 읽기 순서 수정

2단 컬럼 학술 논문에서 좌·우 컬럼 텍스트가 섞이는 문제 수정. 수정 파일: `internal/processors/xycut_plus_plus_sorter.go`.

### 문제

XYCutPlusPlusSorter가 수직 갭을 기준으로 페이지를 분할하지만 2단 레이아웃에서 동일 Y 위치의 좌·우 블록이 번갈아 출력됨.

### 수정 방향

1. 페이지 너비의 20~80% 사이에 수직 갭이 존재하면 2-column 레이아웃 판정
2. 좌 컬럼(x < gap) 블록 전체를 먼저, 우 컬럼(x ≥ gap) 블록을 나중에 정렬
3. 각 컬럼 내부는 기존 Y-기반 정렬 유지
4. 3단 이상도 동일 원리 (재귀적 수직 분할)

기대 결과: NID 벤치마크 점수 ≥ 0.85 달성

---

## Phase 6 완료 기준

- [x] `pkg/pdfbox/` 전체 빌드 성공
- [x] 샘플 PDF에서 텍스트 청크 추출 (좌표, 폰트 정보 포함) 동작
- [x] 샘플 PDF에서 이미지 추출 동작
- [x] 샘플 PDF에서 수평/수직 선분 추출 동작
- [x] `PDFWriter` (annotated PDF 출력) 동작
- [x] `DocumentProcessor` 스텁 제거 후 실제 PDF 처리 시 Markdown 내용 생성
- [ ] T16: TJ kerning + TrimSpace + WinAnsi 3버그 수정
- [ ] T17: 2단 컬럼 읽기 순서 정상화
- [ ] 벤치마크 NID ≥ 0.85, TEDS ≥ 0.40, MHS ≥ 0.55 달성
