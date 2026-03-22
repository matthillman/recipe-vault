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
