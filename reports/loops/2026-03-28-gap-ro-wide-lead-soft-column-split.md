## Gap ID

`GAP-RO-WIDE-LEAD-SOFT-COLUMN-SPLIT`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

- After the kept two-column, balanced-column, sidebar, depth-stability, and right-column image-grid fixes, one real `T17` family still reproduced on `samples/pdf/2408.02509v1.pdf`.
- A wide abstract lead block was followed by two visually distinct columns whose line boxes slightly overlapped, so Go fell back to row-wise `Y/X` ordering and interleaved the right column into the left abstract body.

## Evidence before

- Real local fixture replay on `samples/pdf/2408.02509v1.pdf` showed this abstract-band ordering:
  - selected page-1 object order before:
    `3546, 3547, 3548, 3549, 3531, 3550, 3551, 3532, 3552, 3533, 3553, 3534, 3554, 3555, 3535, 3556, 3536, 3557, 3537, 3538, 3558, 3539, 3559, 3540, 3560, 3541, 3561, 3562, 3542, 3563, 3543`
- The narrow regression built from those real coordinates failed with the same row-wise pattern:
  - expected:
    `3546, 3548, 3531, 3550, 3532, 3533, 3534, 3554, 3535, 3536, 3557, 3538, 3539, 3540, 3541, 3542, 3543, 3547, 3549, 3551, 3552, 3553, 3555, 3556, 3537, 3558, 3559, 3560, 3561, 3562, 3563`
  - actual before fix:
    `3546, 3547, 3548, 3549, 3531, 3550, 3551, 3532, 3552, 3533, 3553, 3534, 3554, 3555, 3535, 3556, 3536, 3557, 3537, 3538, 3558, 3539, 3559, 3540, 3560, 3541, 3561, 3562, 3542, 3563, 3543`

## What changed

- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
  - added a narrow soft-column fallback that only activates when:
    - a very wide centered lead block exists
    - the remaining objects form two strongly center-separated, vertically overlapping bands
  - reused the existing balanced-column merge and deferred-neutral reinsertion path after that fallback split
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
  - added `TestXYCutKeepsWideAbstractLeadBeforeLeftThenRightColumns` from the real `2408.02509v1` abstract geometry

## Before/after evidence

- Focused regression:
  - before: failed with row-wise interleaving
  - after: passed with `3546 -> full left abstract column -> full right column`
- Real fixture replay on `samples/pdf/2408.02509v1.pdf`:
  - selected page-1 object order after:
    `3546, 3548, 3531, 3550, 3532, 3533, 3534, 3554, 3535, 3536, 3557, 3538, 3539, 3540, 3541, 3542, 3543, 3556, 3547, 3549, 3551, 3552, 3553, 3555, 3537, 3558, 3559, 3560, 3561, 3562, 3563`
  - the main improvement is that the left abstract column no longer row-interleaves with the right column immediately after `3546`
  - one narrow early-right residual remains (`3556`), but the original wide-lead row-wise mixing family is no longer the dominant behavior

## Files changed

- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
- `reports/loops/2026-03-28-gap-ro-wide-lead-soft-column-split.md`

## Tests / commands run

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/readingorder/... -run TestXYCutKeepsWideAbstractLeadBeforeLeftThenRightColumns -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
cd go && ../bin/opendataloader-pdf ../samples/pdf/2408.02509v1.pdf --output-dir /tmp/odl-ro-eval/2408-after --format markdown,json --table-method cluster --image-output off --quiet
```

Results:

- focused wide-lead regression: passed
- targeted reading-order tests: passed
- XYCut-targeted processor/unit tests: passed
- build: passed
- real local fixture replay: completed and showed the expected directional improvement

## Remaining uncertainty

- This fix is intentionally scoped to the wide-lead-plus-soft-columns family.
- The remaining early `3556` placement on the full fixture looks narrower than the row-wise interleaving that motivated this loop; if it still matters, it should be opened as a fresh follow-up gap rather than folded back into this heuristic.

## Commit

`aa8e925`
