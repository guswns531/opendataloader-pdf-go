# Codex Gap Task Packet — GAP-RO-XYCUT-DEPTH-STABILITY

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-RO-XYCUT-DEPTH-STABILITY`

## Category

`reading-order-parity`

---

## Objective

Implement the smallest change that improves Go reading-order stability for complex layouts where recursive XYCut partitioning becomes overly fragmented, unstable across reruns, or risks runaway/deep recursion.

---

## Reference behavior (Java)

The Java reference handles complex page segmentation deterministically and without catastrophic recursion failures.

---

## Current Go behavior

Go now has targeted fixes for basic two-column, balanced-column, and marginal-sidebar ordering. Remaining T17 risk appears to be recursion/partition stability: pathological splits, repeated over-fragmentation, or unstable merge order on complex layouts.

This gap must stay focused on **XYCut depth / stability**, not sidebar heuristics or text extraction.

---

## Why this matters

- unstable recursion can undo recent reading-order gains
- pathological splits can hurt deterministic output quality on difficult pages
- this is the next narrow T17 parity gap after the current kept column/sidebar work

---

## Relevant files

### Go
- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
- nearby reading-order helpers/tests only if directly needed

### Java reference
- corresponding Java XYCut recursion / partition logic

### Fixtures / evidence
- existing XYCut tests
- add one focused regression if you can reproduce instability or over-fragmentation deterministically

---

## Required work

- inspect current recursive partition / merge logic for instability risks
- patch only the narrow depth/stability issue you can justify from code/test evidence
- add or strengthen focused regression coverage
- keep previously fixed two-column / balanced-column / sidebar cases green

---

## Acceptance criteria

1. Targeted complex-layout ordering is deterministic across reruns.
2. No recursion-related crash or obvious runaway behavior on the targeted regression.
3. Basic two-column, balanced-column, and sidebar regressions do not regress.
4. `cd go && go build ./...` succeeds.

---

## Validation commands

Run these before finishing:

```bash
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go test ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go build ./...
```

If you add a deterministic rerun check, run it more than once and report the outcome.

---

## Non-goals

Do **not** attempt these in the same patch:

- new sidebar / mixed-layout heuristics unrelated to depth stability
- broad benchmark tuning without a focused stability signal
- text extraction / serializer / table gaps
- unrelated cleanup/refactors

---

## Output format

When done, report:

1. files changed
2. behavior changed
3. tests/commands run and their result
4. remaining uncertainty
5. whether the gap appears fully fixed, partially fixed, or needs follow-up
6. exact commit message to use if the result looks KEEP-worthy
