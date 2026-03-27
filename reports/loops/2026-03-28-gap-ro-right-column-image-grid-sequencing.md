## Gap ID

`GAP-RO-RIGHT-COLUMN-IMAGE-GRID-SEQUENCING`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

- The remaining fixture-derived T17 mismatch was the `1901.03003` right-column image-grid/caption band.
- Go was still pulling the far-right image subcolumn (`660..662`) forward and interleaving the marginal `667` block into the body/grid band before `650` and `656` were placed correctly.

## Evidence before

- Focused regression reproduced the exact bad Go order called out in the packet:
  - `646, 647, 648, 649, 660, 661, 662, 667, 650, 653, 657, 654, 658, 655, 659, 656, 663, 664, 665, 666`
- Narrow root-cause probe showed the divergence was happening in the deferred-floating path, not a broad XYCut rewrite:
  - `identifyMarginalFloatingElements` was classifying both the real marginal sidebar `667` and the stacked far-right image subcolumn `660..662` as floating
  - after that, generic `mergeDeferredFloatingElements` reinsertion pulled `660..662` back too early

## What changed

- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
  - split deferred floating candidates into:
    - true marginal floating/sidebar items
    - stacked edge-column groups such as the dense far-right `660..662` image subcolumn
  - merged stacked edge-column groups back before the first below-band overlapping continuation block, which keeps the image-grid band together ahead of caption `656`
  - left the existing generic deferred-floating merge path in place for ordinary sidebar cases
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
  - added a focused `1901.03003`-derived regression for the right-column image-grid/caption sequencing family

## Before/after evidence

- Before:
  - `646, 647, 648, 649, 660, 661, 662, 667, 650, 653, 657, 654, 658, 655, 659, 656, 663, 664, 665, 666`
- After:
  - the new regression now passes with these enforced parity signals:
    - `650` stays before the entire `653..662` image cluster
    - `653..662` remain consecutive as one grouped band
    - the grouped image band stays before caption `656`
    - `667` no longer lands inside the `650..666` body/grid band
    - `663 -> 664 -> 665 -> 666` remains intact after `656`

## Validation

Focused reproduction:

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/readingorder/... -run TestXYCutKeepsRightColumnImageGridGroupedAfterLeftTrailBlock -count=1
```

Packet validation:

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
```

Results:

- focused right-column image-grid regression: passed
- targeted reading-order tests: passed
- XYCut-targeted processor/unit tests: passed
- build: passed

## Regressions checked

- previously kept T17 behaviors remained green under the packet validator set:
  - basic two-column sequencing
  - balanced-column ordering
  - marginal sidebar deferral
  - deep-recursion stability fallback

## Remaining uncertainty

- The kept fix is intentionally narrow and does not try to normalize the internal order of every dense image grid; it only prevents the stacked edge subcolumn from being merged ahead of the surrounding body/caption flow in this residual family.

## Follow-up

- No split needed from this packet.
- If another real fixture shows a different dense-grid family, treat that as a fresh gap rather than widening this heuristic.

## Commit

`313e399`
