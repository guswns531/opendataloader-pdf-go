## Gap ID

`GAP-TEXT-SOFT-HYPHEN-LINE-JOIN`

## Date

`2026-03-28`

## Decision

`KEEP`

## Why this loop existed

A fresh local markdown render still showed a wrapped-line hyphenation artifact on `samples/pdf/1901.03003.pdf` page 1:

- `More- over, methods ...`

The Java-parity target is:

- `Moreover, methods ...`

This is a text-parity gap in paragraph line joining, distinct from the already-kept same-line spacing fixes.

## Evidence before

- `/tmp/odl-sample-1901-next-gap/1901.03003.md:92`
- `41, 45, 50] have achieved notable success. More- over, methods based on convolutional neural networks [3, 22, 50] have been broadly applied.`

## Change kept

Scoped markdown paragraph joining so consecutive lines can collapse a discretionary line-wrap hyphen when all of the following hold:

1. previous line ends with `-`
2. next line begins with lowercase
3. the trailing fragment looks like a simple title-case word fragment (for example `More-`)

This keeps the fix narrow enough to repair the live sample without collapsing genuine compounds like `attention-based`.

## Files changed

- `go/internal/generators/markdown/markdown_generator.go`
- `go/internal/generators/markdown/markdown_generator_test.go`
- `go/internal/processors/document_processor_test.go`

## Validation

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/generators/markdown/... ./internal/processors/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 1 --output-dir /tmp/odl-sample-1901-soft-hyphen --format markdown --image-output off --quiet
rg -n "Moreover|More- over|attention-based|attentionbased" /tmp/odl-sample-1901-soft-hyphen/1901.03003.md
```

Results:

- targeted generator + processor tests: passed
- `go build ./...`: passed
- sample render now contains `Moreover, methods ...`
- sample render still contains `attention-based sequence recognition`
- sample render no longer contains `More- over`
- sample render does not contain `attentionbased`

## Remaining uncertainty

- The heuristic intentionally only handles a narrow discretionary line-wrap family. Broader hyphenation parity may still need follow-up gaps if new fixtures show lowercase→lowercase or more complex wrap cases.

## Follow-up recommendation

Return to the next best residual parity gap after this KEEP, likely either:

- another narrow text wrap/join gap discovered from fixture diffs, or
- the next reading-order gap from the T17 family once text parity wins flatten out.
