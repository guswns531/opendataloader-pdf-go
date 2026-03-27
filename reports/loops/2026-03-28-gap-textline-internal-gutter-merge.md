# GAP-TEXTLINE-INTERNAL-GUTTER-MERGE

Result: `KEEP`

Gap family:
- parent: `GAP-RO-INTERNAL-GUTTER-STACK`
- child: `GAP-TEXTLINE-INTERNAL-GUTTER-MERGE`

What changed:
- added a guarded same-baseline merge check in `go/internal/processors/text_line_processor.go`
- prevented very narrow chunks separated by a large internal gutter gap from being fused into a wide body line
- added a focused regression in `go/tests/unit/processors/text_line_processor_test.go`

Evidence before:
- exploratory retry work had shown `contracers` could be split away from `tiﬁed attention network...`, proving a real pre-semantic line-merge issue on `samples/pdf/1901.03003.pdf` pages 12-13
- baseline markdown still leaked `west`, `united`, `arsenal`, `football`, `manchester messageid`, `briogestone`, and `contracers`

Evidence after:
- focused regression now proves the narrow gutter chunk no longer merges into the wide body line:
  - `TestTextLineProcessorKeepsNarrowInternalGutterChunkOutOfWideBodyLine`
- targeted processor tests for this geometry passed
- `cd go && GOCACHE=/tmp/odl-gocache go build ./...` passed
- fixture markdown still leaks the gutter tokens, confirming the remaining issue lives later in semantic grouping / continuation handling rather than this line-merge stage

Validation run:
- `cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors/... -run 'TestTextLineProcessorKeepsNarrowInternalGutterChunkOutOfWideBodyLine|TestTextLineProcessorSuppressesIntrawordSpaceAfterSingleLowercaseFragment' -count=1` ✅
- `cd go && GOCACHE=/tmp/odl-gocache go build ./...` ✅
- `./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet` + token grep: residual leak still present
- `cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./internal/processors/readingorder/... ./tests/unit/processors/... -count=1` ❌ due to pre-existing unrelated table/special-cluster failures, not this patch

Files changed:
- `go/internal/processors/text_line_processor.go`
- `go/tests/unit/processors/text_line_processor_test.go`

Remaining uncertainty:
- final markdown parity for the MORAN fixture is still blocked by later list/paragraph continuation behavior that absorbs already-separated gutter tokens into body flow

Follow-up recommendation:
- next packet should target `GAP-RO-LIST-INTERNAL-GUTTER-CONTINUATION`
- keep this line-stage win as an incremental parity improvement rather than rebundling it into a wider reading-order patch
