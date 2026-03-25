# T04: 텍스트 처리 프로세서 포팅 (TextLine, Paragraph, ContentFilter)

## 선행 태스크
T01, T02, T03 완료

## 참조 Java 파일
- `processors/TextLineProcessor.java` (127줄)
- `processors/ParagraphProcessor.java`
- `processors/ContentFilterProcessor.java`
- `utils/TextNodeStatistics.java`
- `utils/ModeWeightStatistics.java`

## 지시사항

### 1. internal/utils/text_node_statistics.go

Java `TextNodeStatistics.java` 포팅.
텍스트 노드들의 폰트 크기 분포 통계 계산:
- 폰트 크기별 빈도 수집
- 최빈값(mode), 평균(mean), 표준편차 계산
- 사용처: HeadingProcessor의 폰트 크기 임계값 결정

### 2. internal/utils/mode_weight_statistics.go

Java `ModeWeightStatistics.java` 포팅.
가중치 기반 최빈값(mode) 통계:
- 값-가중치 쌍 수집
- 가중 최빈값 계산
- 사용처: 폰트 크기 그룹화

### 3. internal/processors/content_filter_processor.go

Java `ContentFilterProcessor.java` 포팅.
plan/04-phase2-processors.md의 Task 2-2 참조.

필터 기준:
- hidden-text: TextChunk.IsHidden == true
- off-page: BBox가 (0,0,pageWidth,pageHeight) 밖
- tiny: FontStyle.FontSize < 1.0 pt (Java 소스에서 실제 임계값 확인)
- hidden-ocg: OCG 레이어 숨김 속성

입력: []*entities.TextChunk, *api.FilterConfig, pageWidth/Height float64
출력: []*entities.TextChunk (필터링된 결과)

### 4. internal/processors/text_line_processor.go

Java `TextLineProcessor.java` (127줄) 완전 포팅.
plan/04-phase2-processors.md의 Task 2-3 참조.

핵심 로직:
1. baseline Y 좌표 기준 청크 그룹화 (허용 오차: 0.5pt)
2. X 좌표 오름차순 정렬
3. 청크 간 수평 간격 계산 → 공백/탭 문자 삽입 여부 결정
   - 간격 > avgCharWidth * 0.3 → 공백 삽입
   - 간격 > avgCharWidth * 2.0 → 탭 삽입
4. LineArtChunk(불릿)와 텍스트 라인 연결
   - 불릿 X 좌표가 텍스트 라인 leftX - 20pt 이내
   - 불릿 Y 좌표가 텍스트 라인 baseline ±5pt 이내

입력: []*entities.TextChunk, []*entities.LineArtChunk, *containers.ProcessorContext
출력: []*entities.TextLine

### 5. internal/processors/paragraph_processor.go

Java `ParagraphProcessor.java` 완전 포팅.
plan/04-phase2-processors.md의 Task 2-6 참조.

병합 확률 임계값: **0.75** (75%)

정렬 감지:
- 양쪽 정렬: 라인 너비가 컬럼 너비의 90% 이상이면서 두 라인 이상
- 왼쪽 정렬: 첫 글자 X 좌표가 일치
- 오른쪽 정렬: 마지막 글자 X 좌표가 일치
- 가운데 정렬: 중심 X 좌표가 일치

`canMerge(line1, line2 *entities.TextLine) bool` 구현:
- 같은 페이지 여부
- 정렬 일치 (or 양쪽 정렬)
- 폰트 크기 유사도 (±10%)
- 수평 겹침 비율 > 50%
- 수직 간격 < lineHeight * 1.5

### 6. 테스트 파일 생성

`go/tests/unit/processors/text_line_processor_test.go`:
- 단순 텍스트 청크 2개 → 1개 TextLine 생성 테스트
- 다른 baseline 청크 → 2개 TextLine 생성 테스트
- LineArtChunk 연결 테스트

`go/tests/unit/processors/paragraph_processor_test.go`:
- 2개 연속 라인 → 1개 문단 병합 테스트
- 폰트 크기 다른 라인 → 병합 안됨 테스트

## 완료 기준
- `go build ./internal/processors/...` 성공
- `go test ./tests/unit/processors/...` 통과
- 실제 PDF 샘플의 텍스트 라인 추출 결과 확인 (수동 검증)
