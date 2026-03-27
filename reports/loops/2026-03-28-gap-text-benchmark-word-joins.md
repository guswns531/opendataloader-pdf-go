## Gap ID

`GAP-TEXT-BENCHMARK-WORD-JOINS`

## Date

`2026-03-28`

## Decision

`KEEP`

## Evidence

1. where the space was being lost

- The surviving joined-word family is still being created in `go/internal/processors/text_line_processor.go`, not in paragraph merging or markdown serialization.
- `addSyntheticSpacing(...)` only inferred a space for small positive same-baseline gaps when the gap exceeded `avgCharWidth * 0.12`.
- That floor was still too high for dense prose-style runs like the benchmark joins around `theCreationofLife` and `Oneofthecharacteristics`, so adjacent word chunks with real but very small positive gaps were concatenated.
- The kept change narrows that floor to `avgCharWidth * 0.05`, while retaining the earlier singleton/lowercase continuation guards that prevented false intra-word splits on `1901.03003.pdf`.

2. files changed

- `go/internal/processors/text_line_processor.go`
- `go/tests/unit/processors/text_line_processor_test.go`
- `go/internal/processors/document_processor_test.go`
- `go/tests/unit/pdfbox/extractor_test.go`

3. before/after evidence

- Before:
  - checked-in benchmark predictions still show:
    - `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md:3` -> `# Growth and theCreationofLife`
    - `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md:5` -> `Oneofthecharacteristicsoflivingthingsistheability ...`
    - `tests/benchmark/prediction/opendataloader/markdown/01030000000113.md:37` -> `GrowthandtheCreationofLife...`
  - the new benchmark-shaped line assembly regression failed under the old `0.12` floor because all inter-word gaps were positive but still below the old cutoff.
- After:
  - `TestTextLineProcessorRestoresBenchmarkDenseWordJoins` now reproduces the target root-cause family locally and passes with:
    - `Growth and the Creation of Life`
  - previously kept live-fixture protections still pass:
    - bibliography boundary remains `In Proceedings`, not `InProceedings`
    - `1901.03003.pdf` fixture still renders `multi-object rectiﬁed attention` without reintroducing `m ulti-`, `multi- object`, `objectr ectiﬁed`, or `a ttention`
- The checked-in benchmark markdown files remain stale because the benchmark PDFs in this checkout are Git LFS pointer files, not readable PDFs, so a fresh local regeneration was not possible in this environment.

4. tests/commands run

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors -run 'TestTextLineProcessor(GroupsChunksOnSameBaseline|SeparatesDifferentBaselines|InsertsSpaceAndLinksLineArtBullet|InsertsSpaceForSmallPositiveWordBoundaryGap|RestoresBenchmarkDenseWordJoins|SuppressesIntrawordSpaceAfterSingleLowercaseFragment)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -run 'TestProcessJavaDocument(PreservesBibliographyBoundarySpaceInFixture|SuppressesIntrawordSyntheticSpacesInFixture)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/pdfbox -run 'Test(ExtractTextChunks|ExtractTextChunksAvoidFalseOverlapForSurveyIEEEFixture)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
rg -n "theCreationofLife|Oneofthecharacteristics|GrowthandtheCreationofLife" tests/benchmark/prediction/opendataloader/markdown/*.md || true
file tests/benchmark/pdfs/01030000000118.pdf tests/benchmark/pdfs/01030000000113.pdf
```

Results:

- focused `TextLineProcessor` regressions: passed
- live `1901.03003.pdf` processor regressions: passed
- focused `pdfbox` tests: passed
- packet-listed `pkg/pdfbox` and `tests/unit/pdfbox` suites: passed
- `go build ./...`: passed
- benchmark prediction grep still shows stale joined words in checked-in artifacts
- `file` confirms `tests/benchmark/pdfs/01030000000118.pdf` and `tests/benchmark/pdfs/01030000000113.pdf` are ASCII Git LFS pointer files in this checkout, so direct local regeneration is blocked

5. KEEP / DISCARD / SPLIT

- `KEEP`
