T16 태스크: PDF 텍스트 추출 품질 수정 (공백 손실, 인코딩, TJ 오프셋)
작업 위치: tripoli/go/ 디렉토리.

먼저 아래 파일들을 직접 읽어라:
1. go/pkg/pdfbox/extractor/text_extractor.go
2. go/pkg/pdfbox/extractor/content_parser.go
3. go/pkg/pdfbox/model/document.go

읽은 후 아래 3가지 버그를 수정하라.

---

## 버그 1: TJ 배열 kerning 오프셋 무시 → 단어 사이 공백 손실

현재 `decodeTJText` 함수는 TJ 배열의 숫자 항목을 완전히 무시한다.
PDF TJ 연산자에서 숫자 항목은 글리프 간격(음수 = 글리프 당김, 양수 = 밀기)이다.
값이 충분히 음수이면 시각적으로 단어 사이 공백을 표현한다.

**수정**: `decodeTJText` 에서 숫자 항목이 -100 (text space units) 이하이면 공백(" ") 삽입.
threshold는 상수로 정의: `const tjSpaceThreshold = -100.0`

수정 예:
```go
func decodeTJText(tok streamToken) string {
    if tok.kind != "array" {
        return ""
    }
    const tjSpaceThreshold = -100.0
    var sb strings.Builder
    for _, item := range tok.items {
        if item.kind == "string" || item.kind == "hex" {
            sb.WriteString(item.value)
        } else if item.kind == "number" {
            if v, ok := parseFloatToken(item); ok && v < tjSpaceThreshold {
                sb.WriteByte(' ')
            }
        }
    }
    return sb.String()
}
```

---

## 버그 2: appendText에서 strings.TrimSpace → 의도적 공백 제거

`text_extractor.go`의 `appendText` 함수에서:
```go
text = strings.TrimSpace(text)
if text == "" || !ts.inText {
```

`strings.TrimSpace`가 "(Hello ) (World)"처럼 텍스트 끝/시작의 공백을 제거한다.
이로 인해 단어 경계 공백이 사라진다.

**수정**: TrimSpace 제거, 빈 문자열 체크만 유지:
```go
if text == "" || !ts.inText {
    return
}
```

---

## 버그 3: font encoding 미처리 → 8비트 문자열 깨짐

현재 `normalizePDFString`은 UTF-16 BOM만 처리한다.
PDF에서 가장 흔한 인코딩인 WinAnsiEncoding(Windows-1252)을 처리하지 않아서
0x80~0xFF 범위 바이트가 깨진다.

**수정**: `normalizePDFString`에 Windows-1252 → Unicode 변환 추가.

```go
func normalizePDFString(b []byte) string {
    if len(b) >= 2 {
        // UTF-16 BE BOM
        if b[0] == 0xFE && b[1] == 0xFF {
            runes := make([]rune, 0, (len(b)-2)/2)
            for i := 2; i+1 < len(b); i += 2 {
                runes = append(runes, rune(uint16(b[i])<<8|uint16(b[i+1])))
            }
            return string(runes)
        }
        // UTF-16 LE BOM
        if b[0] == 0xFF && b[1] == 0xFE {
            runes := make([]rune, 0, (len(b)-2)/2)
            for i := 2; i+1 < len(b); i += 2 {
                runes = append(runes, rune(uint16(b[i+1])<<8|uint16(b[i])))
            }
            return string(runes)
        }
    }
    // Windows-1252 (WinAnsiEncoding) → Unicode 변환
    // 0x00-0x7F: ASCII와 동일
    // 0x80-0x9F: Windows-1252 특수 매핑
    // 0xA0-0xFF: Latin-1과 동일
    return decodeWinAnsi(b)
}

// Windows-1252 특수 매핑 (0x80~0x9F 범위)
var win1252Extras = map[byte]rune{
    0x80: '\u20AC', // €
    0x82: '\u201A', // ‚
    0x83: '\u0192', // ƒ
    0x84: '\u201E', // „
    0x85: '\u2026', // …
    0x86: '\u2020', // †
    0x87: '\u2021', // ‡
    0x88: '\u02C6', // ˆ
    0x89: '\u2030', // ‰
    0x8A: '\u0160', // Š
    0x8B: '\u2039', // ‹
    0x8C: '\u0152', // Œ
    0x8E: '\u017D', // Ž
    0x91: '\u2018', // '
    0x92: '\u2019', // '
    0x93: '\u201C', // "
    0x94: '\u201D', // "
    0x95: '\u2022', // •
    0x96: '\u2013', // –
    0x97: '\u2014', // —
    0x98: '\u02DC', // ˜
    0x99: '\u2122', // ™
    0x9A: '\u0161', // š
    0x9B: '\u203A', // ›
    0x9C: '\u0153', // œ
    0x9E: '\u017E', // ž
    0x9F: '\u0178', // Ÿ
}

func decodeWinAnsi(b []byte) string {
    // 모든 바이트가 ASCII(0x00-0x7F)이면 바로 반환
    allASCII := true
    for _, ch := range b {
        if ch > 0x7F {
            allASCII = false
            break
        }
    }
    if allASCII {
        return string(b)
    }

    runes := make([]rune, 0, len(b))
    for _, ch := range b {
        if ch < 0x80 {
            runes = append(runes, rune(ch))
        } else if r, ok := win1252Extras[ch]; ok {
            runes = append(runes, r)
        } else {
            // 0xA0-0xFF: Latin-1과 동일
            runes = append(runes, rune(ch))
        }
    }
    return string(runes)
}
```

---

## 수정 후 검증

1. 테스트 실행:
```bash
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... (통과해야 함)
```

2. 샘플 PDF 처리 후 품질 확인:
```bash
go build -o /tmp/odl-test ./cmd/opendataloader-pdf/
/tmp/odl-test --format markdown --output-dir /tmp/odl-out2 \
  /Users/hyeonjun/conductor/workspaces/opendataloader-pdf-go/tripoli/samples/pdf/1901.03003.pdf
head -20 /tmp/odl-out2/1901.03003.md
```

기대 결과: "AMulti-ObjectRectiedAttentionNetwork" → "A Multi-Object Rectified Attention Network" (공백 복구)

3. 전체 빌드:
```bash
cd go && go build ./...
```
