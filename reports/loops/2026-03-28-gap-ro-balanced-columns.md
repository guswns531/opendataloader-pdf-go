## Gap ID

`GAP-RO-BALANCED-COLUMNS`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

- Go reading-order still mis-sequenced uneven two-column pages after the basic split-detection fix.
- Center-biased headings and columns that start or end at different heights could still confuse the merge order compared with Java.

## Evidence before

- The active Codex loop produced a focused XYCut++ patch plus two regressions covering imbalanced column boundaries and gutter-spanning neutral blocks.
- Existing basic two-column coverage did not explicitly protect these asymmetric merge cases.

## What changed

- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
  - split detected columns into left/right plus gutter-spanning neutral elements
  - allowed split detection to retry on filtered non-neutral objects when center-spanning blocks masked the gutter
  - added a balanced-column merge path that preserves leading/trailing asymmetric content while keeping overlapping column cores in column order
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
  - added a regression for uneven column starts/ends
  - added a regression for center-heading / gutter-spanning neutral content during column merge

## Validation

```bash
cd go && go test ./internal/processors/readingorder/... -count=1
cd go && go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && go build ./...
```

Results:

- focused reading-order tests: passed
- XYCut-targeted processor/unit tests: passed
- build: passed

## Regressions checked

- previously added basic two-column XYCut regressions still pass
- targeted gutter-spanning and imbalanced-column cases now pass without widening scope into sidebar/mixed-layout heuristics

## Remaining uncertainty

- This is still synthetic regression evidence; fixture-level Java-vs-Go comparisons for real mixed-layout PDFs are the next trustable signal.
- `go test ./...` remains noisy in unrelated areas, so this KEEP is based on the scoped validator ladder for `T17`.

## Follow-up

- Continue with `GAP-RO-SIDEBAR-MIXED-LAYOUT` or another fixture-backed `T17` child gap, using real Java-vs-Go output diffs to confirm where balanced-column logic still falls short.

## Commit

`N/A`
