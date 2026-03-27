## Gap ID

`GAP-TEXT-INLINE-GAP-SPACING`

## Date

`2026-03-28`

## Decision

`SPLIT`

## Why this loop existed

- The packet targeted missing word-boundary recovery across adjacent same-baseline text runs without literal spaces.
- The main fixture evidence was the checked-in benchmark markdown pairs and `samples/pdf/1901.03003.pdf`.

## Evidence before

- Checked-in benchmark pairs still show collapsed boundaries such as `theCreationofLife`, `Oneofthecharacteristicsoflivingthingsistheability`, `#nucleardivisions`, and `#daughtercellsproduced`.
- A live markdown run on `samples/pdf/1901.03003.pdf` still reproduced:
  - `m ulti-objectrectiﬁedattention`
  - `More- over`
  - `survey.IEEE`
- Current extractor logic only folds explicit boundary whitespace; it does not have a fixture-proven same-baseline gap recovery path for these joined runs.

## What changed

- Prototyped a narrow extractor heuristic plus focused tests for synthetic positioned runs.
- Discarded the prototype after validation because it did not move the live sample output.
- Left the worktree clean; no code changes were kept.

## Validation

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-inline-gap-spacing --format markdown --image-output off --quiet
```

Additional scoped checks during the discarded prototype:

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/extractor -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors -run 'TestTextLineProcessor(InsertsSpace|Groups|Separates)' -count=1
```

Results:

- packet validation suites: passed
- build: passed
- sample markdown output: unchanged for the target joined-word snippets
- candidate patch: not KEEP-worthy because fixture evidence did not improve

## Regressions checked

- baseline PDFBox extractor tests passed
- baseline PDFBox unit tests passed
- no code changes remain in the worktree

## Remaining uncertainty

- The discarded heuristic improved only synthetic unit cases; it did not reach the real fixture symptom.
- That suggests the remaining gap is likely deeper than simple post-append space inference, potentially involving chunk geometry, run segmentation, or a later stage that still collapses boundaries.

## Follow-up

- Split this into a narrower next gap that identifies where the sample’s `objectrectiﬁedattention` and `survey.IEEE` joins originate.
- Recommended next probe: dump extracted chunk geometry around the affected sample spans before `TextLineProcessor` so the next packet can target the real layer instead of adding another blind spacing heuristic.

## Commit

`N/A`
