## Gap ID

`EVAL-TEXT-PARITY-DISCOVERY`

## Date

`2026-03-28`

## Decision

`SPLIT`

## Why this loop existed

- `GAP-TEXT-TJ-OFFSET` was already stale, so the next step was a fixture-backed discovery pass for real remaining `T16` text mismatches.
- The goal of this loop was evidence gathering and packetization only, not a broad extractor patch.

## Evidence before

- Checked-in benchmark markdown fixtures still show repeated word-boundary collapse that does not fit the now-kept whitespace-only `TRIMSPACE` refinement:
  - `tests/benchmark/ground-truth/markdown/01030000000118.md` expects `Growth and the Creation of Life` and `One of the characteristics of living things is the ability`, while `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md` contains `Growth and theCreationofLife` and `Oneofthecharacteristicsoflivingthingsistheability`.
  - `tests/benchmark/ground-truth/markdown/01030000000119.md` expects `nuclear divisions` and `daughter cells`, while `tests/benchmark/prediction/opendataloader/markdown/01030000000119.md` contains `#nucleardivisions` and `#daughtercellsproduced`.
- A current readable sample still reproduces the same family on the live worktree:
  - `samples/pdf/1901.03003.pdf` currently renders `m ulti-objectrectiﬁedattention` in generated markdown.
- Extractor inspection shows current append logic only preserves explicit leading whitespace or whitespace-only runs:
  - `foldLeadingBoundaryWhitespace(...)`
  - `appendBoundaryWhitespace(...)`
  - there is no same-baseline geometric gap heuristic for adjacent non-whitespace runs.

## What changed

- added this evaluator report
- added one new narrow packet: `harness/packets/codex-gap-text-inline-gap-spacing.md`

## Validation

```bash
sed -n '1,220p' plan/gaps/EVAL-TEXT-PARITY-CORE.md
sed -n '1,220p' reports/loops/2026-03-28-gap-text-tj-offset.md
sed -n '1,220p' reports/loops/2026-03-28-gap-text-trimspace-boundary-whitespace.md
sed -n '1,220p' reports/loops/2026-03-28-gap-text-winansi.md
sed -n '1,220p' reports/loops/2026-03-28-gap-text-tounicode.md
command -v go && go version && command -v java && java -version
GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf/
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-sample-1901 --format markdown --image-output off --quiet
python3 - <<'PY' ... benchmark pair scan for joined words ... PY
```

Additional probes attempted:

```bash
./bin/opendataloader-pdf tests/benchmark/pdfs/01030000000118.pdf --output-dir /tmp/odl-eval-01030000000118 --format markdown --table-method cluster --image-output off --quiet
./bin/opendataloader-pdf tests/benchmark/pdfs/01030000000119.pdf --output-dir /tmp/odl-eval-01030000000119 --format markdown --table-method cluster --image-output off --quiet
python3 tests/benchmark/src/evaluator.py tests/benchmark/ground-truth /tmp/odl-eval-01030000000182 --doc-id 01030000000182
```

Results:

- Java reference generation was blocked locally because this environment has no runnable Java runtime.
- The benchmark evaluator could not run because `rapidfuzz` is unavailable in the environment.
- Benchmark PDFs `01030000000118.pdf` and `01030000000119.pdf` are present, but the current Go CLI fails to read them with `xRefTable failed: the file may be damaged`.
- The checked-in benchmark markdown pairs still provide strong fixture evidence of the remaining mismatch family.
- The current Go CLI reproduces the same family on `samples/pdf/1901.03003.pdf`.

## Regressions checked

- inspected current extractor boundary-space logic in `go/pkg/pdfbox/extractor/text_extractor.go`
- confirmed this loop did not modify extractor behavior or tests
- confirmed the current Go CLI still builds with `GOCACHE=/tmp/odl-gocache`

## Remaining uncertainty

- Fresh Java-vs-Go output could not be regenerated locally because `java` is missing.
- Some benchmark documents also show table/layout corruption, so this loop intentionally isolates only the repeated inline word-boundary symptom rather than broader reading-order/table issues.
- `1901.03003.pdf` also contains other artifacts (`More- over`, `survey.IEEE`), but those may require separate packetization beyond this narrow gap.

## Follow-up

- Split the stale broad `GAP-TEXT-TRIMSPACE` follow-up into a narrower next packet: `GAP-TEXT-INLINE-GAP-SPACING`.
- The next implementation loop should target missing inferred spaces between adjacent non-whitespace text runs on the same baseline, using `1901.03003.pdf` plus the checked-in benchmark markdown pairs as fixture evidence.
- Outcome: `SPLIT`, not `KEEP`.

## Commit

`N/A`
