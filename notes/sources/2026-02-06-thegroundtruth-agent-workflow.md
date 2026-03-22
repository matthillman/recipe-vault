# Source Summary: “My Claude Code Workflow and Personal Tips” (thegroundtruth.media)

## Why This Matters Here

We want future threads (human or agent) to be productive immediately, with no prior chat context. The workflow in this article emphasizes **lightweight, file-based memory**: a roadmap, scoped task plans, and durable repo conventions.

## Key Takeaways (Distilled)

- Maintain a single **roadmap** as the entry point for planned work.
- Use **numbered task plans** for multi-step initiatives so execution is repeatable.
- Store reusable context and conventions in repo files so the process survives across sessions.

## How We Applied It In This Repo

- Planning + workflow live in `project/`:
  - `project/ROADMAP.md`
  - `project/WORKFLOW.md`
- Multi-step initiatives live in `tasks/` (see `tasks/000-sample.md`).
- Small follow-ups and refactors are folded back into `project/ROADMAP.md` instead of separate queue files.
- “What we learned” is captured in `notes/` so decisions and technique insights aren’t lost in chat.

## Link

See: `https://thegroundtruth.media/p/my-claude-code-workflow-and-personal-tips`
