# Loop Report

## Gap ID

`GAP-RO-TEXTLINE-SOURCE-ORDER-1901`

## Date

`2026-03-28`

## Decision

`SPLIT`

## Why this loop existed

- Compare the original Go page-12 object stream entering text-line formation against Go `TextLineProcessor.Process` output on `samples/pdf/1901.03003.pdf`.
- Decide whether the first parity break is caused by:
  - `splitTextChunks(...)` dropping mixed-object separators before text-line formation, or
  - Go text-line geometry sorting introducing the bad gutter/body adjacency.
- Keep scope limited to the page-12 MORAN contamination pattern already isolated in the prior split.

## Evidence before

- Prior tracing had already proved the bad adjacency exists before `textLinesToObjects`, `sortObjects`, `ListProcessor`, and paragraph conversion.
- The open question from the prior split was whether the first break inside the narrowed text-line stage comes from:
  - `DocumentProcessor` stripping text chunks out of a mixed stream, or
  - `TextLineProcessor` globally reordering text chunks/lines by geometry.

## What changed

- No source changes kept.
- I added a temporary focused trace test in `go/internal/processors/document_processor_test.go`, ran it against PDF page `12`, captured the ordering evidence below, then reverted the test.

## What the trace proved

- On zero-based page `11` / PDF page `12`, the page object stream entering Go text-line formation already contains the representative contaminated adjacencies in neighboring text chunks:
  - `idx=056` `west`
  - `idx=057` `The experiments above are all based on cropped text`
  - `idx=074` `football`
  - `idx=075` `of multi-oriented text detection. Therefore, scene`
  - `idx=085` `briogestone`
  - `idx=086` `In this paper, we presented a multi-object rec-`
  - `idx=087` `contracers`
  - `idx=088` `tiﬁed attention network (MORAN) for scene text`
- After `TableBorderProcessor`, the page had `0` non-text objects remaining, so `splitTextChunks(...)` did not discard any separators or mixed-object context on this page.
- The `textChunks` slice returned from `splitTextChunks(...)` preserved the same local order as the incoming `pageElements` stream for all representative tokens above.
- `TextLineProcessor.Process(...)` emitted those representative lines in the same local sequence:
  - `west` before `The experiments above are all based on cropped text`
  - `football` before `of multi-oriented text detection. Therefore, scene`
  - `briogestone` before `In this paper, we presented a multi-object rec-`
  - `contracers` before `tiﬁed attention network (MORAN) for scene text`
- That means this loop ruled out both candidate causes named in the packet for the first page-12 break:
  - `splitTextChunks(...)` is not the first break on this page because there was no mixed stream left to split.
  - `TextLineProcessor.Process(...)` is not the first break on this page because the contaminated order was already present at its input.

## Java comparison used

- Java still establishes the reference expectation because it keeps a mixed `pageContents` stream through `TextLineProcessor.processTextLines(...)`.
- For this loop, Java was only needed to confirm the parity target. The decisive evidence came from the Go-side trace showing that page-12 contamination is already present before the two suspected Go stages have any opportunity to change it.

## Validation

```bash
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md-off --format markdown --image-output off --quiet --reading-order off
rg -n "west|united|arsenal|football|manchester|messageid|briogestone|contracers" /tmp/odl-moran-md/1901.03003.md /tmp/odl-moran-md-off/1901.03003.md
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -run TestTraceTextLineSourceOrderOnMORANPage12 -count=1 -v
git diff -- go/internal/processors/document_processor_test.go
git status --short
```

Results:

- fixture markdown reproduction: passed and reproduced the page-12 contamination both with default reading order and with `--reading-order off`
- temporary focused trace test: passed and showed the contaminated chunk order already present before `splitTextChunks(...)` and `TextLineProcessor.Process(...)`
- temporary trace reverted cleanly
- final tree state: clean except loop artifacts and the pre-existing untracked packet/plan files

## Regressions checked

- No code was kept, so there are no new behavioral regressions from this loop.

## Remaining uncertainty

- The first break is now narrowed one stage earlier than this packet assumed.
- This loop did not yet prove whether the wrong page-12 order originates in:
  - `extractor.ExtractTextChunks(...)`,
  - the Go PDFBox adapter that materializes those extracted chunks,
  - or an earlier loss of source-order context before `page.Chunks` is assembled.

## Follow-up

- Next packet should move upstream of `DocumentProcessor.processJavaDocument(...)` and trace the first bad order in the Go page-12 chunk feed itself.
- The narrow next task is:
  - compare the original order of Go `page.Chunks` on PDF page `12` against Java’s pre-text-line page content order,
  - determine whether Go extraction/materialization is already emitting the gutter tokens adjacent to main-body/conclusion chunks before any text-line processing,
  - identify the smallest safe parity fix in the extractor/input-order stage if that upstream break is confirmed.

## Commit

`N/A`
