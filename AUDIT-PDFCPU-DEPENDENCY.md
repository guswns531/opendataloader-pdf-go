# AUDIT-PDFCPU-DEPENDENCY.md

## Purpose

This audit reframes the repository around **Plan A**:

> Stop treating `pdfcpu` as the long-term low-level engine for PDFBox parity, and instead move toward an **Apache PDFBox direct-port** architecture in Go.

The central concern is that repeated benchmark failures may not be isolated processor bugs. They may be downstream symptoms of a base-layer mismatch caused by using `pdfcpu` semantics where Java relies on PDFBox semantics.

---

## Current state

The current Go implementation uses `pkg/pdfbox/` as a PDFBox-like layer, but many files explicitly state that they are **implemented using pdfcpu**.

Observed examples include:

- `go/pkg/pdfbox/model/document.go`
- `go/pkg/pdfbox/loader/loader.go`
- `go/pkg/pdfbox/extractor/content_parser.go`
- `go/pkg/pdfbox/extractor/text_extractor.go`
- `go/pkg/pdfbox/extractor/image_extractor.go`
- `go/pkg/pdfbox/extractor/line_art_extractor.go`

This means the repository currently behaves as:

- Java: OpenDataLoader on top of Apache PDFBox
- Go: OpenDataLoader on top of a PDFBox-shaped adapter backed by pdfcpu

---

## Risk statement

If `pdfcpu` emits different low-level semantics than Apache PDFBox, the following classes of bugs can recur indefinitely even after processor patches:

1. text run segmentation mismatches
2. incorrect glyph decoding or fallback behavior
3. bounding-box / baseline / geometry drift
4. missing line-art or table cues
5. operator handling differences for content streams
6. chunk ordering or grouping differences before processors even run

When that happens, higher-level fixes in:

- text cleanup
- reading order
- heading detection
- table reconstruction

may repeatedly overfit to symptoms without solving the base mismatch.

---

## Why Plan A exists

Full benchmark results now suggest broad quality issues remain:

- NID: `0.7391`
- TEDS: `0.0399`
- MHS: `0.4012`
- Table Detection F1: `0.2373`

This is too broad to confidently attribute only to `T16/T17`-style surface gaps.

The project must seriously consider that:

- `T13` (pdfbox extractor)
- `T15` (real document processor path)
- and related base-layer assumptions

may not be complete enough for true Java parity.

---

## Plan A decision

### Strategic change

Treat `pdfcpu` as a **temporary bootstrap dependency**, not the target architecture.

### New architectural direction

- preserve `pkg/pdfbox/` as the Go-facing package boundary
- gradually remove `pdfcpu`-specific behavior from its core semantics
- replace adapter logic with **directly ported Apache PDFBox behavior** where benchmark impact is highest

This is not a one-shot rewrite; it is a deliberate migration from:

- `pdfcpu-backed PDFBox adapter`

to:

- `direct-port PDFBox semantics implemented in Go`

---

## Priority migration candidates

The first candidates for direct-port replacement should be the low-level pieces that most strongly affect benchmark parity.

### Tier 1 — immediate

1. `go/pkg/pdfbox/extractor/content_parser.go`
2. `go/pkg/pdfbox/extractor/font_decoder.go`
3. `go/pkg/pdfbox/extractor/text_extractor.go`

These govern:

- text-showing operators
- ToUnicode / CMap usage
- WinAnsi / fallback decoding
- text run segmentation
- baseline and geometry
- chunk boundaries feeding processors

### Tier 2 — next

4. `go/pkg/pdfbox/extractor/line_art_extractor.go`
5. `go/pkg/pdfbox/extractor/image_extractor.go`

These influence:

- table cues
- layout segmentation
- mixed-layout detection

### Tier 3 — later

6. `go/pkg/pdfbox/loader/loader.go`
7. `go/pkg/pdfbox/model/document.go`
8. annotation / OCG compatibility surfaces

---

## Milestones to reopen or downgrade in confidence

Under Plan A, these milestones should no longer be treated as unquestionably complete:

- `T13` — pdfbox extractor real implementation
- `T15` — document processor real implementation
- `T16` — text extraction quality
- `T17` — reading order quality

Reason:

Their current status may rest on a base-layer model that is still semantically too far from Apache PDFBox.

---

## Audit questions

Before each direct-port loop, answer:

1. Which Java/PDFBox class or behavior is the reference here?
2. What current Go behavior is still inherited from pdfcpu rather than from PDFBox?
3. Which benchmark failure pattern is plausibly caused by that mismatch?
4. Can the mismatch be replaced with a direct-port implementation without breaking package boundaries?
5. What regression tests and benchmark slices will validate the replacement?

---

## Success criteria for Plan A

Plan A is working if:

1. repeated benchmark symptoms stop reappearing under new names
2. text, heading, table, and reading-order parity improve together rather than in isolation
3. fewer fixes require heuristic cleanup in upper layers
4. direct-port behavior becomes the main explanation for successful parity gains

---

## Non-goals

Plan A does **not** mean:

- deleting `pkg/pdfbox/`
- rewriting the whole repository in one pass
- abandoning existing regression tests
- ignoring benchmark evidence while chasing theoretical purity

The goal is practical direct-port migration, not a purity rewrite detached from metrics.
