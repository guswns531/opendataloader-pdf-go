# Codex Gap Task Packet — GAP-TEXT-INTRAWORD-SPACING

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`GAP-TEXT-INTRAWORD-SPACING`

## Category

`text-parity`

---

## Objective

Implement the smallest change that reduces false synthetic spaces inserted inside a single word while preserving the recently fixed bibliography boundary spaces.

---

## Reference behavior (Java)

Java preserves ordinary word boundaries without breaking a single word into fragments like `m ulti- objectr ectiﬁed a ttention` in normal prose/title text.

---

## Current Go behavior

After the recent inline boundary-space fix, the live fixture output for `samples/pdf/1901.03003.pdf` still shows false intra-word spaces near the start of the document, for example:

- `## MORAN: A Multi-Object Rectiﬁed Attention Network` should remain intact in the heading area
- abstract prose currently contains `m ulti- objectr ectiﬁed a ttention network`

This suggests the current small-positive-gap synthetic spacing fallback is too permissive for some adjacent chunks/runs.

Keep scope focused on this false-positive spacing behavior. Do not revisit unrelated reading-order or encoding work.

---

## Why this matters

- visible markdown/text quality regresses even when joined-word cases improve
- benchmark alignment suffers when one word is split into multiple tokens
- this is likely the next narrow parity gap in the text-spacing family

---

## Relevant files

### Go
- `go/internal/processors/text_line_processor.go`
- nearby tests under `go/tests/unit/processors/`
- any minimal fixture-backed regression test if needed

### Prior loop context
- `reports/loops/2026-03-28-gap-text-inline-boundary-space.md`
- `reports/loops/2026-03-28-gap-text-extractor-width-geometry.md`

### Fixture
- `samples/pdf/1901.03003.pdf`

---

## Required work

- reproduce the false intra-word spacing on `samples/pdf/1901.03003.pdf`
- inspect why the current boundary-space heuristic inserts spaces inside a single word
- implement the minimum narrow change that suppresses those false positives while keeping legitimate bibliography-style spaces such as `In Proceedings`
- add focused regression coverage
- keep the work scoped; do not mix in ToUnicode, WinAnsi, TJ spacing, paragraph reflow, or reading-order changes

---

## Acceptance criteria

1. Live fixture output no longer shows the target false splits such as `m ulti- objectr ectiﬁed a ttention` if the root cause is in scope.
2. Previously fixed bibliography spacing cases (for example `In Proceedings`) remain correct.
3. Targeted processor/pdfbox tests pass.
4. `go build ./...` succeeds.

---

## Validation commands

Run these before finishing:

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./tests/unit/processors -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-intraword-spacing --format markdown --image-output off --quiet
rg -n "m ulti-|objectr ecti|a ttention|InProceedings|In Proceedings" /tmp/odl-intraword-spacing/1901.03003.md || true
```

---

## Non-goals

Do **not** attempt these in the same patch:

- reading-order changes
- ToUnicode / WinAnsi decoding work
- TJ numeric offset spacing
- broad extractor refactors unless strictly required by the observed root cause
- unrelated markdown cleanup

---

## Output format

When done, report:

1. files changed
2. behavior changed
3. tests/commands run and their result
4. remaining uncertainty
5. whether this gap is a KEEP, DISCARD, or SPLIT
