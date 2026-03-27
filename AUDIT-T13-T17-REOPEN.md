# AUDIT-T13-T17-REOPEN.md

## Purpose

This audit records why milestone confidence from `T13` through `T17` must be reconsidered under Plan A.

Plan A assumption:

> If the base PDF extraction layer is still semantically closer to `pdfcpu` than to Apache PDFBox, then downstream milestone completion claims may be overstated.

---

## T13 — pdfbox extractor real implementation

### Prior claim

`T13` was treated as complete because `pkg/pdfbox/` existed and the extractor path was wired up.

### Why confidence is now lower

- multiple files explicitly state they are implemented using `pdfcpu`
- recent text-parity fixes repeatedly landed in `pkg/pdfbox/extractor/*`
- benchmark failures suggest chunk / text / geometry semantics still diverge substantially from Java behavior

### New interpretation

`T13` should be reinterpreted as:

> initial usable extractor implementation, not necessarily PDFBox-semantic parity

### Reopen questions

- which content-stream semantics still differ from PDFBox?
- how much extractor geometry is shaped by pdfcpu limitations?
- are table / heading / reading-order failures already baked into chunk output before processors run?

---

## T14 — PDFWriter (annotations / OCG)

### Current view

No immediate evidence says T14 is the primary benchmark bottleneck.

### Caveat

T14 can remain tentatively complete, but should not be used as evidence that the broader PDFBox layer is mature.

---

## T15 — DocumentProcessor stub → real implementation

### Prior claim

`DocumentProcessor` was switched from stub behavior to the real extraction path.

### Why confidence is now lower

If the underlying extractor semantics are still mismatched, then “real implementation” may only mean:

- real wiring exists
- not that the pipeline is receiving PDFBox-equivalent inputs

### New interpretation

`T15` should be reinterpreted as:

> pipeline connected to the current extractor, pending validation against direct-port semantics

---

## T16 — text extraction quality

### Prior claim

T16 was the remaining text-quality task family.

### What we learned

Substantial useful work did land:

- TJ spacing
- boundary whitespace handling
- WinAnsi improvements
- ToUnicode multiline array support
- several benchmark-driven spacing fixes

### Why confidence is still limited

Many of these fixes operate on symptoms after extraction logic has already diverged.
If the base extractor differs from PDFBox behavior, T16 may keep reappearing as many small text gaps.

### New interpretation

`T16` should remain active and partially complete, but also be treated as a signal that `T13` may not be semantically done.

---

## T17 — reading order quality

### Prior claim

T17 covered multi-column reading-order quality.

### What we learned

Useful reading-order work landed:

- basic two-column split tightening
- balanced column merge handling
- sidebar deferral
- XYCut recursion capping
- benchmark-driven reading-order follow-ups

### Why confidence is still limited

Reading order depends on upstream chunk geometry, segmentation, and layout cues.
If those are already wrong, processor-only fixes can only go so far.

### New interpretation

`T17` remains active, but should also be partially reframed as a downstream consumer of unresolved `T13`-class semantics.

---

## Revised confidence summary

| Task | Old status | Revised confidence |
|------|------------|-------------------|
| T13 | complete | **reopen / low confidence** |
| T14 | complete | **tentatively keep** |
| T15 | complete | **reopen / medium-low confidence** |
| T16 | pending/active | **active, but partly rooted below itself** |
| T17 | pending/active | **active, but partly rooted below itself** |

---

## Practical consequence

From this point forward, benchmark failures should not automatically be assigned only to `T16/T17`.

Instead, each new failure should first be classified as possibly belonging to:

- base extraction semantics (`T13` reopened)
- pipeline validity (`T15` reopened)
- text quality cleanup (`T16`)
- reading-order cleanup (`T17`)

---

## Recommended next work

1. create direct-port packets for Tier-1 extractor code
2. classify benchmark failures against `T13/T15/T16/T17`
3. stop treating processor-layer fixes as sufficient evidence that the base extractor is done
