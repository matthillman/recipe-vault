import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

import { decodeImportIssue } from "./decode-import-issue.mjs";

function issueFor(payload) {
  const encoded = Buffer.from(JSON.stringify(payload), "utf8").toString("base64");
  return { body: `Queued by recipe capture.\n\n<!-- recipe-import:v1\n${encoded}\n-->` };
}

test("decodes a versioned import payload", () => {
  assert.deepEqual(
    decodeImportIssue(issueFor({ url: "https://example.com/recipe", client: "ios-shortcut" })),
    { url: "https://example.com/recipe", client: "ios-shortcut" },
  );
});

test("rejects unsupported fields and insecure URLs", () => {
  assert.throws(() => decodeImportIssue(issueFor({ url: "http://example.com" })), /must use https/);
  assert.throws(() => decodeImportIssue(issueFor({ sourceText: "recipe", extra: "no" })), /unsupported/);
});

test("workflow writes the draft PR description with real newlines", () => {
  const workflow = fs.readFileSync(new URL("../.github/workflows/import-recipe.yml", import.meta.url), "utf8");
  assert.match(workflow, /> \.tmp\/pr-body\.md/);
  assert.match(workflow, /--body-file \.tmp\/pr-body\.md/);
  assert.doesNotMatch(workflow, /--body .*\\n/);
});
