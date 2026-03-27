# Codex Gap Task Packet — GAP-TEXT-TOUNICODE

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-TEXT-TOUNICODE`

## Category

`text-parity`

---

## Objective

Implement the smallest change that improves Go parity for ToUnicode-driven text decoding.

---

## Reference behavior (Java)

The Java implementation decodes ToUnicode-mapped text runs into intended Unicode text for downstream extraction.

---

## Current Go behavior

The Go implementation may partially support ToUnicode but still diverge on edge cases where mapped Unicode output should override or enrich raw byte decoding.

This gap should stay focused on ToUnicode behavior, not TJ spacing, boundary trim handling, or reading order.

---

## Why this matters

- Unicode output quality depends on correct ToUnicode use
- Java parity suffers for non-ASCII or mapped glyph cases
- markdown/json/text outputs degrade when ToUnicode is incomplete

---

## Relevant files

### Go
- `go/pkg/pdfbox/extractor/font_decoder.go`
- `go/pkg/pdfbox/extractor/content_parser.go`
- related CMap parsing / decode helpers

### Java reference
- corresponding Java ToUnicode / font-decoding path

### Tests / fixtures
- existing extractor tests
- add focused regression coverage for ToUnicode decoding if needed

---

## Required work

- inspect current ToUnicode decode path and where it diverges from expected Unicode output
- patch only the narrow ToUnicode gap
- add focused regression tests
- avoid mixing in WinAnsi/TJ/TrimSpace changes unless strictly necessary for this gap

---

## Acceptance criteria

1. Targeted ToUnicode-mapped text decodes to expected Unicode output.
2. Existing extractor behavior does not regress.
3. Related tests pass.
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

- TJ spacing changes
- boundary-space / TrimSpace fixes
- broad reading-order / XYCut work
- unrelated serializer changes

---

## Output format

When done, report:

1. files changed
2. behavior changed
3. tests/commands run and their result
4. remaining uncertainty
5. whether the gap appears fully fixed, partially fixed, or needs follow-up
