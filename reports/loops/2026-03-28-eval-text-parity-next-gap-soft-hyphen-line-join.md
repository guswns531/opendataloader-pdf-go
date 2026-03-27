## Gap ID

`GAP-TEXT-SOFT-HYPHEN-LINE-JOIN`

## Date

`2026-03-28`

## Decision

`KEEP`

## Scope

- Verified the single text-parity gap for discretionary wrapped-line hyphen joining on `samples/pdf/1901.03003.pdf`.
- Did not touch reading-order / XYCut code.

## Existing implementation confirmed

- `go/internal/generators/markdown/markdown_generator.go` already joins paragraph lines by removing a trailing discretionary line-wrap hyphen only when:
  - the accumulated paragraph text ends in `-`
  - the continuation starts with a lowercase letter
  - the trailing fragment matches the discretionary hyphen pattern
- This preserves genuine compounds such as `attention-based`.
- The tracked implementation is already present in `HEAD` via commit `8cd1568` (`gap(text): join discretionary wrapped hyphen fragments`).

## Evidence before

- The stale earlier loop report captured the original failing output:
  - `/tmp/odl-sample-1901-next-gap/1901.03003.md:92`
  - `41, 45, 50] have achieved notable success. More- over, methods based on convolutional neural networks [3, 22, 50] have been broadly applied. Integrating`

## Evidence after

- Fresh local render now shows the target sentence joined correctly:
  - `/tmp/odl-sample-1901-soft-hyphen/1901.03003.md:92`
  - `41, 45, 50] have achieved notable success. Moreover, methods based on convolutional neural networks [3, 22, 50] have been broadly applied. Integrating`
- Fresh render also preserves the genuine compound:
  - `/tmp/odl-sample-1901-soft-hyphen/1901.03003.md:23`
  - `... a multi-object rectiﬁed attention network (MORAN) ... an attention-based sequence recognition`
- No `More- over` or `attentionbased` artifact is present in the rendered markdown.

## Focused regression coverage

- `go/internal/generators/markdown/markdown_generator_test.go`
  - `TestRenderLinesJoinsDiscretionaryLineWrapHyphen`
  - `TestRenderLinesPreservesRealCompoundHyphen`
- `go/internal/processors/document_processor_test.go`
  - `TestProcessJavaDocumentJoinsDiscretionarySoftHyphenLineWrapInFixture`

## Validation

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/generators/markdown -run 'TestRenderLines(JoinsDiscretionaryLineWrapHyphen|PreservesRealCompoundHyphen)$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -run 'TestProcessJavaDocumentJoinsDiscretionarySoftHyphenLineWrapInFixture$' -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-sample-1901-soft-hyphen --format markdown --image-output off --quiet
rg -n "Moreover|More- over|attention-based|attentionbased" /tmp/odl-sample-1901-soft-hyphen/1901.03003.md -n -C 1
```

Results:

- `pkg/pdfbox` and `tests/unit/pdfbox`: passed
- focused markdown generator regressions: passed
- focused processor fixture regression: passed
- `go build ./...`: passed
- fresh sample render: passed

## Notes

- `cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./tests/unit/processors/... -count=1` is not a reliable validation command in this sandbox because `tests/unit/processors` includes an unrelated `httptest.NewServer` case that fails to bind a local port (`operation not permitted`). The gap-specific processor regression itself passes.

## KEEP / SPLIT / NEEDS-HUMAN-REVIEW

- `KEEP`
