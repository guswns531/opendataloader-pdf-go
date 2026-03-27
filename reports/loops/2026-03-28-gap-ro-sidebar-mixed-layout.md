## Gap ID

`GAP-RO-SIDEBAR-MIXED-LAYOUT`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

- Go reading-order still let narrow edge sidebars / floating marginal blocks interrupt the main body too early after the basic two-column and balanced-column fixes.
- Mixed-layout pages remained a likely residual T17 parity gap even though simpler column cases had already improved.

## Evidence before

- Existing T17 work had already fixed basic two-column detection and balanced-column merge order, but not marginal sidebar timing.
- The new focused regression reproduced the failure mode where a very narrow left-edge sidebar should trail the surrounding body band instead of being read before it.
- Java tests already treated ArXiv-style sidebars as special/flexible rather than forcing them into ordinary column order.

## What changed

- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
  - added a narrow mixed-layout pre-pass to peel off very narrow edge-aligned, vertically-overlapping floating/sidebar objects before ordinary column detection
  - added a marginal-sidebar split path for column partitions where one side is effectively a narrow sidebar rather than a full column
  - merged neutral / deferred floating elements back by bottom-edge position so they follow the surrounding main body band instead of jumping ahead of it
- `go/internal/processors/readingorder/xycut_plus_plus_sorter_test.go`
  - added `TestXYCutDefersMarginalSidebarUntilAfterMainBodyBand`

## Validation

```bash
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go test ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go build ./...
```

Results:

- focused reading-order tests: passed
- XYCut-targeted processor/unit tests: passed
- build: passed

## Regressions checked

- previously kept basic two-column and balanced-column XYCut regressions still pass under the targeted validator set
- the new sidebar regression passes with the deferred-floating merge behavior

## Remaining uncertainty

- This KEEP is still based on synthetic regression coverage plus Java test intent, not yet on a real Java-vs-Go fixture diff for a mixed-layout PDF page.
- The local `openclaw system event` completion notification failed during the Codex run because the local gateway was unavailable.

## Follow-up

- Next best T17 step is fixture-backed evaluation for real mixed-layout/sidebar pages, or a narrower follow-up gap if real samples show remaining failures (for example captions / callouts that are not edge-aligned sidebars).

## Commit

`N/A`
