# Codex Gap Task Packet — GAP-RO-TEXTLINE-SOURCE-ORDER-1901

## Role

You are implementing one scoped parity investigation/fix in the Go port of OpenDataLoader PDF.
The Java implementation is the reference behavior unless explicitly stated otherwise.
Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-RO-TEXTLINE-SOURCE-ORDER-1901`

## Category

`reading-order-parity`

---

## Objective

Determine whether the first parity break on `samples/pdf/1901.03003.pdf` page 12 is caused by:

1. `DocumentProcessor` splitting text chunks out of the mixed object stream before text-line formation, or
2. the explicit global geometry sort inside Go `TextLineProcessor.Process`.

Implement the smallest defensible KEEP-sized fix only if the trace makes the root cause clear.

This is the direct follow-up to `reports/loops/2026-03-28-gap-ro-prelist-seed-trace-1901.md`.

---

## Proven prior evidence

From the previous split report:

- the bad adjacency already exists in Go `TextLineProcessor.Process` output on zero-based page `11` / PDF page `12`
- the contamination is present before `textLinesToObjects`, `sortObjects`, `ListProcessor`, and paragraph conversion
- representative wrong order includes:
  - `west`
  - `The experiments above are all based on cropped text`
  - `...`
  - `football`
  - `of multi-oriented text detection. Therefore, scene`
  - `...`
  - `briogestone`
  - `In this paper, we presented a multi-object rec-`
  - `contracers`
  - `tiﬁed attention network (MORAN) for scene text`

Java-side prior evidence:

- Java `DocumentProcessor.processDocument(...)` passes the mixed `pageContents` stream into `TextLineProcessor.processTextLines(...)`
- Go currently extracts text chunks, forms lines from text chunks only, and returns lines globally sorted by baseline/X

---

## Reference behavior (Java)

Java should preserve source-order separation well enough that the gutter tokens above do not become adjacent to the body/conclusion flow on the fixture.
Use Java only as needed to justify the first wrong stage and the smallest safe fix.

---

## Relevant files

### Go
- `go/internal/processors/text_line_processor.go`
- `go/internal/processors/document_processor.go`
- `go/internal/processors/document_processor_test.go`
- focused nearby tests only if directly needed

### Java
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TextLineProcessor.java`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/DocumentProcessor.java`

### Prior report
- `reports/loops/2026-03-28-gap-ro-prelist-seed-trace-1901.md`

---

## Required work

1. Add the lightest possible temporary/focused trace or regression scaffolding to compare:
   - the original mixed object order entering Go text-line formation,
   - the order after Go `TextLineProcessor.Process`, and
   - Java behavior only as far as needed to identify the first parity break.
2. Decide whether the first break is caused mainly by `splitTextChunks(...)` losing mixed-order context or by the explicit geometry sort inside Go `TextLineProcessor.Process`.
3. If a small safe fix becomes clear, implement it with focused regression coverage on the MORAN page-12 contamination pattern.
4. If the root cause is still not narrow enough for a safe KEEP, leave the tree clean except loop artifacts and write a precise SPLIT report.

---

## Acceptance criteria

A successful KEEP requires **all** of:

1. the first parity break is identified with concrete evidence,
2. a scoped code fix prevents or materially improves the MORAN page-12 bad adjacency,
3. targeted tests pass,
4. `cd go && GOCACHE=/tmp/odl-gocache go build ./...` passes,
5. the tree is commit-ready.

A successful SPLIT requires:

1. the tree is left clean except for loop artifacts,
2. the report clearly names the next narrower source-order root cause.

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

If sandboxed tests hit known restrictions, use focused processor tests plus outer validation rather than broad noisy sweeps.

---

## Non-goals

Do not mix in:

- broad benchmark sweeps
- unrelated text-decoding fixes
- late list/paragraph heuristics unless the text-line/source-order evidence directly proves that is still the first wrong stage
- serializer cleanup beyond what the fixture proves

---

## Output format

When done, report:

1. files changed
2. what first parity break was proven (or still uncertain)
3. tests/commands run and their result
4. whether result is KEEP or SPLIT
5. exact next follow-up if split

When completely finished, run this command to notify me:

```bash
openclaw system event --text "Done: evaluated GAP-RO-TEXTLINE-SOURCE-ORDER-1901 in opendataloader-pdf-go" --mode now
```
