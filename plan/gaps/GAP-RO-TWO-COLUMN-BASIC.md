# GAP-RO-TWO-COLUMN-BASIC

## Category

reading-order-parity

## Why this gap exists

Academic and report PDFs often use two-column layouts. A naïve Y-based ordering can interleave left and right columns line-by-line instead of reading the full left column before the right column.

This is the core sub-gap inside `T17`.

---

## Reference behavior (Java)

The Java reference implementation should preserve a more human-correct reading order for standard two-column pages, typically yielding left-column content before right-column content where appropriate.

---

## Current Go behavior

Go may mix columns based primarily on Y ordering, producing output like:

- left row 1
- right row 1
- left row 2
- right row 2

instead of:

- left column top-to-bottom
- right column top-to-bottom

---

## Impact

- severe markdown/text readability degradation
- lower NID / reading-order quality
- worse downstream chunking and retrieval behavior

---

## Candidate source files

### Go
- `go/internal/processors/xycut_plus_plus_sorter.go`
- related text chunk / page entities used during ordering

### Java reference
- corresponding Java reading-order / XYCut++ logic

---

## Fixtures / Reproduction

Use a known two-column paper, including:

- `samples/pdf/1901.03003.pdf`

Capture:

- Java markdown/text output
- Go markdown/text output before change
- expected section ordering around abstract / introduction or comparable regions

---

## Acceptance criteria

1. Two-column fixtures no longer interleave columns line-by-line.
2. Output is measurably closer to Java reference behavior.
3. Targeted XYCut / reading-order tests pass.
4. Single-column ordering does not regress.
5. `go build ./...` and relevant tests succeed.

---

## Evaluation commands

```bash
cd go && go test ./internal/processors/... -run TestXYCut -v
cd go && go build ./...
```

Then compare markdown/text output on targeted two-column fixtures.

---

## Non-goals

Out of scope here:

- sidebars and mixed-layout special cases
- deep recursion stability
- broader text extraction bugs
- table ordering unless directly required by two-column behavior

---

## Keep / Discard guidance

### KEEP

- clear improvement on standard two-column fixtures
- no obvious regression for single-column pages

### DISCARD

- columns still interleave materially
- single-column behavior regresses

### SPLIT if needed

If multiple causes emerge, split into:

- `GAP-RO-COLUMN-SPLIT-DETECTION`
- `GAP-RO-COLUMN-MERGE-ORDER`
