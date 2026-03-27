# Codex Gap Task Packet — GAP-TEXT-BENCHMARK-JOINED-WORDS

## Role

You are fixing one narrow Java→Go text-parity gap in the Go port of OpenDataLoader PDF. Fresh context, one gap only, evidence first.

## Gap ID

`GAP-TEXT-BENCHMARK-JOINED-WORDS`

## Objective

Investigate and, if justified by evidence, fix the next best live text-parity family after the already-kept spacing and soft-hyphen joins.

Start from the benchmark-backed residual joined-word cases called out in prior reports, especially patterns like:

- `theCreationofLife`
- `Oneofthecharacteristics...`
- lingering `InProceedings`-style joins if still live on the current tree

Your job is to:

1. inspect the current kept text-parity reports,
2. identify one exact residual joined-word family that is still live on the current worktree,
3. reproduce it with the narrowest possible unit/fixture/benchmark-shaped evidence,
4. implement one scoped fix only if Java-parity evidence is strong,
5. otherwise stop at `SPLIT` with an exact narrower follow-up gap.

## Context to review first

- `plan/gaps/EVAL-TEXT-PARITY-CORE.md`
- `reports/loops/2026-03-28-gap-text-benchmark-word-joins.md`
- `reports/loops/2026-03-28-gap-text-benchmark-prose-spacing.md`
- `reports/loops/2026-03-28-gap-text-benchmark-heading-spacing.md`
- `reports/loops/2026-03-28-gap-text-inline-gap-spacing.md`
- `reports/loops/2026-03-28-gap-text-singleton-continuation-boundary.md`
- `reports/loops/2026-03-28-gap-text-soft-hyphen-line-join.md`
- `reports/loops/2026-03-28-eval-text-parity-next-gap-soft-hyphen-line-join.md`

## Constraints

- One residual gap only.
- Do not touch reading-order code.
- Prefer fixture-backed or checked-in benchmark evidence over synthetic-only heuristics.
- If the benchmark markdown files are enough to isolate the gap, use them.
- If live PDF reproduction is needed, keep it narrow.
- If evidence is ambiguous, return `SPLIT` instead of widening spacing heuristics.

## Acceptance criteria for KEEP

1. exact residual mismatch identified
2. focused regression or fixture comparison added/strengthened
3. one scoped fix only
4. targeted tests pass
5. `go build ./...` passes
6. loop report written under `reports/loops/`

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors ./tests/unit/processors ./internal/generators/markdown/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
```

If broader runs are unnecessary, use the narrowest trustworthy subset for the chosen gap and explain why.

## Deliverable

Report:
1. chosen gap ID
2. evidence before
3. files changed
4. before/after evidence
5. tests/commands run
6. KEEP / SPLIT / NEEDS-HUMAN-REVIEW

When completely finished, run this command to notify me:
openclaw system event --text "Done: evaluated joined-word text parity gap in opendataloader-pdf-go" --mode now
