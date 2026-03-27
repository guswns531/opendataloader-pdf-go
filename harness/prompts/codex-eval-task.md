# Codex Evaluator Task Packet Template

Use this template when OpenClaw asks Codex to evaluate a candidate patch or current parity state.

---

## Role

You are acting as an evaluator for a **single parity gap**.

Your job is not to broaden the implementation. Your job is to determine whether the current state should be:

- **KEEP**
- **DISCARD**
- **SPLIT**
- **NEEDS-HUMAN-REVIEW**

The Java implementation is the reference behavior unless explicitly documented otherwise.

---

## Gap ID

`<GAP-ID>`

## Candidate scope

`<describe branch / commit / working tree state being evaluated>`

---

## Reference behavior (Java)

`<describe expected Java behavior>`

---

## Current Go claim

`<describe what the candidate says it fixed>`

---

## Evidence to inspect

### Files
- `<file>`
- `<file>`

### Tests / fixtures
- `<fixture>`
- `<test>`

### Commands
```bash
<command>
<command>
```

---

## Evaluation questions

1. Does the change move behavior closer to Java?
2. Do targeted tests pass?
3. Is there evidence of regression elsewhere in the same subsystem?
4. Is the patch scoped and defensible, or ad hoc and fragile?
5. Does this need a narrower follow-up split?

---

## Required output

Return exactly these sections:

### Decision
`KEEP | DISCARD | SPLIT | NEEDS-HUMAN-REVIEW`

### Evidence
- `<bullet>`
- `<bullet>`

### Regressions checked
- `<bullet>`
- `<bullet>`

### Remaining risks
- `<bullet>`
- `<bullet>`

### Suggested next action
- `<one concrete next step>`
