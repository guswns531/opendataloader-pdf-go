# GAP-RO-INTERNAL-GUTTER-STACK retry

Result: `SPLIT`

Worktree outcome:
- exploratory edits reverted
- no repo files left changed by this retry

Why split:
- the fixture leak is not one root cause
- a narrow text-line patch improved one symptom but did not clear the page 12-13 corruption
- the remaining leakage survives through separate pre-reading-order structures, so one more mixed patch would widen scope beyond this packet

Evidence:
- Baseline Go markdown still leaks the target tokens into the page 12-13 flow:
  - `west	The experiments above are all based on cropped text ...`
  - `manchester messageid`
  - `briogestone In this paper, we presented a multi-object rec-`
  - `contracers	tiﬁed attention network (MORAN) for scene text`
- With the exploratory same-baseline line-splitting patch applied locally, the fixture changed but still failed:
  - `contracers` was separated from `tiﬁed attention network...`
  - `west` was separated from part of the list body
  - but the markdown still contained `west`, `united`, `arsenal`, `football`, `manchester messageid`, `briogestone`, and `contracers`
- Generated JSON from that exploratory run showed two distinct residual mechanisms on page 12:
  - standalone gutter paragraphs remained as separate semantic nodes at x≈221-278:
    - `west`, `united`, `arsenal`, `football`, `manchester\nmessageid`, `contracers`
  - list structure still absorbed gutter text into body-facing items:
    - `—————————————————————– \nwest\nThe experiments above are all based on cropped text`
    - `————— \nbriogestone\nIn this paper, we presented a multi-object rec-`

Suggested narrower follow-ups:
- `GAP-TEXTLINE-INTERNAL-GUTTER-MERGE`
  - scope: prevent same-baseline internal gutter tokens from being merged into wide body lines
  - evidence: exploratory patch split `contracers` away from `tiﬁed attention...`, showing a real pre-semantic text-line issue
- `GAP-RO-LIST-INTERNAL-GUTTER-CONTINUATION`
  - scope: stop already-separated gutter tokens from being absorbed into list/paragraph flow before final reading-order sort
  - evidence: even after line splitting, list items still contained `west` and `briogestone` alongside body text

Validation run during this retry:
- focused exploratory unit tests for the temporary text-line patch passed under `GOCACHE=/tmp/odl-gocache`
- fixture rerun with `GOCACHE=/tmp/odl-gocache go run ./cmd/opendataloader-pdf ...` still showed the target tokens, so the candidate was not KEEP-worthy
