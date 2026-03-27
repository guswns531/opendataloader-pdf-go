# GAP-RO-BALANCED-COLUMNS

## Category

reading-order-parity

## Why this gap exists

Some two-column documents are not perfectly symmetric: headings, figures, abstract blocks, or uneven text lengths can produce layouts where a simple column split still yields imperfect ordering. Balanced-column handling is a refinement gap beyond basic two-column detection.

This is one of the follow-up reading-order gaps under `T17`.

---

## Reference behavior (Java)

The Java reference implementation appears to contain additional balanced-column reading-order work beyond basic XYCut column splitting.

---

## Current Go behavior

Go may improve on basic column separation yet still mis-sequence content when:

- one column starts or ends earlier than the other
- block heights differ significantly
- headings or abstract sections span near the center or bias one side

Related commit history suggests this area has already seen targeted work and should be verified explicitly.

---

## Impact

- residual reading-order errors in academic and report PDFs
- lower NID even after basic two-column fixes
- confusing section transitions in markdown/text output

---

## Candidate source files

### Go
- `go/internal/processors/xycut_plus_plus_sorter.go`
- related page/chunk partition helpers

### Java reference
- corresponding Java balanced-column or XYCut++ reading-order logic

---

## Fixtures / Reproduction

Use multi-column PDFs where columns are uneven or structurally imbalanced, including papers with:

- abstract/header asymmetry
- figures intruding on one column
- varying text density across columns

Compare Java vs Go output order around transition boundaries.

---

## Acceptance criteria

1. Targeted imbalanced-column fixtures move closer to Java ordering.
2. Basic two-column ordering does not regress.
3. Relevant XYCut / reading-order tests pass.
4. `go build ./...` succeeds.

---

## Evaluation commands

```bash
cd go && go test ./internal/processors/... -v
cd go && go build ./...
```

Then compare targeted fixture outputs against Java baselines.

---

## Non-goals

Out of scope here:

- first-pass column split detection from scratch
- sidebar-specific mixed layout handling
- text extraction decoding issues

---

## Keep / Discard guidance

### KEEP

- imbalanced-column ordering improves without harming simpler pages

### DISCARD

- heuristics only help one fixture and regress ordinary two-column cases
