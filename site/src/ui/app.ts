import DOMPurify from "dompurify";
import { marked } from "marked";

type RecipeMeta = {
  slug: string;
  title: string;
  tags: string[];
  yieldLines: string[];
  sectionHeadings: string[];
  searchText: string;
  updatedISO: string;
  updatedMs: number;
};

type Manifest = {
  generatedAtISO: string;
  count: number;
  recipes: RecipeMeta[];
};

type SortMode = "updated_desc" | "title_asc" | "title_desc";

type Route =
  | { kind: "list"; q: string; tag: string }
  | { kind: "recipe"; slug: string };

type RouteMode = "push" | "replace";

type RecipeSection = {
  heading: string;
  body: string;
  kind: "ingredients" | "process" | "formula" | "notes" | "other";
};

function parseRoute(): Route {
  const raw = window.location.hash.replace(/^#\/?/, "");
  if (!raw) return { kind: "list", q: "" };

  const [path, query] = raw.split("?", 2);
  const params = new URLSearchParams(query ?? "");

  if (path.startsWith("recipe/")) {
    const slug = decodeURIComponent(path.slice("recipe/".length));
    return { kind: "recipe", slug };
  }

  return { kind: "list", q: params.get("q") ?? "", tag: params.get("tag") ?? "" };
}

function setListRoute({ q, tag }: { q: string; tag?: string }) {
  const params = new URLSearchParams();
  if (q.trim()) params.set("q", q.trim());
  if (tag?.trim()) params.set("tag", tag.trim());
  const suffix = params.toString() ? `?${params.toString()}` : "";
  return `#/list${suffix}`;
}

function setRecipeRoute(slug: string) {
  return `#/recipe/${encodeURIComponent(slug)}`;
}

function formatDate(iso: string) {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return new Intl.DateTimeFormat(undefined, { dateStyle: "medium" }).format(d);
}

async function fetchManifest(): Promise<Manifest> {
  const cacheBust = import.meta.env.DEV ? `?t=${Date.now()}` : "";
  const res = await fetch(`${withBase("recipes/manifest.json")}${cacheBust}`, {
    headers: { accept: "application/json" },
  });
  if (!res.ok) throw new Error(`Failed to load manifest: ${res.status}`);
  return res.json();
}

async function fetchRecipeMarkdown(slug: string): Promise<string> {
  const cacheBust = import.meta.env.DEV ? `?t=${Date.now()}` : "";
  const recipePath = `recipes/${encodeURIComponent(slug)}.md`;
  const res = await fetch(`${withBase(recipePath)}${cacheBust}`, {
    headers: { accept: "text/markdown,text/plain" },
  });
  if (!res.ok) throw new Error(`Missing recipe: ${slug}`);
  return res.text();
}

function withBase(pathname: string) {
  const base = import.meta.env.BASE_URL || "/";
  const normalizedBase = base.endsWith("/") ? base : `${base}/`;
  return `${normalizedBase}${pathname.replace(/^\/+/, "")}`;
}

function titleFromMarkdown(md: string) {
  for (const line of md.split(/\r?\n/)) {
    const m = /^#\s+(.+?)\s*$/.exec(line);
    if (m) return m[1];
  }
  return null;
}

function renderMarkdown(md: string) {
  marked.setOptions({
    gfm: true,
    breaks: true,
  });
  const unsafe = marked.parse(md) as string;
  return DOMPurify.sanitize(unsafe, { USE_PROFILES: { html: true } });
}

function el<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  attrs: Record<string, string> = {},
  children: Array<Node | string> = [],
): HTMLElementTagNameMap[K] {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) node.setAttribute(k, v);
  for (const child of children) node.append(child);
  return node;
}

function svgEl<K extends keyof SVGElementTagNameMap>(
  tag: K,
  attrs: Record<string, string> = {},
  children: Array<Node> = [],
): SVGElementTagNameMap[K] {
  const node = document.createElementNS("http://www.w3.org/2000/svg", tag);
  for (const [k, v] of Object.entries(attrs)) node.setAttribute(k, v);
  for (const child of children) node.append(child);
  return node;
}

function textNode(text: string) {
  return document.createTextNode(text);
}

const KEEP_AWAKE_KEY = "recipeBox.keepAwake";
const SORT_MODE_KEY = "recipeBox.sortMode";

function supportsWakeLock() {
  return Boolean((navigator as unknown as { wakeLock?: unknown }).wakeLock);
}

function parseSortMode(raw: string | null): SortMode {
  if (raw === "title_asc" || raw === "title_desc" || raw === "updated_desc") return raw;
  return "updated_desc";
}

function sortIcon() {
  // Simple "sort" icon (bars + chevrons), no external deps.
  return svgEl(
    "svg",
    {
      class: "icon",
      viewBox: "0 0 24 24",
      fill: "none",
      stroke: "currentColor",
      "stroke-width": "2",
      "stroke-linecap": "round",
      "stroke-linejoin": "round",
      "aria-hidden": "true",
    },
    [
      svgEl("path", { d: "M3 6h10" }),
      svgEl("path", { d: "M3 12h14" }),
      svgEl("path", { d: "M3 18h18" }),
      svgEl("path", { d: "M17 8l2-2 2 2" }),
      svgEl("path", { d: "M19 6v8" }),
      svgEl("path", { d: "M21 16l-2 2-2-2" }),
    ],
  );
}

function wakeIcon() {
  return svgEl(
    "svg",
    {
      class: "icon",
      viewBox: "0 0 24 24",
      fill: "none",
      stroke: "currentColor",
      "stroke-width": "2",
      "stroke-linecap": "round",
      "stroke-linejoin": "round",
      "aria-hidden": "true",
    },
    [
      svgEl("rect", { x: "7", y: "3", width: "10", height: "18", rx: "2" }),
      svgEl("path", { d: "M10 7h4" }),
      svgEl("path", { d: "M12 15v2" }),
    ],
  );
}

function slugifyHeading(label: string) {
  return label
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

function sectionKind(heading: string): RecipeSection["kind"] {
  const lower = heading.toLowerCase();
  if (lower.startsWith("ingredients")) return "ingredients";
  if (lower.startsWith("process")) return "process";
  if (lower.startsWith("formula")) return "formula";
  if (lower.startsWith("notes")) return "notes";
  return "other";
}

function parseRecipeMarkdown(md: string): { title: string | null; yieldLines: string[]; sections: RecipeSection[] } {
  const lines = md.split(/\r?\n/);
  const sections: RecipeSection[] = [];
  let title: string | null = null;
  let yieldLines: string[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];
    if (!title) {
      const match = /^#\s+(.+?)\s*$/.exec(line);
      if (match) {
        title = match[1];
        i += 1;
        continue;
      }
    }

    if (/^\*\*Yield \/ (Target|Pan Target)\*\*$/.test(line.trim())) {
      const out: string[] = [];
      i += 1;
      while (i < lines.length) {
        const trimmed = lines[i].trim();
        if (!trimmed) {
          if (out.length) break;
          i += 1;
          continue;
        }
        const bullet = /^[-*]\s+(.+)$/.exec(trimmed);
        if (!bullet) break;
        out.push(bullet[1].trim());
        i += 1;
      }
      yieldLines = out;
      continue;
    }

    const sectionMatch = /^##\s+(.+?)\s*$/.exec(line);
    if (sectionMatch) {
      const heading = sectionMatch[1].trim();
      i += 1;
      const body: string[] = [];
      while (i < lines.length && !/^##\s+/.test(lines[i])) {
        body.push(lines[i]);
        i += 1;
      }
      sections.push({
        heading,
        body: body.join("\n").trim(),
        kind: sectionKind(heading),
      });
      continue;
    }

    i += 1;
  }

  return { title, yieldLines, sections };
}

export function mountApp(root: HTMLElement) {
  root.replaceChildren();

  const shell = el("div", { class: "shell" });
  const topbar = el("div", { class: "topbar" });
  const topbarInner = el("div", { class: "topbar-inner" });
  const topbarHeader = el("div", { class: "topbar-header" });
  const topbarLeft = el("div", { class: "topbar-left" });
  const topbarCenter = el("div", { class: "topbar-center" });
  const topbarRight = el("div", { class: "topbar-right" });

  const logoBtn = el("button", {
    class: "logo-btn",
    type: "button",
    title: "Home",
    "aria-label": "Go to recipe list",
  });
  const logoImg = el("img", {
    class: "logo",
    alt: "Recipe Box",
    src: withBase("logo.svg"),
    width: "30",
    height: "30",
    decoding: "async",
    loading: "eager",
  }) as HTMLImageElement;
  logoBtn.append(logoImg);

  const brand = el("div", { class: "brand" }, [textNode("Recipe Box")]);

  const controls = el("div", { class: "controls" });
  const searchWrap = el("div", { class: "search" });
  const searchInput = el("input", {
    type: "search",
    placeholder: "Search recipes…",
    autocomplete: "off",
    enterkeyhint: "search",
    "aria-label": "Search recipes",
  });
  searchWrap.append(searchInput);

  const keepAwakeBtn = el("button", {
    class: "pill icon-pill wake-pill",
    type: "button",
    title: supportsWakeLock()
      ? "Prevent the screen from sleeping while this tab is visible"
      : "Wake Lock not supported in this browser",
    "aria-label": "Toggle keep screen awake",
  });
  const keepAwakeLabel = el("span", { class: "sr-only" });
  keepAwakeBtn.append(wakeIcon(), keepAwakeLabel);

  const sortDetails = el("details") as HTMLDetailsElement;
  const sortSummary = el("summary", { class: "pill icon-pill", role: "button" });
  const sortA11yLabel = el("span", { class: "sr-only" });
  sortSummary.append(sortIcon(), sortA11yLabel);
  const sortMenu = el("div", { class: "sort-menu", role: "menu" });
  sortDetails.append(sortSummary, sortMenu);
  const sortWrap = el("div", { class: "sort" });
  sortWrap.append(sortDetails);

  controls.append(searchWrap, sortWrap);

  topbarLeft.append(logoBtn);
  topbarCenter.append(brand);
  topbarRight.append(keepAwakeBtn);
  topbarHeader.append(topbarLeft, topbarCenter, topbarRight);

  topbarInner.append(topbarHeader, controls);
  topbar.append(topbarInner);

  const content = el("main", { class: "content" });
  shell.append(topbar, content);
  root.append(shell);

  let wakeLockSentinel: unknown | null = null;
  let keepAwake = window.localStorage.getItem(KEEP_AWAKE_KEY) === "1";
  let sortMode: SortMode = parseSortMode(window.localStorage.getItem(SORT_MODE_KEY));
  let lastListRoute = setListRoute({ q: "", tag: "" });

  const updateRoute = (hash: string, mode: RouteMode) => {
    const url = new URL(window.location.href);
    url.hash = hash;
    if (mode === "replace") {
      window.history.replaceState(null, "", url);
    } else {
      window.history.pushState(null, "", url);
    }
    void render();
  };

  const updateKeepAwakeUI = () => {
    const on = keepAwake && Boolean(wakeLockSentinel);
    const label = on ? "Keep screen awake: on" : "Keep screen awake: off";
    keepAwakeLabel.textContent = label;
    keepAwakeBtn.title = label;
    keepAwakeBtn.setAttribute("aria-label", label);
    keepAwakeBtn.setAttribute("aria-pressed", on ? "true" : "false");
    keepAwakeBtn.classList.toggle("on", on);
    keepAwakeBtn.hidden = !supportsWakeLock();
  };

  const collator = new Intl.Collator(undefined, { sensitivity: "base", numeric: true });
  const sortModeLabel = (mode: SortMode) => {
    if (mode === "updated_desc") return "Updated";
    if (mode === "title_asc") return "Title A→Z";
    return "Title Z→A";
  };

  const updateSortUI = () => {
    const label = `Sort: ${sortModeLabel(sortMode)}`;
    sortA11yLabel.textContent = label;
    sortSummary.title = label;
    sortSummary.setAttribute("aria-label", label);
    for (const btn of sortMenu.querySelectorAll<HTMLButtonElement>("button[data-sort]")) {
      btn.setAttribute("aria-checked", btn.dataset.sort === sortMode ? "true" : "false");
    }
  };

  const setSortMode = (mode: SortMode) => {
    sortMode = mode;
    window.localStorage.setItem(SORT_MODE_KEY, mode);
    updateSortUI();
    sortDetails.open = false;
    void render();
  };

  const addSortItem = (mode: SortMode) => {
    const btn = el("button", {
      class: "sort-item",
      type: "button",
      role: "menuitemradio",
      "aria-checked": mode === sortMode ? "true" : "false",
      "data-sort": mode,
    });
    btn.append(textNode(sortModeLabel(mode)));
    btn.addEventListener("click", () => setSortMode(mode));
    sortMenu.append(btn);
  };

  addSortItem("updated_desc");
  addSortItem("title_asc");
  addSortItem("title_desc");

  document.addEventListener("click", (ev) => {
    if (!sortDetails.open) return;
    const target = ev.target as Node | null;
    if (!target) return;
    if (!sortDetails.contains(target)) sortDetails.open = false;
  });

  const releaseWakeLock = async () => {
    const sentinel = wakeLockSentinel as { release?: () => Promise<void> } | null;
    wakeLockSentinel = null;
    updateKeepAwakeUI();
    if (sentinel?.release) {
      try {
        await sentinel.release();
      } catch {
        // ignore
      }
    }
  };

  const acquireWakeLock = async () => {
    if (!supportsWakeLock()) return;
    if (!keepAwake) return;
    if (document.visibilityState !== "visible") return;
    if (wakeLockSentinel) return;

    const navAny = navigator as unknown as {
      wakeLock: { request: (type: "screen") => Promise<unknown> };
    };

    try {
      wakeLockSentinel = await navAny.wakeLock.request("screen");
      const sentinelAny = wakeLockSentinel as unknown as {
        addEventListener?: (name: string, cb: () => void) => void;
      };
      sentinelAny.addEventListener?.("release", () => {
        wakeLockSentinel = null;
        updateKeepAwakeUI();
      });
      updateKeepAwakeUI();
    } catch {
      keepAwake = false;
      window.localStorage.setItem(KEEP_AWAKE_KEY, "0");
      wakeLockSentinel = null;
      updateKeepAwakeUI();
    }
  };

  let manifest: Manifest | null = null;
  let renderVersion = 0;

  const render = async () => {
    const current = ++renderVersion;

    const routeAtStart = parseRoute();
    if (!manifest) {
      content.replaceChildren(el("div", { class: "empty" }, ["Loading recipes…"]));
      try {
        manifest = await fetchManifest();
      } catch (e) {
        if (current !== renderVersion) return;
        content.replaceChildren(
          el("div", { class: "empty" }, [
            "Failed to load recipe manifest. Run `npm run sync` and reload.",
          ]),
        );
        return;
      }
    }

    if (current !== renderVersion) return;

    const route = manifest ? parseRoute() : routeAtStart;
    content.replaceChildren();

    controls.classList.toggle("hidden", route.kind === "recipe");
    if (route.kind === "recipe") sortDetails.open = false;

    if (route.kind === "recipe") {
      searchInput.value = "";
      const meta = manifest.recipes.find((r) => r.slug === route.slug) ?? null;

      const header = el("div", { class: "recipe-header" });
      const backBtn = el("button", { class: "back", type: "button" }, ["Back"]);
      backBtn.addEventListener("click", () => updateRoute(lastListRoute, "replace"));
      const titleBlock = el("div");
      const title = el("h1", { class: "recipe-title" }, [
        textNode(meta?.title ?? route.slug),
      ]);
      const metaLine = el(
        "div",
        { class: "recipe-meta" },
        meta?.updatedISO ? [`Updated ${formatDate(meta.updatedISO)}`] : [],
      );
      titleBlock.append(title, metaLine);
      header.append(backBtn, titleBlock);

      content.append(header);

      const detailShell = el("div", { class: "recipe-detail" });
      content.append(detailShell);

      const detailTop = el("div", { class: "recipe-detail-top" });
      detailShell.append(detailTop);

      if (meta?.yieldLines?.length) {
        const summary = el("div", { class: "recipe-summary" });
        for (const line of meta.yieldLines) {
          summary.append(el("div", { class: "recipe-chip" }, [textNode(line)]));
        }
        detailTop.append(summary);
      }

      if (meta?.tags?.length) {
        const tags = el("div", { class: "recipe-tags" });
        for (const tag of meta.tags) {
          tags.append(el("div", { class: "card-chip" }, [textNode(tag)]));
        }
        detailTop.append(tags);
      }

      const nav = el("nav", { class: "recipe-nav", "aria-label": "Recipe sections" });
      detailShell.append(nav);

      const layout = el("div", { class: "recipe-layout" });
      const aside = el("aside", { class: "recipe-aside" });
      const main = el("article", { class: "recipe-main" });
      layout.append(aside, main);
      detailShell.append(layout);

      const loading = el("div", { class: "empty" }, ["Loading…"]);
      main.append(loading);

      try {
        const md = await fetchRecipeMarkdown(route.slug);
        const parsed = parseRecipeMarkdown(md);
        const derivedTitle = parsed.title ?? titleFromMarkdown(md);
        if (derivedTitle && derivedTitle !== title.textContent) title.textContent = derivedTitle;

        if (!meta?.yieldLines?.length && parsed.yieldLines.length) {
          detailTop.prepend(
            el(
              "div",
              { class: "recipe-summary" },
              parsed.yieldLines.map((line) => el("div", { class: "recipe-chip" }, [textNode(line)])),
            ),
          );
        }

        aside.replaceChildren();
        main.replaceChildren();
        nav.replaceChildren();

        const sections = parsed.sections.filter((section) => section.body);
        const primary = sections.filter((section) => section.kind === "ingredients" || section.kind === "formula");
        const secondary = sections.filter((section) => section.kind !== "ingredients" && section.kind !== "formula");
        const orderedSections = primary.length ? [...primary, ...secondary] : sections;

        for (const section of orderedSections) {
          const id = `section-${slugifyHeading(section.heading)}`;
          nav.append(
            el("a", { class: "recipe-nav-link", href: `#${id}` }, [textNode(section.heading)]),
          );

          const wrap = el("section", { class: "recipe-section", id });
          const heading = el("h2", { class: "recipe-section-title" }, [textNode(section.heading)]);
          const body = el("div", { class: "md recipe-section-body" });
          body.innerHTML = renderMarkdown(section.body);
          wrap.append(heading, body);

          if (section.kind === "ingredients" || section.kind === "formula") {
            aside.append(wrap);
          } else {
            main.append(wrap);
          }
        }

        if (!sections.length) {
          const mdWrap = el("article", { class: "md" });
          mdWrap.innerHTML = renderMarkdown(md);
          main.append(mdWrap);
        }
      } catch {
        main.replaceChildren(
          el("div", { class: "empty" }, [
            "Recipe not found (did you run `npm run sync`?).",
          ]),
        );
      }

      return;
    }

    const q = route.q;
    const activeTag = route.tag;
    lastListRoute = setListRoute({ q, tag: activeTag });
    searchInput.value = q;

    const normalized = q.trim().toLowerCase();
    const filtered = manifest.recipes
      .slice()
      .filter((r) => (!normalized ? true : r.searchText.includes(normalized)))
      .filter((r) => (!activeTag ? true : r.tags.includes(activeTag)));

    const recipes = filtered.slice();
    if (sortMode === "updated_desc") {
      recipes.sort((a, b) => b.updatedMs - a.updatedMs);
    } else if (sortMode === "title_asc") {
      recipes.sort((a, b) => collator.compare(a.title, b.title));
    } else {
      recipes.sort((a, b) => collator.compare(b.title, a.title));
    }

    if (!recipes.length) {
      const availableTags = Array.from(new Set(manifest.recipes.flatMap((r) => r.tags))).sort();
      if (availableTags.length) {
        const filters = el("div", { class: "filter-bar" });
        const allChip = el(
          "button",
          {
            class: `filter-chip${activeTag ? "" : " active"}`,
            type: "button",
          },
          ["All"],
        );
        allChip.addEventListener("click", () => updateRoute(setListRoute({ q, tag: "" }), "replace"));
        filters.append(allChip);
        for (const tag of availableTags) {
          const chip = el(
            "button",
            {
              class: `filter-chip${activeTag === tag ? " active" : ""}`,
              type: "button",
            },
            [textNode(tag)],
          );
          chip.addEventListener("click", () =>
            updateRoute(setListRoute({ q, tag: activeTag === tag ? "" : tag }), "replace"),
          );
          filters.append(chip);
        }
        content.append(filters);
      }
      content.append(el("div", { class: "empty" }, ["No matches."]));
      return;
    }

    const availableTags = Array.from(new Set(manifest.recipes.flatMap((r) => r.tags))).sort();
    if (availableTags.length) {
      const filters = el("div", { class: "filter-bar" });
      const allChip = el(
        "button",
        {
          class: `filter-chip${activeTag ? "" : " active"}`,
          type: "button",
        },
        ["All"],
      );
      allChip.addEventListener("click", () => updateRoute(setListRoute({ q, tag: "" }), "replace"));
      filters.append(allChip);

      for (const tag of availableTags) {
        const chip = el(
          "button",
          {
            class: `filter-chip${activeTag === tag ? " active" : ""}`,
            type: "button",
          },
          [textNode(tag)],
        );
        chip.addEventListener("click", () =>
          updateRoute(setListRoute({ q, tag: activeTag === tag ? "" : tag }), "replace"),
        );
        filters.append(chip);
      }
      content.append(filters);
    }

    const list = el("div", { class: "list" });
    for (const r of recipes) {
      const a = el("a", { class: "card", href: `#/recipe/${encodeURIComponent(r.slug)}` });
      a.addEventListener("click", (ev) => {
        ev.preventDefault();
        updateRoute(setRecipeRoute(r.slug), "push");
      });

      const h = el("h2", { class: "card-title" }, [textNode(r.title)]);
      const summary = el("div", { class: "card-summary" });
      if (r.yieldLines[0]) {
        summary.append(el("div", { class: "card-chip card-chip-primary" }, [textNode(r.yieldLines[0])]));
      }
      if (r.yieldLines[1]) {
        summary.append(el("div", { class: "card-chip" }, [textNode(r.yieldLines[1])]));
      }
      for (const tag of r.tags.slice(0, 2)) {
        summary.append(el("div", { class: "card-chip" }, [textNode(tag)]));
      }
      summary.append(
        el("div", { class: "card-chip card-chip-subtle" }, [
          textNode(`Updated ${formatDate(r.updatedISO)}`),
        ]),
      );
      a.append(h, summary);
      list.append(a);
    }
    content.append(list);
  };

  let searchTimer: number | undefined;
  const onSearch = () =>
    updateRoute(
      setListRoute({
        q: searchInput.value,
        tag: parseRoute().kind === "list" ? parseRoute().tag : "",
      }),
      "replace",
    );
  searchInput.addEventListener("input", () => {
    if (searchTimer) window.clearTimeout(searchTimer);
    searchTimer = window.setTimeout(onSearch, 120);
  });
  searchInput.addEventListener("keydown", (ev) => {
    if (ev.key === "Escape") {
      searchInput.value = "";
      const route = parseRoute();
      updateRoute(
        setListRoute({ q: "", tag: route.kind === "list" ? route.tag : "" }),
        "replace",
      );
      searchInput.blur();
    }
  });

  const goHome = () => {
    updateRoute(setListRoute({ q: "", tag: "" }), "replace");
    searchInput.focus();
  };
  logoBtn.addEventListener("click", goHome);

  keepAwakeBtn.addEventListener("click", async () => {
    if (!supportsWakeLock()) return;
    keepAwake = !keepAwake;
    window.localStorage.setItem(KEEP_AWAKE_KEY, keepAwake ? "1" : "0");
    if (!keepAwake) await releaseWakeLock();
    await acquireWakeLock();
  });

  document.addEventListener("visibilitychange", () => void acquireWakeLock());

  window.addEventListener("popstate", () => void render());
  window.addEventListener("hashchange", () => void render());
  updateKeepAwakeUI();
  updateSortUI();
  void acquireWakeLock();
  void render();
}
