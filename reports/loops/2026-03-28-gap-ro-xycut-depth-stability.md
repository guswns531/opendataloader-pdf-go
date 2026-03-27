## Gap ID

`GAP-RO-XYCUT-DEPTH-STABILITY`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

- Go reading-order still had a residual `T17` stability risk after the basic two-column, balanced-column, and mixed-layout/sidebar fixes.
- Recursive XYCut partitioning could keep fragmenting a degenerate layout without an explicit depth guard, risking unstable ordering or runaway recursion on pathological pages.

## Evidence before

- The active worktree contained a focused patch that adds a recursion-depth fallback plus a synthetic regression for deep degenerate segmentation.
- Before this change, `recursiveSegment` had no explicit maximum depth and always re-entered itself when cuts kept producing nested groups.
- The drafted gap packet for `GAP-RO-XYCUT-DEPTH-STABILITY` explicitly called out deep-recursion / determinism as the next narrow `T17` parity step.

## What changed

- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
  - introduced `maxSegmentDepth`
  - routed recursive segmentation through `recursiveSegmentWithDepth(..., depth)`
  - added a depth guard that falls back to deterministic `sortByYThenX` ordering once segmentation gets too deep
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
  - added `TestXYCutFallsBackAfterDeepDegenerateRecursion`
  - the regression creates a long degenerate row, verifies all objects survive sorting, and reruns the sorter multiple times to confirm stable order

## Validation

```bash
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go test ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go build ./...
```

Results:

- focused reading-order tests: passed
- XYCut-targeted processor/unit tests: passed
- build: passed

## Regressions checked

- previously kept two-column / balanced-column / sidebar XYCut coverage stays green under the targeted validator set
- the new deep-recursion regression is deterministic across repeated reruns inside the test

## Remaining uncertainty

- This KEEP is still based on synthetic stability coverage rather than a real Java-vs-Go fixture that demonstrates a depth blow-up on a production PDF.
- The current fallback uses deterministic ordering rather than a more Java-specific segmentation strategy, so future fixture diffs may still reveal a narrower parity refinement beyond this safety guard.

## Follow-up

- Next best step is fixture-backed `T17` evaluation on real complex pages to see whether any remaining instability or merge-order mismatch survives after the current XYCut series.
- If real evidence shows a different failure mode, split it into a narrower follow-up gap instead of widening this one.

## Commit

`5565108`
