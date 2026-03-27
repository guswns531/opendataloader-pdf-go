# Codex Gap Task Packet — GAP-TEXT-TJ-OFFSET

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-TEXT-TJ-OFFSET`

## Category

`text-parity`

---

## Objective

Implement the smallest change that improves Go parity for word-boundary recovery when PDF `TJ` arrays encode spacing through numeric kerning offsets.

---

## Reference behavior (Java)

The Java implementation preserves intended word separation better on PDFs where `TJ` numeric offsets visually represent spaces between words.

---

## Current Go behavior

The Go implementation appears to ignore numeric items inside `TJ` arrays, which can collapse words together.

Known symptom from a sample paper:

- expected: `A Multi-Object Rectified Attention Network`
- actual: `AMulti-ObjectRectifiedAttentionNetwork`

---

## Why this matters

- directly degrades text extraction quality
- harms markdown readability
- reduces parity with Java
- can hurt downstream RAG/chunking quality

---

## Relevant files

### Go
- `go/pkg/pdfbox/extractor/content_parser.go`
- `go/pkg/pdfbox/extractor/text_extractor.go`

### Java reference
- Java PDFBox text extraction logic corresponding to TJ handling and text-run accumulation

### Tests / fixtures
- `samples/pdf/1901.03003.pdf`
- any existing text extraction tests under `go/tests/unit/pdfbox/` or related locations

---

## Required work

- inspect how `TJ` array items are currently decoded in Go
- implement only the minimum spacing recovery needed for this gap
- add or tighten focused regression coverage if a clean test hook exists
- keep the patch focused; do not combine unrelated text-decoding fixes

---

## Acceptance criteria

1. The targeted TJ-spacing symptom improves on the known sample or an equivalent focused test.
2. Output moves closer to Java behavior for the affected case.
3. Targeted PDFBox/text extraction tests pass.
4. No obvious regression appears in ordinary text extraction.
5. `go build ./...` succeeds.

---

## Validation commands

Run these before finishing:

```bash
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...
cd go && go build ./...
```

If practical, also run a targeted sample conversion on `samples/pdf/1901.03003.pdf` and report the before/after snippet near the collapsed title text.

---

## Non-goals

Do **not** attempt these in the same patch:

- WinAnsi decoding
- TrimSpace cleanup changes
- broad ToUnicode work
- reading-order / XYCut changes

---

## Output format

When done, report:

1. files changed
2. behavior changed
3. tests/commands run and their result
4. remaining uncertainty
5. whether the gap appears fully fixed, partially fixed, or needs follow-up
