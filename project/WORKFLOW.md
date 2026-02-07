# Coworking Workflow (Humans + Agents)

This repo is optimized for short, repeatable loops: **capture → test → document → file**.

## Core Principles

- **Assume no memory**: anything important must land in a file (not just chat).
- **Prefer small diffs**: add a new note/recipe rather than heavily rewriting an existing “stable” one.
- **Separate sources from conclusions**:
  - Put raw/reference material in `reference/`.
  - Put our distilled takeaways in `notes/` (link back to the source file).

## What To Create (Choose One)

- **Recipe** (`recipes/*.md`): production-ready, repeatable. Use `templates/recipe.md`.
- **Formula** (`notes/formulas/*.md`): ratios, baker’s % tables, calculators, conversion notes.
- **Experiment log** (`notes/experiments/YYYY-MM-DD-*.md`): what we tried, variables, results, next steps.
- **Technique / Ingredient / Equipment note** (`notes/**/`): concise “how/why/when” guidance.
- **Task plan** (`tasks/NNN-*.md`): multi-step work with checkboxes and acceptance criteria.

## Definition Of Done (For Any Change)

- The change is in the right folder with a clear file name.
- The file answers:
  - What is it?
  - How do I use it?
  - What can go wrong?
  - What did we learn / decide?
- If we referenced an external source, there’s a pointer in `reference/` and a summary in `notes/sources/`.

## Minimal Review Checklist

- Titles are meaningful (`# ...`) and sections are skimmable.
- Quantities are consistent (especially baker’s % and dough weights).
- “Framework” notes don’t masquerade as tested recipes (label them clearly).

