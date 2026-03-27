# Codex Gap Task Packet — GAP-TEXT-BENCHMARK-HEADING-SPACING

## Role

You are implementing one narrow Java→Go parity gap in the Go port of OpenDataLoader PDF. Keep scope tight.

## Gap ID

`GAP-TEXT-BENCHMARK-HEADING-SPACING`

## Category

`text-parity`

## Objective

Continue from the freshly kept prose-spacing fixes and target the next obvious benchmark join family that now looks heading / display-text shaped rather than ordinary body prose.

Primary live symptoms to investigate:

- `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md:1` -> `CellularCycle andReplication`
- `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md:3` -> `# Growth and theCreationofLife`

If those two symptoms do **not** share a single defensible root cause, split them and keep only the narrower one.

## Reference behavior (Java)

Java preserves heading/display-text boundaries well enough that these read more like:

- `Cellular Cycle and Replication`
- `# Growth and the Creation of Life`

## Current Go behavior

Checked-in benchmark predictions still contain joined heading/display text even after the recent body-prose fixes. This suggests a remaining root cause distinct from the dense-prose packet.

## Relevant files

### Go
- `go/internal/processors/text_line_processor.go`
- `go/internal/processors/paragraph_processor.go`
- `go/internal/processors/document_processor.go`
- `go/pkg/pdfbox/extractor/text_extractor.go`
- focused tests under `go/tests/unit/` or `go/internal/processors/`

### Reports
- `reports/loops/2026-03-28-gap-text-benchmark-word-joins.md`
- `reports/loops/2026-03-28-gap-text-benchmark-prose-spacing.md`
- nearby text reports from the same date under `reports/loops/`

### Evidence files
- `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md`
- `tests/benchmark/ground-truth/markdown/01030000000118.md`
- source PDF(s) if actually available in this checkout

## Required work

- inspect the remaining heading/display-text joins and trace where spaces are still being lost
- determine whether the issue is in extraction, line assembly, paragraph merging, or markdown serialization
- keep the fix scoped to one root cause family
- add at least one focused regression test if you keep a fix
- validate with a reproducible local check against the targeted heading/display-text join
- if the benchmark PDFs are still LFS pointers, construct the tightest benchmark-shaped regression you can justify from the code path and explicitly document the limitation
- if the remaining symptom is actually broader layout/reading-order behavior rather than text spacing, split instead of broadening scope

## Acceptance criteria

1. At least one targeted heading/display-text benchmark join is corrected by a reproducible local check or a tightly justified benchmark-shaped regression.
2. Targeted tests/build pass.
3. A loop report records KEEP / DISCARD / SPLIT with evidence and any LFS limitations.
4. If root cause differs from expectations, document and split instead of broadening scope.

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors ./internal/processors ./tests/unit/pdfbox -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
rg -n "CellularCycle andReplication|theCreationofLife|Growth and theCreationofLife" tests/benchmark/prediction/opendataloader/markdown/*.md || true
file tests/benchmark/pdfs/01030000000118.pdf
```

Add any narrower reproduction command you discover for the exact source PDF / fixture.

## Non-goals

- reading-order changes
- ordinary body-prose spacing unless you formally split back to that narrower gap
- broad normalization passes unrelated to the chosen root cause
- unrelated encoding or ligature work unless directly required by the identified root cause

## Output format

Report:
1. where the space was being lost
2. files changed
3. before/after evidence
4. tests/commands run
5. KEEP / DISCARD / SPLIT

When completely finished, run this command to notify me:
openclaw system event --text "Done: evaluated GAP-TEXT-BENCHMARK-HEADING-SPACING in opendataloader-pdf-go" --mode now
