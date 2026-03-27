# Codex Gap Task Packet — GAP-RO-RIGHT-COLUMN-IMAGE-GRID-SEQUENCING

## Role

You are implementing one narrow Java→Go reading-order parity fix in the Go port of OpenDataLoader PDF. Keep it fixture-backed, local, and tightly scoped.

## Gap ID

`GAP-RO-RIGHT-COLUMN-IMAGE-GRID-SEQUENCING`

## Category

`reading-order-parity`

## Objective

Fix the residual `T17` reading-order mismatch where a dense right-column image grid is merged too early relative to the surrounding body/caption flow.

This is **not** a generic mixed-layout rewrite. It is a narrow follow-up to the split from `GAP-RO-FIXTURE-MIXED-LAYOUT-EVAL`.

## Reference behavior

The Java-style expected order for the investigated `1901.03003`-style fixture band is:

`646, 647, 648, 649, 650, 653, 654, 655, 657, 658, 659, 660, 661, 662, 656, 663, 664, 665, 666`

The key parity signal is that the right-column image-grid cluster (`653..655`, `657..659`, `660..662`) should remain grouped **after** `650`, and the caption/body continuation (`656`) should follow the whole grouped cluster instead of being interrupted by premature grid merging.

## Current Go behavior

The current Go order for the same modeled case is:

`646, 647, 648, 649, 660, 661, 662, 667, 650, 653, 657, 654, 658, 655, 659, 656, 663, 664, 665, 666`

Residual problems to target:

1. the far-right subcolumn (`660..662`) jumps too early
2. the gutter/neutral sidebar-like block (`667`) is interleaved incorrectly
3. the grid group is not preserved as the expected Java-style band before `656`

## Relevant files

### Prior split report
- `reports/loops/2026-03-28-gap-ro-fixture-mixed-layout-eval.md`

### Existing T17 reports
- `reports/loops/2026-03-28-gap-ro-two-column-basic.md`
- `reports/loops/2026-03-28-gap-ro-balanced-columns.md`
- `reports/loops/2026-03-28-gap-ro-sidebar-mixed-layout.md`
- `reports/loops/2026-03-28-gap-ro-xycut-depth-stability.md`

### Go code
- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`

## Required work

1. Reproduce the residual ordering mismatch with a focused regression in `xycut_plus_plus_sorter_test.go`.
2. Identify the narrow root cause in the column/neutral/grid merge path.
3. Implement **one** scoped fix only if it is clearly justified by the expected Java-style order.
4. Keep the previously kept T17 behaviors intact:
   - basic two-column sequencing
   - balanced-column ordering
   - marginal sidebar deferral
   - deep-recursion stability fallback
5. If the root cause still fans out into multiple behaviors, stop and record a further `SPLIT` instead of widening heuristics.

## Acceptance criteria

A successful `KEEP` must include all of:

1. a focused regression that captures the current bad ordering
2. the regression passes with Java-closer ordering after the fix
3. targeted reading-order tests pass
4. `go build ./...` passes
5. the fix stays scoped to this grid/caption sequencing family

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
```

## Non-goals

- broad XYCut rewrites
- generic text-parity changes
- speculative benchmark-wide heuristics
- touching unrelated processors/subsystems

## Output format

Report:
1. exact ordering mismatch reproduced
2. root cause found
3. files changed
4. before/after evidence
5. tests/commands run
6. KEEP / SPLIT / NEEDS-HUMAN-REVIEW

When completely finished, run this command to notify me:
openclaw system event --text "Done: evaluated GAP-RO-RIGHT-COLUMN-IMAGE-GRID-SEQUENCING in opendataloader-pdf-go" --mode now
