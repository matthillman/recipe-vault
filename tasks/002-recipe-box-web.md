# Task 002: Recipe Box Web UI

## Progress Summary

**Status**: In Progress (v0 scaffold exists)

- [x] Step 1: Add Vite app scaffold under `site/`
- [x] Step 2: Add recipe sync + manifest generation
- [x] Step 3: Add mobile-first list + recipe view
- [x] Step 4: Add Cloudflare Pages deploy notes
- [ ] Step 5: Iterate on “how I think” UX (favorites, recents, tags)

## Overview

Create a mobile-first “recipe box” web page that makes it fast to browse and read the Markdown files in `recipes/`.

## Current State

- Recipes live as Markdown files in `recipes/`.
- A minimal web UI exists under `site/` and pulls in recipes via a generated manifest.
- Deployment target is Cloudflare Pages (static site).

## Target State

- Open the site on a phone and quickly:
  - search recipes
  - open a recipe with minimal taps/scroll
  - keep a short set of “go-to” recipes pinned

## Implementation Steps (Next Iterations)

1. Add **favorites** (localStorage) and show a “Pinned” section at the top.
2. Add **recently viewed** list (localStorage, bounded).
3. Add optional **front matter metadata** (`tags`, `source`, `yield`) and display chips in list view.
4. Add **recipe scaling** helper (e.g., 0.5× / 1× / 2×) with lightweight parsing for common `X g` quantities.
5. Add a “**keep screen awake**” toggle while cooking (Wake Lock API, best-effort).
6. Optional: add offline support (PWA + cache) once the UX feels right.

## Acceptance Criteria (For “v1”)

- [ ] Favorites can be toggled from list and recipe view
- [ ] Favorites persist across reloads (localStorage)
- [ ] Recently viewed list is visible and capped (e.g., 10)
- [ ] Search remains fast and usable on mobile

## Files Involved

- Modified:
  - `site/src/ui/app.ts`
