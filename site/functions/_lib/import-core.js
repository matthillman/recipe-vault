const MAX_URL_LENGTH = 2048;
const MAX_SOURCE_TEXT_BYTES = 32 * 1024;
const MAX_NOTE_BYTES = 1024;
const MAX_BODY_BYTES = 48 * 1024;

const encoder = new TextEncoder();

export async function handleImport(request, env, fetcher = fetch) {
  if (request.method !== "POST") return jsonResponse({ error: "method_not_allowed" }, 405, { Allow: "POST, OPTIONS" });
  if (!env?.RECIPE_CAPTURE_TOKEN || !env?.GITHUB_ISSUES_TOKEN || !env?.GITHUB_REPOSITORY) {
    return jsonResponse({ error: "service_not_configured" }, 503);
  }
  if (!(await authorized(request, env.RECIPE_CAPTURE_TOKEN))) {
    return jsonResponse({ error: "unauthorized" }, 401, { "WWW-Authenticate": "Bearer" });
  }

  const declaredLength = Number(request.headers.get("content-length") || 0);
  if (declaredLength > MAX_BODY_BYTES) return jsonResponse({ error: "request_too_large" }, 413);
  let bodyText;
  try {
    bodyText = await request.text();
  } catch {
    return jsonResponse({ error: "invalid_request_body" }, 400);
  }
  if (byteLength(bodyText) > MAX_BODY_BYTES) return jsonResponse({ error: "request_too_large" }, 413);

  let input;
  try {
    input = JSON.parse(bodyText);
  } catch {
    return jsonResponse({ error: "invalid_json" }, 400);
  }
  let capture;
  try {
    capture = validateCapture(input);
  } catch (error) {
    return jsonResponse({ error: "invalid_capture", message: error.message }, 400);
  }

  const fingerprint = await captureFingerprint(capture);
  try {
    const existing = await findExistingIssue(fetcher, env, fingerprint);
    if (existing) {
      return jsonResponse({ status: "queued", duplicate: true, issueNumber: existing.number, issueUrl: existing.html_url }, 200);
    }

    const encoded = bytesToBase64(encoder.encode(JSON.stringify(capture)));
    const body = [
      "Recipe import queued by the capture API.",
      "",
      `Client: ${capture.client || "unknown"}`,
      capture.url ? `Source: ${capture.url}` : "Source: shared text",
      "",
      `<!-- recipe-import-key:${fingerprint} -->`,
      "<!-- recipe-import:v1",
      encoded,
      "-->",
    ].join("\n");
    const title = captureTitle(capture);
    const created = await githubJSON(fetcher, env, `https://api.github.com/repos/${env.GITHUB_REPOSITORY}/issues`, {
      method: "POST",
      body: JSON.stringify({ title: `Recipe import: ${title}`, body, labels: ["recipe-import"] }),
    });
    return jsonResponse({ status: "queued", duplicate: false, issueNumber: created.number, issueUrl: created.html_url }, 202);
  } catch {
    return jsonResponse({ error: "queue_unavailable" }, 502);
  }
}

export function validateCapture(input) {
  if (!input || typeof input !== "object" || Array.isArray(input)) throw new Error("request must be an object");
  const allowed = new Set(["url", "sourceText", "note", "client", "idempotencyKey"]);
  for (const [key, value] of Object.entries(input)) {
    if (!allowed.has(key)) throw new Error(`unsupported field: ${key}`);
    if (typeof value !== "string") throw new Error(`${key} must be a string`);
  }
  const capture = {
    url: input.url?.trim() || "",
    sourceText: input.sourceText?.trim() || "",
    note: input.note?.trim() || "",
    client: input.client?.trim() || "",
    idempotencyKey: input.idempotencyKey?.trim() || "",
  };
  if (!capture.url && !capture.sourceText) throw new Error("provide url and/or sourceText");
  if (capture.url) capture.url = normalizeSourceURL(capture.url);
  if (capture.url.length > MAX_URL_LENGTH) throw new Error("url exceeds 2048 characters");
  if (byteLength(capture.sourceText) > MAX_SOURCE_TEXT_BYTES) throw new Error("sourceText exceeds 32 KiB");
  if (byteLength(capture.note) > MAX_NOTE_BYTES) throw new Error("note exceeds 1 KiB");
  if (capture.client.length > 64) throw new Error("client exceeds 64 characters");
  if (capture.idempotencyKey.length > 128) throw new Error("idempotencyKey exceeds 128 characters");
  return Object.fromEntries(Object.entries(capture).filter(([, value]) => value !== ""));
}

export function normalizeSourceURL(raw) {
  let parsed;
  try {
    parsed = new URL(raw);
  } catch {
    throw new Error("url is invalid");
  }
  if (parsed.protocol !== "https:") throw new Error("url must use https");
  if (!publicHostname(parsed.hostname)) throw new Error("url must use a public hostname");
  parsed.hash = "";
  for (const key of [...parsed.searchParams.keys()]) {
    const lower = key.toLowerCase();
    if (lower.startsWith("utm_") || lower === "fbclid" || lower === "gclid") parsed.searchParams.delete(key);
  }
  return parsed.toString();
}

function publicHostname(hostname) {
  const host = hostname.toLowerCase().replace(/^\[|\]$/g, "");
  if (!host || host === "localhost" || host === "::1" || host === "0:0:0:0:0:0:0:1") return false;
  if (/^(?:0|10|127)\./.test(host) || /^169\.254\./.test(host) || /^192\.168\./.test(host)) return false;
  const match172 = /^172\.(\d+)\./.exec(host);
  if (match172 && Number(match172[1]) >= 16 && Number(match172[1]) <= 31) return false;
  if (host.includes(":") && /^(?:fc|fd|fe8|fe9|fea|feb)/.test(host)) return false;
  return true;
}

async function authorized(request, expected) {
  const header = request.headers.get("authorization") || "";
  const supplied = header.startsWith("Bearer ") ? header.slice(7) : "";
  const [a, b] = await Promise.all([crypto.subtle.digest("SHA-256", encoder.encode(supplied)), crypto.subtle.digest("SHA-256", encoder.encode(expected))]);
  const left = new Uint8Array(a);
  const right = new Uint8Array(b);
  let mismatch = left.length ^ right.length;
  for (let i = 0; i < Math.max(left.length, right.length); i += 1) mismatch |= (left[i] || 0) ^ (right[i] || 0);
  return mismatch === 0 && supplied.length > 0;
}

async function captureFingerprint(capture) {
  const material = capture.url ? `url:${capture.url}` : `text:${capture.sourceText}`;
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", encoder.encode(material)));
  return [...digest].map((byte) => byte.toString(16).padStart(2, "0")).join("");
}

async function findExistingIssue(fetcher, env, fingerprint) {
  const query = `repo:${env.GITHUB_REPOSITORY} is:issue is:open "recipe-import-key:${fingerprint}"`;
  const result = await githubJSON(fetcher, env, `https://api.github.com/search/issues?q=${encodeURIComponent(query)}`);
  return Array.isArray(result.items) && result.items.length > 0 ? result.items[0] : null;
}

async function githubJSON(fetcher, env, url, init = {}) {
  const response = await fetcher(url, {
    ...init,
    headers: {
      Accept: "application/vnd.github+json",
      Authorization: `Bearer ${env.GITHUB_ISSUES_TOKEN}`,
      "Content-Type": "application/json",
      "User-Agent": "recipe-vault-capture/1.0",
      "X-GitHub-Api-Version": "2022-11-28",
      ...(init.headers || {}),
    },
  });
  const text = await response.text();
  let payload = {};
  try {
    payload = text ? JSON.parse(text) : {};
  } catch {
    payload = { message: text };
  }
  if (!response.ok) throw new Error(`GitHub API ${response.status}: ${payload.message || "request failed"}`);
  return payload;
}

function captureTitle(capture) {
  if (capture.url) return new URL(capture.url).hostname.replace(/^www\./, "").slice(0, 80);
  return capture.sourceText.split(/\r?\n/, 1)[0].replace(/\s+/g, " ").slice(0, 80) || "shared text";
}

function bytesToBase64(bytes) {
  let binary = "";
  for (let i = 0; i < bytes.length; i += 1) binary += String.fromCharCode(bytes[i]);
  return btoa(binary);
}

function byteLength(value) {
  return encoder.encode(value).length;
}

function jsonResponse(payload, status, extraHeaders = {}) {
  return new Response(`${JSON.stringify(payload)}\n`, {
    status,
    headers: { "Content-Type": "application/json; charset=utf-8", "Cache-Control": "no-store", ...extraHeaders },
  });
}
