# Task 003: Recipe Import Pipeline

## Goal

Capture recipe pages, public PDF/image URLs, or shared text from a CLI, iPhone Share Sheet, or private Custom GPT; normalize them conservatively; and produce a validated draft pull request for human review.

## Progress Summary

**Status**: Implemented locally; GitHub labels/secrets, Cloudflare secrets, deployment, and client setup remain operational activation steps.

## Acceptance Criteria

- [x] `make recipe-import URL=...` creates a canonical recipe from supported JSON-LD pages.
- [x] OpenAI normalization uses structured output and never writes model-authored Markdown directly.
- [x] Failed or ambiguous imports do not create invalid recipe files.
- [x] An authenticated Cloudflare Pages endpoint queues imports as GitHub issues.
- [x] A GitHub workflow converts queued issues into validated draft pull requests.
- [x] iPhone Shortcut and private Custom GPT setup are documented against the same API.
- [x] Public PDF and supported image URLs use the existing URL capture contract and multimodal normalization.
- [x] Recipe, Go, site, and API tests pass.

## Implementation Steps

1. [x] Add provenance conventions to the recipe template.
2. [x] Build deterministic extraction, conservative normalization, rendering, and duplicate detection in Go.
3. [x] Add GitHub issue decoding and issue-to-draft-PR automation.
4. [x] Add the authenticated Pages Function and OpenAPI document.
5. [x] Document iPhone Shortcut, Custom GPT, secrets, and deployment setup.
6. [x] Exercise fixtures, validation, Go tests, and the site build.

## Decisions

- GitHub Issues are the durable job queue and audit trail.
- Imports remain `Imported, untested` and reach the vault only through a draft PR.
- Recipe JSON-LD is preferred; supplied text is the fallback; PDF/image sources use multimodal normalization and require review for OCR errors.
- Missing facts are reported, never invented.
- Raw webpages are not archived.
