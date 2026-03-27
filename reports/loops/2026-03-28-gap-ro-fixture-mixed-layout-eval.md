## Gap ID

`GAP-RO-FIXTURE-MIXED-LAYOUT-EVAL`

## Date

`2026-03-28`

## Decision

`SPLIT`

## Why this loop existed

- After the synthetic `T17` keeps for two-column, balanced-column, sidebar/mixed-layout, and XYCut depth stability, the next autopilot step was to look for a real fixture-backed Java↔Go reading-order mismatch.
- The goal was to avoid widening heuristics without a concrete production-style page showing the remaining divergence.

## Evidence gathered

- The benchmark evaluator path was identifiable, but trustworthy full benchmark reruns were blocked in this checkout because the local Python environment is missing `rapidfuzz`.
- Several `tests/benchmark/pdfs/*` entries are placeholder ASCII files rather than usable local PDFs, so broad fixture exploration from the benchmark set is not currently reliable here.
- One real local PDF fixture (`samples/pdf/2408.02509v1.pdf`) was still useful for narrowing the residual family.
- The temporary regression probe for a second real reading-order case modeled on `1901.03003` exposed a concrete residual mismatch:
  - expected Java-style order for the right-column image grid/caption band was:
    `... 650, 653, 654, 655, 657, 658, 659, 660, 661, 662, 656, 663, 664, 665, 666`
  - current Go order came out as:
    `... 649, 660, 661, 662, 667, 650, 653, 657, 654, 658, 655, 659, 656, 663, 664, 665, 666`
- A first exploratory guard around balanced-column merging did not isolate this cleanly; it fixed neither the grid ordering nor the caption timing without risking broader XYCut churn.

## Files changed during evaluation

- exploratory edits to:
  - `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
  - `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
- those exploratory code changes were reverted after validation showed the candidate was not clean enough to KEEP.
- tracking artifacts kept from this loop:
  - `harness/packets/codex-gap-ro-fixture-mixed-layout-eval.md`
  - `harness/packets/codex-gap-ro-right-column-image-grid-sequencing.md`

## Validation

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
```

Additional evaluation attempts:

```bash
cd tests/benchmark && python3 src/evaluator.py --prediction-dir prediction/opendataloader --ground-truth-dir ground-truth/markdown --output-filename evaluation.codex.json
file tests/benchmark/pdfs/01030000000160.pdf tests/benchmark/pdfs/01030000000178.pdf tests/benchmark/pdfs/01030000000182.pdf tests/benchmark/pdfs/01030000000185.pdf samples/pdf/2408.02509v1.pdf
```

Results:

- focused reading-order tests: passed on the reverted clean tree
- XYCut-targeted processor/unit tests: passed on the reverted clean tree
- build: passed on the reverted clean tree
- benchmark evaluator rerun: blocked locally by missing `rapidfuzz`
- several benchmark PDF entries: placeholder ASCII files, limiting trustworthy local real-fixture replay

## Why this is a SPLIT, not a KEEP

- The explored heuristic change was not narrow enough to fix the real residual mismatch cleanly.
- The remaining live issue is more specific than generic mixed-layout handling: it is a right-column image-grid / caption sequencing problem where a dense visual grid is being merged too early relative to nearby body/caption blocks.
- That deserves its own fresh-context packet instead of stretching the broader mixed-layout evaluator loop.

## Follow-up

- Next gap: `GAP-RO-RIGHT-COLUMN-IMAGE-GRID-SEQUENCING`
- Scope it to the real residual ordering around the `1901.03003`-style right-column grid/caption band.
- Keep validation narrow: fixture-derived ordering evidence + focused XYCut tests + `go build ./...`.

## Commit

`N/A`
