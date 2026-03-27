## Gap ID

`GAP-RO-LIST-INTERNAL-GUTTER-CONTINUATION`

## Date

`2026-03-28`

## Decision

`SPLIT`

## Why this loop was rechecked

- An active Codex loop (`amber-meadow`) was still running, but it had devolved into a long raw evidence dump on `1901.03003` rather than converging on a bounded patch.
- I killed that run and rechecked the current Go worktree directly.
- The tree was clean before and after the evaluation; there was no pending KEEP-worthy patch to validate or commit.

## Reproduction

```bash
rm -rf /tmp/odl-moran-md && mkdir -p /tmp/odl-moran-md
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
rg -n "west|united|arsenal|football|manchester|messageid|briogestone|contracers" /tmp/odl-moran-md/1901.03003.md
sed -n '60,150p' /tmp/odl-moran-md/1901.03003.md
```

Observed output is unchanged and still contains both:

- gutter tokens as standalone lines:
  - `west`
  - `united`
  - `arsenal`
  - `football`
  - `manchester messageid`
  - `contracers`
- contaminated list/body continuation lines:
  - `- west The experiments above are all based on cropped text`
  - `- in more application scenarios, irregular and multi-`
  - `- instance, Liu et al. [29] and Ch'ng et al. [7] released`
  - `————— briogestone In this paper, we presented a multi-object rec-`

## What was tested during this recheck

I tried one narrow hypothesis locally: tighten `ListProcessor` continuation acceptance so a narrow non-overlapping line would not be absorbed into an existing list item.

Result:

- the targeted synthetic unit test shape could be made stricter,
- but the real `1901.03003` markdown output did not change at all,
- so the change was reverted and not kept.

## What this implies

This is probably **not primarily a `ListProcessor.isListContinuation` bug**.

The unchanged fixture output suggests the wrong grouping is already present earlier, likely in one of these stages:

1. text-line / semantic paragraph formation before list construction,
2. line-art-bullet / list-item seed creation,
3. document-level element ordering or carry-over that feeds list processing a pre-corrupted sequence.

In other words, the packet name is too narrow now: list continuation is a symptom, not yet the proven root cause.

## Files changed during evaluation

- no kept source changes
- new report only:
  - `reports/loops/2026-03-28-eval-list-internal-gutter-continuation-recheck.md`

## Validation

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
```

Results:

- internal processor tests: passed
- internal reading-order tests: passed
- build: passed

## Exact next best parity step

Split this into a root-cause packet focused on the seed stage that creates the bad sequence on `1901.03003` pages 12-13, for example:

`GAP-RO-GUTTER-TOKEN-SEED-GROUPING-1901`

That packet should:

- trace page 12-13 objects immediately before paragraph/list processing,
- confirm whether `west/united/arsenal/football/...` are already interleaved into the main flow before list construction,
- compare Java vs Go object grouping at that pre-list stage,
- only then decide whether the fix belongs in paragraph grouping, list seeding, or upstream reading-order ordering.

## Commit

`N/A`
