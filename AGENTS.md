# Repository Guidelines

## Project Structure & Module Organization

This repo is a small Markdown-based cookbook/knowledge base.

- `recipes/`: One recipe (or component/framework) per file, split out for easier browsing and linking.
  - Naming: `recipes/<kebab-case-slug>.md` (example: `recipes/overnight-sourdough-focaccia.md`).
- `recipe-list.md`: Optional canonical “all-in-one” rollup (curated; don’t update by accident).
- `notes/`: Knowledge base (techniques, experiments, formulas, distilled learnings).
- `notes/modernist/`, `notes/science/`, `notes/beverages/`: Modernist + science cooking and drink work.
- `templates/`: File templates for recipes/notes/experiments.
- `project/`: Coworking workflow + backlog (start with `project/WORKFLOW.md`).
- `tasks/`: Numbered multi-step work plans (start with `tasks/000-sample.md`).
- `reference/`: Supporting notes/materials (treat as read-mostly unless you’re intentionally updating references).

## Build, Test, and Development Commands

There is no build system or app runtime in this repository.

- `ls recipes/`: Quick browse of recipe files.
- `rg "focaccia" recipes/ recipe-list.md`: Search across the collection.
- `rg "hydration" recipes/ notes/`: Search recipes plus our knowledge base.

## Coding Style & Naming Conventions

- Use Markdown (`.md`) with clear headings: `# Title`, `## Ingredients`, `## Process`, etc.
- Keep quantities metric-first (grams) when possible; include helpful targets (yield/pan size, dough weight).
- Prefer short, actionable steps; avoid long paragraphs.
- File names: lowercase kebab-case; avoid spaces and punctuation.

## Testing Guidelines

No automated tests are currently defined.

- Sanity-check new/edited recipes by ensuring headings render well and quantities are internally consistent.
- Keep “base recipes” stable; add variants as separate files rather than mutating a working baseline.

## Commit & Pull Request Guidelines

Git history is minimal (currently a single commit: “Starting files”), so there is no established convention yet.

- Commits: use imperative, scoped summaries when helpful (e.g., `recipes: add standard sourdough loaf`).
- PRs (if used): include a brief description, list of files changed, and any rationale for recipe changes (why/what improved).

## Agent-Specific Notes

- Prefer adding/editing individual files in `recipes/` and only updating `recipe-list.md` when you intend to change the canonical rollup.
- Avoid reorganizing `reference/` without a clear reason and a short note in the commit message.
