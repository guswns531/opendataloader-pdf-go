# GAP-TEXT-TOUNICODE

## Category

text-parity

## Why this gap exists

PDF text extraction quality depends heavily on correct ToUnicode mapping. Even when raw bytes or font encodings are ambiguous, ToUnicode CMaps often provide the intended Unicode text. If the Go port underuses or misapplies this mapping, extracted text can diverge from Java behavior.

This is a focused text-parity sub-gap in the broader `T16` family.

---

## Reference behavior (Java)

The Java reference implementation decodes ToUnicode-mapped text runs so that extracted text reflects intended Unicode output for downstream markdown/text/json generation.

---

## Current Go behavior

Go may partially support ToUnicode but still diverge on edge cases such as:

- mapped glyph sequences not being decoded consistently
- fallback paths overriding ToUnicode-derived text
- mixed encoded streams producing incomplete Unicode output

Related history suggests some ToUnicode work has already landed, so this gap is primarily about parity verification and residual mismatches.

---

## Impact

- broken or incomplete Unicode output
- divergence from Java text extraction behavior
- reduced text quality for multilingual PDFs and non-ASCII content
- potential benchmark and fixture mismatches

---

## Candidate source files

### Go
- `go/pkg/pdfbox/extractor/content_parser.go`
- `go/pkg/pdfbox/extractor/text_extractor.go`
- related PDF string / glyph decoding helpers

### Java reference
- corresponding Java PDFBox text extraction and ToUnicode mapping logic

---

## Fixtures / Reproduction

Use PDFs and/or unit cases where ToUnicode mapping materially affects text output.

Prefer fixtures that include:

- non-ASCII glyphs
- ligatures or special symbols
- multilingual text where ToUnicode is required for correct extraction

Compare Java vs Go outputs directly.

---

## Acceptance criteria

1. Targeted ToUnicode fixtures move closer to Java output.
2. Existing ToUnicode regression tests pass.
3. Ordinary ASCII extraction is not regressed.
4. Related text extraction tests pass.
5. `go build ./...` succeeds.

---

## Evaluation commands

```bash
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...
cd go && go build ./...
```

If available, add or run direct fixture comparisons against Java output for known ToUnicode-sensitive samples.

---

## Non-goals

Out of scope here:

- TJ spacing logic
- TrimSpace boundary handling
- WinAnsi decoding not directly tied to ToUnicode use
- reading-order / XYCut behavior

---

## Keep / Discard guidance

### KEEP

- ToUnicode-sensitive output clearly improves
- targeted regressions remain green

### DISCARD

- behavior remains ambiguous
- changes reduce correctness on existing covered cases
