# Codex Gap Task Packet — GAP-RO-GUTTER-TOKEN-SEED-GROUPING-1901

## Role

You are implementing one scoped parity gap in the Go port of OpenDataLoader PDF.
The Java implementation is the reference behavior unless explicitly stated otherwise.
Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-RO-GUTTER-TOKEN-SEED-GROUPING-1901`

## Category

`reading-order-parity`

---

## Objective

Find and fix the earliest defensible stage that causes narrow gutter / figure-adjacent tokens to be seeded into the main article flow on `samples/pdf/1901.03003.pdf` pages 12-13.

This packet is **not** limited to late list continuation. The current hypothesis is that the corruption may already exist in one of these stages before final list construction:

- text-line merging / splitting
- semantic paragraph or intermediate grouping
- document-order seeding before list construction

Keep the patch small and root-cause-driven.

---

## Evidence

From `reports/loops/2026-03-28-eval-list-internal-gutter-continuation-recheck.md`:

- The previous focused recheck proved a narrow `ListProcessor.isListContinuation` tweak did **not** change the real `1901.03003` markdown output.
- The output still contains both standalone gutter tokens and contaminated continuation lines:
  - `west`
  - `united`
  - `arsenal`
  - `football`
  - `manchester messageid`
  - `contracers`
  - `- west The experiments above are all based on cropped text`
  - `- in more application scenarios, irregular and multi-`
  - `- instance, Liu et al. [29] and Ch'ng et al. [7] released`
  - `————— briogestone In this paper, we presented a multi-object rec-`
- That means the packet should target the upstream seed/grouping stage that feeds list construction the wrong sequence.

---

## Reference behavior (Java)

Java does not weave those gutter tokens into the main article flow on the MORAN fixture pages 12-13.

---

## Current Go behavior

Go still produces corrupted markdown on those pages, suggesting the bad ordering/grouping exists before or during intermediate structure formation.

---

## Relevant files

### Go
- `go/internal/processors/text_line_processor.go`
- `go/internal/processors/document_processor.go`
- `go/internal/processors/hybrid_document_processor.go`
- `go/internal/processors/readingorder/xycut_plus_plus_sorter.go`
- nearby focused tests only if directly needed

### Java
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TextLineProcessor.java`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/DocumentProcessor.java`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/HybridDocumentProcessor.java`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/readingorder/XYCutPlusPlusSorter.java`

### Prior reports
- `reports/loops/2026-03-28-eval-list-internal-gutter-continuation-recheck.md`
- `plan/gaps/GAP-RO-INTERNAL-GUTTER-STACK.md`

---

## Required work

1. Trace the target pages through the earliest practical Go stage before list construction.
2. Determine whether the gutter tokens are already interleaved into the main flow before list processing.
3. Compare with Java behavior only as far as needed to justify the smallest defensible upstream fix.
4. Implement a scoped fix at the actual root-cause stage.
5. Add focused regression coverage for that root cause.
6. Do not widen the fix beyond this MORAN-shaped failure mode unless the Java parity evidence clearly supports it.

---

## Acceptance criteria

1. Trace or regression evidence identifies the actual upstream stage creating the bad seed/grouping.
2. The fixture output materially improves on `1901.03003` pages 12-13, or a focused geometry regression proves the corruption no longer occurs.
3. Focused regression coverage is added.
4. Existing reading-order regressions continue to pass.
5. `cd go && GOCACHE=/tmp/odl-gocache go build ./...` passes.

---

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./internal/processors/readingorder/... ./tests/unit/processors/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
rm -rf /tmp/odl-moran-md && mkdir -p /tmp/odl-moran-md
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
rg -n "west|united|arsenal|football|manchester|messageid|briogestone|contracers" /tmp/odl-moran-md/1901.03003.md
sed -n '60,150p' /tmp/odl-moran-md/1901.03003.md
```

---

## Non-goals

Do not mix in:

- unrelated text-decoding fixes
- broad benchmark sweeps
- speculative mixed-layout/sidebar heuristics without fixture proof
- serializer cleanup beyond what this fixture proves

---

## Output format

When done, report:

1. files changed
2. behavior changed
3. tests/commands run and their result
4. remaining uncertainty
5. whether the gap appears fully fixed, partially fixed, or needs follow-up

When completely finished, run this command to notify me:

```bash
openclaw system event --text "Done: evaluated GAP-RO-GUTTER-TOKEN-SEED-GROUPING-1901 in opendataloader-pdf-go" --mode now
```
