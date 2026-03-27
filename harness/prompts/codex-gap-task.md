# Codex Gap Task Packet Template

Use this template when OpenClaw sends a single parity gap to Codex.

---

## Role

You are implementing **one scoped parity gap** in the Go port of OpenDataLoader PDF.

The Java implementation is the reference behavior unless explicitly stated otherwise.

Do not broaden scope beyond this packet.

---

## Gap ID

`<GAP-ID>`

## Category

`<text-parity | reading-order-parity | table-parity | heading-parity | serializer-parity | cli-parity | infra>`

---

## Objective

Implement the smallest change that improves Go parity for this gap.

---

## Reference behavior (Java)

`<describe what Java currently does>`

---

## Current Go behavior

`<describe current mismatch>`

---

## Why this matters

`<user-visible impact / benchmark impact / regression impact>`

---

## Relevant files

### Go
- `<file>`
- `<file>`

### Java reference
- `<file>`
- `<file>`

### Tests / fixtures
- `<file>`
- `<pdf fixture>`

---

## Required work

- implement only what is necessary for this gap
- add or tighten regression coverage if appropriate
- keep the patch focused and explain any uncertainty

---

## Acceptance criteria

1. `<criterion>`
2. `<criterion>`
3. `<criterion>`

---

## Validation commands

Run these before finishing:

```bash
<command>
<command>
```

---

## Non-goals

Do **not** attempt these in the same patch:

- `<non-goal>`
- `<non-goal>`

---

## Output format

When done, report:

1. files changed
2. behavior changed
3. tests/commands run and their result
4. remaining uncertainty
5. whether the gap appears fully fixed, partially fixed, or needs follow-up
