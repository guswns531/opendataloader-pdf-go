# Pure Go Core Migration Plan

Actionable phase tracking lives in `docs/go/pure-go-phase-checklist.md`.
Compatibility freeze details live in `docs/go/pure-go-contract-matrix.md`.

## 1. Goal

Replace the current Java + veraPDF local extraction core with a Pure Go implementation while preserving the public compatibility surface that matters:

- CLI option names and exit semantics
- Node/Python wrapper behavior
- local-mode JSON schema shape
- local-mode Markdown benchmark quality and speed floors

This document defines the migration boundary, the order of work, and the parallel workstreams.

---

## 2. Summary

The main rewrite risk is not Java itself. The main risk is that the current engine is built on top of `veraPDF` as a runtime substrate:

- PDF parsing and chunk extraction
- structure-tree access
- semantic object model
- geometry and text heuristics
- table/list/heading helper utilities

Adapters such as Jackson, OkHttp, and the current wrapper CLIs are straightforward to replace. The hard problem is replacing the `org.verapdf.*` model and helper layer that the processors assume is already available.

The practical rewrite strategy is:

1. Freeze the compatibility contract first.
2. Define a Go-native document model and state model.
3. Build a Go PDF ingestion layer that replaces the current veraPDF-fed artifact pipeline.
4. Port the deterministic local heuristics in dependency order.
5. Reintroduce output generators and wrappers.
6. Defer hybrid, enrichments, and annotated-PDF generation until the local core is stable.

---

## 3. Compatibility Boundary

### 3.1 Public Contract To Preserve

- `schema.json` is the strongest published JSON output contract.
- `options.json` is the strongest published CLI option contract and is code-generated into Node, Python, and docs.
- Java CLI behavior in `CLIMain.java` and `CLIOptions.java` defines traversal, validation, default paths, and exit semantics.
- Node and Python packages are thin wrappers over the CLI, so preserving CLI semantics preserves most external compatibility.

### 3.2 MVP Compatibility Target

The MVP target for the Pure Go rewrite is:

- deterministic local mode only
- CLI-compatible binary for current option names/defaults
- compatible JSON output for the current local schema
- Markdown quality and speed sufficient to pass current benchmark thresholds
- existing Node/Python `convert()` and deprecated `run()` wrappers kept working against the new CLI

### 3.3 Explicitly Deferred From MVP

- hybrid mode and backend enrichments
- formula/picture enrichment parity
- annotated PDF output
- full tagged-PDF and structure-tree parity

These areas already show contract drift between `README.md`, `schema.json`, tests, and hybrid docs. They should be normalized after the local Go core is stable.

---

## 4. Architecture Findings

### 4.1 Current Pipeline Shape

The untagged local pipeline runs roughly in this order:

1. PDF preprocessing and artifact extraction
2. content filtering
3. optional cluster-table detection
4. border-table assignment and cell recursion
5. chunk-to-line grouping
6. special table rewrite
7. header/footer filtering
8. list detection
9. paragraph grouping
10. heading detection
11. neighbor stitching and level assignment
12. optional final reading-order sort
13. JSON/Markdown/HTML/Text/PDF generation

### 4.2 Self-Contained Parts

These are the best early port targets because they are mostly geometry or repo-local heuristics:

- `XYCutPlusPlusSorter`
- `BulletedParagraphUtils`
- `TextNodeStatistics`
- `LevelProcessor` and level metadata helpers
- `SpecialTableProcessor`

### 4.3 Moderately Coupled Parts

These are good rewrite targets once the Go object model and metric layer exist:

- `TextLineProcessor`
- `ParagraphProcessor`
- `HeadingProcessor`
- list neighbor stitching in `ListProcessor`
- table recursion/assembly in `TableBorderProcessor`

### 4.4 Most Coupled Parts

These are the hard blockers for a Pure Go core:

- preprocessing and raw extraction in `DocumentProcessor`
- tagged/structure-tree traversal in `TaggedDocumentProcessor`
- table-border collection and cluster-table detection
- global mutable state patterns in `StaticLayoutContainers` and related static containers
- helper logic currently hidden behind veraPDF utilities such as chunk merge, node scoring, list labeling, and caption detection

---

## 5. Replacement Layers Required In Go

The Pure Go implementation needs these layers before parity work is realistic.

### 5.1 Go Native Document Model

Create Go-native equivalents for the semantic/layout types the processors rely on:

- bounding boxes and multi-bounding boxes
- pages and page metadata
- raw text/image/line artifacts
- text chunks, text lines, text blocks
- paragraphs, headings, captions
- lists and list items
- tables, rows, cells, border metadata
- IDs, indexes, nesting levels, linked-object references

This model must be mutable enough to support the current pipeline, because the Java processors assign IDs, indexes, levels, and neighbor links throughout processing.

### 5.2 PDF Ingestion Layer

Replace the current veraPDF-fed ingestion path with a Go API that can provide:

- page count and page size
- raw page artifacts with geometry and style
- optional structure-tree access
- metadata for JSON root fields
- hooks for table-border candidates and image extraction

This is the largest technical blocker. The rewrite will fail if the ingestion layer cannot supply stable raw inputs for downstream heuristics.

### 5.3 Geometry And Semantics Helper Layer

Replace the helper behavior currently buried inside veraPDF utilities with a Go metrics layer:

- chunk merge and line-join decisions
- alignment and spacing signals
- heading probability and style rarity scoring
- list label recognition and interval logic
- caption association rules
- neighbor relationship and overlap utilities

The goal is not to transliterate Java classes one by one. The goal is to reproduce the signals the processors need.

### 5.4 Output Layer

Once the semantic model exists, implement:

- JSON writer against `schema.json`
- Markdown writer
- HTML writer
- text writer

Annotated PDF output is optional and should not block the local extractor MVP.

---

## 6. Recommended Implementation Order

### Phase 0: Contract Freeze

Deliverables:

- document the local-mode compatibility contract
- explicitly mark deferred features
- define the MVP benchmark gates

Tasks:

- use `schema.json` as the JSON contract
- use `options.json` as the CLI contract
- use Java CLI tests for exit-code and traversal semantics
- use benchmark thresholds as the MVP quality floor

Success criteria:

- MVP scope and deferred scope are explicit
- no feature is assumed in scope without a concrete acceptance rule

### Phase 1: Go Core Skeleton

Deliverables:

- `go.mod`
- package layout for model, parser, metrics, processors, emitters, and CLI
- Go-native document model
- processing context replacing static global state

Tasks:

- define semantic/layout types
- define per-document processing state
- define interfaces between ingestion, heuristics, and emitters

Success criteria:

- processors can be written without static globals
- the model is sufficient to express current schema output

### Phase 2: Deterministic Geometry And Text Pipeline

Deliverables:

- reading-order sorter
- chunk-to-line grouping
- line-to-paragraph grouping
- basic content filtering

Tasks:

- port `XYCutPlusPlusSorter`
- implement text merge metrics in Go
- implement paragraph grouping metrics in Go
- implement minimal header/footer filtering hooks

Success criteria:

- single-page text-heavy PDFs produce stable line/paragraph structure
- benchmark Markdown is meaningful on simple documents

### Phase 3: Semantic Structure Layer

Deliverables:

- list reconstruction
- heading detection and heading levels
- nesting level assignment
- table assembly from already-detected cells/borders

Tasks:

- port list neighbor stitching
- recreate heading scoring using Go-native style statistics
- port level assignment helpers
- support semantic table output from precomputed cells

Success criteria:

- headings and lists appear in Markdown consistently
- JSON structure is close enough to validate against the published schema

### Phase 4: Raw Extraction And Table Detection

Deliverables:

- raw artifact extraction parity
- border-table detection
- cluster-table replacement or acceptable approximation
- optional structure-tree extraction hooks

Tasks:

- build or integrate the real ingestion layer
- reproduce border detection signals
- decide whether cluster-table parity is exact or approximate for MVP

Success criteria:

- benchmark table and reading-order scores reach threshold floors
- the local Go core can run end-to-end without Java

### Phase 5: Emitters And Wrapper Rebinding

Deliverables:

- JSON/Markdown/HTML/Text output
- CLI-compatible Go binary
- Node/Python wrappers pointed to the Go binary instead of the JAR

Tasks:

- implement emitters
- preserve CLI exit codes and traversal semantics
- keep wrapper arg mapping unchanged where possible

Success criteria:

- wrapper integration tests pass
- benchmark thresholds pass
- local mode no longer depends on Java

### Phase 6: Deferred Features

After the local core is stable:

- hybrid mode
- enrichments
- annotated PDF output
- tagged-PDF parity

---

## 7. Parallel Workstreams

The rewrite should be split by write ownership, not by language mirror classes.

### Workstream A: Contract And Acceptance

Owner scope:

- CLI contract
- schema contract
- benchmark harness
- wrapper compatibility

Tasks:

- lock down option/exit semantics
- add Go-core benchmark entrypoint
- add schema validation tests for local JSON output
- define parity dashboards

### Workstream B: Core Model And Processing Context

Owner scope:

- semantic/layout types
- per-document context
- ID/index/level/link handling

Tasks:

- define stable internal model
- remove need for static mutable containers
- define serialization-facing structures

### Workstream C: Geometry And Text Heuristics

Owner scope:

- reading order
- chunk merge
- line grouping
- paragraph grouping
- heading/list metrics

Tasks:

- port self-contained algorithms first
- build Go-native scoring helpers
- compare outputs against benchmark PDFs

### Workstream D: Extraction And Tables

Owner scope:

- PDF ingestion
- raw artifact extraction
- border detection
- cluster-table replacement
- struct-tree access

Tasks:

- decide extraction backend approach
- generate stable raw artifact fixtures
- unblock downstream processors with test fixtures before full parser parity

### Workstream E: Emitters And Wrapper Integration

Owner scope:

- JSON/Markdown/HTML/Text emitters
- Go CLI
- Node/Python wrapper rebinding

Tasks:

- preserve output file naming and directory semantics
- preserve wrapper behavior
- keep deprecated `run()` entrypoints functional during transition

---

## 8. Acceptance Criteria

The MVP should be considered successful only if all of the following are true:

- local extraction runs without Java
- CLI-compatible option names and exit codes are preserved
- Node/Python wrappers still work against the local core
- JSON output validates against the published local schema
- benchmark floors are met:
  - `nid >= 0.85`
  - `teds >= 0.40`
  - `mhs >= 0.55`
  - `table_detection_f1 >= 0.55`
  - `elapsed_per_doc <= 2.0s`

Anything weaker than this is a prototype, not a replacement.

---

## 9. Decision Points

These decisions should be made explicitly before implementation gets too deep.

### 9.1 Extraction Backend Strategy

Open question:

- write a native PDF extraction layer from scratch
- build on an existing Go PDF library
- use an interim offline fixture pipeline for downstream heuristics while extraction matures

Impact:

- determines overall schedule and parity ceiling

### 9.2 Table Detection Strategy

Open question:

- exact parity with current border and cluster behavior
- or threshold-based compatibility for MVP

Impact:

- biggest risk to benchmark success after raw extraction

### 9.3 Struct-Tree Scope

Open question:

- completely defer tagged/struct-tree handling
- or expose parser hooks early but keep behavior disabled for MVP

Impact:

- affects architecture of the ingestion API, but should not block MVP delivery

### 9.4 Output Scope For MVP

Open question:

- whether `html`, `text`, and annotated `pdf` are implemented in MVP or explicitly deferred

Impact:

- affects wrapper compatibility and release messaging more than core extraction quality

---

## 10. Main Risks

- Exact behavior today is partly hidden inside veraPDF helpers, not just visible Java code.
- A weak ingestion layer will make every downstream heuristic unstable.
- Contract drift already exists across README, schema, hybrid docs, and serializers.
- Recreating tables and list semantics will likely require behavior approximation, not line-for-line parity.
- Static global state in the Java design should not be copied into Go.

---

## 11. Immediate Next Steps

1. Create the Go module and internal package layout.
2. Implement the Go-native document model and processing context.
3. Add a benchmark-facing Go CLI stub so the benchmark harness can target the future core early.
4. Port `XYCutPlusPlusSorter` and basic level helpers first.
5. Build fixture-driven tests for chunk-to-line and line-to-paragraph grouping before attempting full parser parity.

This sequence keeps early work focused on reusable core pieces instead of prematurely binding the rewrite to a still-unstable extraction backend.
