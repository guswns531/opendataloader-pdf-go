# Codex Gap Task Packet — GAP-RO-BALANCED-COLUMNS

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-RO-BALANCED-COLUMNS`

## Category

`reading-order-parity`

---

## Objective

Implement the smallest change that improves Go reading-order parity for imbalanced two-column layouts.

---

## Reference behavior (Java)

The Java implementation preserves a human-correct reading order on uneven two-column pages where one column starts earlier, ends later, or contains asymmetric blocks such as abstract text, headings, or figures.

---

## Current Go behavior

Go already improved basic two-column split detection, but may still mis-sequence content when:

- one column starts or ends earlier than the other
- block heights differ significantly
- center-biased headings or abstract blocks distort merge order

This gap must stay focused on **balanced/imbalanced column merge order**, not sidebar handling, deep XYCut stability, or text decoding.

---

## Why this matters

- residual reading-order errors in academic/report PDFs
- lower NID after basic split detection is fixed
- confusing section transitions in markdown/text output

---

## Relevant files

### Go
- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
- nearby reading-order helpers/tests only if directly needed

### Java reference
- corresponding Java balanced-column / XYCut++ reading-order logic

### Fixtures / evidence
- existing reading-order tests and benchmark markdown samples
- if needed, add one focused regression for an imbalanced two-column case

---

## Required work

- inspect current merge/sequencing behavior after the basic two-column split
- patch only the narrow imbalanced-column ordering gap
- add or strengthen focused regression coverage
- avoid broad sidebar / mixed-layout / deep-recursion changes unless strictly required for this gap

---

## Acceptance criteria

1. Targeted imbalanced-column cases move closer to Java behavior.
2. Basic two-column ordering does not regress.
3. Targeted XYCut / reading-order tests pass.
4. `cd go && go build ./...` succeeds.

---

## Validation commands

Run these before finishing:

```bash
cd go && go test ./internal/processors/readingorder/... -count=1
cd go && go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && go build ./...
```

---

## Non-goals

Do **not** attempt these in the same patch:

- sidebar and mixed-layout special cases
- deep recursion stability work
- broader text extraction gaps
- table ordering changes

---

## Output format

When done, report:

1. files changed
2. behavior changed
3. tests/commands run and their result
4. remaining uncertainty
5. whether the gap appears fully fixed, partially fixed, or needs follow-up
