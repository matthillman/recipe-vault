# Task 001: Recipe Cards PDF Generator (Go)

## Progress Summary

**Status**: Not Started

**Status**: Implemented (v1)

- [x] Confirm card size/orientation and paper size defaults
- [x] Define a “recipe card” extraction convention for `recipes/*.md`
- [x] Implement CLI: list/select/build
- [x] Implement PDF renderer + 8-up page layout
- [x] Add baker’s % support (render explicit formula sections when present)
- [x] Add tests for parsing + layout math
- [x] Document usage in `README.md`

## Follow-Ups (Optional)

- [ ] Add `--all` to build cards for every recipe in `recipes/`
- [ ] Improve yield extraction by scanning in-order headings (currently heuristic)
- [ ] Add an explicit `## Card` section to the most-used recipes for “perfect” card content

## Overview

Build a small Go CLI that selects recipes from `recipes/` and generates a printable PDF of cut-out cards for a notebook. The PDF should lay out **~8 cards per page** and produce consistently readable output (title, key ingredients, key steps, and optionally baker’s %).

## Current State

- Recipes are Markdown files in `recipes/` with headings like `## Ingredients`, `## Process`, and sometimes explicit baker’s % sections.
- There is no existing codebase, build tooling, or test setup.

## Target State

- A Go program at `cmd/recipecards/` that can:
  - List available recipes (`list`)
  - Generate a PDF (`build`) from a chosen subset
- Output PDF supports:
  - Letter (default) and A4
  - 8-up layout when feasible
  - A predictable “card template”

## Design Decisions (Proposed)

### Card Size vs “8 Per Page”

On US Letter (8.5x11):

- A true **2x4 inch portrait** card (2" wide x 4" tall) *can* fit 8-up **if the page is landscape**:
  - Page (landscape): 11" wide x 8.5" tall
  - Layout: 4 columns (4 x 2" = 8") and 2 rows (2 x 4" = 8")
  - Remaining space: 3" horizontal (margins + gutters) and 0.5" vertical (margins and/or a row gutter)
- A wider **2.5x4 portrait** card can also fit 8-up on landscape letter (4 x 2.5" = 10"), but vertical clearance is still tight (only 0.5" total for top/bottom margins + any row gutter). This is doable, but needs small margins and/or zero row gutter.

Action:

- Default card: **2.5x4 inches, portrait cards on a landscape page** (4 columns x 2 rows = 8-up)
- Provide `--card 2x4p` (optional) for a smaller card with more margin headroom
- Provide `--page-orient portrait|landscape` (default `landscape` for 8-up)

Default layout constants (for `2.5x4p` on letter landscape):

- Left/right margins: 0.25"
- Top/bottom margins: 0.25"
- Column gutters: 1/6" (~0.167")
- Row gutter: 0" (rows share a cut line)

### Extraction Convention

Priority order when building a card:

1. If a recipe contains `## Card`, use that block verbatim (best for hand-authored cards).
2. Else auto-extract:
   - Title from first `# ...`
   - Yield/target from the `**Yield / Target**` or `**Yield / Pan Target**` block (first 1–3 bullets)
   - Ingredients from `## Ingredients` (first N lines, prefer gram lines)
   - Process from `## Process` (first N steps)
3. If present, include baker’s % from:
   - `## Formula` / `## Formula (Baker's %)` sections, or lines containing `%`

### Baker’s % (V1 vs V2)

- V1: render **explicit** baker’s % sections when present.
- V2 (optional): compute baker’s % when a recipe includes:
  - `Total flour: <g>` and ingredient weights in grams
  - Recognized flour lines (contain `flour`)

## Implementation Plan

### 1) CLI

Binary: `recipecards`

- `recipecards list`
  - prints slug + title + path
- `recipecards build --out out/cards.pdf --recipes overnight-sourdough-focaccia,standard-sourdough-loaf`
  - selects by slug (filename without `.md`)
- Flags:
  - `--paper letter|a4`
  - `--page-orient portrait|landscape`
  - `--card 2x4p|2.5x4p|4x2` (optional)
  - `--per-page 8` (default; adjusts if geometry can’t fit)
  - `--font-size` (optional)

### 2) Markdown Parsing (No Heavy Dependencies)

- Implement a small line-based parser in `internal/recipes/`:
  - Heading detection (`#`, `##`)
  - Bullet list parsing (`- `, `* `)
  - Numbered list parsing (`1.`)
  - Block extraction between headings

### 3) PDF Generation

- Use a dependency-free PDF writer in `internal/pdf/` (this environment does not allow downloading Go modules):
  - Built-in Type1 fonts (Helvetica/Courier)
  - Text wrapping to card width
  - Thin cut lines / crop marks optional

### 4) Layout Engine

- Compute card grid based on:
  - Paper dimensions in points (1in = 72pt)
  - Margins + gutters
  - Card width/height
- If the requested geometry doesn’t fit (card size + margins + gutters), the renderer should:
  - Reduce gutters toward zero first
  - Then reduce cards-per-page (or suggest `--page-orient landscape`)
- Place cards row-major; create new pages as needed.

### 5) Tests

- Parser unit tests:
  - Extract title, ingredients, process, formula blocks from known recipe files
- Layout tests:
  - 4x2 cards on letter produce 2x4 grid
  - Portrait card reduces cards-per-page

## Acceptance Criteria

- [ ] `go test ./...` passes
- [ ] `go run ./cmd/recipecards list` shows recipe slugs from `recipes/`
- [ ] `go run ./cmd/recipecards build --out out/cards.pdf --recipes standard-sourdough-loaf` generates a valid PDF
- [ ] Default settings produce 8-up on US Letter with readable text
- [ ] Baker’s % appears on cards when an explicit formula section exists

## Files Involved

- New:
  - `cmd/recipecards/main.go`
  - `internal/recipes/*`
  - `internal/pdf/*` (if dependency-free)
  - `tasks/001-recipe-cards-pdf-generator.md`
- Modified:
  - `README.md` (usage)
