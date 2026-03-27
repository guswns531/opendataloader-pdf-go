## Gap ID

`GAP-RO-GUTTER-TOKEN-SEED-GROUPING-1901`

## Date

`2026-03-28`

## Decision

`SPLIT`

## Summary

- Reproduced the packet fixture on `samples/pdf/1901.03003.pdf` pages `12-13`.
- Confirmed the corruption still appears with `--reading-order off`, so final page-level XYCut sorting is not the root cause.
- Tested one upstream hypothesis: preserve source/extractor order as the tie-break for near-baseline text objects before paragraph/list grouping.
- That hypothesis was not KEEP-worthy: focused tests passed, but the real fixture markdown was unchanged.

## What was reproduced

Commands run:

```bash
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
rg -n "west|united|arsenal|football|manchester|messageid|briogestone|contracers" /tmp/odl-moran-md/1901.03003.md
sed -n '60,150p' /tmp/odl-moran-md/1901.03003.md

./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md-off --format markdown --image-output off --quiet --reading-order off
rg -n "west|united|arsenal|football|manchester|messageid|briogestone|contracers" /tmp/odl-moran-md-off/1901.03003.md
sed -n '60,150p' /tmp/odl-moran-md-off/1901.03003.md
```

Observed in both modes:

- standalone gutter tokens such as `west`, `united`, `arsenal`, `football`, `manchester messageid`, `contracers`
- contaminated body/list lines such as:
  - `- west The experiments above are all based on cropped text`
  - `- in more application scenarios, irregular and multi-`
  - `- instance, Liu et al. [29] and Ch'ng et al. [7] released`
  - `————— briogestone In this paper, we presented a multi-object rec-`

## What was tried

- I changed the pre-grouping sort tie-break in `go/internal/processors/document_processor.go` so same-band text objects preferred source/extractor order over pure `X` ordering.
- I added focused tests for that sort behavior and for the MORAN fixture shape.
- After rebuilding `./bin/opendataloader-pdf`, the real page `12-13` markdown output did not change at all.
- I reverted the code and tests.

## What this proves

- The bad seed/grouping is established earlier than the pre-paragraph/pre-list sort tie-break that I changed.
- Because the corruption survives `--reading-order off`, the final `sortDocumentContents` XYCut pass is not the responsible stage.
- The next loop should trace one stage earlier, most likely:
  1. `TextLineProcessor.Process` line ordering or line creation on these pages
  2. the mixed object sequence produced immediately before `ListProcessor.Process`
  3. whether the relevant gutter tokens are already adjacent to bullet/body lines before any paragraph transform

## Validation

Commands run:

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -run 'TestSortPreGroupingObjectsUsesSourceOrderWithinSharedTextBand|TestProcessJavaDocumentKeepsMORANGutterTokensOutOfConclusionFlow' -count=1
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
```

Results:

- focused temporary processor tests: passed before revert
- CLI rebuild: passed
- broader `./tests/unit/processors/...` sweep is not clean in this sandbox because `TestHybridDocumentProcessorWritesTriageLog` opens an `httptest` listener and fails with `bind: operation not permitted`

## Tree state

- No kept source changes.
- This report is the only new loop artifact from this evaluation.

## Commit

`N/A`
