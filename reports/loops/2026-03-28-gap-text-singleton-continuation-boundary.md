## Gap ID

`GAP-TEXT-SINGLETON-CONTINUATION-BOUNDARY`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

- After `GAP-TEXT-INTRAWORD-SPACING`, the live fixture no longer emitted `m ulti- objectr ectiﬁed a ttention`, but it still collapsed one real boundary into `multi-objectrectiﬁed attention`.
- The remaining miss was narrower than the earlier intraword fix: a single lowercase fragment (`r`) followed by a lowercase continuation (`ectiﬁed`) needed a preserved boundary to the left without reintroducing earlier false positives.

## Evidence before

- Fresh markdown on `samples/pdf/1901.03003.pdf` still contained:
  - `we thus propose a multi-objectrectiﬁed attention network`
- The same fixture already preserved prior bibliography fixes such as `In Proceedings`, so this needed to stay scoped.

## What changed

- `go/internal/processors/text_line_processor.go`
- `go/tests/unit/processors/text_line_processor_test.go`
- `go/internal/processors/document_processor_test.go`

Implementation summary:

- added a narrow pre-check in `addSyntheticSpacing` that restores a boundary before a singleton lowercase continuation fragment when:
  - the previous chunk ends at a word boundary,
  - the current chunk is a single lowercase rune,
  - the next chunk starts with a multi-rune lowercase continuation,
  - the current gap is small-positive, and
  - the following gap is large enough (or otherwise boundary-like) to justify the missing left boundary.
- updated the focused unit and fixture-backed regressions to assert `multi-object rectiﬁed attention` rather than the older collapsed form.

## Validation

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -run 'TestProcessJavaDocumentSuppressesIntrawordSyntheticSpacesInFixture' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors -run 'TestTextLineProcessorSuppressesIntrawordSpaceAfterSingleLowercaseFragment' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-inline-boundary-check --format markdown --image-output off --quiet
rg -n "multi-object|survey\.IEEE|InProceedings|theCreationofLife|Oneofthecharacteristics" /tmp/odl-inline-boundary-check/1901.03003.md tests/benchmark/prediction/opendataloader/markdown/*.md || true
```

Results:

- focused `internal/processors` regression: passed
- focused `tests/unit/processors` regression: passed
- `go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...`: passed
- `go build ./...`: passed
- CLI rebuild: passed
- live fixture conversion: passed
- live fixture now shows `multi-object rectiﬁed attention network`
- prior root-cause join `survey.IEEE` remains fixed (no hit in the fresh fixture output)
- unrelated benchmark prediction files still show residual joins such as `theCreationofLife`, `Oneofthecharacteristics`, and some `InProceedings` cases; those remain follow-up gaps outside this packet.

## Regressions checked

- page-1 intro paragraph no longer emits `multi-objectrectiﬁed`
- prior `In Proceedings` bibliography behavior remains intact in the same fixture
- no extractor-geometry regression was needed for this packet

## Follow-up

- Next best text-parity gap remains benchmark-backed joined-word cleanup outside this sample-specific fragment pattern, likely starting from `theCreationofLife` / `Oneofthecharacteristics` style joins.
- Keep it as a fresh packet instead of widening this singleton-fragment heuristic further.

## Commit

`N/A`
