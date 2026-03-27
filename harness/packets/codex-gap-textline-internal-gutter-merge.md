# Codex Gap Task Packet — GAP-TEXTLINE-INTERNAL-GUTTER-MERGE

## Role

You are implementing one scoped parity gap in the Go port of OpenDataLoader PDF.
The Java implementation is the reference behavior unless explicitly stated otherwise.
Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-TEXTLINE-INTERNAL-GUTTER-MERGE`

## Category

`reading-order-parity`

---

## Objective

Prevent narrow internal-gutter tokens from being merged into wide body text lines before later semantic grouping on `samples/pdf/1901.03003.pdf` pages 12-13.

This packet is only about the pre-semantic text-line merge stage.

---

## Evidence

From `reports/loops/2026-03-28-gap-ro-internal-gutter-stack-retry-split.md`:

- Even after exploratory work, fixture markdown still leaked tokens such as `west`, `united`, `arsenal`, `football`, `manchester`, `messageid`, `briogestone`, and `contracers`.
- The retry established two distinct residual mechanisms.
- One of them is a text-line stage issue: same-baseline internal gutter tokens are merged into wide body-facing lines too early.
- Example evidence:
  - `contracers` was separable from `tiﬁed attention network...` with a narrow line-splitting patch.
  - `west` was separable from part of the list body.

---

## Reference behavior (Java)

Java does not merge these narrow gutter tokens into the main body line structures in the same way on the MORAN fixture pages 12-13.

---

## Current Go behavior

Go still merges some narrow gutter-adjacent tokens into broad body lines before list/paragraph grouping, which then makes later reading-order cleanup harder.

---

## Relevant files

### Go
- `go/internal/processors/text_line_processor.go`
- nearby focused tests if directly needed

### Evidence/report
- `reports/loops/2026-03-28-gap-ro-internal-gutter-stack-retry-split.md`

---

## Required work

- inspect the current same-baseline / same-line merge logic in `text_line_processor.go`
- make the smallest defensible change that prevents obvious narrow gutter tokens from being fused into wide body lines on the target fixture
- add focused regression coverage for the line-merging behavior if possible
- keep the patch limited to this pre-semantic text-line merge mechanism

---

## Acceptance criteria

1. The narrow text-line merge issue materially improves on the MORAN pages 12-13 fixture or a focused geometry regression proves the merge no longer happens.
2. The change is backed by focused regression coverage.
3. Existing reading-order tests still pass.
4. `cd go && GOCACHE=/tmp/odl-gocache go build ./...` passes.

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

- list/paragraph continuation fixes after lines are already split
- broad XYCut / final sorter changes
- text decoding / TJ / ToUnicode work
- benchmark-wide cleanup

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
openclaw system event --text "Done: evaluated GAP-TEXTLINE-INTERNAL-GUTTER-MERGE in opendataloader-pdf-go" --mode now
```
