# Pure Go Contract Matrix

This document freezes the compatibility surface for the Pure Go local-core migration
 and records the current status of the Go skeleton against that contract.

## MVP Contract Sources

| Surface | Source of truth | Current Go status | Notes |
| --- | --- | --- | --- |
| CLI option names/defaults | `options.json`, Java `CLIOptions.java` | Partial | Most flags parse, many are placeholders only |
| CLI exit and traversal behavior | Java CLI tests, `CLIMain.java` | Partial | Basic discovery and exit codes exist, parity is not proven |
| Local JSON output shape | `schema.json` | Partial | Schema-shaped emitter exists, full parity not validated |
| Local Markdown output | benchmark harness + Java output behavior | Partial | Deterministic Markdown exists, quality gap remains |
| Node/Python wrapper behavior | wrapper integration tests | Partial | Go binary is benchmark-addressable, wrappers still target Java packaging |
| Local benchmark floor | `tests/benchmark/thresholds.json` | Not met | Smoke only; replacement floor not reached |

## Explicit MVP Scope

- Pure Go local extraction core
- CLI-compatible local binary
- JSON compatibility for local mode
- Markdown quality sufficient for benchmark thresholds
- Existing wrapper surface preserved against the new local binary

## Explicitly Deferred

- Hybrid mode parity
- Formula/picture enrichments
- Tagged PDF / struct-tree parity
- Annotated PDF output
- PDF writer parity

## CLI Status Matrix

### Implemented Or Mostly Wired

| Option / behavior | Status | Notes |
| --- | --- | --- |
| `--output-dir`, `-o` | Implemented | Output path is honored |
| `--format`, `-f` for `json`, `markdown`, `html`, `text` | Implemented | Unsupported formats are silently dropped today |
| `--quiet`, `-q` | Implemented | Console logging suppression works |
| `--pages` | Implemented | Parsed and applied after pipeline run |
| `--sanitize` | Implemented | Sanitization filter runs |
| `--replace-invalid-chars` | Implemented | Text cleaner applies replacement |
| `--include-header-footer` | Implemented | Controls header/footer filtering pass |
| `--reading-order` | Implemented | `xycut` and `off` behavior exists |
| `--table-method` | Partial | Heuristic toggle exists, but not Java parity |
| `--content-safety-off` | Partial | Only layout-filter gating is wired, not full Java behavior |

### Parsed But Not Functionally Delivered

| Option / behavior | Status | Gap |
| --- | --- | --- |
| `--password` | Placeholder | No encrypted PDF handling in Go ingestion |
| `--keep-line-breaks` | Placeholder | Parsed into decisions, not reflected in emitters/pipeline |
| `--use-struct-tree` | Placeholder | Parsed only; no tagged-PDF ingestion path |
| `--detect-strikethrough` | Placeholder | Parsed only; no equivalent processing pass |
| page separator options | Placeholder | Parsed but not applied in emitters |
| image output options | Placeholder | Parsed but no image extraction/output path |

### Deferred / Unsupported In MVP Today

| Option / behavior | Status | Gap |
| --- | --- | --- |
| `--format=pdf` | Unsupported | No PDF writer in Go |
| `--format=markdown-with-html` | Unsupported | No equivalent Markdown mode |
| `--format=markdown-with-images` | Unsupported | No equivalent Markdown image mode |
| hybrid options | Deferred | No Pure Go hybrid parity path yet |

## Native Ingestion Contract

The future Pure Go ingestion layer must provide at least:

- document metadata and page count
- page geometry including width, height, rotation, and page number/index
- raw text artifacts with stable bounds, order, font, font size, and content
- image and path/line artifacts needed for layout and table reconstruction
- hooks for encrypted PDFs and password handling
- optional struct-tree hooks without forcing MVP completion on tagged PDFs

## Current Go Skeleton Reuse

These areas are already worth keeping:

- Go-native document graph and IDs
- processing context and option resolution
- CLI discovery and page-range parsing
- runtime selector lives outside `cmd`; native PDF backend is preferred for `.raw.json`, semantic fixture loader handles `.json`, and `.pdf` falls back to the temporary pdftext bridge
- fixture ingestion path for downstream algorithm tests
- fixture-backed native loader for raw-artifact goldens
- JSON / Markdown / HTML / text emitters as scaffolding
- heuristic packages as temporary or starter implementations

## Benchmark Gates

Target replacement floor from `tests/benchmark/thresholds.json`:

- `nid >= 0.85`
- `teds >= 0.40`
- `mhs >= 0.55`
- `table_detection_f1 >= 0.55`
- `elapsed_per_doc <= 2.0s`

Current documented Go smoke signal:

- 1 document smoke only
- `nid ~= 0.739`
- `mhs = 0.0`
- `teds = N/A`
- speed is acceptable for smoke, quality is not

## Immediate Use

Use this matrix before every phase commit:

1. update implemented vs placeholder behavior
2. move only proven behavior from partial to implemented
3. keep deferred items explicit instead of silently parsing them
