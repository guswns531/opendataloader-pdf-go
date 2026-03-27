## Gap ID

`GAP-TEXT-TRIMSPACE`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

- A small uncommitted text-parity patch remained in the worktree after earlier trim/boundary-space work.
- The patch preserves whitespace-only runs by appending boundary whitespace to the prior extracted text chunk without duplicating trailing spaces.

## Evidence before

- Dirty worktree showed `go/pkg/pdfbox/extractor/text_extractor.go` with a focused whitespace-preservation change.
- The earlier trimspace family already had parity value, and this leftover patch was narrow enough to validate independently.

## What changed

- `go/pkg/pdfbox/extractor/text_extractor.go`
  - replaced direct whitespace-only run concatenation with `appendBoundaryWhitespace`
  - preserved boundary whitespace while avoiding duplicate trailing-space growth
  - kept width estimates in sync after appending preserved whitespace

## Validation

```bash
cd go && go test ./pkg/pdfbox/extractor -run 'Test.*Boundary|Test.*Space|Test.*Trim' -count=1
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && go build ./...
```

Results:

- focused extractor whitespace tests: passed
- packet-level pdfbox/unit suites: passed
- build: passed

## Remaining uncertainty

- This is a small refinement inside the existing trimspace family, not a fresh fixture-driven gap closure by itself.
- Broader text parity still depends on fixture-level evidence already tracked in the T16 family.

## Follow-up

- Commit and push this narrow KEEP so the next reading-order loop starts from a clean tree.
- Resume `GAP-RO-BALANCED-COLUMNS` on top of the cleaned baseline.
