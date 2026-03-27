# Codex Gap Task Packet — GAP-TEXT-TRIMSPACE

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-TEXT-TRIMSPACE`

## Category

`text-parity`

---

## Objective

Implement the smallest change that improves Go parity when boundary spaces are lost because text runs are trimmed too aggressively during extraction or early normalization.

---

## Reference behavior (Java)

The Java implementation preserves meaningful boundary spaces across adjacent text runs well enough that intended word separation survives in extracted text.

---

## Current Go behavior

The Go implementation may remove meaningful leading/trailing whitespace too early in the extraction path, causing adjacent runs that should remain separated to collapse.

This gap should focus on **trim-related boundary loss**, not TJ numeric spacing or encoding work.

---

## Why this matters

- degrades plain text and markdown readability
- reduces Java parity in text output
- hurts downstream chunking and search quality

---

## Relevant files

### Go
- `go/pkg/pdfbox/extractor/text_extractor.go`
- any nearby helpers directly involved in early text normalization / append behavior

### Java reference
- corresponding Java text-run accumulation behavior

### Tests / fixtures
- existing tests under `go/pkg/pdfbox/extractor/` and `go/tests/unit/pdfbox/`
- add a focused regression test if needed

---

## Required work

- inspect whether text extraction still trims meaningful spaces too early
- make the minimum focused change for trim-related boundary preservation
- add or strengthen regression coverage
- do not mix this patch with TJ offset or WinAnsi fixes

---

## Acceptance criteria

1. A focused adjacent-run spacing case preserves intended word boundaries.
2. Output moves closer to Java behavior for the trim-related case.
3. Targeted PDFBox/text extraction tests pass.
4. No obvious noisy-whitespace regression is introduced.
5. `go build ./...` succeeds.

---

## Validation commands

Run these before finishing:

```bash
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...
cd go && go build ./...
```

---

## Non-goals

Do **not** attempt these in the same patch:

- TJ numeric offset spacing work
- WinAnsi decoding
- broader ToUnicode changes
- reading-order / XYCut changes

---

## Output format

When done, report:

1. files changed
2. behavior changed
3. tests/commands run and their result
4. remaining uncertainty
5. whether the gap appears fully fixed, partially fixed, or needs follow-up
