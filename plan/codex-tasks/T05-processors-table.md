# T05: 테이블 프로세서 포팅 (TableBorder, Special, Cluster)

## 선행 태스크
T04 완료

## 참조 Java 파일
- `processors/TableBorderProcessor.java` (252줄)
- `processors/AbstractTableProcessor.java`
- `processors/SpecialTableProcessor.java`
- `processors/ClusterTableProcessor.java`

## 지시사항

### 1. internal/processors/abstract_table_processor.go

Java `AbstractTableProcessor.java` 포팅.
공통 테이블 처리 로직:
- 테이블 행/열 병합 (rowspan/colspan 계산)
- 셀 텍스트 내용 정리
- `TableProcessor` 인터페이스 정의:
  ```go
  type TableProcessor interface {
      Process(elements []entities.IObject, ctx *containers.ProcessorContext) []entities.IObject
  }
  ```

### 2. internal/processors/table_border_processor.go

Java `TableBorderProcessor.java` (252줄) 완전 포팅.
plan/04-phase2-processors.md의 Task 2-5 참조.

**핵심**: Java ThreadLocal depth → Go 함수 매개변수 depth int
최대 재귀 깊이: **10** (maxTableDepth 상수)

알고리즘:
1. LineArtChunk에서 수평/수직 선분 추출
2. 선분들로 직사각형 테이블 경계 감지
   - 닫힌 직사각형 구성 가능한 선분 집합 찾기
   - 최소 2행 × 2열 구조 요구
3. 테이블 경계 내 텍스트 청크를 행/열에 할당
4. 청크가 테이블 경계에 걸친 경우 분할
5. 중첩 테이블 재귀 처리 (depth+1)
6. `SemanticTable` 생성

```go
const maxTableDepth = 10

func (p *TableBorderProcessor) processNode(
    elements []entities.IObject,
    lineArts []*entities.LineArtChunk,
    ctx *containers.ProcessorContext,
    depth int,
) []entities.IObject {
    if depth >= maxTableDepth {
        return elements
    }
    // ...
}
```

### 3. internal/processors/special_table_processor.go

Java `SpecialTableProcessor.java` 포팅.
경계선 없는 테이블 감지 (텍스트 정렬 패턴 기반):
- 여러 텍스트 라인에서 공통 X 좌표 클러스터 감지
- 열 수 ≥ 2, 행 수 ≥ 2 인 경우 테이블로 간주
- 정렬 허용 오차: ±3pt

### 4. internal/processors/cluster_table_processor.go

Java `ClusterTableProcessor.java` 포팅.
`--table-method cluster` 옵션 시 사용되는 클러스터링 기반 테이블 감지.

알고리즘 (DBSCAN 스타일):
- 텍스트 청크의 X/Y 좌표로 밀도 기반 클러스터링
- epsilon: 텍스트 높이의 0.5배
- minPoints: 2
- 클러스터 → 행/열 그리드 구조 감지

### 5. 테스트 파일

`go/tests/unit/processors/table_border_processor_test.go`:
- 4개 선분으로 구성된 단순 테이블 감지 테스트
- 재귀 깊이 제한 테스트 (depth >= 10에서 중단)
- 중첩 테이블 (2레벨) 처리 테스트

## 완료 기준
- `go build ./internal/processors/...` 성공
- `go test ./tests/unit/processors/...` (테이블 관련) 통과
- 테두리가 있는 샘플 PDF 테이블 감지 수동 확인
