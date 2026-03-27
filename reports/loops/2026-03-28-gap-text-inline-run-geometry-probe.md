## Gap ID

`GAP-TEXT-INLINE-RUN-GEOMETRY-PROBE`

## Date

`2026-03-28`

## Decision

`SPLIT`

## Evidence

- Live markdown still reproduces the target sample on `samples/pdf/1901.03003.pdf`:
  - `/tmp/odl-inline-run-probe/1901.03003.md:1874` contains `A survey.IEEE Trans. Pattern Anal.`
  - `/tmp/odl-inline-run-probe/1901.03003.md:23` still contains `m ulti-objectrectiﬁedattention`
- A temporary scoped probe on page 16 isolated the `survey.IEEE` join:
  - extractor chunk 83: `tion in imagery: A survey.` at `x=72.164`, `width=129.514`, `baseline=92.955`
  - extractor chunk 84: `IEEE Trans. Pattern Anal.` at `x=181.439`, `width=124.532`, `baseline=92.955`
  - chunk 83 ends at `201.678`, so extractor geometry reports a `-20.239` overlap against chunk 84 even though these are separate text runs
  - `TextLineProcessor` then emits one line: `tion in imagery: A survey.IEEE Trans. Pattern Anal.`
- This means the visible join does not exist as joined text immediately after extraction, but it is created when `TextLineProcessor` concatenates same-baseline chunks with no synthetic space because the extractor width estimate already says they overlap.

## Validation

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-inline-run-probe --format markdown --image-output off --quiet
```

Results:

- targeted Go tests: passed
- Go build: passed
- live fixture conversion: passed and still reproduces the target joins

## Why No Fix Was Kept

- A narrow `TextLineProcessor` heuristic would be another blind spacing patch; the observed trigger is bad chunk geometry, not missing grouping logic alone.
- A trustworthy fix now points at extractor geometry and width estimation, which is a different narrower root-cause packet.
- The separate page-2 `m ulti-objectrectiﬁedattention` symptom appears to involve later paragraph assembly across multiple runs/lines and should not be bundled into this geometry probe.

## Follow-up

- Next gap: `GAP-TEXT-EXTRACTOR-WIDTH-GEOMETRY`
- Scope for that packet:
  - validate width computation against Java/PDFBox behavior for adjacent same-baseline runs
  - improve extractor geometry so `A survey.` and `IEEE ...` no longer overlap in extracted chunks
  - re-check `survey.IEEE` plus `InProceedings`-style joins without touching reading-order work

## Commit

`N/A`
