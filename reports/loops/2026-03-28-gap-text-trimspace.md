## Gap ID

`GAP-TEXT-TRIMSPACE`

## Date

`2026-03-28`

## Decision

`DISCARD`

## Evidence

- The current extractor already contains the trimspace-family boundary preservation helpers in [go/pkg/pdfbox/extractor/text_extractor.go](/Users/hj/.openclaw/workspace/opendataloader-pdf-go/go/pkg/pdfbox/extractor/text_extractor.go).
- Existing focused tests for trailing, leading, whitespace-only, and duplicate-boundary cases were already present and passing in [go/pkg/pdfbox/extractor/content_parser_test.go](/Users/hj/.openclaw/workspace/opendataloader-pdf-go/go/pkg/pdfbox/extractor/content_parser_test.go).
- `git log` shows prior KEEP-worthy trimspace commits already landed: `6841e53 gap(text): preserve boundary spaces across runs` and `3b8032a gap(text): preserve whitespace-only boundary runs`.

## What changed

- Added one narrow regression test in [go/pkg/pdfbox/extractor/content_parser_test.go](/Users/hj/.openclaw/workspace/opendataloader-pdf-go/go/pkg/pdfbox/extractor/content_parser_test.go) to assert that a whitespace-only trailing run after `Hello ` does not duplicate the final boundary space.

## Validation

```bash
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...
cd go && GOCACHE=/tmp/opendataloader-go-build-cache go build ./...
```

Results:

- tests: passed
- build: passed

## Follow-up

- Treat the original packet as stale rather than reopening trimspace logic.
- If outer evaluation still finds collapsed words, split that evidence into a narrower new gap tied to a concrete fixture and failure mode rather than broad trimspace handling.
