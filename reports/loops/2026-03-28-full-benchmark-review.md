# Full Benchmark Review — 2026-03-28

## Context

Ran the full local benchmark after disabling the OpenClaw autopilot cron job and repairing the benchmark environment.

Environment fixes required before the run:

- installed `git-lfs`
- ran `git lfs pull`
- installed missing Python benchmark dependencies:
  - `rapidfuzz`
  - `apted`
  - `beautifulsoup4`
  - `lxml`
  - `py-cpuinfo`

## Cron status

Autopilot cron job `ed999f3d-b5a8-4a69-9dad-c0b440b9aef6` was disabled before the benchmark run.

## Benchmark summary

Source: `tests/benchmark/prediction/opendataloader/evaluation.json`

- NID (Reading Order): `0.7391`
- TEDS (Table Structure): `0.0399`
- MHS (Heading Structure): `0.4012`
- Table Detection F1: `0.2373`
- Speed: `0.01s/doc` (`1.1s / 200 docs`)

## High-level interpretation

### What improved enough to matter

- Reading order is no longer the only obvious bottleneck.
- The recent text and reading-order loops likely helped stabilize many non-table pages, as many documents still score well on NID.
- Speed remains excellent and is not the current limiting factor.

### Main bottlenecks now

1. **Table structure parity is extremely weak**
   - TEDS `0.0399`
   - Table Detection F1 `0.2373`
2. **Heading structure parity is still weak**
   - MHS `0.4012`
3. **Residual reading-order failures remain**, but they are no longer the only dominant source of pain.

## Worst documents by metric

Extracted from `tests/benchmark/prediction/opendataloader/evaluation.csv`.

### Worst overall (sample)

- `01030000000003`
- `01030000000015`
- `01030000000103`
- `01030000000114`
- `01030000000127`
- `01030000000128`
- `01030000000130`
- `01030000000132`
- `01030000000133`
- `01030000000141`
- `01030000000163`
- `01030000000170`
- `01030000000172`
- `01030000000200`
- `01030000000199`

### Worst TEDS (sample)

Documents with zero table-structure score include:

- `01030000000045`
- `01030000000047`
- `01030000000051`
- `01030000000052`
- `01030000000053`
- `01030000000064`
- `01030000000081`
- `01030000000082`
- `01030000000083`
- `01030000000084`
- `01030000000088`
- `01030000000089`
- `01030000000090`
- `01030000000110`
- `01030000000117`
- `01030000000119`
- `01030000000120`
- `01030000000122`
- `01030000000127`
- `01030000000128`

### Worst MHS (sample)

Documents with zero heading-structure score include:

- `01030000000003`
- `01030000000036`
- `01030000000044`
- `01030000000067`
- `01030000000103`
- `01030000000105`
- `01030000000107`
- `01030000000123`
- `01030000000133`
- `01030000000141`
- `01030000000148`
- `01030000000157`
- `01030000000163`
- `01030000000196`
- `01030000000197`
- `01030000000200`

### Worst NID (sample)

- `01030000000003`
- `01030000000015`
- `01030000000103`
- `01030000000114`
- `01030000000127`
- `01030000000128`
- `01030000000130`
- `01030000000132`
- `01030000000133`
- `01030000000141`
- `01030000000163`
- `01030000000170`
- `01030000000172`
- `01030000000200`
- `01030000000199`

## Concrete failure pattern observed

Spot-checking one low-scoring table-heavy document showed a clear pattern:

- table HTML / structure from ground truth was flattened into loose text lines
- row/column relationships were lost
- some heading lines were over-promoted into markdown headings

Representative symptom from a sampled failure:

- ground truth preserved a structured table and normal prose flow
- prediction emitted fragmented region labels / numeric cells as loose stacked text
- nearby prose was partially promoted to headings (`##`) too aggressively

This suggests the current biggest remaining wins are likely in:

1. table detection recall
2. table cell grouping / structure reconstruction
3. heading over-promotion suppression near table-like layouts

## Recommended next gap backlog

### Table family

- `GAP-TABLE-DETECTION-LOW-RECALL`
- `GAP-TABLE-STRUCTURE-TEDS-CORE`
- `GAP-TABLE-ROW-COLUMN-GROUPING`
- `GAP-TABLE-TEXT-LINE-VS-CELL-ASSIGNMENT`

### Heading family

- `GAP-HEADING-HIERARCHY-MHS-CORE`
- `GAP-HEADING-OVERPROMOTION-NEAR-TABLES`
- `GAP-HEADING-LIST-TABLE-CONFUSION`

### Residual reading-order family

- `GAP-RO-ZERO-SCORE-DOC-CLUSTER-A`
- `GAP-RO-MIXED-LAYOUT-DOC-TAIL`

## Suggested immediate next move

Do **not** keep squeezing only text-parity gaps.

Best next move:

1. create a focused table benchmark gap from one zero-TEDS doc
2. create a focused heading benchmark gap from one zero-MHS doc
3. keep reading-order follow-ups limited to the zero-NID cluster docs

This should produce larger full-benchmark gains than continuing another generic text-parity cleanup pass.
