## Gap ID

`GAP-RO-TWO-COLUMN-BASIC`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

- Go reading-order still had a basic two-column detection hole when a real gutter landed exactly on a histogram bucket boundary.
- That caused XYCut++ to miss an otherwise valid split and fall back toward interleaved ordering.

## Evidence before

- Dirty worktree contained a focused XYCut++ patch and new regression tests targeting boundary-sensitive gutter detection.
- The prior implementation marked coverage buckets inclusively, so an element ending exactly on a bucket boundary could erase the next bucket and hide the gutter.

## What changed

- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
  - switched object-span coverage to half-open bucket handling using `math.Ceil(...)-1`
  - ignored leading empty space when selecting the best gap so page margins do not masquerade as column gutters
  - reused the sorter’s existing `minGapThreshold` instead of the narrower local ratio constant
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
  - added regression coverage for a gutter that lands on a histogram boundary
  - added a modest-gutter two-column regression to keep the basic grouping behavior stable

## Validation

```bash
cd go && go test ./internal/processors/readingorder/... -count=1
cd go && go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && go build ./...
```

Results:

- focused reading-order package tests: passed
- XYCut-targeted processor/unit tests: passed
- build: passed

## Regressions checked

- basic XYCut ordering coverage in internal and unit reading-order tests
- single-column behavior indirectly preserved by existing targeted XYCut suite
- full `go test ./...` was sampled and still fails in pre-existing table-processor tests unrelated to this patch

## Remaining uncertainty

- This closes a narrow split-detection hole inside basic two-column handling, but broader `T17` parity still needs fixture-level comparison for balanced columns and mixed layouts.

## Follow-up

- Start a fresh loop for the next best parity gap with crisp evidence, likely `GAP-TEXT-TRIMSPACE` if fixture output still shows residual boundary-space collapse, otherwise continue into the next reading-order sub-gap.

## Commit

`N/A`
