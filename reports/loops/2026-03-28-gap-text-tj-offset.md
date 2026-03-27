## Gap ID

`GAP-TEXT-TJ-OFFSET`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

- The packet claimed Go still collapsed word boundaries when `TJ` numeric offsets represented spaces.
- The named sample was `samples/pdf/1901.03003.pdf`.

## Evidence before

- A live extractor probe against `1901.03003.pdf` already returned `MORAN: A Multi-Object Rectiﬁed Attention Network` on page 1.
- Existing focused unit coverage in `go/pkg/pdfbox/extractor/content_parser_test.go` already covered synthetic `TJ` spacing recovery.

## What changed

- `go/tests/unit/pdfbox/extractor_test.go`
  - added a focused regression that asserts the extracted title chunk from `1901.03003.pdf` preserves the recovered `TJ` spacing

## Validation

```bash
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...
cd go && go build ./...
```

Additional scoped checks:

```bash
cd go && go test ./pkg/pdfbox/extractor -run 'TestDecodeTJText|TestExtractTextChunksRecoversSpaceFromTJOffsets'
cd go && GOCACHE=/tmp/odl-gocache go run /tmp/tjprobe/main.go
```

Results:

- packet validation suites: passed
- targeted extractor tests: passed
- live sample probe: page 1 title already extracts with correct `TJ` word spacing

## Regressions checked

- synthetic `TJ` spacing recovery tests still pass
- sample-based title extraction is now locked by regression coverage
- full packet build completed successfully

## Remaining uncertainty

- The packet's exact historical failure is not reproducible on the current worktree.
- Other text-joining artifacts in the sample exist, but they are outside this packet's scoped `TJ` title-spacing symptom.

## Follow-up

- Close this packet as stale unless a fresh reproducer shows a remaining `TJ` numeric-spacing failure.
- If a new reproducer appears, split it from broader glyph/joining issues before changing extraction heuristics.

## Commit

`N/A`
