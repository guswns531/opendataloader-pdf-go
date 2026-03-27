# GAP-TEXT-WINANSI

## Category

text-parity

## Why this gap exists

Many PDFs contain 8-bit strings encoded with WinAnsi / Windows-1252 semantics. If Go only handles UTF-16 BOM cases and raw ASCII, extended characters can decode incorrectly.

This is one of the encoding-related sub-gaps under `T16`.

---

## Reference behavior (Java)

The Java reference implementation decodes common WinAnsi text into correct Unicode characters for downstream extraction.

---

## Current Go behavior

Go may mishandle bytes in the `0x80-0xFF` range when no UTF-16 BOM is present, producing broken punctuation or incorrect characters.

Typical affected characters include:

- smart quotes
- en/em dashes
- bullet
- euro sign
- ligature-related symbols and accented characters

---

## Impact

- broken human-readable text
- lower text parity with Java
- reduced usefulness for markdown/json outputs and downstream NLP

---

## Candidate source files

### Go
- `go/pkg/pdfbox/extractor/content_parser.go`
- related string normalization helpers

### Java reference
- corresponding PDF string decoding logic

---

## Fixtures / Reproduction

Use PDFs or synthetic tests containing Windows-1252 bytes in extracted strings.

At minimum, include cases covering:

- `0x80-0x9F` mapped characters
- ordinary ASCII passthrough
- `0xA0-0xFF` Latin-1 aligned range

---

## Acceptance criteria

1. Extended WinAnsi characters decode to correct Unicode output.
2. ASCII-only strings remain unchanged.
3. BOM-based UTF-16 decoding still works.
4. Targeted unit tests pass.
5. `go build ./...` succeeds.

---

## Evaluation commands

```bash
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...
cd go && go build ./...
```

If possible, compare targeted string outputs directly against Java baseline behavior.

---

## Non-goals

Out of scope here:

- ToUnicode cmap behavior beyond this narrow decoding path
- TJ spacing
- TrimSpace issues
- reading order fixes

---

## Keep / Discard guidance

### KEEP

- WinAnsi decoding improves while ASCII and UTF-16 behavior remain stable

### DISCARD

- fix corrupts ASCII or BOM-decoded strings
- behavior is inconsistent across targeted cases
