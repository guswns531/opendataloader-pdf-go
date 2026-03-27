# GAP-RO-GUTTER-TOKEN-SEED-GROUPING-1901

## Category

reading-order-parity

## Why this gap exists

On `samples/pdf/1901.03003.pdf` pages 12-13, narrow gutter / figure-adjacent tokens such as `west`, `united`, `arsenal`, `football`, `manchester`, `messageid`, `briogestone`, and `contracers` still leak into the main article flow.

Recent rechecks showed this is not convincingly explained by late list continuation logic alone. The corruption likely begins earlier, when pre-list text lines, semantic groups, or document-order seeds are formed.

## Reference behavior (Java)

The Java pipeline does not weave those gutter tokens into the main body flow on the same fixture pages. The offending tokens do not appear as standalone bullet/list contaminations in the corresponding article flow.

## Current Go behavior

Go markdown for pages 12-13 still contains:

- standalone gutter-like tokens (`west`, `united`, `arsenal`, `football`, `manchester messageid`, `contracers`)
- contaminated continuation lines such as:
  - `- west The experiments above are all based on cropped text`
  - `- in more application scenarios, irregular and multi-`
  - `- instance, Liu et al. [29] and Ch'ng et al. [7] released`
  - `————— briogestone In this paper, we presented a multi-object rec-`

## Impact

- visible markdown corruption on a real fixture
- remaining T17 parity gap after earlier XYCut/sidebar keeps
- likely indicates upstream grouping/seeding mismatch, not just final reading-order cleanup

## Candidate source files

### Go
- `go/internal/processors/text_line_processor.go`
- `go/internal/processors/document_processor.go`
- `go/internal/processors/hybrid_document_processor.go`
- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`

### Java
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TextLineProcessor.java`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/DocumentProcessor.java`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/HybridDocumentProcessor.java`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/readingorder/XYCutPlusPlusSorter.java`

## Fixtures / Reproduction

Primary fixture:
- `samples/pdf/1901.03003.pdf` pages 12-13

Repro:

```bash
rm -rf /tmp/odl-moran-md && mkdir -p /tmp/odl-moran-md
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
rg -n "west|united|arsenal|football|manchester|messageid|briogestone|contracers" /tmp/odl-moran-md/1901.03003.md
sed -n '60,150p' /tmp/odl-moran-md/1901.03003.md
```

## Acceptance criteria

1. Trace or regression evidence identifies whether the corruption is already present before list construction.
2. The smallest defensible upstream fix materially improves the fixture output or a focused geometry regression proves the bad seed/group no longer happens.
3. Focused regression coverage is added where the root cause is fixed.
4. Existing reading-order regressions continue to pass.
5. `cd go && GOCACHE=/tmp/odl-gocache go build ./...` passes.

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./internal/processors/readingorder/... ./tests/unit/processors/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
rm -rf /tmp/odl-moran-md && mkdir -p /tmp/odl-moran-md
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
rg -n "west|united|arsenal|football|manchester|messageid|briogestone|contracers" /tmp/odl-moran-md/1901.03003.md
sed -n '60,150p' /tmp/odl-moran-md/1901.03003.md
```

## Non-goals

- unrelated text decoding / TJ / ToUnicode work
- broad benchmark sweeps
- speculative XYCut heuristic widening without fixture evidence
- serializer cleanup beyond what the fixture proves
