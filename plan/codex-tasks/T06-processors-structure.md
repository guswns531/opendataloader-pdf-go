# T06: 구조 프로세서 포팅 (Heading, List, Level, Caption, ReadingOrder)

## 선행 태스크
T04, T05 완료

## 참조 Java 파일
- `processors/HeadingProcessor.java` (210줄)
- `processors/ListProcessor.java`
- `processors/LevelProcessor.java`
- `processors/CaptionProcessor.java`
- `processors/HeaderFooterProcessor.java`
- `processors/readingorder/XYCutPlusPlusSorter.java` (677줄)
- `utils/BulletedParagraphUtils.java`
- `utils/levels/LevelInfo.java` + 관련 파일들

## 지시사항

### 1. internal/utils/levels/ 포팅

Java `utils/levels/` 디렉토리 전체 포팅:

`level_info.go`:
```go
type LevelInfo struct {
    Level     int
    FontSize  float64
    IsBold    bool
    Count     int
}
```

`list_level_info.go`, `table_level_info.go`, `text_bullet_paragraph_level_info.go`, `line_art_bullet_paragraph_level_info.go` 각각 포팅.

### 2. internal/utils/bulleted_paragraph_utils.go

Java `BulletedParagraphUtils.java` 포팅.
불릿 문단 감지 유틸리티:
- 텍스트 불릿 패턴: `^(\d+\.|[a-zA-Z]\.|[①-⑳]|[ⓐ-ⓩ]|\([0-9]+\))`
- 라인아트 불릿: `LineArtBullet != nil`
- 들여쓰기 레벨 계산

### 3. internal/processors/heading_processor.go

Java `HeadingProcessor.java` (210줄) 완전 포팅.
plan/04-phase2-processors.md의 Task 2-4 참조.

알고리즘:
1. 모든 TextLine의 폰트 크기 분포 수집 (TextNodeStatistics 사용)
2. 본문 폰트 크기 추정 (최빈값)
3. 제목 후보 기준:
   - 폰트 크기 > 본문 * 1.1 OR 굵은 폰트
   - 단독 라인 (이전/이후 빈 줄)
   - 길이 < 120자
4. HeadingCandidate 스코어링 (폰트 크기, 굵기, 위치, 길이 가중치)
5. 임계 스코어 이상이면 SemanticHeading 생성 (레벨은 LevelProcessor가 결정)

### 4. internal/processors/level_processor.go

Java `LevelProcessor.java` 포팅.
제목 레벨(1-6) 할당:
- 폰트 크기 내림차순으로 레벨 1 = 가장 큰 폰트
- 동일 폰트 크기는 동일 레벨
- 최대 6레벨 (H1~H6)

### 5. internal/processors/list_processor.go

Java `ListProcessor.java` 완전 포팅.
plan/04-phase2-processors.md의 Task 2-7 참조.

정규식 패턴:
```go
var (
    orderedBulletPattern    = regexp.MustCompile(`^(\d+\.|\([a-zA-Z0-9]+\)|[a-zA-Z]\.)`)
    unorderedBulletPattern  = regexp.MustCompile(`^[•\-※◦▪▸►\*]`)
    koreanAttachPattern     = regexp.MustCompile(`^붙\s*임`)
)

const baselineDiffThreshold = 1.2
```

다중 페이지 목록 연속 처리: 이전 페이지의 마지막 목록과 현재 페이지의 첫 목록이 같은 불릿 패턴이면 병합.

### 6. internal/processors/caption_processor.go

Java `CaptionProcessor.java` 포팅.
이미지/테이블 캡션 감지:
- 캡션 패턴: `^(그림|Fig|Figure|표|Table|도표)\s*[\d\-\.]+`
- 이미지/테이블 바로 위/아래 라인에서 감지
- `SemanticCaption` 생성 (RefType: "figure" or "table")

### 7. internal/processors/header_footer_processor.go

Java `HeaderFooterProcessor.java` 포팅.
페이지 상/하단 영역 필터링:
- 상단 헤더: Y > pageHeight * 0.9
- 하단 푸터: Y < pageHeight * 0.1
- `--include-header-footer` 옵션이 false이면 제거

### 8. internal/processors/readingorder/xycut_plus_plus_sorter.go

Java `XYCutPlusPlusSorter.java` (677줄) 완전 포팅.
plan/04-phase2-processors.md의 Task 2-8 참조.
arXiv:2504.10258 알고리즘 구현.

**주의**: 최근 커밋(8e3f74a, 6b42c7a)에서 narrow outlier filtering 관련 코드 리뷰 수정이 있었음.
Java 소스 최신 상태를 정확히 반영할 것.

```go
type XYCutPlusPlusSorter struct {
    MinWidthRatio float64  // narrow outlier 필터 (기본: 0.05)
}

// Sort returns elements in reading order.
func (s *XYCutPlusPlusSorter) Sort(
    elements []entities.IObject,
    pageWidth, pageHeight float64,
) []entities.IObject
```

5단계 알고리즘:
1. Narrow outlier 감지 및 분리 (너비 < pageWidth * MinWidthRatio)
2. Cross-layout 감지 (spanning elements)
3. Density-based axis selection (X축 vs Y축 분할 선택)
4. 재귀적 이진 분할 + 정렬
5. Outlier 재삽입 (위치 기반)

### 9. 테스트 파일

`go/tests/unit/processors/heading_processor_test.go`
`go/tests/unit/processors/list_processor_test.go`
`go/tests/unit/processors/readingorder/xycut_test.go`

## 완료 기준
- `go build ./internal/processors/...` 성공
- `go test ./tests/unit/processors/...` 전체 통과
- XYCutPlusPlusSorter: 간단한 2컬럼 레이아웃 정렬 테스트 통과
