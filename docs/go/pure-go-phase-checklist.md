# Pure Go Port Phase Checklist

## Porting Rules

- [ ] Local mode must stop depending on Java.
- [ ] Do not use `github.com/ledongthuc/pdf` or another high-level shortcut as the final ingestion core.
- [ ] Preserve `options.json` CLI semantics, `schema.json` JSON shape, and Node/Python wrapper behavior.
- [ ] Treat fixture ingestion as a test harness only, not the final runtime backend.
- [ ] Runtime selection is now split out of `cmd`; `.raw.json` -> nativepdf fixture loader, `.json` -> semantic fixture loader, `.pdf` -> temporary pdftext bridge.
- [ ] Keep hybrid, enrichments, annotated PDF, and tagged-PDF parity out of MVP unless a phase explicitly pulls them in.

## Area Classification

| Area | Classification | Why | Representative files |
| --- | --- | --- | --- |
| CLI contract and wrapper-facing behavior | 바로 이식 가능 | The public contract is already explicit in option metadata and CLI tests. The work is mostly semantic parity, validation, traversal, and exit behavior. | `java/opendataloader-pdf-cli/src/main/java/org/opendataloader/pdf/cli/CLIOptions.java`, `java/opendataloader-pdf-cli/src/main/java/org/opendataloader/pdf/cli/CLIMain.java`, `options.json` |
| Small repo-local utilities with geometry/statistics focus | 바로 이식 가능 | These are mostly deterministic helpers and do not require full PDF ingestion if the Go model exposes equivalent fields. | `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/utils/BulletedParagraphUtils.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/utils/TextNodeStatistics.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/readingorder/XYCutPlusPlusSorter.java` |
| Markdown and HTML emitters | Go 모델 먼저 필요 | Output logic is not the blocker, but it assumes a stable semantic tree with links, levels, tables, and image references. | `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/markdown/MarkdownGenerator.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/markdown/MarkdownHTMLGenerator.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/html/HtmlGenerator.java` |
| JSON serializers over semantic nodes | Go 모델 먼저 필요 | Serialization can move early once the Go-native semantic model is fixed, but exact output depends on stable node types and metadata fields. | `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/json/ObjectMapperHolder.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/json/serializers/*.java` |
| Text-line, paragraph, heading, list, level logic | Go 모델 먼저 필요 | These processors are portable, but they depend on a mutable semantic/layout model with IDs, indexes, bounds, links, levels, and text style statistics. | `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TextLineProcessor.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/ParagraphProcessor.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/HeadingProcessor.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/ListProcessor.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/LevelProcessor.java` |
| Special-table rewriting and table emitters from already-known cells | Go 모델 먼저 필요 | This logic is easier after table cells/rows exist in the model, but it should not wait for full parser parity. | `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/SpecialTableProcessor.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/utils/levels/TableLevelInfo.java` |
| Main document orchestration and preprocessing | veraPDF 대체 계층 먼저 필요 | The current entrypoint constructs `PDDocument`, `GFSAPDFDocument`, static veraPDF containers, artifact extraction, and output-side PDF metadata directly on top of veraPDF. | `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/DocumentProcessor.java` |
| Tagged PDF / structure-tree processing | veraPDF 대체 계층 먼저 필요 | The code walks the veraPDF structure tree and semantic node graph directly. Without a native struct-tree API, there is nothing meaningful to port against. | `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TaggedDocumentProcessor.java` |
| Raw filtering, hidden text, captions, and border/cluster tables | veraPDF 대체 계층 먼저 필요 | These areas depend on raw artifacts, line art, table borders, chunk merge utilities, list-label utilities, and node helpers currently provided by veraPDF. | `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/ContentFilterProcessor.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/HiddenTextProcessor.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/CaptionProcessor.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/AbstractTableProcessor.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/TableBorderProcessor.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/processors/ClusterTableProcessor.java` |
| JSON root writer, image extraction, PDF rewrite | veraPDF 대체 계층 먼저 필요 | These components use PDF object model access, COS/XMP metadata, image extraction, and PDF rewriting that are currently tied to veraPDF internals. | `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/json/JsonWriter.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/utils/ImagesUtils.java`, `java/opendataloader-pdf-core/src/main/java/org/opendataloader/pdf/pdf/PDFWriter.java` |

## Recommended Phase Order

### Phase 0: Contract Freeze

- [ ] Lock MVP scope to deterministic local mode only.
- [ ] Freeze compatibility targets: `options.json`, `schema.json`, Node/Python wrapper behavior, CLI exit semantics.
- [ ] Maintain `docs/go/pure-go-contract-matrix.md` as the explicit source for implemented vs placeholder vs deferred behavior.
- [ ] Mark non-MVP features as deferred in one place: hybrid, enrichments, annotated PDF, tagged-PDF parity.
- [ ] Record benchmark gates as acceptance criteria.
- [ ] Create a per-phase commit policy and benchmark/test gate.

### Phase 1: Go Core Foundation

- [ ] Finalize the Go-native document model for pages, raw artifacts, semantic nodes, tables, lists, headings, captions, links, IDs, and levels.
- [ ] Replace any remaining static/global-state assumptions with per-document processing context.
- [ ] Define interfaces between ingestion, heuristics, and emitters.
- [ ] Keep fixture ingestion as the stable test seam for downstream work.

### Phase 2: CLI and Emitter Parity on Fixtures

- [ ] Reach CLI parity for traversal, defaults, validation, output naming, and exit codes.
- [ ] Make Markdown, HTML, text, and schema JSON emit against the Go model.
- [ ] Add schema validation tests for Go JSON output.
- [ ] Rebind benchmark and wrapper smoke tests to the Go CLI path only.

### Phase 3: Directly Portable Deterministic Heuristics

- [ ] Port and harden reading order.
- [ ] Port utility-level statistics and bullet/list helper logic.
- [ ] Improve chunk-to-line and line-to-paragraph grouping using fixture goldens.
- [ ] Build style-statistics support needed later by heading and level detection.

### Phase 4: Model-Coupled Semantic Processors

- [ ] Port `TextLineProcessor` semantics against the Go model.
- [ ] Port paragraph assembly and list-from-text reconstruction.
- [ ] Port heading detection and heading-level assignment.
- [ ] Port level metadata helpers and neighbor stitching.
- [ ] Port special-table rewriting over precomputed table cells.

### Phase 5: Native Pure-Go Ingestion

- [ ] Remove `ledongthuc/pdf` from the runtime path.
- [ ] Design a native ingestion API for page geometry, raw text/image/line artifacts, style, page metadata, and optional struct-tree hooks.
- [ ] Implement a first native ingestion backend that feeds stable raw artifacts into the existing Go pipeline.
- [ ] Add raw-artifact fixture dumps for debugging native ingestion regressions.

### Phase 6: Table and Raw-Artifact Subsystem

- [ ] Recreate border-table detection and assignment in Go.
- [ ] Decide whether cluster-table behavior is exact parity or threshold-compatible MVP behavior.
- [ ] Port raw content filtering, hidden-text detection, caption association, and image-reference plumbing.
- [ ] Bring benchmark table metrics above threshold.

### Phase 7: Stretch / Deferred Parity

- [ ] Add struct-tree hooks for tagged PDFs if still required.
- [ ] Revisit JSON root metadata parity that depends on low-level PDF objects.
- [ ] Decide on annotated PDF and PDF rewrite scope.
- [ ] Restore hybrid integration only after local-mode replacement is stable.

## Phase Exit Gates

| Phase | Must be true before commit/push |
| --- | --- |
| Phase 0 | Scope, contracts, and deferred features are documented in repo |
| Phase 1 | Go model/context can express the current schema without Java runtime state |
| Phase 2 | CLI smoke tests and emitter/schema tests pass against fixtures |
| Phase 3 | Reading order and paragraph grouping are stable on fixture goldens |
| Phase 4 | Headings/lists/levels work on fixture goldens and simple PDFs |
| Phase 5 | Local extraction runs on PDFs without Java and without `ledongthuc/pdf` |
| Phase 6 | Benchmark thresholds are met for local mode |
| Phase 7 | Deferred features have explicit acceptance tests or remain deferred |

## Commit Policy

- [ ] Commit once per completed phase or sub-phase gate, not per tiny edit.
- [ ] Use commit subjects like `phase-0 contract freeze`, `phase-1 go model`, `phase-3 reading-order`.
- [ ] Push after the local gate for that phase passes.
- [ ] Record benchmark/test evidence in commit message body or adjacent `.context` note.

## Current Completed Checkpoints

- [x] `phase-0: add pure-go phase tracker`
- [x] `phase-1: draft native ingestion api`
- [x] `phase-1: add native ingestion interface scaffold`
- [x] `phase-1: add native handle adapter and raw fixture schema`
- [x] `phase-1: add fixture-backed native loader`
- [x] `phase-1: route .raw.json through nativepdf and mark pdftext temporary`
- [x] `phase-1: add native raw goldens and stop cross-column paragraph collapse`
- [x] `phase-1: split runtime selection out of cmd and isolate pdftext as a bridge`
- [x] `phase-1: prefer native backend and keep pdftext as fallback bridge`
- [x] `phase-1: native pdf path now flows through a loader placeholder via nativepdf.NewIngestor(loader)`
- [x] `phase-1: add real native pdf loader skeleton and handle shell`
