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

This is set up as a static Vite build, which fits Cloudflare Pages’ free tier well.

Cloudflare Pages settings:

- **Root directory:** `site`
- **Build command:** `npm run build`
- **Build output directory:** `dist`

Workflow:

- Connect the repo to Cloudflare Pages (Git integration).
- Pushes to your default branch trigger automatic builds + deploys.

