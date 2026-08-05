# Recipe Box (Web)

Mobile-first web UI to browse the Markdown recipes in `../recipes/`.

## Local iteration

```bash
cd site
npm install
npm run dev
```

Notes:

- `npm run dev` runs `npm run sync` first, which copies `../recipes/*.md` into `site/public/recipes/` and generates `site/public/recipes/manifest.json`.
- If you add/edit recipes while the dev server is running, re-run `npm run sync` (then refresh the page).

## Deploy (Cloudflare Pages)

This is deployed on Cloudflare Pages from the GitHub repo `matthillman/recipe-vault`.

Current Cloudflare Pages configuration:

- **Git repository:** `matthillman/recipe-vault`
- **Root directory:** `site`
- **Build command:** `npm run build`
- **Build output directory:** `dist`
- **Production branch:** `main`
- **Automatic deployments:** Enabled
- **Build comments:** Enabled
- **Build cache:** Disabled
- **Build watch paths:** `*`
- **Build system version:** Version 3
- **Deploy hooks:** None defined

Workflow:

- Pushes to `main` trigger production deploys automatically.
- The build runs from `site/`, so changes to the web UI or Markdown content can both affect the deployed site.

## Recipe capture API

Cloudflare Pages Functions under `functions/api/` add two routes to the same deployment:

- `POST /api/import`: authenticated recipe capture
- `GET /api/openapi.json`: OpenAPI document for a private Custom GPT Action

Configure these encrypted production secrets/variables in the Pages project:

- `RECIPE_CAPTURE_TOKEN`: a random bearer token shared with approved clients
- `GITHUB_ISSUES_TOKEN`: a fine-grained GitHub token scoped to `matthillman/recipe-vault`, with Issues read/write and Metadata read
- `GITHUB_REPOSITORY`: `matthillman/recipe-vault`

The function can create and search issues but cannot read or write repository contents. Complete setup and smoke tests are in `../docs/recipe-import.md`.
