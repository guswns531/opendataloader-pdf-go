# T08: Markdown/HTML/Text 생성기 포팅

## 선행 태스크
T02, T03, T07 완료

## 참조 Java 파일
- `markdown/MarkdownGenerator.java` (359줄)
- `markdown/MarkdownGeneratorFactory.java`
- `markdown/MarkdownHTMLGenerator.java`
- `markdown/MarkdownSyntax.java`
- `html/HtmlGenerator.java`
- `html/HtmlGeneratorFactory.java`
- `html/HtmlSyntax.java`
- `text/TextGenerator.java`
- `utils/Base64ImageUtils.java`
- `utils/ImagesUtils.java`

## 지시사항

### 1. internal/utils/base64_image_utils.go

Java `Base64ImageUtils.java` 포팅.
이미지 Base64 인코딩 유틸리티:
```go
// EncodeImageBase64 returns a data URL string for use in embedded images.
func EncodeImageBase64(data []byte, format string) string {
    encoded := base64.StdEncoding.EncodeToString(data)
    mimeType := "image/png"
    if strings.ToLower(format) == "jpeg" || strings.ToLower(format) == "jpg" {
        mimeType = "image/jpeg"
    }
    return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
}
```

### 2. internal/utils/images_utils.go

Java `ImagesUtils.java` 포팅.
PDF에서 이미지 추출 (pdfcpu 기반):
- 페이지에서 임베디드 이미지 추출
- PNG/JPEG 형식으로 변환 (stdlib image/png, image/jpeg)
- 외부 파일로 저장 (--image-output external)

### 3. internal/generators/markdown/markdown_syntax.go

Java `MarkdownSyntax.java` 포팅.
plan/05-phase3-generators.md의 Task 3-4 참조.
모든 Markdown 포맷 상수 정의.

### 4. internal/generators/markdown/markdown_generator.go

Java `MarkdownGenerator.java` (359줄) 완전 포팅.
plan/05-phase3-generators.md의 Task 3-5 참조.

`MarkdownGenerator` struct:
```go
type MarkdownGenerator struct {
    config     *api.Config
    withHTML   bool
    withImages bool
}
```

각 엔티티 타입별 쓰기 메서드:
- `writeHeading()`: `### 텍스트` (레벨에 맞게)
- `writeParagraph()`: 텍스트 + 줄바꿈
- `writeTable()`: GFM 테이블 형식 (`| col1 | col2 |`)
- `writeList()`: `- item` 또는 `1. item`
- `writeImage()`: `![alt](src)` (embedded/external/off 분기)
- `writeFormula()`: `$latex$` (inline) 또는 `$$latex$$` (block)
- `writeCaption()`: 이탤릭체 텍스트

**--keep-line-breaks 처리**:
config.KeepLineBreaks가 true이면 TextLine 끝에 `\n` 추가.

**페이지 구분자**:
```go
if i > 0 && g.config.MarkdownPageSeparator != "" {
    sb.WriteString(g.config.MarkdownPageSeparator + "\n")
}
```

### 5. internal/generators/markdown/markdown_html_generator.go

Java `MarkdownHTMLGenerator.java` 포팅.
`markdown-with-html` 포맷: 복잡한 표는 HTML `<table>` 태그로 출력.

### 6. internal/generators/html/html_syntax.go + html_generator.go

Java `HtmlSyntax.java` + `HtmlGenerator.java` 포팅.
plan/05-phase3-generators.md의 Task 3-6 참조.

전체 HTML 문서 생성:
- `<!DOCTYPE html><html><body>` 래퍼
- 헤딩: `<h1>~<h6>`
- 문단: `<p>`
- 표: `<table><tr><td>`
- 목록: `<ul><li>` / `<ol><li>`
- 이미지: `<img src="" alt="">`
- 수식: `<span class="formula">$latex$</span>`

### 7. internal/generators/text/text_generator.go

Java `TextGenerator.java` 포팅.
plan/05-phase3-generators.md의 Task 3-7 참조.
서식 없이 순수 텍스트만 추출. 페이지 구분자 지원.

### 8. internal/generators/factory.go

plan/05-phase3-generators.md의 Task 3-8 참조.
`Generator` 인터페이스 + `GetGenerator(format string, config *api.Config) Generator`.

### 9. 테스트

`go/tests/unit/generators/markdown_generator_test.go`:
- SemanticHeading H1/H2 → `# / ##` 변환 테스트
- SemanticTable 2x2 → GFM 테이블 변환 테스트
- SemanticImage embedded → `data:image/png;base64,...` 포함 테스트
- 페이지 구분자 삽입 테스트

`go/tests/integration/end_to_end_test.go`:
- `samples/` 디렉토리의 샘플 PDF 처리 → Markdown 출력 생성 테스트

## 완료 기준
- `go build ./internal/generators/...` 성공
- `go test ./tests/unit/generators/...` 통과
- `go test ./tests/integration/...` 통과 (샘플 PDF 처리)
- Markdown 출력에 `#` 헤딩, `|` 테이블, `![]()` 이미지 포함 확인
