# Gap Packet — GAP-RO-INTERNAL-GUTTER-STACK

## Gap ID

`GAP-RO-INTERNAL-GUTTER-STACK`

## Category

reading-order

## Reference behavior

Java does not weave narrow figure-adjacent gutter labels into the main prose on `samples/pdf/1901.03003.pdf` pages 12-13.

## Current Go behavior

Go still emits leaked gutter-stack tokens in markdown around the page 12-13 conclusion flow, including:

- `west`
- `united`
- `arsenal`
- `football`
- `manchester`
- `messageid`
- `briogestone`
- `contracers`

The previous kept marginal-sidebar sorter fix was not enough; the leak likely happens before or during intermediate text-line / semantic grouping.

## Impact

- real fixture corruption remains on a high-signal academic sample
- residual T17 reading-order parity gap

## Relevant files

Go:
- `go/internal/processors/text_line_processor.go`
- `go/internal/processors/document_processor.go`
- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
- nearby focused tests only if directly required

Java:
- matching pre-semantic grouping and reading-order code paths

## Fixture / evidence

```bash
rm -rf /tmp/odl-moran-md && mkdir -p /tmp/odl-moran-md
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
rg -n "west|united|arsenal|football|manchester|messageid|briogestone|contracers" /tmp/odl-moran-md/1901.03003.md
sed -n '60,150p' /tmp/odl-moran-md/1901.03003.md
```

## Acceptance criteria

1. The leaked gutter-stack tokens above no longer appear woven into the main body flow, or the output moves materially closer to Java on the fixture.
2. Focused regression coverage is added.
3. Existing reading-order regressions continue to pass.
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

- unrelated text-parity work
- broad benchmark runs
- speculative cleanup without fixture evidence

## Deliverable

Implement one narrow fix plus regression coverage, then report files changed, validation run, remaining uncertainty, and whether this gap is KEEP / DISCARD / SPLIT.
