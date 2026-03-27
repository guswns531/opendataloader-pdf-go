## Gap ID

`GAP-RO-SIDEBAR-MIXED-LAYOUT`

## Date

`2026-03-28`

## Decision

`SPLIT`

## Why this loop existed

- A fresh Codex loop tried to close the remaining mixed-layout reading-order leak on `samples/pdf/1901.03003.pdf` pages 12-13 after the earlier kept marginal-sidebar sorter work.

## Evidence before

- Go markdown still leaked narrow tokens such as `west`, `united`, `arsenal`, `football`, `manchester`, `messageid`, `briogestone`, and `contracers` into the main body flow around the conclusion section.

## What changed

- exploratory candidate edits in `go/internal/processors/document_processor.go` and `go/internal/processors/text_line_processor.go`
- added a focused test for splitting large baseline-aligned horizontal gaps
- candidate built and targeted unit tests passed
- candidate was discarded from the worktree because fixture parity did not improve enough

## Validation

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./...
cd go && GOCACHE=/tmp/odl-gocache go build ./...
rm -rf /tmp/odl-moran-md && mkdir -p /tmp/odl-moran-md
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
rg -n "west|united|arsenal|football|manchester|messageid|briogestone|contracers" /tmp/odl-moran-md/1901.03003.md
```

Results:

- build: passed
- targeted test additions: passed
- fixture parity: failed acceptance; leaked gutter-stack tokens still remained in output

## Regressions checked

- no KEEP was taken, so no new regression risk was accepted

## Remaining uncertainty

- The leak appears to happen earlier than the already-kept edge-sidebar ordering logic, likely in intermediate text-line / list / semantic grouping rather than only final XYCut ordering.

## Follow-up

- Split into `GAP-RO-INTERNAL-GUTTER-STACK` with a fresh packet focused on pre-semantic grouping and internal gutter text leakage.

## Commit

`N/A`
