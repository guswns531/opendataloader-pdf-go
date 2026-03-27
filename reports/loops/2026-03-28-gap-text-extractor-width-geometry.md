## Gap ID

`GAP-TEXT-EXTRACTOR-WIDTH-GEOMETRY`

## Date

`2026-03-28`

## Decision

`KEEP`

## Evidence

- Before this loop, extractor chunk geometry on `samples/pdf/1901.03003.pdf` page 15 reported a false overlap for the target bibliography line:
  - `tion in imagery: A survey.` at `x=72.164`, `width=129.514`, `end=201.678`
  - `IEEE Trans. Pattern Anal.` at `x=181.439`
  - overlap: `-20.239`
- After switching extractor advancement/width from rune-count heuristics to font width data and clamping same-baseline chunk width to the next chunk boundary, the extractor now reports:
  - `tion in imagery: A survey.` at `x=72.164`, `width=107.282`, `end=179.446`
  - `IEEE Trans. Pattern Anal.` at `x=181.439`
  - positive gap: `+1.993`
- Live markdown output improved on the target fixture after rebuilding the CLI:
  - before: `Text detection and recogni- tion in imagery: A survey.IEEE Trans. Pattern Anal.`
  - after: `Text detection and recogni- tion in imagery: A survey. IEEE Trans. Pattern Anal.`
- Residual `InProceedings` joins still remain in the same sample, so this loop fixed one extractor-width root cause but did not finish the whole inline-spacing family.

## Files Changed

- `go/pkg/pdfbox/extractor/font_decoder.go`
- `go/pkg/pdfbox/extractor/text_extractor.go`
- `go/pkg/pdfbox/extractor/content_parser_test.go`
- `go/tests/unit/pdfbox/extractor_test.go`

## Validation

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-width-geometry-probe2 --format markdown --image-output off --quiet
rg -n "survey\.IEEE|InProceedings|theCreationofLife|Oneofthecharacteristics" /tmp/odl-width-geometry-probe2/1901.03003.md || true
```

Results:

- targeted Go tests: passed
- Go build: passed
- rebuilt CLI fixture conversion: passed
- `survey.IEEE` no longer reproduces
- `InProceedings` still reproduces

## Remaining Uncertainty

- Some surviving joins appear to come from line/paragraph assembly or boundary-space synthesis rather than raw extractor width inflation.
- The current width clamping uses a small baseline epsilon; that solved the observed overlap without harming the focused tests, but broader bibliography-heavy fixtures may still reveal edge cases.

## Follow-up

- Next gap: narrow the surviving bibliography join family, likely around missing boundary-space synthesis for adjacent chunks that now have a small positive gap but still normalize into `InProceedings`.
- Keep scope separate from reading-order work.
