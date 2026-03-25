# T15: DocumentProcessor 스텁 제거 → pdfbox 실구현 교체

## 선행 태스크
T13, T14 완료

## 참조 파일
- go/internal/processors/document_processor.go (현재 스텁)
- java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/DocumentProcessor.java (454줄)
- go/pkg/pdfbox/extractor/ (T13에서 생성됨)

## 지시사항

### 1. go/internal/processors/document_processor.go 수정

현재 `loadPDFContent` 또는 PDF 로딩 관련 스텁 코드를 찾아
pkg/pdfbox 실구현으로 교체하라.

document_processor.go 파일을 직접 읽어 스텁 위치 파악 후:

```go
import (
    pdfbox_loader "github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/loader"
    "github.com/opendataloader-project/opendataloader-pdf-go/pkg/pdfbox/extractor"
)

func (p *DocumentProcessor) loadDocument(path string, config *api.Config) (*entities.Document, error) {
    // pdfcpu로 PDF 열기
    doc, err := pdfbox_loader.Open(path, config.Password)
    if err != nil {
        return nil, fmt.Errorf("failed to open PDF %s: %w", path, err)
    }
    defer doc.Close()

    pageCount := doc.PageCount()
    document := &entities.Document{
        Metadata: entities.DocumentMetadata{PageCount: pageCount},
    }

    // 페이지 범위 파싱
    pageNums := parsePageRange(config.Pages, pageCount)

    for _, pageIdx := range pageNums {
        page, err := doc.GetPage(pageIdx - 1)  // 1-based → 0-based
        if err != nil {
            continue
        }

        entityPage := &entities.Page{
            PageMetadata: entities.PageMetadata{
                Number: pageIdx,
                Width:  page.Width,
                Height: page.Height,
            },
        }

        // 텍스트 청크 추출
        texts, err := extractor.ExtractTextChunks(doc, pageIdx-1)
        if err == nil {
            for _, t := range texts {
                entityPage.Chunks = append(entityPage.Chunks, &entities.TextChunk{
                    BaseObject: entities.BaseObject{
                        ID:   ctx.NextID(),
                        BBox: entities.BoundingBox{X: t.X, Y: t.Y, Width: t.Width, Height: t.Height, Page: pageIdx},
                    },
                    Text:     t.Text,
                    Baseline: t.Baseline,
                    FontStyle: entities.FontStyle{
                        FontName: t.FontName,
                        FontSize: t.FontSize,
                        Bold:     t.Bold,
                        Italic:   t.Italic,
                        Color:    t.Color,
                    },
                })
            }
        }

        // 이미지 추출
        images, err := extractor.ExtractImages(doc, pageIdx-1, config.ImageDir)
        if err == nil {
            for _, img := range images {
                entityPage.Elements = append(entityPage.Elements, &entities.SemanticImage{
                    BaseObject: entities.BaseObject{
                        ID:   ctx.NextID(),
                        BBox: entities.BoundingBox{X: img.X, Y: img.Y, Width: img.Width, Height: img.Height, Page: pageIdx},
                    },
                    Data:         img.Data,
                    ExternalPath: img.ExternalPath,
                })
            }
        }

        // 선분 추출
        lineArts, err := extractor.ExtractLineArts(doc, pageIdx-1)
        if err == nil {
            for _, la := range lineArts {
                entityPage.LineArts = append(entityPage.LineArts, &entities.LineArtChunk{
                    BaseObject: entities.BaseObject{
                        BBox: entities.BoundingBox{X: la.X, Y: la.Y, Width: la.Width, Height: la.Height, Page: pageIdx},
                    },
                    IsHorizontal: la.IsHorizontal,
                    IsVertical:   la.IsVertical,
                    LineWidth:    la.LineWidth,
                })
            }
        }

        document.Pages = append(document.Pages, entityPage)
    }

    return document, nil
}
```

### 2. ProcessorContext 연동

ProcessorContext를 loadDocument에 전달하여 NextID() 사용:
- loadDocument 시그니처: `(path string, config *api.Config, ctx *containers.ProcessorContext)`

### 3. Process() 메서드 완성

document_processor.go의 Process() 메서드가 loadDocument를 호출하고
전체 파이프라인이 실제 데이터로 동작하는지 확인.

### 4. THIRD_PARTY_LICENSES.md 업데이트

`THIRD_PARTY/THIRD_PARTY_LICENSES.md` 파일의 Go Dependencies 섹션 뒤에 추가:
```markdown
## PDFBox Equivalent Implementation

The go/pkg/pdfbox/ package provides functionality equivalent to Apache PDFBox 3.0.4
and is implemented using pdfcpu. Licensed under Apache-2.0.

| Original Library | Version | License | URL |
|-----------------|---------|---------|-----|
| Apache PDFBox | 3.0.4 | Apache-2.0 | https://pdfbox.apache.org/ |
```

### 5. go/tests/integration/end_to_end_test.go 강화

실제 PDF 처리 시 출력 내용 검증:
```go
func TestEndToEndWithContent(t *testing.T) {
    // samples/ 에서 PDF 찾기
    // 처리 후 output Markdown이 빈 문자열이 아닌지 확인
    // (텍스트가 실제로 추출됐는지)
    assert.NotEmpty(t, markdownContent, "markdown output should not be empty for real PDF")
}
```

## 완료 기준
- `go build ./...` 성공
- `go test ./...` 전체 통과
- 실제 PDF 파일 처리 시 Markdown 출력에 텍스트 내용 포함
- `go test ./tests/integration/... -v` 에서 PDF 처리 결과 출력 확인
