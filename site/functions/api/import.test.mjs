import test from "node:test";
import assert from "node:assert/strict";

import { handleImport, normalizeSourceURL, validateCapture } from "../_lib/import-core.js";
import { onRequestGet as getOpenAPIDocument } from "./openapi.json.js";
import { decodeImportIssue } from "../../../scripts/decode-import-issue.mjs";

const env = {
  RECIPE_CAPTURE_TOKEN: "capture-secret",
  GITHUB_ISSUES_TOKEN: "github-secret",
  GITHUB_REPOSITORY: "matthillman/recipe-vault",
};

function request(payload, token = "capture-secret") {
  return new Request("https://recipes.example/api/import", {
    method: "POST",
    headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

test("queues a normalized capture as a decodable GitHub issue", async () => {
  const calls = [];
  const fetcher = async (url, init) => {
    calls.push({ url: String(url), init });
    if (String(url).includes("/search/issues")) return Response.json({ items: [] });
    return Response.json({ number: 42, html_url: "https://github.com/matthillman/recipe-vault/issues/42" }, { status: 201 });
  };
  const response = await handleImport(
    request({ url: "https://EXAMPLE.com/soup?utm_source=phone#ingredients", client: "ios-shortcut" }),
    env,
    fetcher,
  );
  assert.equal(response.status, 202);
  assert.equal(calls.length, 2);
  assert.equal(calls[1].init.headers.Authorization, "Bearer github-secret");
  const issue = JSON.parse(calls[1].init.body);
  assert.deepEqual(issue.labels, ["recipe-import"]);
  assert.deepEqual(decodeImportIssue(issue), { url: "https://example.com/soup", client: "ios-shortcut" });
  assert.deepEqual(await response.json(), {
    status: "queued",
    duplicate: false,
    issueNumber: 42,
    issueUrl: "https://github.com/matthillman/recipe-vault/issues/42",
  });
});

test("returns an existing queued issue without creating another", async () => {
  let calls = 0;
  const fetcher = async () => {
    calls += 1;
    return Response.json({ items: [{ number: 7, html_url: "https://github.com/example/issues/7" }] });
  };
  const response = await handleImport(request({ sourceText: "Cake\n200 g flour", client: "chatgpt-action" }), env, fetcher);
  assert.equal(response.status, 200);
  assert.equal(calls, 1);
  assert.equal((await response.json()).duplicate, true);
});

test("rejects unauthorized and invalid requests before calling GitHub", async () => {
  const never = async () => assert.fail("GitHub should not be called");
  assert.equal((await handleImport(request({ url: "https://example.com" }, "wrong"), env, never)).status, 401);
  assert.equal((await handleImport(request({ url: "http://example.com" }), env, never)).status, 400);
  assert.equal((await handleImport(request({}), env, never)).status, 400);
});

test("normalizes tracking parameters and enforces input limits", () => {
  assert.equal(normalizeSourceURL("https://EXAMPLE.com/r?utm_campaign=x&keep=1#step"), "https://example.com/r?keep=1");
  assert.throws(() => normalizeSourceURL("https://localhost/r"), /public hostname/);
  assert.throws(() => normalizeSourceURL("https://127.0.0.1/r"), /public hostname/);
  assert.throws(() => normalizeSourceURL("https://192.168.1.10/r"), /public hostname/);
  assert.throws(() => validateCapture({ sourceText: "x".repeat(32 * 1024 + 1) }), /32 KiB/);
  assert.throws(() => validateCapture({ url: "https://example.com", unexpected: "value" }), /unsupported/);
});

test("turns GitHub failures into a stable gateway error", async () => {
  const fetcher = async () => Response.json({ message: "rate limited" }, { status: 403 });
  const response = await handleImport(request({ url: "https://example.com/recipe" }), env, fetcher);
  assert.equal(response.status, 502);
  assert.deepEqual(await response.json(), { error: "queue_unavailable" });
});

test("serves a Custom GPT-compatible OpenAPI contract for the deployed origin", async () => {
  const response = getOpenAPIDocument({ request: new Request("https://recipes.example/api/openapi.json") });
  assert.equal(response.status, 200);
  const document = await response.json();
  assert.deepEqual(document.servers, [{ url: "https://recipes.example" }]);
  assert.equal(document.paths["/api/import"].post.operationId, "captureRecipe");
  assert.deepEqual(document.paths["/api/import"].post.security, [{ bearerAuth: [] }]);
  assert.deepEqual(document.components.securitySchemes.bearerAuth, { type: "http", scheme: "bearer" });
});
