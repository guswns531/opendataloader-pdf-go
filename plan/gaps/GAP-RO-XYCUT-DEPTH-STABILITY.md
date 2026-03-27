# GAP-RO-XYCUT-DEPTH-STABILITY

## Category

reading-order-parity

## Why this gap exists

Even if column detection improves, recursive XYCut-style ordering can still fail through instability, excessive recursion depth, or pathological layout partitioning.

This is a stability-oriented sub-gap under `T17`.

---

## Reference behavior (Java)

The Java reference implementation handles complex page segmentation without catastrophic recursion failures and produces stable reading order across repeated runs.

---

## Current Go behavior

Potential symptoms may include:

- stack overflow or runaway recursion
- unstable ordering on the same input
- highly fragmented partitions that reduce reading-order quality

Relevant history already suggests this area has seen fixes and may still need parity verification.

---

## Impact

- crashes or severe instability on complex layouts
- unpredictable reading order quality
- difficulty trusting reading-order fixes in broader rollout

---

## Candidate source files

### Go
- `go/internal/processors/xycut_plus_plus_sorter.go`
- recursion / partition helpers related to XYCut

### Java reference
- corresponding Java XYCut recursion and partition heuristics

---

## Fixtures / Reproduction

Use complex layouts that stress recursion and partition logic, including:

- multi-column pages
- mixed block density pages
- previously failing or stack-overflow-associated samples

---

## Acceptance criteria

1. No recursion-related crash on targeted fixtures.
2. Reading order remains deterministic across repeated runs.
3. Two-column improvements are not undone by stability fixes.
4. Relevant processor tests pass.
5. `go build ./...` succeeds.

---

## Evaluation commands

```bash
cd go && go test ./internal/processors/... -v
cd go && go build ./...
```

Run targeted fixture conversions more than once to confirm stable ordering.

---

## Non-goals

Out of scope here:

- all column split heuristics
- text decoding / spacing bugs
- table extraction parity unrelated to reading order recursion

---

## Keep / Discard guidance

### KEEP

- targeted fixtures no longer crash or destabilize
- ordering is stable across reruns

### DISCARD

- recursion instability remains
- fix harms ordinary reading-order quality without clear benefit
