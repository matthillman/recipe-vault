# Cooking Lab Notes

This repository is a small content mono-repo:

- `recipes/` and `notes/` are the source-of-truth knowledge base
- `cmd/` and `internal/` contain small Go tools built on top of that content
- `site/` is a Vite app for browsing the recipes on the web

The operating principle is simple: if we learn something, we write it down in a durable file so a future human or agent can resume work with minimal context.

## Repository Layout

- `recipes/`: production-ready recipes and recipe components
- `recipe-list.md`: curated rollup; update only intentionally
- `notes/`: distilled learnings, experiments, formulas, techniques, ingredients, equipment
- `reference/`: raw external material and supporting documents
- `templates/`: starting points for recipes, notes, experiments, and protocols
- `project/`: workflow rules and the single backlog entry point
- `tasks/`: numbered multi-step execution plans
- `cmd/`, `internal/`: Go tooling for recipe cards and parsing
- `site/`: static recipe browser
- `scripts/`: lightweight repo checks and maintenance helpers

## Common Commands

```bash
# Validate recipe structure
make validate-recipes

# Validate recipes and run Go tests when Go is available
make check

# List recipe slugs for the card generator
make cards-list

# Build a PDF for selected recipes
make cards-build RECIPES=standard-sourdough-loaf,overnight-sourdough-focaccia

# Run the web UI locally
make site-dev

# Import a recipe page into a reviewable local Markdown file
make recipe-import URL=https://example.com/recipe
```

Notes:

- `make check` skips Go tests if `go` is not installed.
- The site sync step copies `recipes/*.md` into `site/public/recipes/`; those generated files are not source-of-truth.
- The card generator may require `GOCACHE=/tmp/gocache`; the `Makefile` handles that.
- Recipe import uses JSON-LD without an API key when possible. Set `OPENAI_API_KEY` for conservative AI normalization and text-only imports.

## Recipe Schema

The validator treats the following layout as canonical:

- One `# Title`
- One `**Yield / Target**` or `**Yield / Pan Target**` block followed by bullet lines
- One `## Ingredients` section with bullet items
- One `## Process` section with numbered steps
- Optional sections such as `## Formula`, `## Notes`, and `## Card`

Preferred conventions:

- Ingredient bullets use `name: amount`
- Quantities are metric-first when practical
- `## Ingredients` appears before `## Process`
- `## Notes` captures variations, failure modes, storage, or decisions worth retaining

## Workflow

1. Add or update the right artifact:
   - recipe in `recipes/`
   - note in `notes/`
   - multi-step initiative in `tasks/` after adding it to `project/ROADMAP.md`
2. Record the decision or learning in a durable file, not just chat.
3. Run `make validate-recipes` for recipe changes and `make check` when touching tooling.
4. Track larger work in `project/ROADMAP.md`.

Start with [project/WORKFLOW.md](/Users/matt/Code/recipe-vault/project/WORKFLOW.md) for collaboration rules and [project/ROADMAP.md](/Users/matt/Code/recipe-vault/project/ROADMAP.md) for the active queue.

## Tooling Surfaces

### Recipe Card PDFs

The Go CLI under [cmd/recipecards/main.go](/Users/matt/Code/recipe-vault/cmd/recipecards/main.go) generates printable 2.5"x4" portrait recipe cards laid out 8-up on US Letter landscape. It prefers an explicit `## Card` section when a recipe needs a hand-authored card layout.

### Recipe Box Web UI

The mobile-first web UI under [site/](/Users/matt/Code/recipe-vault/site) browses the Markdown recipes directly from generated copies under `site/public/recipes/`. It is deployed on Cloudflare Pages from `matthillman/recipe-vault` with root directory `site`, build command `npm run build`, output directory `dist`, and automatic production deploys from `main`. See [site/README.md](/Users/matt/Code/recipe-vault/site/README.md) for the full deployment configuration.

### Recipe Capture and Import

The importer under `cmd/recipeimport/` accepts recipe URLs or supplied text, extracts structured recipe facts, normalizes them conservatively, and renders vault-format Markdown. The hosted capture path queues an authenticated request as a GitHub issue; GitHub Actions turns successful jobs into draft pull requests. See [docs/recipe-import.md](/Users/matt/Code/recipe-vault/docs/recipe-import.md) for setup and operations.
