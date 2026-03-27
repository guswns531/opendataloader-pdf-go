# Codex Gap Task Packet — EVAL-TEXT-PARITY-NEXT-GAP

## Role

You are evaluating the next narrow Java→Go text-parity step in the Go port of OpenDataLoader PDF. Fresh context, one gap only, evidence first.

## Objective

Select and investigate the next best credible `T16` parity gap after the already-kept fixes on 2026-03-28.

Do **not** broadly rewrite text extraction. Your job is to:

1. inspect the current kept text-parity reports,
2. identify one residual mismatch family that still looks live,
3. reproduce it with the narrowest possible regression or fixture-backed evidence,
4. either implement one scoped fix or return `SPLIT` with an exact follow-up gap.

## Context to review first

- `plan/gaps/EVAL-TEXT-PARITY-CORE.md`
- `reports/loops/2026-03-28-gap-text-tj-offset.md`
- `reports/loops/2026-03-28-gap-text-trimspace.md`
- `reports/loops/2026-03-28-gap-text-trimspace-boundary-whitespace.md`
- `reports/loops/2026-03-28-gap-text-winansi.md`
- `reports/loops/2026-03-28-gap-text-tounicode.md`
- `reports/loops/2026-03-28-gap-text-benchmark-word-joins.md`
- `reports/loops/2026-03-28-gap-text-benchmark-prose-spacing.md`
- `reports/loops/2026-03-28-gap-text-benchmark-heading-spacing.md`
- `reports/loops/2026-03-28-eval-text-parity-discovery-inline-gap-spacing.md`

## Constraints

- One residual gap only.
- Prefer a fixture-backed or benchmark-shaped regression over speculative cleanup.
- If benchmark PDFs are blocked by LFS pointers, derive the narrowest unit-level reproduction from existing evidence.
- Do not touch reading-order code.
- If evidence is ambiguous, stop at `SPLIT` instead of widening heuristics.

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
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors ./internal/processors -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
```

If the broad commands are blocked by known sandbox/network/listener issues, run the narrowest trustworthy subset for the chosen gap and explain why.

## Deliverable

Report:
1. chosen gap ID
2. evidence before
3. files changed
4. before/after evidence
5. tests/commands run
6. KEEP / SPLIT / NEEDS-HUMAN-REVIEW

When completely finished, run this command to notify me:
openclaw system event --text "Done: evaluated next T16 text-parity gap in opendataloader-pdf-go" --mode now
