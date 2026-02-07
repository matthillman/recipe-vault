import DOMPurify from "dompurify";
import { marked } from "marked";

type RecipeMeta = {
  slug: string;
  title: string;
  updatedISO: string;
  updatedMs: number;
};

type Manifest = {
  generatedAtISO: string;
  count: number;
  recipes: RecipeMeta[];
};

type Route =
  | { kind: "list"; q: string }
  | { kind: "recipe"; slug: string };

function parseRoute(): Route {
  const raw = window.location.hash.replace(/^#\/?/, "");
  if (!raw) return { kind: "list", q: "" };

  const [path, query] = raw.split("?", 2);
  const params = new URLSearchParams(query ?? "");

  if (path.startsWith("recipe/")) {
    const slug = decodeURIComponent(path.slice("recipe/".length));
    return { kind: "recipe", slug };
  }

  return { kind: "list", q: params.get("q") ?? "" };
}

function setListRoute(q: string) {
  const params = new URLSearchParams();
  if (q.trim()) params.set("q", q.trim());
  const suffix = params.toString() ? `?${params.toString()}` : "";
  window.location.hash = `#/list${suffix}`;
}

function setRecipeRoute(slug: string) {
  window.location.hash = `#/recipe/${encodeURIComponent(slug)}`;
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

function textNode(text: string) {
  return document.createTextNode(text);
}

const KEEP_AWAKE_KEY = "recipeBox.keepAwake";

function supportsWakeLock() {
  return Boolean((navigator as unknown as { wakeLock?: unknown }).wakeLock);
}

export function mountApp(root: HTMLElement) {
  root.replaceChildren();

  const shell = el("div", { class: "shell" });
  const topbar = el("div", { class: "topbar" });
  const topbarInner = el("div", { class: "topbar-inner" });
  const brand = el("div", { class: "brand" }, [textNode("Recipe Box")]);
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
    class: "pill",
    type: "button",
    title: supportsWakeLock()
      ? "Prevent the screen from sleeping while this tab is visible"
      : "Wake Lock not supported in this browser",
    "aria-label": "Toggle keep screen awake",
  });

  topbarInner.append(brand, searchWrap, keepAwakeBtn);
  topbar.append(topbarInner);

  const content = el("main", { class: "content" });
  shell.append(topbar, content);
  root.append(shell);

  let wakeLockSentinel: unknown | null = null;
  let keepAwake = window.localStorage.getItem(KEEP_AWAKE_KEY) === "1";

  const updateKeepAwakeUI = () => {
    const on = keepAwake && Boolean(wakeLockSentinel);
    keepAwakeBtn.textContent = on ? "Awake: On" : "Awake: Off";
    keepAwakeBtn.classList.toggle("on", on);
    keepAwakeBtn.toggleAttribute("disabled", !supportsWakeLock());
  };

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

  const render = async () => {
    content.replaceChildren();

    const route = parseRoute();
    if (!manifest) {
      content.append(el("div", { class: "empty" }, ["Loading recipes…"]));
      try {
        manifest = await fetchManifest();
      } catch (e) {
        content.replaceChildren(
          el("div", { class: "empty" }, [
            "Failed to load recipe manifest. Run `npm run sync` and reload.",
          ]),
        );
        throw e;
      }
    }

    if (route.kind === "recipe") {
      searchInput.value = "";
      const meta = manifest.recipes.find((r) => r.slug === route.slug) ?? null;

      const header = el("div", { class: "recipe-header" });
      const backBtn = el("button", { class: "back", type: "button" }, ["Back"]);
      backBtn.addEventListener("click", () => window.history.back());
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

      const mdWrap = el("article", { class: "md" });
      mdWrap.append(el("div", { class: "empty" }, ["Loading…"]));
      content.append(mdWrap);

      try {
        const md = await fetchRecipeMarkdown(route.slug);
        const html = renderMarkdown(md);
        mdWrap.innerHTML = html;

        const derivedTitle = titleFromMarkdown(md);
        if (derivedTitle && derivedTitle !== title.textContent) {
          title.textContent = derivedTitle;
        }
      } catch {
        mdWrap.replaceChildren(
          el("div", { class: "empty" }, [
            "Recipe not found (did you run `npm run sync`?).",
          ]),
        );
      }

      return;
    }

    const q = route.q;
    searchInput.value = q;

    const normalized = q.trim().toLowerCase();
    const recipes = manifest.recipes
      .slice()
      .sort((a, b) => b.updatedMs - a.updatedMs)
      .filter((r) => (!normalized ? true : r.title.toLowerCase().includes(normalized)));

    if (!recipes.length) {
      content.append(el("div", { class: "empty" }, ["No matches."]));
      return;
    }

    const list = el("div", { class: "list" });
    for (const r of recipes) {
      const a = el("a", { class: "card", href: `#/recipe/${encodeURIComponent(r.slug)}` });
      a.addEventListener("click", (ev) => {
        ev.preventDefault();
        setRecipeRoute(r.slug);
      });

      const h = el("h2", { class: "card-title" }, [textNode(r.title)]);
      const metaLine = el("div", { class: "card-meta" }, [
        `Updated ${formatDate(r.updatedISO)}`,
      ]);
      a.append(h, metaLine);
      list.append(a);
    }
    content.append(list);
  };

  let searchTimer: number | undefined;
  const onSearch = () => setListRoute(searchInput.value);
  searchInput.addEventListener("input", () => {
    if (searchTimer) window.clearTimeout(searchTimer);
    searchTimer = window.setTimeout(onSearch, 120);
  });
  searchInput.addEventListener("keydown", (ev) => {
    if (ev.key === "Escape") {
      searchInput.value = "";
      setListRoute("");
      searchInput.blur();
    }
  });

  keepAwakeBtn.addEventListener("click", async () => {
    if (!supportsWakeLock()) return;
    keepAwake = !keepAwake;
    window.localStorage.setItem(KEEP_AWAKE_KEY, keepAwake ? "1" : "0");
    if (!keepAwake) await releaseWakeLock();
    await acquireWakeLock();
  });

  document.addEventListener("visibilitychange", () => void acquireWakeLock());

  window.addEventListener("hashchange", () => void render());
  updateKeepAwakeUI();
  void acquireWakeLock();
  void render();
}
