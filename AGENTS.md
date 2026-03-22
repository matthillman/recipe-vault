# Repository Guidelines

## Project Structure & Module Organization

This repo is a small content mono-repo: Markdown recipes/notes, a Go card generator, and a Vite recipe browser.

- `recipes/`: One recipe (or component/framework) per file, split out for easier browsing and linking.
  - Naming: `recipes/<kebab-case-slug>.md` (example: `recipes/overnight-sourdough-focaccia.md`).
- `recipe-list.md`: Optional canonical “all-in-one” rollup (curated; don’t update by accident).
- `notes/`: Knowledge base (techniques, experiments, formulas, distilled learnings).
- `notes/modernist/`, `notes/science/`, `notes/beverages/`: Modernist + science cooking and drink work.
- `templates/`: File templates for recipes/notes/experiments.
- `project/`: Workflow rules and the single backlog entry point (`project/ROADMAP.md`).
- `tasks/`: Numbered multi-step work plans (start with `tasks/000-sample.md` after linking work from the roadmap).
- `reference/`: Supporting notes/materials (treat as read-mostly unless you’re intentionally updating references).
- `cmd/`, `internal/`: Go tooling built on the Markdown content.
- `site/`: Vite app for browsing recipes.
- `scripts/`: Repo checks and maintenance helpers.

## Build, Test, and Development Commands

Prefer the root `Makefile` so the common workflows stay discoverable.

- `make validate-recipes`: Validate recipe filenames, required sections, section order, and duplicate titles.
- `make check`: Run recipe validation and Go tests when `go` is installed.
- `make cards-list`: List recipe slugs for the Go recipe card CLI.
- `make cards-build RECIPES=slug1,slug2`: Build selected recipe cards into `out/cards.pdf`.
- `make site-dev`: Start the Vite dev server from `site/`.
- `rg "focaccia" recipes/ recipe-list.md`: Search across the recipe corpus.
- `rg "hydration" recipes/ notes/`: Search recipes plus the knowledge base.

## Coding Style & Naming Conventions

- Use Markdown (`.md`) with clear headings: `# Title`, `## Ingredients`, `## Process`, etc.
- Recipe files should include exactly one title, a `**Yield / Target**` or `**Yield / Pan Target**` block, `## Ingredients`, and `## Process`.
- Prefer `Ingredient name: amount` inside ingredient bullets; keep `## Process` as a numbered list.
- Keep quantities metric-first (grams) when possible; include helpful targets (yield/pan size, dough weight).
- Prefer short, actionable steps; avoid long paragraphs.
- File names: lowercase kebab-case; avoid spaces and punctuation.

## Testing Guidelines

- Run `make validate-recipes` after editing recipe files.
- Run `make check` after touching the Go tooling or shared parsing logic.
- Keep “base recipes” stable; add variants as separate files rather than mutating a working baseline.
- If Go is unavailable in the environment, note that tests were not run.

## Commit & Pull Request Guidelines

Git history is minimal (currently a single commit: “Starting files”), so there is no established convention yet.

- Commits: use imperative, scoped summaries when helpful (e.g., `recipes: add standard sourdough loaf`).
- PRs (if used): include a brief description, list of files changed, and any rationale for recipe changes (why/what improved).

## Agent-Specific Notes

- Prefer adding/editing individual files in `recipes/` and only updating `recipe-list.md` when you intend to change the canonical rollup.
- Treat `recipes/*.md` as source-of-truth; `site/public/recipes/` is generated output.
- Use `project/ROADMAP.md` as the single backlog. Create or update a `tasks/NNN-*.md` file only when the work needs a multi-step execution plan.
- Avoid reorganizing `reference/` without a clear reason and a short note in the commit message.
