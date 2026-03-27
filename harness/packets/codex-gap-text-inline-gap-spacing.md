# Codex Gap Task Packet — GAP-TEXT-INLINE-GAP-SPACING

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-TEXT-INLINE-GAP-SPACING`

## Category

`text-parity`

---

## Objective

Implement the smallest change that improves Go parity when adjacent non-whitespace text runs on the same baseline should produce a word boundary, but the current extractor concatenates them without inserting a space.

---

## Reference behavior (Java)

The Java implementation preserves ordinary word boundaries across adjacent text-showing runs well enough that headings and prose do not collapse into `theCreationofLife`, `daughtercells`, or similar joined tokens when the PDF uses multiple positioned runs instead of literal space characters.

---

## Current Go behavior

The current Go extractor preserves explicit boundary whitespace, but it does not appear to infer a space when two adjacent non-whitespace runs are separated by a visible geometric gap on the same baseline.

This leads to joined words such as:

- `Growth and theCreationofLife`
- `Oneofthecharacteristicsoflivingthingsistheability`
- `#nucleardivisions`
- `#daughtercellsproduced`
- `m ulti-objectrectiﬁedattention`

This gap should stay focused on **same-baseline inline word-boundary inference**, not TJ numeric spacing, encoding, or reading-order work.

---

## Why this matters

- degrades readability in plain text and markdown
- reduces Java parity in ordinary prose and headings
- hurts downstream search, chunking, and benchmark alignment

---

## Relevant files

### Go
- `go/pkg/pdfbox/extractor/text_extractor.go`
- nearby helpers directly involved in chunk append / same-baseline accumulation

### Java reference
- corresponding Java text accumulation behavior for adjacent positioned text runs

### Tests / fixtures
- `samples/pdf/1901.03003.pdf`
- `tests/benchmark/ground-truth/markdown/01030000000118.md`
- `tests/benchmark/prediction/opendataloader/markdown/01030000000118.md`
- `tests/benchmark/ground-truth/markdown/01030000000119.md`
- `tests/benchmark/prediction/opendataloader/markdown/01030000000119.md`
- existing tests under `go/pkg/pdfbox/extractor/` and `go/tests/unit/pdfbox/`

---

## Required work

- inspect how adjacent same-baseline text runs are appended today
- implement the minimum focused change that restores likely missing spaces between non-whitespace runs when the geometric gap clearly indicates a word boundary
- add or strengthen regression coverage for this narrow condition
- do not mix this patch with TJ offset, ToUnicode, WinAnsi, reading-order, or hyphen reflow work

---

## Acceptance criteria

1. A focused same-baseline adjacent-run case can recover a missing word boundary without requiring literal whitespace in the source run.
2. Existing explicit-whitespace handling remains stable.
3. Targeted fixture output moves closer to Java behavior for the joined-word examples above.
4. Targeted PDFBox/text extraction tests pass.
5. `go build ./...` succeeds.

---

## Validation commands

Run these before finishing:

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-inline-gap-spacing --format markdown --image-output off --quiet
```

If practical, also report before/after snippets for:

- `Growth and the Creation of Life`
- `nuclear divisions`
- `daughter cells`
- `m ulti-objectrectiﬁedattention`

---

## Non-goals

Do **not** attempt these in the same patch:

- whitespace-only run preservation already covered by `GAP-TEXT-TRIMSPACE`
- TJ numeric offset spacing
- WinAnsi decoding
- ToUnicode mapping changes
- reading-order or table reconstruction fixes
- hyphenation / line-wrap reflow

---

## Output format

When done, report:

1. files changed
2. behavior changed
3. tests/commands run and their result
4. remaining uncertainty
5. whether this narrow gap appears fixed, partially fixed, or needs another split
