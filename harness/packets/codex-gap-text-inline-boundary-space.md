# Codex Gap Task Packet — GAP-TEXT-INLINE-BOUNDARY-SPACE

## Role

You are implementing one narrow Java→Go parity gap in the Go port of OpenDataLoader PDF. Keep scope tight.

## Gap ID

`GAP-TEXT-INLINE-BOUNDARY-SPACE`

## Category

`text-parity`

## Objective

Fix the remaining inline word joins where adjacent extracted chunks now have sane geometry but final line/paragraph assembly still drops an expected boundary space.

Primary live symptom to target:

- `InProceedings` on `samples/pdf/1901.03003.pdf`

This is the next split after `GAP-TEXT-EXTRACTOR-WIDTH-GEOMETRY`, which fixed `survey.IEEE` by correcting extractor widths but left bibliography joins.

## Reference behavior (Java)

Java/PDFBox preserves the ordinary word boundary in bibliography prose, emitting `In Proceedings` rather than `InProceedings` when layout does not justify collapsing the space.

## Current Go behavior

After commit `43c5eb4`, the rebuilt CLI still reproduces examples such as:

- `InProceedings of the IEEE Conference on Computer`
- `tion. InProceedings of International Conference on`

So the remaining issue is narrower than extractor width inflation.

## Relevant files

### Go
- `go/internal/processors/text_line_processor.go`
- `go/internal/processors/paragraph_processor.go`
- `go/pkg/pdfbox/extractor/text_extractor.go` (read-only unless required)
- focused tests under `go/tests/unit/` or `go/internal/processors/`

### Reports
- `reports/loops/2026-03-28-gap-text-inline-run-geometry-probe.md`
- `reports/loops/2026-03-28-gap-text-extractor-width-geometry.md`

### Fixture
- `samples/pdf/1901.03003.pdf`

## Required work

- inspect the surviving `InProceedings` case at extraction + line assembly boundaries
- determine whether the missing space is lost in `TextLineProcessor`, `ParagraphProcessor`, or a narrower adjacent stage
- keep the fix specific to boundary-space synthesis; do not broaden into reading order or unrelated cleanup
- add at least one focused regression test
- validate the live fixture after rebuilding the CLI

## Acceptance criteria

1. At least one live `InProceedings` symptom becomes `In Proceedings` after the fix.
2. Targeted tests/build pass.
3. A loop report records KEEP / DISCARD / SPLIT with evidence.
4. If the issue turns out to be a different narrower root cause, write that down and split instead of hand-waving a heuristic.

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-inline-boundary-space --format markdown --image-output off --quiet
rg -n "InProceedings|In Proceedings" /tmp/odl-inline-boundary-space/1901.03003.md || true
```

## Non-goals

- reading-order changes
- table/heading work
- broad text normalization unrelated to this boundary-space symptom
- new benchmark-wide cleanup passes

## Output format

Report:
1. where the space was being lost
2. files changed
3. live fixture behavior before/after
4. tests/commands run
5. KEEP / DISCARD / SPLIT

When completely finished, run this command to notify me:
openclaw system event --text "Done: evaluated GAP-TEXT-INLINE-BOUNDARY-SPACE in opendataloader-pdf-go" --mode now
