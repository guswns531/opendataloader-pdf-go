# GAP-RO-INTERNAL-GUTTER-STACK

## Category

reading-order-parity

## Why this gap exists

On mixed-layout academic pages, some narrow gutter labels / figure-adjacent words are not just standalone sidebar objects. They are grouped into the same intermediate text-line / list structures as nearby body content before final reading-order sorting. That means edge-sidebar ordering fixes alone cannot fully remove the leak.

## Reference behavior (Java)

The Java implementation does not weave narrow gutter-stack tokens like `west`, `united`, `arsenal`, `football`, `manchester`, `messageid`, `briogestone`, and `contracers` into the main body flow on `samples/pdf/1901.03003.pdf` pages 12-13.

## Current Go behavior

Go markdown for `samples/pdf/1901.03003.pdf` pages 12-13 still includes those tokens inline or as bullets before and around the page 12-13 conclusion flow, indicating the leak happens before or during intermediate semantic grouping.

## Impact

- visible markdown corruption on a real fixture
- residual T17 parity gap after the kept marginal-sidebar sorter work
- likely harms reading-order benchmark quality on mixed-layout pages with figure-adjacent gutter text

## Candidate source files

### Go
- `go/internal/processors/text_line_processor.go`
- `go/internal/processors/document_processor.go`
- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
- nearby focused tests only if directly needed

### Java
- matching text-line / paragraph / reading-order stages used before final page ordering

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

1. The above leaked gutter-stack tokens are removed from the main article flow or moved materially closer to Java behavior.
2. The fix is backed by focused regression coverage.
3. Existing kept reading-order regressions continue to pass.
4. `cd go && GOCACHE=/tmp/odl-gocache go build ./...` passes.

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

- unrelated text-decoding fixes
- broad benchmark sweeps
- serializer cleanup beyond what the fixture proves
