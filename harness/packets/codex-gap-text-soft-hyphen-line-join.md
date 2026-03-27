# Codex Gap Task Packet — GAP-TEXT-SOFT-HYPHEN-LINE-JOIN

## Role

You are implementing one narrow Java→Go text-parity fix in the Go port of OpenDataLoader PDF. Fresh context, one gap only.

## Gap ID

`GAP-TEXT-SOFT-HYPHEN-LINE-JOIN`

## Objective

Fix the live wrapped-line hyphenation artifact on `samples/pdf/1901.03003.pdf` where markdown currently renders:

- `More- over, methods ...`

The desired behavior is to join this discretionary line-wrap hyphenation as:

- `Moreover, methods ...`

## Evidence before

Current local output still shows:

- `/tmp/odl-sample-1901-next-gap/1901.03003.md:92`
- `41, 45, 50] have achieved notable success. More- over, methods based on convolutional neural networks [3, 22, 50] have been broadly applied. Integrating`

A raw extractor probe already showed:

- chunk 80: `41, 45, 50] have achieved notable success. More-`
- chunk 81: `over, methods based on convolutional neural networks`
- consecutive baselines in the same column
- continuation restarts at the same left margin

This is not the already-fixed same-line spacing family.

## Constraints

- One gap only.
- Do not touch reading-order / XYCut code.
- Prefer the smallest fix defensible against Java-like text output.
- Avoid a broad paragraph-reflow rewrite.
- Preserve genuine compounds such as `attention-based`.

## Required safety regression

Add or strengthen a focused regression so the fix proves both:

1. `More-` + `over,` joins to `Moreover,`
2. genuine compounds like `attention-based` are preserved, not collapsed to `attentionbased`

Use the narrowest trustworthy test surface you can justify (processor/generator/unit level).

## Relevant files to inspect

- `reports/loops/2026-03-28-eval-text-parity-next-gap-soft-hyphen-line-join.md`
- text-processing / markdown-generation files actually responsible for paragraph joining on the Go side
- existing text fixture tests around `1901.03003.pdf`

## Acceptance criteria

1. the live `More- over` artifact is removed on the targeted sample path
2. a focused regression exists for discretionary line-wrap hyphen join
3. a focused safety regression preserves real compounds like `attention-based`
4. targeted tests pass
5. `go build ./...` passes
6. if KEEP-worthy, write a concise loop report under `reports/loops/` dated `2026-03-28`, commit, and push

## Validation commands

```bash
cd go && GOCACHE=/tmp/odl-gocache go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go test ./internal/processors/... ./tests/unit/processors/... -count=1
cd go && GOCACHE=/tmp/odl-gocache go build ./...
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-sample-1901-soft-hyphen --format markdown --image-output off --quiet
```

## Deliverable

If KEEP:
- small scoped code/test change
- concise loop report with before/after evidence and commands
- commit + push

If not KEEP:
- leave tree clean/evaluatable
- concise SPLIT/blocker report

When completely finished, run this command to notify me:
openclaw system event --text "Done: evaluated soft-hyphen line-join text gap in opendataloader-pdf-go" --mode now
