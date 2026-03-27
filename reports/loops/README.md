# Loop Reports

This directory stores concise reports for autonomous parity loops.

Purpose:

- preserve evaluator decisions
- prevent repeated dead-end exploration
- record why a change was KEPT / DISCARDED / SPLIT
- make milestone progress auditable without reading raw chat logs

Recommended pattern:

- one short report per meaningful loop result
- especially for KEEP, DISCARD, and SPLIT outcomes
- include commit hash for kept changes

Use `TEMPLATE.md` as the base format.
