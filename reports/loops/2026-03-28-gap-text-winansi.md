## Gap ID

`GAP-TEXT-WINANSI`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

- Go still had a narrow WinAnsi parity gap in the `0x80-0x9F` range.
- The decoder path could surface replacement characters for undefined Windows-1252 bytes instead of matching the raw WinAnsi fallback behavior already used elsewhere.

## Evidence before

- The uncommitted worktree contained a focused `font_decoder` patch plus a new regression for undefined WinAnsi bytes.
- The previous `winAnsiTable()` implementation depended on `charmap.Windows1252.NewDecoder()`, which can yield replacement runes for undefined bytes such as `0x81`, `0x8d`, `0x8f`, `0x90`, and `0x9d`.

## What changed

- `go/pkg/pdfbox/extractor/font_decoder.go`
  - simplified `winAnsiTable()` to build the table directly from `winAnsiRune(byte(i))` for all 256 entries
  - removed the decoder-based path that could inject `U+FFFD` for undefined WinAnsi bytes
- `go/pkg/pdfbox/extractor/content_parser_test.go`
  - added a regression ensuring undefined bytes in the WinAnsi range match `normalizePDFString(raw)` and do not contain replacement characters

## Validation

```bash
cd go && go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...
cd go && go build ./...
```

Additional scoped validation run from the outer environment:

```bash
cd go && go test ./internal/processors/readingorder/... -count=1
cd go && go test ./internal/processors/... ./tests/unit/processors/... -run TestXYCut -count=1
cd go && go test ./pkg/pdfbox/extractor -count=1
cd go && go build ./...
```

Results:

- packet-level pdfbox/unit suites: passed
- extractor package tests: passed
- targeted XYCut regressions: passed
- build: passed

## Regressions checked

- ASCII decoding remains covered by the existing WinAnsi table test
- undefined `0x80-0x9F` bytes now follow the raw fallback path instead of emitting replacement characters
- unrelated reading-order validations still pass on the same worktree

## Remaining uncertainty

- This is a narrow decoder-table parity fix, not a fixture-level end-to-end text extraction comparison.
- Follow-up text-parity work should still check whether remaining `T16` gaps are now primarily ToUnicode/TJ-related rather than WinAnsi-related.

## Follow-up

- Commit and push this scoped KEEP immediately.
- Continue autopilot with the next best parity gap: `GAP-RO-SIDEBAR-MIXED-LAYOUT`.
