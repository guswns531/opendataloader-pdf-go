# T13: PDF 콘텐츠 추출 실구현 (텍스트, 이미지, 선분)

## 선행 태스크
T01 완료 (go.mod에 pdfcpu 있음)

## 라이센스
모든 pkg/pdfbox/ 파일에 아래 헤더 필수:
```
// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// This package provides functionality equivalent to Apache PDFBox 3.0.4
// (https://pdfbox.apache.org/), implemented using pdfcpu.
```

## 참조
- Java DocumentProcessor.java — pdfcpu context에서 추출하는 방식 참조
- pdfcpu API 문서: github.com/pdfcpu/pdfcpu/pkg/api

## 지시사항

### 1. go/pkg/pdfbox/model/document.go

pdfcpu 기반 PDF 문서 모델:
```go
package model

import (
    "github.com/pdfcpu/pdfcpu/pkg/api"
    pdfcpu_model "github.com/pdfcpu/pdfcpu/pkg/model"
)

type PDDocument struct {
    Path     string
    Password string
    Ctx      *pdfcpu_model.Context
}

type PDPage struct {
    Number int
    Width  float64
    Height float64
    Rotate int
}

// Open loads a PDF file and returns a PDDocument.
func Open(path, password string) (*PDDocument, error)

// PageCount returns total page count.
func (d *PDDocument) PageCount() int

// GetPage returns page metadata for given 0-based page index.
func (d *PDDocument) GetPage(pageIdx int) (*PDPage, error)

// Close releases resources.
func (d *PDDocument) Close() error
```

pdfcpu로 PDF 읽기:
```go
conf := pdfcpu_model.NewDefaultConfiguration()
if password != "" {
    conf.UserPW = password
    conf.OwnerPW = password
}
ctx, err := api.ReadContextFile(path)
// 또는 api.ReadContextWithConfiguration()
```

### 2. go/pkg/pdfbox/extractor/text_extractor.go

pdfcpu content stream에서 텍스트 청크 추출:

```go
package extractor

type ExtractedText struct {
    Text      string
    X, Y      float64    // PDF 좌표계 (좌하단 원점)
    Width     float64
    Height    float64
    FontName  string
    FontSize  float64
    Bold      bool
    Italic    bool
    Color     [3]float64 // RGB 0.0~1.0
    Baseline  float64
    Page      int        // 0-based
}

// ExtractTextChunks extracts all text chunks with coordinates from a page.
// Uses pdfcpu's content stream parsing.
func ExtractTextChunks(doc *model.PDDocument, pageIdx int) ([]*ExtractedText, error)
```

pdfcpu로 텍스트 추출하는 방법:
- `api.ExtractContent(rs, conf, []string{pageRange})` 또는
- pdfcpu의 `pdfcpu_model.Page` 구조 탐색
- content stream의 BT/ET 블록에서 Tf(폰트), Tm(변환행렬), Tj/TJ(텍스트) 연산자 파싱

pdfcpu가 직접 텍스트 좌표 추출을 지원하지 않는 경우:
- `api.ExtractPages()` 또는 page dictionary에서 content stream bytes 직접 읽기
- PDF content stream 연산자 수동 파싱 (최소 BT, ET, Tf, Td, TD, Tm, Tj, TJ, T*)

구현이 복잡한 경우 간소화 전략:
- 텍스트를 줄 단위로 추출하고 페이지 너비/높이에서 Y 좌표 추정
- 완전한 구현은 TODO 주석으로 표시하되 기본 동작은 보장

### 3. go/pkg/pdfbox/extractor/image_extractor.go

pdfcpu로 이미지 추출:

```go
type ExtractedImage struct {
    X, Y     float64
    Width    float64
    Height   float64
    Data     []byte
    Format   string  // "png", "jpeg"
    Page     int
}

// ExtractImages extracts all images from a page.
// Saves to outputDir if non-empty, otherwise returns Data in-memory.
func ExtractImages(doc *model.PDDocument, pageIdx int, outputDir string) ([]*ExtractedImage, error)
```

pdfcpu 이미지 추출:
```go
// api.ExtractImages() 사용
err = api.ExtractImagesFile(inFile, outDir, pages, conf)
```

### 4. go/pkg/pdfbox/extractor/lineart_extractor.go

pdfcpu content stream에서 수평/수직 선분 추출:

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

// ExtractLineArts extracts horizontal and vertical line segments from a page.
// Parses PDF graphics operators: m (moveto), l (lineto), re (rectangle),
// w (line width), S/s (stroke)
func ExtractLineArts(doc *model.PDDocument, pageIdx int) ([]*ExtractedLineArt, error)
```

선분 감지 기준:
- 수평선: |height| < 2.0 AND width > 5.0
- 수직선: |width| < 2.0 AND height > 5.0
- rectangle(re) 연산자: width>height이면 수평, height>width이면 수직으로 분류

### 5. go/pkg/pdfbox/loader/loader.go

편의 래퍼:
```go
package loader

// Open은 PDF 파일을 열어 PDDocument를 반환한다.
func Open(path, password string) (*model.PDDocument, error) {
    return model.Open(path, password)
}
```

### 6. go/tests/unit/pdfbox/extractor_test.go

```go
// samples/ 아래 PDF 파일이 없으면 t.Skip()
// 있으면 텍스트 추출 후 len > 0 확인
func TestExtractTextChunks(t *testing.T)
func TestExtractImages(t *testing.T)
func TestExtractLineArts(t *testing.T)
```

## 완료 기준
- `go build ./pkg/pdfbox/...` 성공
- `go test ./tests/unit/pdfbox/...` 통과 (샘플 PDF 없으면 skip)
- 실제 PDF 파일로 수동 테스트: 텍스트 청크 개수 > 0 확인
