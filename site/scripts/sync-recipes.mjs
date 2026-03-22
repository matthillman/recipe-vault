import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const siteDir = path.resolve(__dirname, "..");
const repoRoot = path.resolve(siteDir, "..");

const srcRecipesDir = path.join(repoRoot, "recipes");
const outRecipesDir = path.join(siteDir, "public", "recipes");

function titleizeSlug(slug) {
  return slug
    .split("-")
    .filter(Boolean)
    .map((w) => w.slice(0, 1).toUpperCase() + w.slice(1))
    .join(" ");
}

function titleFromMarkdown(md) {
  for (const line of md.split(/\r?\n/)) {
    const m = /^#\s+(.+?)\s*$/.exec(line);
    if (m) return m[1];
  }
  return null;
}

function extractYieldLines(md) {
  const lines = md.split(/\r?\n/);
  for (let i = 0; i < lines.length; i += 1) {
    if (!/^\*\*Yield \/ (Target|Pan Target)\*\*$/.test(lines[i].trim())) continue;

    const out = [];
    for (let j = i + 1; j < lines.length; j += 1) {
      const line = lines[j].trim();
      if (!line) {
        if (out.length) break;
        continue;
      }
      const bullet = /^[-*]\s+(.+)$/.exec(line);
      if (!bullet) break;
      out.push(bullet[1].trim());
    }
    return out;
  }
  return [];
}

function extractSectionHeadings(md) {
  return md
    .split(/\r?\n/)
    .map((line) => /^##\s+(.+?)\s*$/.exec(line)?.[1]?.trim() ?? null)
    .filter(Boolean);
}

function buildSearchText({ slug, title, yieldLines, sectionHeadings, md }) {
  return [slug, title, ...yieldLines, ...sectionHeadings, md]
    .join("\n")
    .toLowerCase()
    .replace(/[`*_>#-]/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

async function ensureDir(dir) {
  await fs.mkdir(dir, { recursive: true });
}

async function readRecipesFromDisk() {
  const entries = await fs.readdir(srcRecipesDir, { withFileTypes: true });
  return entries
    .filter((e) => e.isFile() && e.name.endsWith(".md"))
    .map((e) => e.name)
    .sort();
}

async function main() {
  await ensureDir(outRecipesDir);

  const recipeFiles = await readRecipesFromDisk();

  const slugs = [];
  const recipes = [];

  for (const file of recipeFiles) {
    const slug = file.replace(/\.md$/, "");
    slugs.push(slug);

    const srcPath = path.join(srcRecipesDir, file);
    const outPath = path.join(outRecipesDir, file);

    const [md, stat] = await Promise.all([fs.readFile(srcPath, "utf8"), fs.stat(srcPath)]);
    const title = titleFromMarkdown(md) ?? titleizeSlug(slug);
    const yieldLines = extractYieldLines(md);
    const sectionHeadings = extractSectionHeadings(md);

    await fs.writeFile(outPath, md);

    recipes.push({
      slug,
      title,
      yieldLines,
      sectionHeadings,
      searchText: buildSearchText({ slug, title, yieldLines, sectionHeadings, md }),
      updatedISO: new Date(stat.mtimeMs).toISOString(),
      updatedMs: stat.mtimeMs,
    });
  }

  const outEntries = await fs.readdir(outRecipesDir, { withFileTypes: true });
  await Promise.all(
    outEntries
      .filter((e) => e.isFile() && e.name.endsWith(".md"))
      .filter((e) => !slugs.includes(e.name.replace(/\.md$/, "")))
      .map((e) => fs.unlink(path.join(outRecipesDir, e.name))),
  );

  const manifest = {
    generatedAtISO: new Date().toISOString(),
    count: recipes.length,
    recipes,
  };

  await fs.writeFile(
    path.join(outRecipesDir, "manifest.json"),
    JSON.stringify(manifest, null, 2) + "\n",
    "utf8",
  );

  process.stdout.write(
    `Synced ${recipes.length} recipes -> ${path.relative(repoRoot, outRecipesDir)}\n`,
  );
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
