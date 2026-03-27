# Loop Report

## Gap ID

`GAP-RO-PRELIST-SEED-TRACE-1901`

## Date

`2026-03-28`

## Decision

`SPLIT`

## Why this loop existed

- Trace the Go object sequence on `samples/pdf/1901.03003.pdf` pages `12-13` immediately before paragraph/list processing.
- Prove the first stage where gutter tokens such as `west`, `united`, `arsenal`, `football`, `manchester`, `messageid`, `briogestone`, and `contracers` become adjacent to main-flow body/conclusion content.
- Keep scope upstream of `ListProcessor.Process` and avoid another late `sortDocumentContents` / XYCut loop.

## Evidence before

- Prior reports had already shown the corruption survives `--reading-order off`, so final page-level XYCut is not the cause.
- Real fixture output still contains:
  - `- west The experiments above are all based on cropped text`
  - `- [18] proposed methods to improve the performance football of multi-oriented text detection`
  - `manchester messageid`
  - `————— briogestone In this paper, we presented a multi-object rec-`
  - `contracers\ttiﬁed attention network (MORAN) for scene text`

## What changed

- No source changes kept.
- I added a temporary focused processor trace in `go/internal/processors/document_processor_test.go`, ran it, captured the pre-list ordering evidence, then reverted it.

## What the trace proved

- On Go page number `11` (the selected PDF page `12`), the bad adjacency already exists in the direct output of `TextLineProcessor.Process`, before:
  - `textLinesToObjects`
  - `sortObjects`
  - `SpecialTableProcessor`
  - `HeaderFooterProcessor`
  - `ListProcessor.Process`
  - paragraph conversion
- Representative `TextLineProcessor.Process` ordering from the temporary trace:
  - `west`
  - `The experiments above are all based on cropped text`
  - `...`
  - `football`
  - `of multi-oriented text detection. Therefore, scene`
  - `...`
  - `briogestone`
  - `In this paper, we presented a multi-object rec-`
  - `contracers`
  - `tiﬁed attention network (MORAN) for scene text`
- The immediate pre-list sequence matched that same wrong adjacency, so the corruption is already baked into the text-line stage rather than introduced by list or paragraph processors.

## Java comparison used

- Java `DocumentProcessor.processDocument(...)` hands the mixed page object stream into `TextLineProcessor.processTextLines(pageContents)` without splitting all text chunks away and without a global Y/X resort before list processing.
- Go instead:
  - splits text chunks out of the mixed object stream,
  - builds lines from text chunks alone,
  - returns lines already globally sorted by baseline/X in `TextLineProcessor.Process`.
- That is enough to justify the next split point, but not yet enough to justify a safe KEEP-sized behavior change in this loop.

## Validation

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -run TestTracePreListSequenceOnMORANFixturePages12To13 -count=1 -v
git diff -- go/internal/processors/document_processor_test.go
git status --short
```

Results:

- focused temporary trace test: passed and emitted the evidence above
- temporary test reverted cleanly
- final tree state: clean except loop artifacts and pre-existing untracked packet/plan files

## Regressions checked

- No code was kept, so there are no new behavioral regressions from this loop.

## Remaining uncertainty

- The first wrong stage is now narrowed to Go text-line formation/ordering, but the smallest safe fix is still unclear.
- The unresolved question is whether the parity fix belongs in:
  - making Go `TextLineProcessor` preserve mixed source order more like Java,
  - changing how `DocumentProcessor` feeds mixed objects into text-line creation,
  - or carrying separator/non-text context through text-line formation so the gutter stream cannot become adjacent in the first place.

## Follow-up

- Next packet should be a narrower text-line parity task:
  - compare Go `TextLineProcessor.Process` ordering on PDF page `12` against Java `TextLineProcessor.processTextLines(...)`,
  - trace the original mixed object stream entering Go text-line formation,
  - determine whether the first parity break is caused by `splitTextChunks(...)` discarding mixed ordering/context or by the explicit global baseline/X sort inside Go `TextLineProcessor.Process`.

## Commit

`N/A`
