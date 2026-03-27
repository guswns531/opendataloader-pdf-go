# GAP-RO-EXTRACTOR-CHUNK-ORDER-1901

## Category

reading-order-parity

## Why this gap exists

On `samples/pdf/1901.03003.pdf` page 12, gutter / figure-adjacent tokens such as `west`, `football`, `briogestone`, and `contracers` are already adjacent to main-flow body/conclusion chunks before Go text-line formation begins.

The latest split reports ruled out these downstream suspects for the first break on that page:

- final page-level reading-order / XYCut sorting
- pre-list / paragraph processors
- `DocumentProcessor.splitTextChunks(...)`
- the explicit geometry sort inside Go `TextLineProcessor.Process(...)`

That leaves an upstream source-order/materialization problem in the chunk feed itself.

## Reference behavior (Java)

Java does not weave those gutter tokens into the same body/conclusion flow on the MORAN fixture page. The pre-text-line page content order preserves enough separation that those tokens do not seed the article flow.

## Current Go behavior

Tracing on zero-based page `11` / PDF page `12` showed representative adjacent chunk pairs already present before text-line processing:

- `west` → `The experiments above are all based on cropped text`
- `football` → `of multi-oriented text detection. Therefore, scene`
- `briogestone` → `In this paper, we presented a multi-object rec-`
- `contracers` → `tiﬁed attention network (MORAN) for scene text`

## Impact

- visible markdown corruption on a real reading-order fixture
- proves remaining T17 work is upstream of list / paragraph / XYCut cleanup
- likely requires a smaller extractor/input-order parity fix rather than more late heuristics

## Candidate source files

### Go
- `go/internal/extractor/` (relevant text chunk extraction/materialization code)
- `go/internal/processors/document_processor.go`
- `go/internal/processors/text_line_processor.go`
- `go/internal/models/page.go`
- nearby focused tests only if directly needed

### Java
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/DocumentProcessor.java`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TextLineProcessor.java`
- relevant Java extraction/materialization code that determines page text chunk order

## Fixtures / Reproduction

Primary fixture:
- `samples/pdf/1901.03003.pdf` page 12 (usually reproduced via `--pages 12-13`)

Repro:

```bash
rm -rf /tmp/odl-moran-md /tmp/odl-moran-md-off && mkdir -p /tmp/odl-moran-md /tmp/odl-moran-md-off
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md-off --format markdown --image-output off --quiet --reading-order off
rg -n "west|football|briogestone|contracers|messageid|arsenal|united" /tmp/odl-moran-md/1901.03003.md /tmp/odl-moran-md-off/1901.03003.md
```

## Acceptance criteria

1. Concrete evidence identifies whether the first parity break is already present in Go `page.Chunks` / extractor materialization order before text-line processing.
2. If a small safe root-cause fix is clear, the real MORAN fixture materially improves or a focused regression proves the bad adjacency no longer occurs.
3. Focused regression coverage is added for the actual upstream cause.
4. Existing nearby reading-order/text-line checks continue to pass.
5. `cd go && GOCACHE=/tmp/odl-gocache go build ./...` passes.

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md-off --format markdown --image-output off --quiet --reading-order off
rg -n "west|football|briogestone|contracers|messageid|arsenal|united" /tmp/odl-moran-md/1901.03003.md /tmp/odl-moran-md-off/1901.03003.md
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -count=1
```

## Non-goals

- more speculative late list/paragraph heuristics
- broad benchmark sweeps
- unrelated text decoding fixes
- serializer cleanup beyond what this fixture proves
