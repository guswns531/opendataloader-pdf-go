# Codex Direct-Port Packet — PORT-PDFBOX-TEXT-SHOWING-OPERATORS

## Role

You are implementing **one direct-port migration step** in the Go port of OpenDataLoader PDF.

This packet follows **Plan A**:

- stop relying on `pdfcpu`-shaped semantics where they differ from Apache PDFBox
- move the Go extractor toward direct Apache PDFBox behavior

The Java implementation is the primary reference.

Do not broaden scope beyond this packet.

---

## Port Unit ID

`PORT-PDFBOX-TEXT-SHOWING-OPERATORS`

## Category

`direct-port / pdfbox / extractor`

---

## Objective

Audit and tighten the handling of text-showing operators so the Go extraction path behaves more like Apache PDFBox and less like an adapter over `pdfcpu` assumptions.

Focus on the operator semantics that most strongly affect:

- text run segmentation
- spacing recovery
- chunk boundaries
- geometry fed into higher-level processors

---

## Why this matters

Recent benchmark and parity work suggests that repeated text, heading, table, and reading-order failures may stem from low-level extractor semantics, not only from processor heuristics.

If text-showing operators are still semantically off, upper-layer fixes will keep chasing symptoms.

---

## Reference behavior (Java)

Use the Java OpenDataLoader + Apache PDFBox extraction path as the primary reference for how text-showing operators should influence:

- emitted text
- run boundaries
- spacing behavior
- text matrix progression
- geometry/baseline tracking

---

## Current Go suspicion

The current Go code in `pkg/pdfbox/extractor/` works, but some behavior may still be shaped more by `pdfcpu`-compatible extraction shortcuts than by direct PDFBox semantics.

This packet is about identifying and correcting one narrow mismatch in the text-showing path if found.

---

## Relevant files

### Go
- `go/pkg/pdfbox/extractor/content_parser.go`
- `go/pkg/pdfbox/extractor/text_extractor.go`
- `go/pkg/pdfbox/extractor/font_decoder.go`

### Java reference
- Java PDFBox/OpenDataLoader text extraction path for text-showing operators

### Existing evidence
- recent text parity fixes in extractor tests
- benchmark-driven joined-word and boundary-space regressions
- `tests/benchmark/prediction/opendataloader/evaluation.*`

---

## Required work

1. inspect how the Go extractor currently handles text-showing operators (`Tj`, `TJ`, quotes variants if relevant)
2. compare this with the Java/PDFBox reference path
3. identify one concrete mismatch that is plausibly still harming parity
4. implement the **smallest direct-port-oriented correction**
5. add or tighten focused regression tests

---

## Acceptance criteria

1. The patch is justified as moving semantics toward PDFBox behavior, not adding an arbitrary heuristic.
2. Targeted extractor tests pass.
3. `go build ./...` succeeds.
4. The final report clearly states what PDFBox-like behavior was tightened.

---

## Validation commands

```bash
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...
cd go && go build ./...
```

If practical, cite one benchmark-shaped symptom that this port step plausibly helps.

---

## Non-goals

Do **not** in this packet:

- rewrite the entire extractor
- fix table processors directly
- fix heading processors directly
- refactor unrelated `internal/processors/*` code
- perform broad benchmark-only heuristics without a PDFBox semantic justification

---

## Output format

When done, report:

1. files changed
2. what PDFBox-like operator behavior was tightened
3. tests/commands run and their result
4. remaining uncertainty
5. whether this appears to be a KEEP, DISCARD, or SPLIT candidate
