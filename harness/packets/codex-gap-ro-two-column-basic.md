# Codex Gap Task Packet — GAP-RO-TWO-COLUMN-BASIC

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-RO-TWO-COLUMN-BASIC`

## Category

`reading-order-parity`

---

## Objective

Implement the smallest change that improves Go reading-order parity for standard two-column layouts.

---

## Reference behavior (Java)

The Java implementation should preserve a human-correct reading order for standard two-column pages, yielding left-column content top-to-bottom before the right column where appropriate.

---

## Current Go behavior

Go may still interleave columns too eagerly or choose the wrong split/merge order for standard two-column layouts.

This gap should stay focused on basic two-column sequencing, not sidebars, deep XYCut stability, or unrelated text extraction issues.

---

## Why this matters

- severe readability degradation in markdown/text output
- lower reading-order parity vs Java
- worse downstream chunking and retrieval behavior

---

## Relevant files

### Go
- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
- `go/tests/unit/processors/readingorder/xycut_test.go`
- nearby reading-order tests/helpers only if directly needed

### Java reference
- corresponding Java reading-order / XYCut++ path

### Fixtures / evidence
- unit coverage already exists for two-column ordering
- if needed, add one focused regression that demonstrates the missing standard two-column behavior

---

## Required work

- inspect the current two-column split / merge behavior
- patch only the narrow basic two-column gap
- add or strengthen focused regression coverage
- avoid broad sidebar / mixed-layout / deep-recursion changes unless strictly required for this gap

---

## Acceptance criteria

1. Standard two-column layouts no longer interleave columns line-by-line.
2. Output moves closer to Java behavior for the targeted case.
3. Targeted XYCut / reading-order tests pass.
4. Single-column ordering does not obviously regress.
5. `cd go && go build ./...` succeeds.

---

## Validation commands

Run these before finishing:

```bash
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
