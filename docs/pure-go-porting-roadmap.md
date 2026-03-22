# Pure Go Porting Roadmap

## Goal

Build a Pure Go PDF core for `opendataloader-pdf-go` without depending on Java,
JNI, subprocess bridges, or third-party PDF parsing libraries such as
`github.com/ledongthuc/pdf`.

The target is not a thin wrapper around the existing Java runtime. The target
is a native Go pipeline that owns:

- PDF container parsing
- object/xref/stream decoding
- text extraction
- image and path extraction
- table and reading-order primitives
- tagged-PDF hooks for future structure-tree work

## Current Status

The Go project already has meaningful downstream infrastructure:

- CLI, option parsing, and discovery
- document and raw-artifact models
- emitters for JSON, Markdown, HTML, and text
- heuristic passes for paragraph assembly, headings, lists, tables, and reading order
- fixture-backed ingestion for tests and goldens

The missing core is the low-level PDF engine. The current `nativepdf`
implementation is still a skeleton: it derives page shells, decodes simple
Flate streams, and extracts minimal text shells from `Tj`/`TJ` patterns.

## What Must Be Ported

The Java core is not just application logic. It is deeply coupled to veraPDF
model types and containers. For the Pure Go effort, separate the work into two
layers:

1. Replacement for veraPDF-backed PDF primitives
2. Port of project-specific semantic processors

### Layer 1: Pure Go PDF primitives

- header/trailer/xref parser
- indirect object loader
- stream decoding
- page tree traversal
- inherited page attributes
- resource lookup
- content stream operator tokenizer
- graphics state stack
- text state and text matrices
- font resource handling
- ToUnicode/CMap decoding
- glyph-to-Unicode resolution fallback chain
- image XObject decoding hooks
- path/line extraction for tables and layout
- annotation/link extraction
- structure-tree hooks

### Layer 2: Project-specific processing

- content filtering
- text chunk cleanup and merge logic
- line grouping
- paragraph assembly
- header/footer detection
- list detection
- table detection and table-border logic
- reading order
- heading detection and level inference
- caption attachment
- hybrid triage integration

## Recommended Milestones

### Milestone 0: Stabilize the seam

Status: in progress

- Keep `internal/ingest/nativepdf` as the sole low-level PDF seam.
- Route current skeleton behavior through parser-oriented types instead of
  continuing ad-hoc extraction inside the loader.
- Preserve fixture-backed tests as the semantic reference layer.

Exit criteria:

- loader code no longer owns parsing logic directly
- page/artifact outputs are produced via a parser result object

### Milestone 1: Real container parser

- parse header, trailer, xref table, and indirect objects
- resolve `/Catalog`, `/Pages`, `/Page`
- read page count, boxes, rotation, and page labels correctly
- decode `FlateDecode`

Exit criteria:

- page count and page metadata come from parsed objects, not regex shell logic
- object traversal works on real PDFs with non-trivial page trees

### Milestone 2: Text artifact extraction

- tokenize page content streams
- implement basic graphics/text operators: `BT`, `ET`, `Tf`, `Tm`, `Td`, `TD`,
  `Tj`, `TJ`, `'`, `"`
- resolve fonts and ToUnicode maps
- emit positioned text artifacts with per-fragment bounds and style

Exit criteria:

- downstream Go paragraph/list/heading pipeline operates on real text artifacts
- skeleton literal-string fallback is no longer needed for standard digital PDFs

### Milestone 3: Graphics and table primitives

- extract stroked path segments and rectangles
- normalize horizontal/vertical lines for table candidates
- extract image artifacts and bounds

Exit criteria:

- `TableCandidates` and raw line/path artifacts are driven by the parser
- table heuristics can use actual page geometry instead of text-only fallback

### Milestone 4: Fidelity and parity work

- duplicate/overprint handling
- hidden text and clipping edge cases
- reading-order fidelity improvements
- font metrics and spacing refinements
- encrypted PDFs and password path

Exit criteria:

- benchmark deltas become attributable to heuristic differences rather than
  parser incompleteness

### Milestone 5: Tagged PDF and structure tree

- parse `/StructTreeRoot`
- map marked content to extracted artifacts
- expose structure-tree nodes through `nativepdf.StructTree`

Exit criteria:

- Go side can support future tagged-PDF logic without the Java runtime

## Implementation Order Inside `nativepdf`

1. `parser.go`
   Define parser result types and the phased parser entry point.
2. `lex_*.go`
   Tokenization for objects, streams, and content operators.
3. `file_*.go`
   Xref, trailer, indirect object, and page tree parsing.
4. `text_*.go`
   Text state, fonts, encodings, ToUnicode, glyph mapping.
5. `graphics_*.go`
   Path, image, and line extraction.
6. `struct_*.go`
   Structure-tree parsing and marked-content linking.

## Non-goals For The First Cut

- full PDF/UA generation
- incremental update writing
- perfect support for every filter and image codec on day one
- scanned-PDF OCR in the native parser layer

## Working Rule

Do not chase Java parity by porting processor classes in source order.
First achieve artifact parity. The existing Go heuristics can only be trusted
once the low-level page artifacts are trustworthy.
