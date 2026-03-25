T17 태스크: 다단 컬럼 읽기 순서 수정 (XYCutPlusPlusSorter 2-column 지원)
작업 위치: tripoli/go/ 디렉토리.

먼저 아래 파일들을 직접 읽어라:
1. go/internal/processors/xycut_plus_plus_sorter.go
2. go/internal/entities/document.go
3. go/internal/entities/text_chunk.go

읽은 후 아래 버그를 수정하라.

---

## 문제: 2단 컬럼 PDF에서 좌·우 컬럼 텍스트가 섞임

학술 논문처럼 2단(two-column) 레이아웃 PDF에서 현재 XYCutPlusPlusSorter가
좌 컬럼과 우 컬럼의 텍스트를 같은 Y 위치 기준으로 섞어서 출력한다.

예시 (잘못된 순서):
```
[좌단 1행] [우단 1행] [좌단 2행] [우단 2행] ...
```

기대 순서:
```
[좌단 1행] [좌단 2행] ... [우단 1행] [우단 2행] ...
```

---

## 수정 방향

### 핵심 아이디어: 수직 갭 감지 → 컬럼 분리 후 좌→우 순서

1. **수직 갭 감지**: 페이지의 모든 텍스트 청크 X 좌표를 분석해서
   페이지 너비의 20%~80% 사이에 청크가 없는 수직 갭(vertical gap)이 있으면
   2-column 레이아웃으로 판정한다.

2. **컬럼 분리**: 갭 위치를 기준으로 좌 컬럼(x < gap_x)과 우 컬럼(x >= gap_x)으로 분리.

3. **정렬 순서**: 좌 컬럼 전체(Y 기준 오름차순) → 우 컬럼 전체(Y 기준 오름차순).

4. **재귀 적용**: 각 컬럼 내부에서도 동일 알고리즘 재귀 적용 (3단 이상 지원).

### 구현 지침

`xycut_plus_plus_sorter.go`에서 청크 정렬 로직을 수정:

```go
// detectColumnSplit은 chunks에서 수직 갭 X 좌표를 감지한다.
// 갭이 없으면 -1 반환.
func detectColumnSplit(chunks []*entities.TextChunk, pageWidth float64) float64 {
    if pageWidth <= 0 || len(chunks) < 4 {
        return -1
    }
    minX := pageWidth * 0.20
    maxX := pageWidth * 0.80
    gapWidth := pageWidth * 0.03  // 최소 갭 너비 (페이지 너비의 3%)

    // X 범위 히스토그램으로 갭 감지
    // [minX, maxX] 범위를 100개 버킷으로 나눠서 각 버킷에 청크가 있는지 확인
    buckets := 100
    occupied := make([]bool, buckets)
    bucketWidth := (maxX - minX) / float64(buckets)
    for _, c := range chunks {
        bIdx := int((c.X - minX) / bucketWidth)
        if bIdx >= 0 && bIdx < buckets {
            // 청크 너비도 커버
            endBIdx := int((c.X + c.Width - minX) / bucketWidth)
            for i := bIdx; i <= endBIdx && i < buckets; i++ {
                occupied[i] = true
            }
        }
    }

    // 연속된 비어있는 버킷 구간 찾기
    bestGapStart := -1
    bestGapLen := 0
    curStart := -1
    curLen := 0
    for i, occ := range occupied {
        if !occ {
            if curStart < 0 {
                curStart = i
            }
            curLen++
        } else {
            if curLen > bestGapLen {
                bestGapLen = curLen
                bestGapStart = curStart
            }
            curStart = -1
            curLen = 0
        }
    }
    if curLen > bestGapLen {
        bestGapLen = curLen
        bestGapStart = curStart
    }

    if bestGapLen <= 0 {
        return -1
    }
    gapX := minX + float64(bestGapStart)*bucketWidth + float64(bestGapLen)*bucketWidth/2
    actualGapWidth := float64(bestGapLen) * bucketWidth
    if actualGapWidth < gapWidth {
        return -1
    }
    return gapX
}

// sortWithColumnAwareness는 컬럼 감지 후 좌→우 순서로 정렬한다.
func sortWithColumnAwareness(chunks []*entities.TextChunk, pageWidth float64) []*entities.TextChunk {
    gapX := detectColumnSplit(chunks, pageWidth)
    if gapX < 0 {
        // 단일 컬럼: Y 기준 정렬
        sort.SliceStable(chunks, func(i, j int) bool {
            return chunks[i].Y < chunks[j].Y
        })
        return chunks
    }

    // 좌·우 컬럼 분리
    var left, right []*entities.TextChunk
    for _, c := range chunks {
        if c.X+c.Width/2 < gapX {
            left = append(left, c)
        } else {
            right = append(right, c)
        }
    }

    // 각 컬럼 재귀 정렬
    left = sortWithColumnAwareness(left, gapX)
    right = sortWithColumnAwareness(right, pageWidth-gapX)

    return append(left, right...)
}
```

XYCutPlusPlusSorter의 `Sort` 또는 `sortPage` 메서드에서
기존 단순 Y-정렬 대신 `sortWithColumnAwareness(chunks, page.Width)` 호출.

---

## 수정 후 검증

1. 단위 테스트 (2단 컬럼 시나리오):
```bash
cd go && go test ./internal/processors/... -run TestXYCut -v
```

2. 샘플 PDF (2단 학술 논문) 처리:
```bash
go build -o /tmp/odl-test ./cmd/opendataloader-pdf/
/tmp/odl-test --format markdown --output-dir /tmp/odl-out3 \
  /Users/hyeonjun/conductor/workspaces/opendataloader-pdf-go/tripoli/samples/pdf/1901.03003.pdf
head -50 /tmp/odl-out3/1901.03003.md
```

기대 결과: Abstract, Introduction 등 섹션이 좌 컬럼 전체 → 우 컬럼 전체 순으로 출력.

3. 전체 빌드 및 테스트:
```bash
cd go && go build ./... && go test ./...
```
