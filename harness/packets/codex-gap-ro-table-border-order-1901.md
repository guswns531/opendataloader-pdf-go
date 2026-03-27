# Codex Packet — GAP-RO-TABLE-BORDER-ORDER-1901

## Goal

Evaluate and, if clearly safe, fix the MORAN page-12 reading-order contamination introduced by Go `TableBorderProcessor.Process(...)` result ordering.

## Why this packet exists

The previous loop proved extractor/materialization order is clean and that the first bad adjacency appears immediately after Go table-border processing. Java preserves encounter order there; Go currently applies a final global Y/X sort before returning.

## Reference files

### Read first
- `plan/gaps/GAP-RO-TABLE-BORDER-ORDER-1901.md`
- `go/internal/processors/table_border_processor.go`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TableBorderProcessor.java`
- `reports/loops/2026-03-28-gap-ro-extractor-chunk-order-1901.md`

## Scope

Do exactly one narrow thing:
- determine whether the final global sort in Go `TableBorderProcessor.processNode(...)` is the parity break on MORAN page 12, and
- if evidence is strong and the change is small/safe, implement the narrowest Java-aligned fix plus focused regression coverage.

## Acceptance criteria

1. Demonstrate before/after evidence for the table-border stage ordering on MORAN page 12.
2. If you keep a code change, it must materially improve the fixture or focused adjacency regression.
3. Keep the change narrow to table-border ordering only.
4. `cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -count=1` passes.
5. `cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf` passes.

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
rm -rf /tmp/odl-moran-md /tmp/odl-moran-md-off && mkdir -p /tmp/odl-moran-md /tmp/odl-moran-md-off
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md-off --format markdown --image-output off --quiet --reading-order off
rg -n "west|football|briogestone|contracers|messageid|arsenal|united" /tmp/odl-moran-md/1901.03003.md /tmp/odl-moran-md-off/1901.03003.md
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -count=1
```

## Non-goals

- extractor/materialization changes
- XYCut or page-level reading-order fixes outside table-border stage
- broad cleanup unrelated to this gap

## Output requirements

When finished, leave:
1. code/tests only if evaluator-clean
2. a concise loop report under `reports/loops/`
3. a short summary of whether the table-border final sort was kept, removed, or constrained

When completely finished, run this command to notify me:
openclaw system event --text "Done: evaluated GAP-RO-TABLE-BORDER-ORDER-1901 in opendataloader-pdf-go" --mode now
