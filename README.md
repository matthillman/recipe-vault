# Cooking Lab Notes

This repository is a lightweight cookbook + cooking knowledge base designed for **human/agent coworking**. The goal is simple: if we learn something, we write it down here so a future thread can pick up with **zero prior context**.

## Where Things Live

- Recipes (one per file): `recipes/`
- Canonical rollup (optional, curated): `recipe-list.md`
- Knowledge base (our learnings): `notes/`
- Modernist + science cooking notes: `notes/modernist/`, `notes/science/`
- Beverage + cocktail work: `notes/beverages/`
- Project workflow + backlog: `project/`
- Larger work items (numbered plans): `tasks/`
- Reusable starting points: `templates/`
- External/reference material: `reference/`

## Quick Start (Humans + Agents)

1. Pick the artifact you’re creating:
   - A recipe: add a file in `recipes/`
   - A new technique/ingredient/equipment note: add a file under `notes/`
   - A bigger initiative (re-org, standardization, audits): create a `tasks/NNN-*.md` plan
2. Record what we learned (even if the recipe itself is unchanged): add/append to `notes/`.
3. Keep the repo navigable: update `recipe-list.md` only when you intend to change the canonical rollup.

See `project/WORKFLOW.md` for the full collaboration rules.

## Recipe Card PDFs

There is a small Go CLI that can generate printable 2.5"x4" portrait recipe cards (8-up on US Letter landscape):

```bash
# List recipe slugs
env GOCACHE=/tmp/gocache go run ./cmd/recipecards list

# Build a PDF for selected recipes
env GOCACHE=/tmp/gocache go run ./cmd/recipecards build \
  --out out/cards.pdf \
  --recipes standard-sourdough-loaf,overnight-sourdough-focaccia
```

Notes:

- This environment may require `GOCACHE=/tmp/gocache` (the default Go build cache path can be unwritable).
- The generator prefers an explicit `## Card` section in a recipe file if you want a hand-authored “perfect” card.
