# Codex Gap Task Packet — GAP-TEXT-INLINE-RUN-GEOMETRY-PROBE

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-TEXT-INLINE-RUN-GEOMETRY-PROBE`

## Category

`text-parity`

---

## Objective

Identify the real source layer for surviving joined-word artifacts in the Go pipeline by instrumenting and inspecting extracted chunk geometry around the known bad sample spans, then keep only a narrowly justified fix if fixture evidence improves.

This is a follow-up split from `GAP-TEXT-INLINE-GAP-SPACING`.

---

## Reference behavior (Java)

Java preserves ordinary word boundaries in these sample spans well enough that words do not collapse into `objectrectifiedattention`-style joins in normal prose.

---

## Current Go behavior

The previous same-baseline gap heuristic improved only synthetic unit cases and did not change the live sample output for:

- `m ulti-objectrectiﬁedattention`
- `survey.IEEE`
- `theCreationofLife`
- `Oneofthecharacteristicsoflivingthingsistheability`

That suggests the real issue is not yet isolated. It may be in extractor chunk geometry, run segmentation, normalization, or a later stage that still collapses boundaries.

---

## Why this matters

- prevents blind heuristic churn
- creates the smallest trustworthy next packet
- increases odds that the next KEEP actually moves fixture output

---

## Relevant files

### Go
- `go/pkg/pdfbox/extractor/text_extractor.go`
- `go/pkg/pdfbox/extractor/content_parser.go`
- `go/internal/processors/text_line_processor.go`
- any minimal helper/test file needed for targeted inspection

### Existing report
- `reports/loops/2026-03-28-gap-text-inline-gap-spacing.md`

### Fixtures
- `samples/pdf/1901.03003.pdf`
- benchmark markdown pairs under `tests/benchmark/{ground-truth,prediction}/opendataloader/markdown/`

---

## Required work

- reproduce at least one surviving joined-word sample from `samples/pdf/1901.03003.pdf`
- determine whether the join already exists in extractor output before `TextLineProcessor`
- if useful, add a temporary or test-scoped probe that captures chunk text/X/width/baseline around the target span
- only if the root cause becomes clear and a narrow fix improves the live fixture, keep the fix with focused regression coverage
- otherwise, leave code clean and write a loop report that names the next narrower root-cause gap

---

## Acceptance criteria

1. The report identifies the exact pipeline layer where at least one target join originates.
2. Any kept change is validated by live fixture evidence, not only synthetic tests.
3. If no validated fix is found, the worktree is left clean except for the required loop report.
4. Targeted validation commands pass for any kept change.

---

## Validation commands

Run these before finishing:

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-inline-run-probe --format markdown --image-output off --quiet
```

Also capture concise evidence showing whether the join exists:
- immediately after extraction
- after text-line grouping if relevant

---

## Non-goals

Do **not** broaden into:
- reading-order work
- ToUnicode/WinAnsi unless directly proven as the root cause of the sampled join
- unrelated hyphenation cleanup
- large refactors

---

## Output format

When done, report:

1. layer where the join originates
2. files changed
3. behavior changed
4. tests/commands run and their result
5. whether this became a KEEP, DISCARD, or another SPLIT
