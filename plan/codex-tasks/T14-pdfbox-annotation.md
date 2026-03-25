# T14: PDFWriter 포팅 (주석, OCG 레이어)

## 선행 태스크
T13 완료 (pkg/pdfbox/model 존재)

## 라이센스
모든 pkg/pdfbox/ 파일에 Apache-2.0 헤더 필수 (T13과 동일).

## 참조 Java 파일 (직접 읽어서 구현)
- java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/pdf/PDFWriter.java
- java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/pdf/PDFLayer.java

## 지시사항

### 1. go/internal/pdf/pdf_layer.go

Java PDFLayer.java (enum) 포팅:
```go
package pdf

type PDFLayer string

const (
    PDFLayerContent              PDFLayer = "content"
    PDFLayerTableCells           PDFLayer = "table cells"
    PDFLayerListItems            PDFLayer = "list items"
    PDFLayerTableContent         PDFLayer = "table content"
    PDFLayerListContent          PDFLayer = "list content"
    PDFLayerTextBlockContent     PDFLayer = "text blocks content"
    PDFLayerHeaderFooterContent  PDFLayer = "header and footer content"
)
```

### 2. go/pkg/pdfbox/ocg/ocg.go

Optional Content Groups (레이어) 관리:

pdfcpu는 OCG를 직접 지원하지 않으므로 PDF COS 딕셔너리를 직접 조작:
```go
package ocg

import (
    pdfcpu_model "github.com/pdfcpu/pdfcpu/pkg/model"
)

type OptionalContentGroup struct {
    Name    string
    Visible bool
    ref     int  // PDF object reference
}

// AddOCG adds an Optional Content Group to the PDF context.
func AddOCG(ctx *pdfcpu_model.Context, name string) (*OptionalContentGroup, error)

// EnableOCG sets the visibility of an OCG.
func EnableOCG(ctx *pdfcpu_model.Context, group *OptionalContentGroup, enabled bool) error
```

OCG 추가는 PDF 스펙에 따라:
- /OCProperties 딕셔너리를 /Catalog에 추가
- /OCGs 배열에 OCG 딕셔너리 추가
- /D (default view) 설정

pdfcpu API가 지원하지 않으면 저수준 `ctx.XRefTable`으로 직접 딕셔너리 조작.

### 3. go/pkg/pdfbox/annotation/annotation.go

사각형 주석 추가:

```go
package annotation

type SquareAnnotation struct {
    X, Y     float64     // 좌하단 좌표 (PDF 좌표계)
    Width    float64
    Height   float64
    Color    [3]float64  // RGB 0.0~1.0
    Opacity  float64     // 0.0~1.0, default 0.4
    Contents string      // tooltip 텍스트
    LayerRef int         // OCG object reference (0이면 레이어 없음)
}

// AddSquareAnnotation adds a square annotation to a page.
// Uses pdfcpu's annotation support or direct PDF dict manipulation.
func AddSquareAnnotation(ctx *pdfcpu_model.Context, pageIdx int, ann *SquareAnnotation) error
```

pdfcpu annotation 지원 확인:
- `github.com/pdfcpu/pdfcpu/pkg/api` 의 annotation 관련 함수 확인
- 없으면 page dictionary의 /Annots 배열에 직접 딕셔너리 추가

PDF 주석 딕셔너리 구조:
```
<< /Type /Annot
   /Subtype /Square
   /Rect [x y x+w y+h]
   /C [r g b]
   /IC [r g b]
   /CA 0.4
   /Contents (tooltip text)
   /OC << /Type /OCMD /OCGs [ocg-ref] >>
>>
```

### 4. go/internal/pdf/pdf_writer.go (T11 스텁 → 실구현)

Java PDFWriter.java 완전 포팅:

```go
package pdf

import (
    "github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
    "github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/annotation"
    "github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/ocg"
    pdfcpu_api "github.com/pdfcpu/pdfcpu/pkg/api"
)

type PDFWriter struct {
    optionalContents map[PDFLayer]*ocg.OptionalContentGroup
}

// UpdatePDF adds semantic annotations to an existing PDF.
// Java: updatePDF(File inputPDF, String password, String outputFolder, List<List<IObject>> contents)
func (w *PDFWriter) UpdatePDF(
    inputPath string,
    password string,
    outputPath string,
    doc *entities.Document,
) error

// getColor returns RGB color for each element type.
// Java getColor(SemanticType) 포팅
func getColor(objectType entities.ObjectType) [3]float64 {
    switch objectType {
    case entities.ObjectTypeHeading:
        return [3]float64{0, 0, 1}      // 파랑
    case entities.ObjectTypeList:
        return [3]float64{0, 1, 0}      // 초록
    case entities.ObjectTypeParagraph:
        return [3]float64{0, 1, 1}      // 시안
    case entities.ObjectTypeImage:
        return [3]float64{1, 0, 0}      // 빨강
    case entities.ObjectTypeTable:
        return [3]float64{1, 0, 1}      // 마젠타
    case entities.ObjectTypeCaption:
        return [3]float64{1, 1, 0}      // 노랑
    default:
        return [3]float64{0.9, 0.9, 0.9} // 연회색
    }
}
```

### 5. go/tests/unit/pdf/pdf_writer_test.go

```go
// getColor 색상 매핑 테스트
func TestGetColor(t *testing.T)
// PDFLayer 상수값 테스트
func TestPDFLayerValues(t *testing.T)
```

## 완료 기준
- `go build ./pkg/pdfbox/... ./internal/pdf/...` 성공
- `go test ./tests/unit/pdf/...` 통과
- 실제 PDF에 주석 추가 후 저장 동작 확인 (수동)
