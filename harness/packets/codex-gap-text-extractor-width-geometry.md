# Codex Gap Task Packet — GAP-TEXT-EXTRACTOR-WIDTH-GEOMETRY

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-TEXT-EXTRACTOR-WIDTH-GEOMETRY`

## Category

`text-parity`

---

## Objective

Fix extractor chunk width/geometry so adjacent same-baseline runs do not falsely overlap and collapse visible spaces later in `TextLineProcessor`.

This is the direct follow-up split from `GAP-TEXT-INLINE-RUN-GEOMETRY-PROBE`.

---

## Reference behavior (Java)

Java/PDFBox preserves ordinary boundaries in these sample spans well enough that prose does not collapse into joins like:

- `survey.IEEE`
- `InProceedings`
- `theCreationofLife`

---

## Current Go behavior

The previous probe established that at least one surviving join is caused upstream of `TextLineProcessor`:

- on `samples/pdf/1901.03003.pdf`, extractor chunk `tion in imagery: A survey.` ends at `x=201.678`
- next chunk `IEEE Trans. Pattern Anal.` starts at `x=181.439`
- reported overlap is `-20.239`, so line assembly correctly sees no gap and emits `survey.IEEE`

That means the next narrow fix target is extractor geometry / width estimation, not another line-join heuristic.

---

## Why this matters

- addresses a root cause instead of patching symptoms
- should improve multiple joined-word artifacts with one scoped fix
- keeps text-parity work moving without mixing reading-order changes

---

## Relevant files

### Go
- `go/pkg/pdfbox/extractor/text_extractor.go`
- `go/pkg/pdfbox/extractor/content_parser.go`
- `go/internal/processors/text_line_processor.go` (read-only unless absolutely needed for validation)
- focused tests under `go/pkg/pdfbox/extractor/` or `go/tests/unit/pdfbox/`

### Existing reports
- `reports/loops/2026-03-28-gap-text-inline-gap-spacing.md`
- `reports/loops/2026-03-28-gap-text-inline-run-geometry-probe.md`

### Fixtures
- `samples/pdf/1901.03003.pdf`
- benchmark markdown pairs under `tests/benchmark/{ground-truth,prediction}/opendataloader/markdown/`

---

## Required work

- inspect current extractor width computation against Java/PDFBox behavior where possible
- reproduce the `survey.IEEE` failure and confirm the false overlap disappears after the fix
- keep scope narrow to extractor geometry / width estimation for adjacent runs
- add focused regression coverage if you keep a fix
- validate against at least these live symptoms after the change:
  - `survey.IEEE`
  - one `InProceedings`-style join from the same sample/report set
- if evidence is mixed, prefer a `SPLIT` report over speculative heuristics

---

## Acceptance criteria

1. Extracted chunk geometry no longer falsely reports overlap for the `survey.` / `IEEE` case, or a narrower root cause is conclusively identified.
2. Any kept change improves live fixture output, not only synthetic tests.
3. Targeted validation commands pass for any kept change.
4. A loop report records KEEP / DISCARD / SPLIT with concise evidence.

---

## Validation commands

Run these before finishing:

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-width-geometry-probe --format markdown --image-output off --quiet
rg -n "survey\.IEEE|InProceedings|theCreationofLife|Oneofthecharacteristics" /tmp/odl-width-geometry-probe/1901.03003.md tests/benchmark/prediction/opendataloader/markdown/*.md || true
```

Also capture concise evidence for the `survey.` / `IEEE` chunk geometry before/after.

---

## Non-goals

Do **not** broaden into:
- reading-order work
- paragraph/serializer cleanup beyond what falls out naturally from corrected extractor geometry
- unrelated encoding work unless directly required for the width fix
- large refactors

---

## Output format

When done, report:

1. what width/geometry behavior was wrong
2. files changed
3. behavior changed on live fixture(s)
4. tests/commands run and their result
5. whether this became a KEEP, DISCARD, or SPLIT
