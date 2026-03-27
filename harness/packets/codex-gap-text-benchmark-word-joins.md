# Codex Gap Task Packet — GAP-TEXT-BENCHMARK-WORD-JOINS

## Role

You are implementing one narrow Java→Go parity gap in the Go port of OpenDataLoader PDF. Keep scope tight.

## Gap ID

`GAP-TEXT-BENCHMARK-WORD-JOINS`

## Category

`text-parity`

## Objective

Reduce the next obvious benchmark-backed joined-word text failures that remain after the recent `1901.03003.pdf` spacing fixes.

Primary live symptoms to investigate:

- `theCreationofLife`
- `Oneofthecharacteristics`
- any closely related missing-boundary joins from the same benchmark pages

Treat this as a fresh packet. Do not broaden into reading-order work.

## Reference behavior (Java)

Java preserves ordinary prose boundaries in these benchmark outputs well enough that chapter/title/body text does not collapse into joined words like:

- `theCreationofLife`
- `Oneofthecharacteristics`

## Current Go behavior

Current benchmark prediction files still contain:

- `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md:3` → `# Growth and theCreationofLife`
- `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md:5` → `Oneofthecharacteristicsoflivingthingsistheability ...`
- `tests/benchmark/prediction/opendataloader/markdown/01030000000113.md:37` → TOC-style joins around `GrowthandtheCreationofLife`

This may be a different root cause from the singleton-fragment fix. Confirm where the space is being lost before patching.

## Relevant files

### Go
- `go/internal/processors/text_line_processor.go`
- `go/internal/processors/paragraph_processor.go`
- `go/internal/processors/document_processor.go`
- `go/pkg/pdfbox/extractor/text_extractor.go`
- focused tests under `go/tests/unit/` or `go/internal/processors/`

### Reports
- `reports/loops/2026-03-28-gap-text-inline-boundary-space.md`
- `reports/loops/2026-03-28-gap-text-intraword-spacing.md`
- `reports/loops/2026-03-28-gap-text-singleton-continuation-boundary.md`

### Evidence files
- `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md`
- `tests/benchmark/prediction/opendataloader/markdown/01030000000113.md`
- matching ground-truth files under `tests/benchmark/ground-truth/opendataloader/markdown/`
- source PDF(s) if discoverable from the repo

## Required work

- inspect the failing benchmark markdown and trace where the missing spaces are lost
- determine whether the issue is in extraction, line assembly, paragraph merging, or serializer logic
- keep the fix scoped to one root cause family
- add at least one focused regression test if you keep a fix
- validate with live reproduction against the targeted benchmark symptom(s)
- if the benchmark joins turn out to be TOC-/heading-specific and distinct from body text, split rather than bundling both

## Acceptance criteria

1. At least one targeted benchmark join is corrected by a reproducible local check.
2. Targeted tests/build pass.
3. A loop report records KEEP / DISCARD / SPLIT with evidence.
4. If root cause differs from expectations, document and split instead of improvising a broad heuristic.

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
rg -n "theCreationofLife|Oneofthecharacteristics|GrowthandtheCreationofLife" tests/benchmark/prediction/opendataloader/markdown/*.md || true
```

Add any narrower reproduction command you discover for the exact source PDF / fixture.

## Non-goals

- reading-order changes
- broad benchmark cleanup unrelated to missing prose boundaries
- large text normalization passes
- unrelated encoding or ligature work unless directly required by the chosen root cause

## Output format

Report:
1. where the space was being lost
2. files changed
3. before/after evidence
4. tests/commands run
5. KEEP / DISCARD / SPLIT

When completely finished, run this command to notify me:
openclaw system event --text "Done: evaluated GAP-TEXT-BENCHMARK-WORD-JOINS in opendataloader-pdf-go" --mode now
