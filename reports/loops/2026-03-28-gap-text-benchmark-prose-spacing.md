## Gap ID

`GAP-TEXT-BENCHMARK-PROSE-SPACING`

## Date

`2026-03-28`

## Decision

`KEEP`

## Evidence

1. where the space was being lost

- The remaining prose join was still being created in `go/internal/processors/text_line_processor.go`.
- The narrowed benchmark-shaped regression for `One of the characteristics...` exposed a distinct residual case from the prior positive-gap fix: `One` and `of` arrived on the same baseline with a small false overlap (`gap = -0.32`), so `shouldInsertSyntheticSpace(...)` returned early on `gap <= 0` and never evaluated the word-boundary heuristics.
- Paragraph merging and markdown serialization were not the loss point for this symptom. Once the line processor restored the missing boundary, the assembled line text matched the Java-style prose spacing.
- The benchmark PDF in this checkout is still a Git LFS pointer (`file tests/benchmark/pdfs/01030000000118.pdf` -> `ASCII text`), so direct end-to-end benchmark regeneration was not possible locally. The kept regression therefore targets the same dense-prose chunk geometry in unit form.

2. files changed

- `go/internal/processors/text_line_processor.go`
- `go/tests/unit/processors/text_line_processor_test.go`
- `reports/loops/2026-03-28-gap-text-benchmark-prose-spacing.md`

3. before/after evidence

- Before:
  - `go test ./tests/unit/processors -run 'TestTextLineProcessorRestoresBenchmarkDenseProseSpacing$' -count=1` failed with:
    - expected: `One of the characteristics of living things is the ability to replicate and pass on genetic information`
    - actual: `Oneof the characteristics of living things is the ability to replicate and pass on genetic information`
  - checked-in benchmark prediction remained:
    - `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md:5` -> `Oneofthecharacteristicsoflivingthingsistheability toreplicateandpasson genetic information to the next`
- After:
  - `TextLineProcessor` now treats tiny extractor-style overlaps as recoverable only for inline word-boundary pairs and only within a capped tolerance (`min(avgCharWidth * 0.08, 0.5)`), preserving the existing lowercase-fragment and hyphenation guards.
  - `go test ./tests/unit/processors -run 'TestTextLineProcessorRestoresBenchmarkDenseProseSpacing$' -count=1` passes.
  - Focused safety regressions still pass for:
    - `In Proceedings`
    - `we thus propose a multi-object rectiﬁed attention`

4. tests/commands run

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors -run 'TestTextLineProcessor(RestoresBenchmarkDense(ProseSpacing|WordJoins)|SuppressesIntrawordSpaceAfterSingleLowercaseFragment|InsertsSpaceForSmallPositiveWordBoundaryGap)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -run 'TestProcessJavaDocument(PreservesBibliographyBoundarySpaceInFixture|SuppressesIntrawordSyntheticSpacesInFixture)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/pdfbox -run 'Test(ExtractTextChunks|ExtractTextChunksAvoidFalseOverlapForSurveyIEEEFixture)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors -run 'TestTextLineProcessorRestoresBenchmarkDenseProseSpacing$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
rg -n "Oneofthecharacteristics|toreplicateandpasson" tests/benchmark/prediction/opendataloader/markdown/*.md || true
file tests/benchmark/pdfs/01030000000118.pdf
```

Results:

- focused processor regressions: passed
- focused `1901.03003.pdf` processor regressions: passed
- focused `pdfbox` tests: passed
- `go build ./...`: passed
- benchmark prediction grep still shows stale joined prose in checked-in artifacts
- benchmark PDF is an LFS pointer in this checkout, so live benchmark regeneration remains blocked

Additional limitation:

- The packet-listed broad test command `go test ./tests/unit/processors ./internal/processors ./tests/unit/pdfbox -count=1` is not clean in this sandbox because `TestHybridDocumentProcessorWritesTriageLog` panics while binding an `httptest` listener (`listen tcp6 [::1]:0: bind: operation not permitted`). The spacing fix itself was validated with the focused commands above.

5. KEEP / DISCARD / SPLIT

- `KEEP`
