# GAP-TEXT-TJ-OFFSET

## Category

text-parity

## Why this gap exists

In PDF `TJ` operators, numeric items encode glyph spacing adjustments. Large negative offsets can visually represent a word break. If the Go implementation ignores those offsets, words that should be separated may collapse together.

This is one of the concrete sub-gaps previously grouped under `T16`.

---

## Reference behavior (Java)

The Java reference implementation preserves intended spacing well enough that words represented via TJ kerning offsets are emitted with correct boundaries in markdown/text output.

---

## Current Go behavior

Go may ignore numeric items inside `TJ` arrays, causing visible word boundaries to disappear.

Example symptom:

- expected: `A Multi-Object Rectified Attention Network`
- actual: `AMulti-ObjectRectifiedAttentionNetwork`

---

## Impact

- degraded text extraction quality
- lower user trust in plain text / markdown output
- potential downstream RAG chunking quality loss
- likely related to text parity regressions in benchmark fixtures

---

## Candidate source files

### Go
- `go/pkg/pdfbox/extractor/content_parser.go`
- `go/pkg/pdfbox/extractor/text_extractor.go`

### Java reference
- corresponding Java PDFBox text extraction logic for TJ handling

---

## Fixtures / Reproduction

Use one or more PDFs where kerning offsets represent word boundaries, including the known sample:

- `samples/pdf/1901.03003.pdf`

Capture both:

- Java output
- Go output before change

---

## Acceptance criteria

A candidate fix is acceptable only if all apply:

1. Known TJ-spacing fixture no longer collapses affected words.
2. Output moves closer to Java baseline for the targeted sample(s).
3. Targeted text extraction tests pass.
4. No obvious regression appears in ordinary ASCII text extraction.
5. `go build ./...` succeeds.

---

## Evaluation commands

Prefer the lightest useful checks first.

```bash
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...
cd go && go build ./...
```

Then run a targeted fixture comparison against the Java baseline or previous output.

---

## Non-goals

This gap does **not** aim to solve all text extraction parity issues.

Out of scope here:

- WinAnsi decoding
- TrimSpace boundary loss
- broader ToUnicode behavior
- reading order / XYCut issues

---

## Keep / Discard guidance

### KEEP

- spacing improves on known TJ fixtures
- no text extraction regressions are introduced elsewhere

### DISCARD

- spacing remains wrong
- fix is overly heuristic and breaks ordinary text extraction

### SPLIT if needed

If the issue turns out to involve multiple causes, split into:

- `GAP-TEXT-TJ-NUMERIC-SPACING`
- `GAP-TEXT-GLYPH-JOIN-BEHAVIOR`
