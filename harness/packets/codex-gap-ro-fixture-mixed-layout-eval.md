# Codex Gap Task Packet — GAP-RO-FIXTURE-MIXED-LAYOUT-EVAL

## Role

You are evaluating one narrow Java→Go reading-order parity gap in the Go port of OpenDataLoader PDF. Keep scope tight and evidence-first.

## Gap ID

`GAP-RO-FIXTURE-MIXED-LAYOUT-EVAL`

## Category

`reading-order-parity`

## Objective

The recent `T17` work kept synthetic fixes for:

- `GAP-RO-TWO-COLUMN-BASIC`
- `GAP-RO-BALANCED-COLUMNS`
- `GAP-RO-SIDEBAR-MIXED-LAYOUT`
- `GAP-RO-XYCUT-DEPTH-STABILITY`

Now do the next best autopilot step: find a **real fixture-backed** residual reading-order mismatch between Java and Go, preferably on a mixed-layout / sidebar / callout style page.

If you cannot get a trustworthy real-fixture comparison in this checkout, do not invent a broad heuristic. Instead, produce the narrowest evidence-backed split packet or report explaining the blocker and the next concrete sub-gap.

## Relevant files

### Existing reports
- `reports/loops/2026-03-28-gap-ro-two-column-basic.md`
- `reports/loops/2026-03-28-gap-ro-balanced-columns.md`
- `reports/loops/2026-03-28-gap-ro-sidebar-mixed-layout.md`
- `reports/loops/2026-03-28-gap-ro-xycut-depth-stability.md`

### Existing packets
- `harness/packets/codex-gap-ro-two-column-basic.md`
- `harness/packets/codex-gap-ro-balanced-columns.md`
- `harness/packets/codex-gap-ro-sidebar-mixed-layout.md`

### Go code
- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
- related reading-order tests under `go/internal/processors/readingorder/` and `go/tests/unit/processors/`

### Benchmark / fixture evidence
- `tests/benchmark/`
- any scripts or docs already used in this repo to compare Java vs Go output parity

## Required work

1. Inspect available real fixtures / benchmark artifacts for a remaining reading-order mismatch that still looks live after the current T17 fixes.
2. Prefer a single narrow failure mode, especially:
   - marginal sidebar timing
   - caption/callout interruption
   - center/gutter neutral block merge order
   - residual mixed-layout column sequencing
3. If you find a reproducible real mismatch:
   - identify where Go diverges from Java
   - implement **one** narrow fix only if the root cause is clear
   - add focused regression coverage
4. If the root cause is not yet clear:
   - do not widen heuristics blindly
   - write a loop report with `SPLIT` or `NEEDS-HUMAN-REVIEW`
   - leave behind the next narrow gap packet if helpful
5. Keep validation trustworthy. Prefer fixture diff evidence plus targeted tests/build.

## Acceptance criteria

A successful KEEP must include all of:

1. a real fixture-backed Java↔Go reading-order mismatch identified
2. one narrow root cause family only
3. targeted regression coverage added or strengthened
4. targeted reading-order tests pass
5. `go build ./...` passes

If KEEP is not justified, return `SPLIT` with concrete evidence and the next narrow gap.

## Validation commands

Start with the lightest trustworthy checks you discover. At minimum, if code changes are kept:

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
```

Also run any fixture comparison or benchmark-diff command you find relevant, and record it exactly.

## Non-goals

- unrelated text-decoding / spacing fixes
- broad reading-order rewrites without fixture evidence
- speculative heuristics without a concrete mismatch
- touching multiple subsystems in one loop

## Output format

Report:
1. exact fixture / mismatch investigated
2. where divergence occurs
3. files changed
4. before/after evidence
5. tests/commands run
6. KEEP / SPLIT / NEEDS-HUMAN-REVIEW

When completely finished, run this command to notify me:
openclaw system event --text "Done: evaluated GAP-RO-FIXTURE-MIXED-LAYOUT-EVAL in opendataloader-pdf-go" --mode now
