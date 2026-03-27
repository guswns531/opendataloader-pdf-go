## Gap ID

`GAP-RO-2408-SIDEBAR-TAIL-SEQUENCING`

## Date

`2026-03-28`

## Decision

`SPLIT`

## Why this loop existed

- After the kept `T17` fixes for basic two-column ordering, balanced columns, mixed-layout sidebars, recursion depth stability, and the `1901.03003` right-column image-grid sequencing family, the next task was to identify one remaining credible reading-order parity gap.
- The goal was to find a live residual mismatch family with fixture-backed or fixture-shaped evidence, without widening XYCut heuristics speculatively.

## Evidence before

- The remaining real local reading-order fixtures in this checkout are still narrow:
  - `samples/pdf/1901.03003.pdf`
  - `samples/pdf/2408.02509v1.pdf`
- The earlier fixture-eval/report chain already consumed the strongest hard parity signal from `1901.03003` and kept `GAP-RO-RIGHT-COLUMN-IMAGE-GRID-SEQUENCING`.
- The Java XYCut suite still contains one additional real-fixture-derived case that Go does not encode exactly: the `2408.02509v1` academic-paper two-column layout with a marginal arXiv sidebar (`95..105` in the Java test).

## Probe run

- A temporary Go regression replayed the Java `2408.02509v1` geometry exactly.
- Current Go ordering for that geometry came out as:

```text
[95 96 97 98 99 105 100 101 102 103 104]
```

- The narrow difference is that Go reinserts the marginal sidebar `105` between the trailing left-column blocks `99` and `100`.

## Why this is not yet a KEEP

- The Java test comments describe `105` as the last item in the expected reading order, but the actual Java assertions explicitly treat the arXiv sidebar position as flexible and do not require `105` to trail `100..104`.
- The current evidence therefore shows a real ordering difference, but not a trustworthy Java-parity violation.
- Tightening Go heuristics on this basis would risk widening the mixed-layout/sidebar logic without a hard parity target.

## Files changed during evaluation

- exploratory edit, reverted:
  - `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
- kept artifact:
  - `reports/loops/2026-03-28-eval-reading-order-next-gap.md`

## Validation

Exploratory reproduction command:

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/readingorder/... -run 'TestXYCutMatchesAcademicPaperTwoColumnWithMarginalSidebar' -count=1
```

Observed exploratory failure before revert:

- expected probe order: `95 96 97 98 99 100 101 102 103 104 105`
- current Go order: `95 96 97 98 99 105 100 101 102 103 104`

Clean-tree validation after reverting the exploratory probe:

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
```

## Results

- targeted reading-order tests: passed
- XYCut-targeted processor/unit tests: passed
- build: passed

## Exact follow-up gap

- `GAP-RO-2408-SIDEBAR-TAIL-SEQUENCING`
- Required evidence before any fix:
  - capture actual Java output order for the `2408.02509v1` geometry or a comparable fixture-backed Java conversion result
  - confirm whether Java truly places `105` after the entire `100..104` body band, or whether the sidebar remains intentionally flexible
- If Java output proves `105` should trail the full body band, the likely root-cause family is still the deferred-floating/sidebar reinsertion path.
- If Java output remains flexible, this should be dropped as a parity gap and not patched heuristically.

## Commit

`N/A`
