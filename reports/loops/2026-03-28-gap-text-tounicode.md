## Gap ID

`GAP-TEXT-TOUNICODE`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

- Go still failed ToUnicode parity on multi-line `beginbfrange` array mappings.
- Decoder output fell back to raw bytes instead of mapped Unicode when the CMap array values were split across lines.

## Evidence before

- `TestParseCMapContentDecodesMultiLineBFRangeArray` failed with empty mappings/code lengths.
- Decoder output for `00 01 00 02 00 03` remained raw bytes instead of `éñ€`.
- Codex found the parser only captured the first part of the candidate fix and still missed the multiline-array opener case.

## What changed

- `go/pkg/pdfbox/extractor/font_decoder.go`
  - taught `parseCMapContent` to keep state for pending `bfrange` arrays across lines
  - allowed multiline array openers that contain only start/end codes plus `[` 
  - added a helper to apply collected array values deterministically
- `go/pkg/pdfbox/extractor/content_parser_test.go`
  - added regression coverage for multiline `beginbfrange` arrays and end-to-end decoder output
- `harness/packets/codex-gap-text-tounicode.md`
  - kept the scoped task packet used for this loop

## Validation

```bash
cd go && go test ./pkg/pdfbox/extractor -run 'TestParseCMapContentDecodesMultiLineBFRangeArray|TestDecodeTextHandlesToUnicodeMultiLineArrayAndStringFallback' -count=1
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && go build ./...
```

Results:

- focused extractor regression: passed
- packet-level test suite: passed
- build: passed

## Regressions checked

- existing extractor/package tests under `./pkg/pdfbox/...`
- unit tests under `./tests/unit/pdfbox/...`
- non-array `bfrange` path preserved by existing CMap regression coverage

## Remaining uncertainty

- This closes the multiline-array parsing hole for the current ToUnicode family, but broader ToUnicode parity may still need fixture-driven checks beyond this unit coverage.

## Follow-up

- Move to the next drafted text gap still showing active worktree evidence: `GAP-TEXT-TRIMSPACE` if the current dirty changes validate cleanly, otherwise start a fresh packet for the highest-signal remaining text mismatch.
- Separately fix `scripts/autopilot-check.sh` so cron has a reliable `go` binary in PATH.

## Commit

`N/A`
