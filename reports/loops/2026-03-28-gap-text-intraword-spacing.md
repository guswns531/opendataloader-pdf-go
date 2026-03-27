## Gap ID

`GAP-TEXT-INTRAWORD-SPACING`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

- Suppress false synthetic spaces inside a single word on `samples/pdf/1901.03003.pdf` without regressing the recently fixed bibliography boundary spaces.

## Evidence before

- Live markdown output showed `m ulti- objectr ectiﬁed a ttention network` near the start of the fixture.
- The same fixture already needed to keep bibliography cases as `In Proceedings`, not `InProceedings`.
- Probing the page-1 line assembly showed the bad line was built from fragment chunks like `m` + `ulti-`, `o` + `bject`, `r` + `ectiﬁed`, and `a` + `ttention`, with synthetic spaces being inserted at the wrong fragment boundaries.

## What changed

- `go/internal/processors/text_line_processor.go`
- `go/tests/unit/processors/text_line_processor_test.go`
- `go/internal/processors/document_processor_test.go`
- Added two narrow guards in `shouldInsertSyntheticSpace`:
  - do not insert a synthetic space before a single lowercase fragment that immediately follows a hyphen
  - do not insert a synthetic space after a single lowercase fragment when the next chunk is a lowercase continuation
- Added a focused unit regression for the fragment pattern and a fixture-backed regression for the sample PDF intro paragraph.

## Validation

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors -run 'TestTextLineProcessor(InsertsSpaceForSmallPositiveWordBoundaryGap|SuppressesIntrawordSpaceAfterSingleLowercaseFragment|InsertsSpaceAndLinksLineArtBullet|GroupsChunksOnSameBaseline|SeparatesDifferentBaselines)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -run 'TestProcessJavaDocument(PreservesBibliographyBoundarySpaceInFixture|SuppressesIntrawordSyntheticSpacesInFixture)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-intraword-spacing-fresh-2 --format markdown --image-output off --quiet
rg -n "m ulti-|multi- object|objectr ecti|a ttention|InProceedings|In Proceedings" /tmp/odl-intraword-spacing-fresh-2/1901.03003.md || true
```

Results:

- `go test ./tests/unit/processors -count=1`: failed for an unrelated existing sandbox issue in `TestHybridDocumentProcessorWritesTriageLog` because `httptest.NewServer` could not bind a local port.
- Focused `TextLineProcessor` tests: passed.
- Focused fixture-backed `internal/processors` tests: passed.
- `go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1`: passed.
- `go build ./...`: passed.
- CLI rebuild: passed.
- Fresh fixture conversion: passed.
- Live grep no longer finds `m ulti-`, `multi- object`, `objectr ecti`, `a ttention`, or `InProceedings`; it still finds `In Proceedings` as expected.

## Regressions checked

- Bibliography spacing remains correct on the same fixture, including `In Proceedings of European Conference on Computer Vision`.
- The page-1 intro paragraph no longer shows the targeted false synthetic spaces.

## Remaining uncertainty

- The intro line still renders as `multi-objectrectiﬁed attention` rather than fully restoring `multi-object rectiﬁed attention`; this packet only addressed the false-positive synthetic spacing family, not the separate missing-boundary side of that extraction.
- The narrowed lowercase-fragment guard is intentionally local, but broader corpora may still reveal other fragment patterns outside this packet’s scope.

## Follow-up

- Split any remaining `multi-objectrectiﬁed`-style missing-boundary issue into a separate packet instead of widening this spacing fix.

## Commit

`N/A`
