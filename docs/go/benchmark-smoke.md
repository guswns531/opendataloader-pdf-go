# Go Benchmark Smoke Results

## Environment

- Date: 2026-03-21
- Host: `HJ-Mac.local`
- OS: `Darwin 25.3.0 arm64`
- Go version: `go1.26.1 darwin/arm64`
- Repo revision: `e03c205cbd104c10452e601553eddf7c35a159c1`
- CPU: `Apple M5`
- Memory: not captured

## Command

- Command: `./scripts/bench-go-smoke.sh --doc-id 01030000000001`
- Working directory: repo root
- Notes: benchmark was executed through the Python harness using engine `opendataloader-go` after installing local Python dependencies required by `tests/benchmark/run.py`.

## Selected Docs

- Doc 1: `01030000000001`

## Key Metrics

- Total docs: `1`
- Successful: `1`
- Failed: `0`
- Total elapsed: `0.38556790351867676s`
- Avg per doc: `0.38556790351867676s`
- NID mean: `0.7389514136700521`
- TEDS mean: `N/A`
- MHS mean: `0.0`
- Table detection accuracy: `0.79`
- Table detection TP / FP / FN / TN: `0 / 0 / 42 / 158`

## Observations

- The harness executed end-to-end with the Go engine and produced `summary.json`, `evaluation.json`, and `evaluation.csv`.
- Reading-order score is usable for a first smoke run, but heading and table quality are still far from Java parity.
- This smoke pass used a single document, so it is only a workflow validation plus an early quality signal, not a representative benchmark.

## Result

- Status: `pass (workflow smoke)` 
- Follow-up: run a small multi-doc smoke set next, then a wider benchmark slice once the remaining accuracy work is in place.
