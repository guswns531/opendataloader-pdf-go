# Codex Gap Task Packet — GAP-RO-SIDEBAR-MIXED-LAYOUT

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-RO-SIDEBAR-MIXED-LAYOUT`

## Category

`reading-order-parity`

---

## Objective

Implement the smallest change that improves Go reading-order parity for mixed-layout pages where sidebars, callouts, captions, or other floating regions should not interrupt the main reading flow too early.

---

## Reference behavior (Java)

The Java implementation keeps main narrative flow stable on mixed-layout pages while still placing side content in a deterministic, human-correct order.

---

## Current Go behavior

Go now has better basic two-column handling and a balanced-column merge path, but mixed-layout pages may still be over-classified as full columns. That can insert sidebar/callout content too early or let local floating regions interrupt the parent flow.

This gap must stay focused on **sidebar / mixed-layout ordering**, not deeper recursion stability or text extraction.

---

## Why this matters

- sidebars and pull quotes can break readability in markdown/text output
- over-eager column heuristics still hurt `T17` parity after the basic column fixes
- mixed layouts are a likely remaining source of NID loss

---

## Relevant files

### Go
- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
- nearby reading-order helpers/tests only if directly needed

### Java reference
- corresponding Java layout partitioning / XYCut++ mixed-layout handling

### Fixtures / evidence
- existing reading-order tests and benchmark markdown samples
- if needed, add one focused regression for sidebar/callout ordering

---

## Required work

- inspect current handling of cross-layout / neutral / floating regions after the balanced-column patch
- patch only the narrow mixed-layout ordering gap
- add or strengthen focused regression coverage
- avoid broad deep-recursion changes unless strictly required for this gap

---

## Acceptance criteria

1. Targeted mixed-layout cases move closer to Java behavior.
2. Basic two-column and balanced-column regressions do not regress.
3. Reading-order remains deterministic across reruns.
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

- sidebar + deep-recursion stability together
- broader benchmark optimization without mixed-layout evidence
- text extraction / serializer gaps
- table ordering changes

---

## Output format

When done, report:

1. files changed
2. behavior changed
3. tests/commands run and their result
4. remaining uncertainty
5. whether the gap appears fully fixed, partially fixed, or needs follow-up
