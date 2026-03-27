# Codex Gap Task Packet — GAP-RO-PRELIST-SEED-TRACE-1901

## Role

You are implementing one scoped parity investigation in the Go port of OpenDataLoader PDF.
The Java implementation is the reference behavior unless explicitly stated otherwise.
Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-RO-PRELIST-SEED-TRACE-1901`

## Category

`reading-order-parity`

---

## Objective

Trace the exact Go object sequence on `samples/pdf/1901.03003.pdf` pages 12-13 **immediately before paragraph/list processing**, confirm where gutter tokens first become adjacent to main-flow content, and implement the smallest defensible upstream fix if the root cause is clear.

This loop is narrower than the previous seed-grouping packet:

- do **not** spend another loop on final `sortDocumentContents` / XYCut cleanup
- do **not** revisit late list-continuation heuristics unless the trace proves the pre-list sequence is still clean
- focus on the first stage where `west`, `united`, `arsenal`, `football`, `manchester`, `messageid`, `briogestone`, or `contracers` become incorrectly attached to nearby body/list candidates

---

## Proven prior evidence

From the previous split reports:

- A pre-grouping source-order tie-break did **not** change the real markdown output.
- Running with `--reading-order off` still shows the same contamination.
- Therefore the corruption exists **before** the final page-level XYCut sort.

Key contaminated output still includes:

- `- west The experiments above are all based on cropped text`
- `- [18] proposed methods to improve the performance football of multi-oriented text detection`
- `manchester messageid`
- `————— briogestone In this paper, we presented a multi-object rec-`
- `contracers\ttiﬁed attention network (MORAN) for scene text`

---

## Reference behavior (Java)

Java does not weave those gutter tokens into the same body/conclusion flow on the same fixture pages.
Use Java only as needed to justify the smallest fix.

---

## Current Go hypothesis

The bad sequence is likely formed in one of these stages before `ListProcessor.Process`:

1. `TextLineProcessor.Process` line creation / ordering
2. intermediate text-line → object conversion in `DocumentProcessor`
3. the mixed object sequence entering paragraph/list processors

---

## Relevant files

### Go
- `go/internal/processors/text_line_processor.go`
- `go/internal/processors/document_processor.go`
- `go/internal/processors/list_processor.go`
- `go/internal/processors/paragraph_processor.go`
- nearby focused tests only if directly needed

### Java
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TextLineProcessor.java`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/DocumentProcessor.java`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/ListProcessor.java`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/ParagraphProcessor.java`

### Prior reports
- `reports/loops/2026-03-28-eval-list-internal-gutter-continuation-recheck.md`
- `reports/loops/2026-03-28-gap-ro-gutter-token-seed-grouping-1901-split.md`

---

## Required work

1. Add the lightest possible temporary/focused trace or test scaffolding to inspect the Go object sequence immediately before paragraph/list processing on pages 12-13.
2. Determine whether the relevant gutter tokens are already adjacent to body/list candidates at that stage.
3. Compare with Java only as far as needed to justify the first wrong stage.
4. If a small upstream fix becomes clear, implement it with focused regression coverage.
5. If the root cause is still not clear, leave the tree clean and record a sharper split point.

---

## Acceptance criteria

A successful KEEP requires **all** of:

1. the first wrong stage is identified with concrete evidence,
2. a scoped code fix improves the real fixture or a focused pre-list regression proves the wrong adjacency no longer occurs,
3. targeted tests pass,
4. `cd go && GOCACHE=/tmp/odl-gocache go build ./...` passes.

A successful SPLIT requires:

1. the tree is left clean except for loop artifacts,
2. the report clearly names the next narrower root cause.

---

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -count=1
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
rm -rf /tmp/odl-moran-md /tmp/odl-moran-md-off && mkdir -p /tmp/odl-moran-md /tmp/odl-moran-md-off
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md-off --format markdown --image-output off --quiet --reading-order off
rg -n "west|united|arsenal|football|manchester|messageid|briogestone|contracers" /tmp/odl-moran-md/1901.03003.md /tmp/odl-moran-md-off/1901.03003.md
sed -n '60,150p' /tmp/odl-moran-md/1901.03003.md
sed -n '60,150p' /tmp/odl-moran-md-off/1901.03003.md
```

If sandboxed tests hit the known `httptest` listener restriction, use focused processor tests plus outer validation rather than broad noisy sweeps.

---

## Non-goals

Do not mix in:

- broad benchmark sweeps
- unrelated text-decoding fixes
- another speculative final-sort tie-break
- serializer cleanup beyond what the fixture proves

---

## Output format

When done, report:

1. files changed
2. what stage was proven wrong (or still uncertain)
3. tests/commands run and their result
4. whether result is KEEP or SPLIT
5. exact next follow-up if split

When completely finished, run this command to notify me:

```bash
openclaw system event --text "Done: evaluated GAP-RO-PRELIST-SEED-TRACE-1901 in opendataloader-pdf-go" --mode now
```
