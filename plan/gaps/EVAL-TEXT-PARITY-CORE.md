# EVAL-TEXT-PARITY-CORE

## Category

evaluator

## Purpose

Provide the core evaluation checklist for text-parity loops in the Go port.

This evaluator is intended to judge candidate patches related to the `T16` family, including:

- ToUnicode
- TJ spacing
- TrimSpace boundary handling
- WinAnsi decoding

---

## Reference

The Java implementation is the primary behavioral reference.

---

## Core questions

1. Does the candidate move extracted text closer to Java output?
2. Does it improve known fixtures without breaking ordinary text extraction?
3. Are new or strengthened regression tests present where appropriate?
4. Is the fix scoped to the gap, rather than mixing multiple unrelated changes?
5. Is the result good enough to KEEP, or should the issue be split further?

---

## Minimum evidence

A text-parity candidate should ideally provide:

- targeted test results
- at least one fixture comparison or reproduction example
- explicit note of remaining uncertainty

---

## Suggested checks

```bash
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...
cd go && go build ./...
```

When possible, also compare Java vs Go outputs for affected sample PDFs.

---

## KEEP when

- target fixture output is measurably closer to Java
- tests pass
- no obvious regressions appear in nearby text extraction paths

## DISCARD when

- tests fail
- fix does not improve parity signal
- change introduces broad whitespace/encoding breakage

## SPLIT when

- one candidate mixes encoding, spacing, and glyph-mapping issues
- evidence suggests multiple root causes remain entangled

## NEEDS-HUMAN-REVIEW when

- Java behavior itself is ambiguous
- fixture evidence conflicts across document types
