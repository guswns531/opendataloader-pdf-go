# Codex Gap Task Packet — GAP-TEXT-WINANSI

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-TEXT-WINANSI`

## Category

`text-parity`

---

## Objective

Implement the smallest change that improves Go parity for WinAnsi / Windows-1252 decoding in extracted text.

---

## Reference behavior (Java)

The Java implementation decodes common WinAnsi bytes into correct Unicode characters for extracted text.

---

## Current Go behavior

The Go implementation may still diverge on WinAnsi edge cases, especially in the `0x80-0x9F` range and other non-ASCII byte paths.

This gap should stay focused on WinAnsi decoding parity, not TJ spacing or reading order.

---

## Why this matters

- human-readable punctuation and symbols break if WinAnsi decoding is incomplete
- Java parity suffers on ordinary PDFs using 8-bit text
- markdown/json outputs degrade for downstream use

---

## Relevant files

### Go
- `go/pkg/pdfbox/extractor/content_parser.go`
- `go/pkg/pdfbox/extractor/font_decoder.go`
- nearby PDF string decoding helpers

### Java reference
- corresponding Java decoding path for PDF strings / font-decoded text

### Tests / fixtures
- existing extractor tests
- add focused regression coverage for WinAnsi byte mappings if needed

---

## Required work

- inspect the current WinAnsi decoding table and fallback paths
- patch only the narrow decoding gap
- add focused regression coverage
- avoid mixing in TJ / TrimSpace / ToUnicode work unless strictly necessary for this gap

---

## Acceptance criteria

1. Targeted WinAnsi characters decode to the expected Unicode output.
2. ASCII behavior remains unchanged.
3. Existing related extractor tests pass.
4. `go build ./...` succeeds.

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

- TJ offset spacing changes
- TrimSpace/boundary-space handling
- broad ToUnicode refactors
- reading-order / XYCut changes

---

## Output format

When done, report:

1. files changed
2. behavior changed
3. tests/commands run and their result
4. remaining uncertainty
5. whether the gap appears fully fixed, partially fixed, or needs follow-up
