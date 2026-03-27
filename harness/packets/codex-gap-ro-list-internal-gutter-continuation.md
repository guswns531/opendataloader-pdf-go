# Codex Gap Task Packet — GAP-RO-LIST-INTERNAL-GUTTER-CONTINUATION

## Role

You are implementing one scoped parity gap in the Go port of OpenDataLoader PDF.
The Java implementation is the reference behavior unless explicitly stated otherwise.
Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-RO-LIST-INTERNAL-GUTTER-CONTINUATION`

## Category

`reading-order-parity`

---

## Objective

Stop already-separated internal-gutter tokens from being absorbed into list / paragraph continuation flow on `samples/pdf/1901.03003.pdf` pages 12-13.

This packet begins after the kept line-stage fix from `GAP-TEXTLINE-INTERNAL-GUTTER-MERGE`. Treat the remaining leak as a later semantic-grouping / continuation problem, not a text decoding or XYCut problem.

---

## Evidence

From the parent split report and the kept line-stage follow-up:

- The line-stage fix now keeps narrow gutter tokens like `contracers` out of wide body lines.
- Despite that, final markdown still leaks tokens such as:
  - `west`
  - `united`
  - `arsenal`
  - `football`
  - `manchester messageid`
  - `briogestone`
  - `contracers`
- Current markdown snippet still shows body/list continuation corruption, for example:
  - `- west The experiments above are all based on cropped text`
  - `————— briogestone In this paper, we presented a multi-object rec-`
  - `contracers` appears as its own line near the conclusion flow instead of being excluded from body progression the way Java does

This indicates a later grouping / continuation mechanism is still absorbing gutter-side nodes into main-flow structures.

---

## Reference behavior (Java)

Java does not weave these narrow gutter tokens into the main article list/paragraph flow on the MORAN fixture pages 12-13.

---

## Current Go behavior

Go still carries already-separated gutter tokens into list items / paragraph continuation structures before final reading-order output.

---

## Relevant files

### Go
- `go/internal/processors/list_processor.go`
- `go/internal/processors/document_processor.go`
- nearby semantic grouping / continuation helpers directly involved in list or paragraph carry-over
- focused tests only if directly needed

### Evidence
- `reports/loops/2026-03-28-gap-ro-internal-gutter-stack-retry-split.md`
- `reports/loops/2026-03-28-gap-textline-internal-gutter-merge.md`

---

## Required work

- inspect the list / paragraph continuation stage that groups already-separated text lines into semantic list items or flowing prose
- make the smallest defensible change that stops narrow gutter-side tokens from being absorbed into body/list continuation on the target fixture
- add focused regression coverage if possible
- keep the patch limited to this post-line, pre-final-output continuation issue

---

## Acceptance criteria

1. On `samples/pdf/1901.03003.pdf` pages 12-13, the leaked gutter tokens move materially closer to Java behavior in final markdown output.
2. Focused regression coverage is added where practical.
3. Existing kept reading-order regressions still pass.
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

- new text-line merge heuristics (already handled by the kept child gap)
- broad XYCut / final sorter changes unless the evidence proves the continuation stage is not the cause
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
openclaw system event --text "Done: evaluated GAP-RO-LIST-INTERNAL-GUTTER-CONTINUATION in opendataloader-pdf-go" --mode now
```
