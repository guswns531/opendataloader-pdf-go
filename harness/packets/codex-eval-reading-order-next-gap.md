# Codex Gap Task Packet — EVAL-READING-ORDER-NEXT-GAP

## Role

You are evaluating the next narrow Java→Go reading-order parity step in the Go port of OpenDataLoader PDF. Fresh context, one gap only, evidence first.

## Objective

Select and investigate the next best credible `T17` parity gap after the already-kept 2026-03-28 reading-order fixes, especially the recent `GAP-RO-RIGHT-COLUMN-IMAGE-GRID-SEQUENCING` keep.

Do **not** broadly rewrite XYCut++. Your job is to:

1. inspect the current kept reading-order reports,
2. identify one residual mismatch family that still looks live,
3. reproduce it with the narrowest possible regression or fixture-backed evidence,
4. either implement one scoped fix or return `SPLIT` with an exact follow-up gap.

## Context to review first

- `plan/gaps/EVAL-READING-ORDER-PARITY.md`
- `reports/loops/2026-03-28-gap-ro-two-column-basic.md`
- `reports/loops/2026-03-28-gap-ro-balanced-columns.md`
- `reports/loops/2026-03-28-gap-ro-sidebar-mixed-layout.md`
- `reports/loops/2026-03-28-gap-ro-xycut-depth-stability.md`
- `reports/loops/2026-03-28-gap-ro-fixture-mixed-layout-eval.md`
- `reports/loops/2026-03-28-gap-ro-right-column-image-grid-sequencing.md`
- `plan/gaps/GAP-RO-BALANCED-COLUMNS.md`
- `plan/gaps/GAP-RO-SIDEBAR-MIXED-LAYOUT.md`
- `plan/gaps/GAP-RO-XYCUT-DEPTH-STABILITY.md`

## Constraints

- One residual gap only.
- Prefer a real-fixture or fixture-shaped regression over speculative cleanup.
- If benchmark PDFs are blocked by LFS pointers, use the best real local PDF fixture already available in this checkout or construct the tightest unit-level reproduction you can justify from existing evidence.
- Do not touch text extraction / spacing code.
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
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/readingorder/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
```

If broader fixture replay is blocked by missing benchmark assets or unrelated sandbox restrictions, run the narrowest trustworthy subset for the chosen gap and explain why.

## Deliverable

Report:
1. chosen gap ID
2. evidence before
3. files changed
4. before/after evidence
5. tests/commands run
6. KEEP / SPLIT / NEEDS-HUMAN-REVIEW

When completely finished, run this command to notify me:
openclaw system event --text "Done: evaluated next T17 reading-order gap in opendataloader-pdf-go" --mode now
