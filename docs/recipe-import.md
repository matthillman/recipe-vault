# Recipe Import Setup and Operations

## Data Flow

1. A CLI, iPhone Shortcut, or private Custom GPT submits a public HTTPS recipe page, PDF, supported image URL, and/or recipe text.
2. The Cloudflare Pages Function authenticates the caller and creates a versioned GitHub issue.
3. GitHub Actions extracts recipe JSON-LD or sends validated PDF/image bytes to OpenAI, asks for conservative structured normalization, renders Markdown, and runs `make check`.
4. A valid import becomes a draft pull request. Failure leaves the issue open with diagnostics.
5. Merging the reviewed PR publishes through the existing Cloudflare Pages deployment.

Raw webpages are not archived. Imported recipes are marked `Imported, untested` and retain source attribution.

## Local CLI

Set an OpenAI API key for text-only sources and reliable normalization:

```bash
export OPENAI_API_KEY="..."
make recipe-import URL="https://example.com/recipe"
```

Public PDF and image URLs use the same command and require `OPENAI_API_KEY`:

```bash
make recipe-import URL="https://example.com/scanned-recipe.pdf"
make recipe-import URL="https://example.com/recipe-card.jpg"
```

Supported image formats are PNG, JPEG, WebP, and non-animated GIF. This feature accepts a public media URL; it does not upload a local file or a photo from the device.

Use supplied text when a page blocks automated fetching:

```bash
make recipe-import \
  URL="https://example.com/recipe" \
  SOURCE_TEXT_FILE="/path/to/recipe.txt"
```

Useful direct flags:

```bash
go run ./cmd/recipeimport import --url "https://example.com/recipe" --dry-run
go run ./cmd/recipeimport import --url "https://example.com/recipe" --no-ai
```

The no-AI path requires usable `schema.org/Recipe` JSON-LD and fails when it cannot parse concrete ingredient amounts conservatively. PDF and image imports always require AI normalization.

## GitHub Setup

Create these repository labels:

```bash
gh label create recipe-import --color 1d76db --description "Queued recipe import"
gh label create import-processing --color fbca04 --description "Recipe import is running"
gh label create import-failed --color d93f0b --description "Recipe import needs correction or retry"
gh label create import-pr-open --color 0e8a16 --description "Recipe import has a draft PR"
```

Add the Actions secret `OPENAI_API_KEY`. Optionally set the Actions variable `RECIPE_IMPORT_MODEL`; it defaults to `gpt-5.6-terra`.

In repository Actions settings, allow GitHub Actions to create pull requests. The workflow itself requests only contents, issues, and pull-request write permissions.

To retry a transiently failed issue, run the `Import recipe` workflow manually and supply its issue number. Submit a new capture if the source URL or supplied text itself needs correction.

## Cloudflare Pages Setup

The existing Pages project discovers `site/functions/` during its normal Git deployment. Add:

- `RECIPE_CAPTURE_TOKEN`: generate a random high-entropy token, for example with `openssl rand -hex 32`.
- `GITHUB_ISSUES_TOKEN`: a fine-grained token for only `matthillman/recipe-vault`, granting Issues read/write and Metadata read.
- `GITHUB_REPOSITORY`: `matthillman/recipe-vault`.

Use separate secret values for preview and production environments. After deployment, smoke-test without exposing the token in shell history by reading it from an environment variable:

```bash
curl --fail-with-body \
  -H "Authorization: Bearer ${RECIPE_CAPTURE_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/recipe","client":"smoke-test"}' \
  "https://YOUR_RECIPE_BOX_DOMAIN/api/import"
```

Expected status is `202` for a new issue or `200` for an already queued source.

## Failure and Security Behavior

- Only HTTPS sources are accepted; redirects are capped and private, loopback, link-local, multicast, and unspecified IP ranges are blocked by the importer.
- Page, PDF, and image content is untrusted input. The model receives no tools and must return strict structured JSON.
- Remote sources are fetched by the importer under the same HTTPS, redirect, and public-IP restrictions. HTML is capped at 5 MiB; PDFs and supported images are capped at 10 MiB. Validated media bytes are embedded in the OpenAI request rather than fetched there by URL.
- Dependable mass, liquid-volume, and temperature conversions are allowed. Ingredient-density guesses are not.
- Items mentioned only in process steps, or ingredients without numeric amounts, remain in the process but are omitted from the canonical ingredient list with an import warning. Missing yield, all-invalid ingredients, or missing process steps still stop the import.
- API payloads are capped at 32 KiB of source text and 1 KiB of notes.
- Duplicate open captures return their existing issue. Existing recipe source URLs and slugs are rejected by the importer.
- PDF and image recipes include an OCR-review warning; quantities and instructions must be checked against the source before marking the recipe ready.
