# Loop Report

## Gap ID

`GAP-RO-EXTRACTOR-CHUNK-ORDER-1901`

## Date

`2026-03-28`

## Decision

`SPLIT`

## Why this loop existed

- Determine whether the first MORAN page-12 parity break is already present in Go extractor/materialization chunk order before text-line formation.
- Apply only a small safe upstream fix if the extractor/materialization stage is clearly the first break.

## What changed

- No source changes kept.
- I added a temporary focused trace in `go/internal/processors/document_processor_test.go`, ran it against `samples/pdf/1901.03003.pdf` page `12`, captured the stage-by-stage neighborhoods below, then reverted the test.

## What the trace proved

- Go raw extractor order is **not** the first break on zero-based page `11` / PDF page `12`.
- `extractor.ExtractTextChunks(...)` and `loadDocument(...).Pages[0].Chunks` matched for the target gutter tokens and still preserved separator/duplicate context rather than the contaminated body adjacency:
  - `west`: `Prediction | —————————————————————– | west | west | —————`
  - `football`: `arsenal | ————— | football | football | —————`
  - `briogestone`: `essageid | ————— | briogestone | contracers | —————————————————————–`
  - `contracers`: `————— | briogestone | contracers | —————————————————————– | Figure 9...`
- The first contaminated adjacency appeared one stage later, after `TableBorderProcessor.Process(...)` and before text-line formation:
  - `west` -> `The experiments above are all based on cropped text`
  - `football` -> `of multi-oriented text detection. Therefore, scene`
  - `briogestone` -> `In this paper, we presented a multi-object rec-`
  - `contracers` -> `tiﬁed attention network (MORAN) for scene text`
- That same contaminated local order then flowed unchanged into `groupTextChunksIntoLines(...)`.

## Java comparison used

- Go [table_border_processor.go](/Users/hj/.openclaw/workspace/opendataloader-pdf-go/go/internal/processors/table_border_processor.go#L117) globally re-sorts the page result by Y/X before returning it.
- Java [TableBorderProcessor.java](/Users/hj/.openclaw/workspace/opendataloader-pdf-go/java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TableBorderProcessor.java#L68) preserves encounter order in `newContents` and returns it directly at [TableBorderProcessor.java](/Users/hj/.openclaw/workspace/opendataloader-pdf-go/java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TableBorderProcessor.java#L96).
- Inference from those source comparisons plus the temporary Go trace: the page-12 contamination is not an extractor/materialization defect; it is introduced by Go’s pre-text-line table-border stage reordering.

## Why this is a split, not a keep

- The packet scoped this loop to extractor/materialization order.
- The proved first break is a different upstream component than the packet hypothesis.
- Removing or changing the global sort inside `TableBorderProcessor` would be a broader behavioral change affecting all border-table pages, so it is not a small safe fix to keep under this packet without a narrower follow-up.

## Validation

Temporary focused trace run:

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -run TestTraceExtractorChunkOrderOnMORANPage12 -count=1 -v
```

Result:

- passed and showed extractor/materialization order remained clean while `TableBorderProcessor` introduced the first bad local adjacency
- temporary trace reverted cleanly afterward

Packet validation run on the clean tree:

```bash
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
mkdir -p /tmp/odl-moran-md-1901 /tmp/odl-moran-md-off-1901
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md-1901 --format markdown --image-output off --quiet
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md-off-1901 --format markdown --image-output off --quiet --reading-order off
rg -n "west|football|briogestone|contracers|messageid|arsenal|united" /tmp/odl-moran-md-1901/1901.03003.md /tmp/odl-moran-md-off-1901/1901.03003.md
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -count=1
```

Results:

- build: passed
- fixture markdown reproduction: passed and still reproduced the known contamination in both default and `--reading-order off` runs
- `go test ./internal/processors -count=1`: passed

## Final tree state

- clean except this loop report

## Exact next follow-up

- New narrow gap: trace and evaluate Go `TableBorderProcessor.Process(...)` result-order parity against Java on MORAN page `12`, specifically the global post-processing sort at [table_border_processor.go](/Users/hj/.openclaw/workspace/opendataloader-pdf-go/go/internal/processors/table_border_processor.go#L117).
- Acceptance for that follow-up should decide whether removing or constraining that sort fixes the page-12 contamination without regressing real border-table handling.
