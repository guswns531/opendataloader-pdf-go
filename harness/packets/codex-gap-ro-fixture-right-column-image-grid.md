# Codex Gap Task Packet — GAP-RO-FIXTURE-RIGHT-COLUMN-IMAGE-GRID

## Role

You are evaluating one narrow Java→Go reading-order parity gap in the Go port of OpenDataLoader PDF. Keep scope tight and evidence-first.

## Gap ID

`GAP-RO-FIXTURE-RIGHT-COLUMN-IMAGE-GRID`

## Category

`reading-order-parity`

## Objective

Follow up the split from `GAP-RO-FIXTURE-MIXED-LAYOUT-EVAL` and target one real fixture-derived residual mismatch only:

- `samples/pdf/1901.03003.pdf`
- the Java real-world XYCut case already encoded in:
  - `java/opendataloader-pdf-core/src/test/java/org/opendataloader/pdf/processors/readingorder/XYCutPlusPlusSorterTest.java`

The live residual family is:

- a trailing left-column block (`650`)
- a right-column 3x3 image grid (`653,654,655,657,658,659,660,661,662`)
- caption `656`
- nearby marginal arXiv sidebar `667`

Current Go ordering observed during fixture-geometry replay:

```text
[646 647 648 649 660 661 662 667 650 653 657 654 658 655 659 656 663 664 665 666]
```

Java intent for this case is narrower:

- `650` should stay before the image group
- the image blocks should remain consecutive as one group
- that image group should stay before caption `656`

Do not broaden into other mixed-layout families unless directly required to explain this exact mismatch.

## Relevant files

### Existing report
- `reports/loops/2026-03-28-gap-ro-fixture-mixed-layout-eval.md`

### Go code
- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`

### Java reference
- `java/opendataloader-pdf-core/src/test/java/org/opendataloader/pdf/processors/readingorder/XYCutPlusPlusSorterTest.java`

### Fixture
- `samples/pdf/1901.03003.pdf`
- `samples/pdf/1901.03003.json`

## Required work

1. Recreate the `1901.03003` Java fixture geometry as focused Go regression coverage.
2. Keep `2408.02509v1` only as a no-regression comparison if helpful.
3. Identify which one root-cause family creates the mis-order:
   - nested mini-column partitioning inside the image grid
   - balanced-column merge order
   - deferred floating/sidebar reinsertion
4. If the root cause is clear:
   - implement one narrow fix only
   - add focused regression coverage
5. If the root cause is not clear:
   - do not widen heuristics blindly
   - return `SPLIT` or `NEEDS-HUMAN-REVIEW` with exact evidence

## Acceptance criteria

A successful KEEP must include all of:

1. fixture-derived `1901.03003` mismatch reproduced in Go regression form
2. one narrow root-cause family only
3. targeted regression coverage added
4. targeted reading-order tests pass
5. `go build ./...` passes

If KEEP is not justified, return `SPLIT` with the exact remaining ambiguity.

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
```

Add the narrowest reproduction command you use for the `1901.03003` fixture-derived case.

## Non-goals

- benchmark-wide reading-order cleanup
- new broad sidebar heuristics
- unrelated text spacing fixes
- touching multiple reading-order subfamilies in one loop

## Output format

Report:
1. exact `1901.03003` mismatch investigated
2. where divergence occurs
3. files changed
4. before/after evidence
5. tests/commands run
6. KEEP / SPLIT / NEEDS-HUMAN-REVIEW
