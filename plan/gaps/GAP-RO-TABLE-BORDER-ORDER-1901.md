# GAP-RO-TABLE-BORDER-ORDER-1901

## Category

reading-order-parity

## Why this gap exists

The 2026-03-28 `GAP-RO-EXTRACTOR-CHUNK-ORDER-1901` split proved the first MORAN page-12 contamination is introduced after extractor/materialization and specifically after Go `TableBorderProcessor.Process(...)`.

Go currently globally re-sorts the page result by Y/X before returning from `processNode(...)`, while Java preserves encounter order in `newContents` and returns it directly.

## Reference behavior (Java)

Java `TableBorderProcessor.processTableBorders(...)` preserves encounter order for non-table content and inserted table objects. It does not apply a final page-level Y/X sort before returning.

## Current Go behavior

Go `TableBorderProcessor.processNode(...)` appends semantic tables, remaining untouched elements, and split remainders, then globally re-sorts the entire result slice by Y descending / X ascending before returning.

On `samples/pdf/1901.03003.pdf` page 12, that stage is the first point where gutter tokens such as `west`, `football`, `briogestone`, and `contracers` become adjacent to main-flow body text.

## Impact

- visible markdown reading-order corruption on a real fixture
- likely upstream root cause for a remaining T17-family parity break
- may be fixable by aligning Go table-border return ordering with Java instead of layering more late heuristics

## Candidate source files

### Go
- `go/internal/processors/table_border_processor.go`
- `go/internal/processors/document_processor.go`
- focused tests under `go/internal/processors/`

### Java
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TableBorderProcessor.java`

## Fixtures / Reproduction

Primary fixture:
- `samples/pdf/1901.03003.pdf` page 12 (usually reproduced via `--pages 12-13`)

## Acceptance criteria

1. Evidence shows whether the final sort in Go `TableBorderProcessor` is the concrete parity break versus Java encounter order.
2. If safe, a narrow change removes or constrains the Go re-sort and materially improves MORAN page-12 output.
3. Focused regression coverage is added for the actual table-border ordering behavior.
4. Nearby processor tests still pass.
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

- unrelated extractor changes
- broader XYCut / paragraph / list heuristics
- full benchmark sweeps
- general table-detection redesign
