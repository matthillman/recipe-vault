# Roadmap

This is the single entry point for planned work. Keep it short, prioritized, and current.

## Workflow

1. Add a new item below (top = highest priority).
2. For multi-step work, create a task plan in `tasks/` and link it here.
3. Refactors and small follow-ups belong here too; do not open a separate queue.
4. When done, mark ✅ and add a one-line outcome.

## Active / Next

- [ ] Backfill canonical `## Notes` or equivalent notes sections across recipe files where useful

## Backlog

- [ ] Define a standard recipe template and migrate older recipes to it (See: `templates/recipe.md`)
- [ ] Add a “baker’s % quick reference” note for common dough types (See: `notes/formulas/`)
- [ ] Add a modernist ingredient “cheat sheet” (typical % ranges, hydration methods, incompatibilities)
- [ ] Write a protocol note for juice clarification aimed at bottled/canned cocktails (haze control + shelf stability)

## Done

- ✅ Create an index for `notes/` so future threads can discover prior learnings quickly
- ✅ Tighten recipe validation beyond headings (yield presence, duplicate titles, and section structure)
- ✅ Establish a root `Makefile`, recipe validator, and a single roadmap-first workflow
- ✅ Build a Go CLI to generate printable 2x4 recipe cards as an 8-up PDF (See: `tasks/001-recipe-cards-pdf-generator.md`)
- ✅ Split `recipe-list.md` sections into individual files in `recipes/`
- ✅ Add a standard sourdough loaf recipe file
- ✅ Add contributor guidelines in `AGENTS.md`
