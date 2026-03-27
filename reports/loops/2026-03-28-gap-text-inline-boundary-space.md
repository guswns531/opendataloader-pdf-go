## Gap ID

`GAP-TEXT-INLINE-BOUNDARY-SPACE`

## Date

`2026-03-28`

## Decision

`KEEP`

## Evidence

1. where the space was being lost

- The surviving bibliography joins were being created in `go/internal/processors/text_line_processor.go`, not `ParagraphProcessor`.
- After the extractor-width fix, adjacent same-baseline chunks could still have a small positive gap that was too small for the generic `avgCharWidth * 0.3` spacing threshold, so `addSyntheticSpacing` concatenated word-boundary chunks directly.
- The kept fix adds a narrower fallback for positive small gaps at inline alphanumeric boundaries, without changing reading order or paragraph merging.

2. files changed

- `go/internal/processors/text_line_processor.go`
- `go/tests/unit/processors/text_line_processor_test.go`
- `go/internal/processors/document_processor_test.go`

3. live fixture behavior before/after

- Before:
  - `/tmp/odl-inline-boundary-space-before/1901.03003.md:1478` had `InProceedings of the IEEE Conference on Computer`
  - `/tmp/odl-inline-boundary-space-before/1901.03003.md:1617` had `InProceedings of International Conference on Com-`
  - `/tmp/odl-inline-boundary-space-before/1901.03003.md:1866` had `[49]... Word spotting in the wild. InProceedings of European Conference on Computer Vision`
- After rebuilding the CLI and rerunning the live fixture:
  - `/tmp/odl-inline-boundary-space-full2/1901.03003.md:1478` has `In Proceedings of the IEEE Conference on Computer`
  - `/tmp/odl-inline-boundary-space-full2/1901.03003.md:1617` has `In Proceedings of International Conference on Com-`
  - `/tmp/odl-inline-boundary-space-full2/1901.03003.md:1866` has `[49]... Word spotting in the wild. In Proceedings of European Conference on Computer Vision`
- The packet acceptance target was met: at least one live `InProceedings` symptom became `In Proceedings`, and in this run the cited fixture cases all flipped to the spaced form.

4. tests/commands run

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors -run 'TestTextLineProcessor(InsertsSpaceForSmallPositiveWordBoundaryGap|InsertsSpaceAndLinksLineArtBullet|GroupsChunksOnSameBaseline|SeparatesDifferentBaselines)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -run TestProcessJavaDocumentPreservesBibliographyBoundarySpaceInFixture -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-inline-boundary-space --format markdown --image-output off --quiet
rg -n "InProceedings|In Proceedings" /tmp/odl-inline-boundary-space/1901.03003.md || true
```

Results:

- focused `TextLineProcessor` tests: passed
- fixture-backed `DocumentProcessor` regression test: passed
- packet-listed pdfbox/unit tests: passed
- Go build: passed
- rebuilt CLI fixture conversion: passed
- live grep confirms the target lines now render as `In Proceedings`

5. KEEP / DISCARD / SPLIT

- `KEEP`
