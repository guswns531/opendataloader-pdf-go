# GAP-TEXT-TRIMSPACE

## Category

text-parity

## Why this gap exists

If the Go extractor trims leading/trailing whitespace too aggressively at append time, it may destroy intentional word boundaries between adjacent text runs.

This is a focused sub-gap of `T16`.

---

## Reference behavior (Java)

The Java implementation preserves text boundaries across adjacent runs without removing meaningful spaces that should survive in final text output.

---

## Current Go behavior

Go may call `strings.TrimSpace` during text append or normalization in a way that removes intended spaces at the boundaries of text runs.

This can turn:

- `Hello World`

into concatenated output when the source run segmentation includes trailing or leading spaces.

---

## Impact

- merged words in text/markdown output
- false degradation in readability
- downstream chunking and alignment quality loss

---

## Candidate source files

### Go
- `go/pkg/pdfbox/extractor/text_extractor.go`

### Java reference
- corresponding Java text-run accumulation logic

---

## Fixtures / Reproduction

Use PDFs or targeted unit tests where adjacent text runs intentionally rely on boundary spaces.

Compare Java vs Go on:

- extracted text
- markdown output

---

## Acceptance criteria

1. Boundary spaces required for word separation are preserved.
2. Empty or meaningless text runs are still safely ignored.
3. Output moves closer to Java baseline for targeted fixtures.
4. Text extraction tests pass.
5. `go build ./...` succeeds.

---

## Evaluation commands

```bash
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...
cd go && go build ./...
```

Add or run a focused regression test for adjacent text-run spacing if available.

---

## Non-goals

Out of scope here:

- TJ numeric spacing logic
- WinAnsi decoding
- ToUnicode mapping issues
- reading order fixes

---

## Keep / Discard guidance

### KEEP

- targeted spacing loss is corrected
- empty-run handling remains safe

### DISCARD

- fix introduces noisy whitespace everywhere
- output diverges further from Java behavior
