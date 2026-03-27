# Codex Gap Task Packet — GAP-TEXT-BENCHMARK-PROSE-SPACING

## Role

You are implementing one narrow Java→Go parity gap in the Go port of OpenDataLoader PDF. Keep scope tight.

## Gap ID

`GAP-TEXT-BENCHMARK-PROSE-SPACING`

## Category

`text-parity`

## Objective

Continue from the freshly kept `GAP-TEXT-BENCHMARK-WORD-JOINS` change and target the next obvious dense-prose joined-word symptom that still appears in benchmark outputs.

Primary live symptom to investigate:

- `Oneofthecharacteristicsoflivingthingsistheability toreplicateandpasson genetic information...`

Secondary symptom (only if it proves to share the same root cause):

- nearby prose-style joins from the same benchmark page where ordinary body text boundaries are still collapsed

If the remaining issue is actually TOC / heading specific rather than prose, split it instead of broadening this packet.

## Reference behavior (Java)

Java preserves ordinary prose word boundaries well enough that body text does not collapse into joined phrases like:

- `One of the characteristics of living things is the ability`
- `to replicate and pass on genetic information`

## Current Go behavior

Checked-in benchmark predictions still contain body-text joins such as:

- `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md:5` -> `Oneofthecharacteristicsoflivingthingsistheability toreplicateandpasson genetic information to the next`

The immediately previous loop already reduced one heading-shaped dense-gap family by lowering the synthetic word-boundary threshold. This packet should only pursue the next residual root cause if it is distinct and defensible.

## Relevant files

### Go
- `go/internal/processors/text_line_processor.go`
- `go/internal/processors/paragraph_processor.go`
- `go/internal/processors/document_processor.go`
- `go/pkg/pdfbox/extractor/text_extractor.go`
- focused tests under `go/tests/unit/` or `go/internal/processors/`

### Reports
- `reports/loops/2026-03-28-gap-text-benchmark-word-joins.md`
- nearby text reports from the same date under `reports/loops/`

### Evidence files
- `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md`
- matching ground-truth file under `tests/benchmark/ground-truth/opendataloader/markdown/`
- source PDF(s) if actually available in this checkout

## Required work

- inspect the remaining benchmark prose-join symptom and trace where spaces are still being lost
- determine whether the issue is in extraction, line assembly, paragraph merging, or markdown serialization
- keep the fix scoped to one root cause family
- add at least one focused regression test if you keep a fix
- validate with a reproducible local check against the targeted prose-style join
- if the remaining symptom is not reproducible because benchmark PDFs are LFS pointers, construct the tightest benchmark-shaped regression you can justify from the code path and explicitly document the limitation
- if the residual issue is actually TOC/heading-specific or would require a broader heuristic, split instead of improvising a broad patch

## Acceptance criteria

1. At least one targeted prose-style benchmark join is corrected by a reproducible local check or a tightly justified benchmark-shaped regression.
2. Targeted tests/build pass.
3. A loop report records KEEP / DISCARD / SPLIT with evidence and any LFS limitations.
4. If root cause differs from expectations, document and split instead of broadening scope.

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors ./internal/processors ./tests/unit/pdfbox -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
rg -n "Oneofthecharacteristics|toreplicateandpasson" tests/benchmark/prediction/opendataloader/markdown/*.md || true
file tests/benchmark/pdfs/01030000000118.pdf
```

Add any narrower reproduction command you discover for the exact source PDF / fixture.

## Non-goals

- reading-order changes
- TOC-only / heading-only spacing unless you formally split to that narrower gap
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
openclaw system event --text "Done: evaluated GAP-TEXT-BENCHMARK-PROSE-SPACING in opendataloader-pdf-go" --mode now
