# Codex Gap Task Packet — GAP-RO-EXTRACTOR-CHUNK-ORDER-1901

## Role

You are implementing one scoped parity investigation/fix in the Go port of OpenDataLoader PDF.
The Java implementation is the reference behavior unless explicitly stated otherwise.
Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-RO-EXTRACTOR-CHUNK-ORDER-1901`

## Category

`reading-order-parity`

---

## Objective

Determine whether the first parity break on `samples/pdf/1901.03003.pdf` page 12 is already present in the Go page chunk feed before text-line formation, and fix it only if a small root-cause patch is clearly justified.

This is the direct upstream follow-up to `GAP-RO-TEXTLINE-SOURCE-ORDER-1901`.

The prior split already ruled out these stages as the first break on page 12:

- final page-level reading-order / XYCut sorting
- pre-list / paragraph processors
- `DocumentProcessor.splitTextChunks(...)`
- the explicit geometry sort inside Go `TextLineProcessor.Process(...)`

So this loop must move upstream into extractor/materialization order.

---

## Proven prior evidence

From `reports/loops/2026-03-28-gap-ro-textline-source-order-1901.md`:

- the page object stream entering Go text-line formation already contains contaminated adjacent chunk pairs on zero-based page `11` / PDF page `12`
- representative adjacent pairs already present before text-line processing:
  - `west` -> `The experiments above are all based on cropped text`
  - `football` -> `of multi-oriented text detection. Therefore, scene`
  - `briogestone` -> `In this paper, we presented a multi-object rec-`
  - `contracers` -> `tiﬁed attention network (MORAN) for scene text`
- `splitTextChunks(...)` preserved that local order and had no non-text separators left to discard on this page
- `TextLineProcessor.Process(...)` preserved those same representative adjacencies rather than creating them

Therefore the next narrow question is whether Go extraction/materialization is already emitting chunks in the wrong local order compared with Java’s pre-text-line page content order.

---

## Reference behavior (Java)

Java should preserve enough pre-text-line separation that those gutter tokens do not seed the body/conclusion flow on the MORAN page.
Use Java only as far as needed to justify the first upstream break and the smallest safe fix.

---

## Relevant files

### Go
- `go/internal/extractor/` (relevant text chunk extraction/materialization code)
- `go/internal/processors/document_processor.go`
- `go/internal/processors/text_line_processor.go`
- `go/internal/models/page.go`
- focused nearby tests only if directly needed

### Java
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/DocumentProcessor.java`
- `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TextLineProcessor.java`
- relevant Java extraction/materialization code that determines page content/chunk order

### Prior reports
- `reports/loops/2026-03-28-gap-ro-prelist-seed-trace-1901.md`
- `reports/loops/2026-03-28-gap-ro-textline-source-order-1901.md`

---

## Required work

1. Add the lightest possible temporary/focused trace or regression scaffolding to inspect the original Go chunk/materialization order for PDF page 12 before text-line processing.
2. Compare that order against Java only as far as needed to identify whether Go extraction/materialization is already the first parity break.
3. If a small safe root-cause fix becomes clear, implement it with focused regression coverage on the MORAN page-12 contamination pattern.
4. If the root cause is still not narrow enough for a safe KEEP, leave the tree clean except loop artifacts and write a precise SPLIT report naming the next upstream component.

---

## Acceptance criteria

A successful KEEP requires **all** of:

1. the first upstream parity break is identified with concrete evidence,
2. a scoped code fix materially improves the MORAN fixture or a focused regression proves the bad adjacency no longer occurs,
3. targeted tests pass,
4. `cd go && GOCACHE=/tmp/odl-gocache go build ./...` passes,
5. the tree is commit-ready.

A successful SPLIT requires:

1. the tree is left clean except for loop artifacts,
2. the report clearly names the next narrower extractor/materialization root cause.

---

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
rm -rf /tmp/odl-moran-md /tmp/odl-moran-md-off && mkdir -p /tmp/odl-moran-md /tmp/odl-moran-md-off
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md --format markdown --image-output off --quiet
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --pages 12-13 --output-dir /tmp/odl-moran-md-off --format markdown --image-output off --quiet --reading-order off
rg -n "west|football|briogestone|contracers|messageid|arsenal|united" /tmp/odl-moran-md/1901.03003.md /tmp/odl-moran-md-off/1901.03003.md
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors -count=1
```

If sandboxed tests hit known restrictions, use focused tests plus outer validation rather than broad noisy sweeps.

---

## Non-goals

Do not mix in:

- more speculative late list/paragraph heuristics
- broad benchmark sweeps
- unrelated text decoding fixes
- serializer cleanup beyond what this fixture proves

---

## Output format

When done, report:

1. files changed
2. what first upstream parity break was proven (or still uncertain)
3. tests/commands run and their result
4. whether result is KEEP or SPLIT
5. exact next follow-up if split

When completely finished, run this command to notify me:

```bash
openclaw system event --text "Done: evaluated GAP-RO-EXTRACTOR-CHUNK-ORDER-1901 in opendataloader-pdf-go" --mode now
```
