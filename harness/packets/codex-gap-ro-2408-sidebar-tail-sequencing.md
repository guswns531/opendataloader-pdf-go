# Codex Gap Task Packet — GAP-RO-2408-SIDEBAR-TAIL-SEQUENCING

## Role

You are evaluating one narrow Java→Go reading-order parity gap in the Go port of OpenDataLoader PDF. Fresh context, one gap only, evidence first.

## Gap ID

`GAP-RO-2408-SIDEBAR-TAIL-SEQUENCING`

## Objective

Decide whether the current Go ordering difference for the `2408.02509v1` academic-paper geometry is a real Java parity miss worth fixing, or only a flexible/sidebar-tolerated difference that should stay split.

Current exploratory evidence says Go orders the marginal sidebar object `105` as:

- `95 96 97 98 99 105 100 101 102 103 104`

The exploratory Java-style expectation used in the probe was:

- `95 96 97 98 99 100 101 102 103 104 105`

But the prior evaluator report also notes the Java test comments/assertions may treat sidebar placement as flexible.

Your job is to gather trustworthy Java-side evidence from the current repo and stop at the narrowest correct conclusion:

1. inspect the prior report and relevant Java/Go XYCut tests,
2. reproduce the `2408.02509v1` geometry on both sides if possible,
3. determine whether Java truly requires `105` to trail the full `100..104` band,
4. if yes, implement one minimal Go fix plus focused regression,
5. otherwise leave code unchanged and write a precise `SPLIT` or `DISCARD` report.

## Context to review first

- `reports/loops/2026-03-28-eval-reading-order-next-gap.md`
- Go XYCut files/tests around the `2408.02509v1` case
- corresponding Java XYCut tests / fixture-backed evidence in this repo

## Constraints

- One gap only.
- Do not touch text-extraction or markdown-generation code.
- Do not widen general sidebar heuristics without hard Java evidence.
- Prefer exact Java test/assertion behavior over comments.
- If trustworthy Java reproduction is not available, stop at `SPLIT` with exact blocker details.

## Acceptance criteria for KEEP

1. exact Java reference behavior for this geometry is shown
2. focused Go regression added/strengthened
3. one scoped reading-order fix only
4. targeted reading-order tests pass
5. `go build ./...` passes
6. concise loop report written under `reports/loops/`

## Validation commands

Use the narrowest trustworthy subset for the chosen outcome, for example:

```bash
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go test ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go build ./...
```

If Java-side reproduction or tests are available locally, run the narrowest corresponding command and include it in the report.

## Deliverable

Report:
1. chosen gap ID
2. exact Java-side evidence found
3. files changed
4. before/after evidence
5. tests/commands run
6. KEEP / SPLIT / DISCARD / NEEDS-HUMAN-REVIEW

If KEEP-worthy, commit and push.

When completely finished, run this command to notify me:
openclaw system event --text "Done: evaluated 2408 sidebar tail sequencing gap in opendataloader-pdf-go" --mode now
