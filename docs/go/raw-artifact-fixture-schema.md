# Raw Artifact Fixture Schema

This document defines the JSON fixture format for Pure Go native ingestion tests.
It is intended for stable, fixture-driven development of the raw artifact layer
before the production parser is complete.

The fixture is not the runtime output format. It is a debugging and regression
format for page-level extraction, geometry, ordering, and style signals.

## Purpose

- Capture raw artifacts in a deterministic, portable form.
- Let downstream heuristics run without depending on a live PDF parser.
- Provide goldens for ingestion, table detection, reading order, and semantic
  reconstruction work.
- Make parser regressions easy to diff in tests.

## Top-Level Structure

The fixture root is a JSON object with these top-level keys:

- `metadata`: document-level information
- `pages`: ordered list of page objects
- `table_candidates`: optional document-level table hints

### `metadata`

Document metadata is required and should contain at least:

- `file_name`: source file name or logical fixture name
- `page_count`: total page count in the fixture

Optional document metadata fields:

- `title`
- `author`
- `producer`
- `creator`
- `language`
- `subject`
- `keywords`
- `created_at`
- `modified_at`

### `pages`

`pages` is an array of page objects in page order. Page order must match the
one-based `number` field and zero-based `index` field.

## Page Object

Each page object has:

- `metadata`: page metadata
- `artifacts`: ordered raw artifacts on the page
- `table_candidates`: optional page-local table hints
- `struct_tree`: optional tagged-PDF hook

### Page Metadata

Required fields:

- `number`: one-based page number
- `index`: zero-based page index

Optional fields:

- `width`
- `height`
- `rotation`
- `label`
- `crop_box`
- `media_box`

Page geometry fields use the same box shape as artifact bounds:

```json
{ "left": 0, "bottom": 0, "right": 612, "top": 792 }
```

## Artifact Object

Each artifact represents one low-level extracted page item.

Required fields:

- `kind`
- `page_index`
- `page_number`
- `sequence`

Recommended fields:

- `bounds`
- `boxes`
- `style`

Optional fields:

- `text`
- `data`
- `format`
- `links`
- `source_id`
- `tags`

### Artifact Kinds

Supported `kind` values are:

- `text`
- `image`
- `line`
- `path`
- `shape`

#### `text`

Required:

- `text`
- `bounds`

Optional:

- `boxes`
- `style`
- `links`
- `tags`

#### `image`

Required:

- `data` or `source_id`
- `bounds`

Optional:

- `format`
- `boxes`
- `style`
- `links`
- `tags`

#### `line`

Required:

- `bounds`

Optional:

- `boxes`
- `style`
- `links`
- `tags`

#### `path`

Required:

- `bounds`

Optional:

- `boxes`
- `style`
- `links`
- `tags`

#### `shape`

Required:

- `bounds`

Optional:

- `boxes`
- `style`
- `links`
- `tags`

## Sequence And Order Rules

- `sequence` is the stable per-page order of artifacts as extracted.
- `sequence` must start at `0` for each page.
- `sequence` values must be unique within a page.
- Artifacts in `pages[].artifacts` must be listed in increasing `sequence`
  order.
- If extraction order is ambiguous, preserve the parser's native stable order
  and do not sort in test fixtures.

## Bounds And Boxes

### `bounds`

`bounds` is the primary bounding box for the artifact.

Required shape:

```json
{ "left": 0, "bottom": 0, "right": 100, "top": 20 }
```

Rules:

- `left <= right`
- `bottom <= top`
- coordinates are in page space
- use the same coordinate convention across all artifacts in the fixture

### `boxes`

`boxes` is an optional list of secondary boxes for multi-segment artifacts.
Use it when a single artifact spans multiple disjoint rectangles.

Rules:

- `boxes` may be omitted when `bounds` is sufficient
- `boxes` should be ordered in visual reading order
- `bounds` should enclose every box in `boxes`
- if `boxes` is present, it must not be empty

## Style Fields

`style` is optional but recommended for text artifacts and useful for later
semantic scoring.

Suggested fields:

- `font`
- `font_size`
- `text_color`
- `bold`
- `italic`
- `underline`
- `hidden_text`
- `content`

Rules:

- `content` should mirror `text` when the parser provides it separately.
- `font_size` should be numeric.
- boolean fields should be explicit booleans, not strings.
- omit unsupported style signals instead of inventing placeholder values.

## Table Candidate Extension Point

The fixture may include table candidate hints to support native ingestion work
before full table reconstruction exists.

Document-level and page-level `table_candidates` may contain:

- `horizontal_lines`
- `vertical_lines`
- `rectangles`

Suggested line segment shape:

```json
{
  "start": { "x": 0, "y": 0 },
  "end": { "x": 100, "y": 0 },
  "width": 1,
  "stroke_rgb": "#000000",
  "page_index": 0
}
```

Rules:

- hints are optional
- hints must not replace raw artifact records
- hints should only describe geometry already present in the source parser
- use them to test table detection logic, not as a second semantic source of
  truth

## Validation Rules

A fixture is valid only if all of the following are true:

- `metadata.file_name` is present
- `metadata.page_count` is present and non-negative
- `pages` length matches `metadata.page_count`
- every page has both `metadata.number` and `metadata.index`
- page numbers are one-based and indices are zero-based
- every artifact has `kind`, `page_index`, `page_number`, and `sequence`
- every artifact belongs to the page that contains it
- artifact sequences are unique and increasing within a page
- required fields are present for the artifact kind
- every `bounds` box is normalized

Optional validation for stronger tests:

- compare `text` content against golden output
- compare `bounds` values with a numeric tolerance
- compare `style` fields only when the source parser is expected to recover
  them
- validate table candidate geometry against expected line counts

## Minimal Example

```json
{
  "metadata": {
    "file_name": "sample.pdf",
    "page_count": 1
  },
  "pages": [
    {
      "metadata": {
        "number": 1,
        "index": 0,
        "width": 612,
        "height": 792,
        "rotation": 0
      },
      "artifacts": [
        {
          "kind": "text",
          "page_index": 0,
          "page_number": 1,
          "sequence": 0,
          "bounds": {
            "left": 72,
            "bottom": 720,
            "right": 180,
            "top": 736
          },
          "text": "Hello world",
          "style": {
            "font": "Helvetica",
            "font_size": 12,
            "bold": false,
            "italic": false,
            "underline": false,
            "hidden_text": false,
            "content": "Hello world"
          }
        }
      ]
    }
  ]
}
```

## Notes

- Keep fixtures small and focused.
- Prefer one behavior per fixture when possible.
- If a parser can emit richer raw signals, add them behind optional fields
  rather than changing the base shape.
- When the runtime native ingestion layer becomes stable, this schema should
  stay as the regression format even if internal types evolve.
