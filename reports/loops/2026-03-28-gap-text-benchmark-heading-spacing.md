## Gap ID

`GAP-TEXT-BENCHMARK-HEADING-SPACING`

## Date

`2026-03-28`

## Decision

`KEEP`

## Evidence

1. where the space was being lost

- The remaining `CellularCycle andReplication` join is still created in `go/internal/processors/text_line_processor.go`.
- This display-text case is narrower than the earlier dense-prose family: the line chunks are large-font alphabetic words with real but extremely small positive gaps, so the current `textLineDenseProseRatio` floor was still too high to recover `Cellular Cycle` and `and Replication`.
- `paragraph_processor.go` and markdown serialization were not the loss point for this packet. Once the line processor restores the per-line boundaries, later stages only flatten the two display lines.
- The other packet symptom, `# Growth and theCreationofLife`, does not remain as an independent root cause here. That join was already covered by the earlier kept dense-word regression; the checked-in benchmark markdown is stale because `tests/benchmark/pdfs/01030000000118.pdf` is still a Git LFS pointer in this checkout.

2. files changed

- `go/internal/processors/text_line_processor.go`
- `go/tests/unit/processors/text_line_processor_test.go`
- `reports/loops/2026-03-28-gap-text-benchmark-heading-spacing.md`

3. before/after evidence

- Before:
  - the new benchmark-shaped display regression failed with:
    - expected line 1: `Cellular Cycle`
    - actual line 1: `CellularCycle`
    - expected line 2: `and Replication`
    - actual line 2: `andReplication`
- After:
  - `TextLineProcessor` applies a tighter word-boundary floor only for larger-font alphabetic display-word runs, which restores the benchmark-shaped heading lines without broadening prose behavior.
  - `go test ./tests/unit/processors -run 'TestTextLineProcessorRestoresBenchmarkDisplayHeadingSpacing$' -count=1` passes and produces:
    - `Cellular Cycle`
    - `and Replication`
  - previously kept safeguards still pass:
    - `Growth and the Creation of Life`
    - `One of the characteristics of living things is the ability ...`
    - `we thus propose a multi-object rectiﬁed attention`
- The checked-in benchmark markdown under `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md` remains stale in this environment because the source benchmark PDF is not locally readable.

4. tests/commands run

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors -run 'TestTextLineProcessor(RestoresBenchmarkDisplayHeadingSpacing|RestoresBenchmarkDenseWordJoins|RestoresBenchmarkDenseProseSpacing|SuppressesIntrawordSpaceAfterSingleLowercaseFragment|InsertsSpaceForSmallPositiveWordBoundaryGap)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -run 'Test(ProcessJavaDocumentPreservesBibliographyBoundarySpaceInFixture|ProcessJavaDocumentSuppressesIntrawordSyntheticSpacesInFixture|PromoteListHeadingsReclassifiesOrderedSectionTitle|HeadingScoreRejectsSingleTokenMathFragments|LooksLikeSectionHeadingTextMatchesPrefixesWithoutTrailingDot|LooksLikeSectionHeadingTextRejectsFootnoteLikeAndProsePrefixes|HeadingScoreBoostsColonSuffixedTitleLines|HeadingScoreDoesNotBoostColonSuffixedNumberedListItems)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/pdfbox -run 'Test(ExtractTextChunks|ExtractTextChunksAvoidFalseOverlapForSurveyIEEEFixture)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors ./internal/processors ./tests/unit/pdfbox -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
rg -n "CellularCycle andReplication|theCreationofLife|Growth and theCreationofLife" tests/benchmark/prediction/opendataloader/markdown/*.md || true
file tests/benchmark/pdfs/01030000000118.pdf
```

Results:

- focused `TextLineProcessor` regressions: passed
- focused adjacent `internal/processors` regressions: passed
- focused `tests/unit/pdfbox` regressions: passed
- `go build ./...`: passed
- packet-listed broad `go test ./tests/unit/processors ./internal/processors ./tests/unit/pdfbox -count=1`: still fails for the existing sandbox-only `TestHybridDocumentProcessorWritesTriageLog` listener panic (`listen tcp6 [::1]:0: bind: operation not permitted`), unrelated to this spacing change
- benchmark prediction grep still finds stale joined headings in checked-in artifacts
- `file tests/benchmark/pdfs/01030000000118.pdf` reports `ASCII text`, confirming the benchmark PDF is a Git LFS pointer in this checkout

5. KEEP / DISCARD / SPLIT

- `KEEP`
